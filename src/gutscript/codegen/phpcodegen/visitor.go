package phpcodegen

import "gutscript/ast"
import "strconv"
import "strings"
import "fmt"

type Visitor struct {
	indent int
}

func (self *Visitor) IndentSpace() string {
	if self.indent == 0 {
		return ""
	}
	return strings.Repeat("    ", self.indent)
}

func (self *Visitor) Visit(n ast.Node) string {
	// fmt.Printf("visit %#v\n", n)
	if stmts, ok := n.(*ast.StatementNodeList); ok {
		var output string
		for _, stmt := range *stmts {
			output += self.IndentSpace() + self.Visit(stmt)
		}
		return output
	}
	if variable, ok := n.(ast.Variable); ok {
		return "$" + variable.Identifier
	}
	if number, ok := n.(ast.Number); ok {
		return strconv.FormatInt(number.Val, 10)
	}
	if floating, ok := n.(ast.FloatingNumber); ok {
		return strconv.FormatFloat(floating.Val, 'e', -1, 64)
	}

	if expr, ok := n.(ast.UnaryExpr); ok {
		if expr.Op != 0 {
			return fmt.Sprintf("%c%s", expr.Op, self.Visit(expr.Val))
		} else {
			return self.Visit(expr.Val)
		}
	}
	if expr, ok := n.(ast.Expr); ok {
		if expr.Parenthesis {
			return fmt.Sprintf("(%s %c %s)", self.Visit(expr.Left), expr.Op, self.Visit(expr.Right))
		}
		return fmt.Sprintf("%s %c %s", self.Visit(expr.Left), expr.Op, self.Visit(expr.Right))
	}
	if stmt, ok := n.(ast.AssignStatement); ok {
		return self.Visit(stmt.Variable) + " = " + self.Visit(stmt.Expr) + ";\n"
	}
	if stmt, ok := n.(*ast.IfStatement); ok {
		var out string = ""
		out += self.IndentSpace() + "if ( " + self.Visit(stmt.Expr) + " ) {\n"
		self.indent++
		out += self.Visit(stmt.Body)
		self.indent--
		out += self.IndentSpace() + "}"

		if len(stmt.ElseIfList) > 0 {
			for _, elseifStmt := range stmt.ElseIfList {
				out += self.Visit(elseifStmt)
			}
		}
		if stmt.ElseBody != nil {
			out += " else {\n"
			self.indent++
			out += self.Visit(stmt.ElseBody)
			self.indent--
			out += "}"
		}
		out += "\n"
		return out
	}
	if stmt, ok := n.(ast.ElseIfStatement); ok {
		var out string = ""
		out += self.IndentSpace() + " elseif ( " + self.Visit(stmt.Expr) + " ) {\n"
		self.indent++
		out += self.Visit(stmt.Body)
		self.indent--
		out += self.IndentSpace() + "}"
		return out
	}
	if stmt, ok := n.(ast.ReturnStatement); ok {
		return "return " + self.Visit(stmt.Expr) + ";\n"
	}
	if fn, ok := n.(ast.Function); ok {
		var out string = ""
		out += self.IndentSpace() + "function " + fn.Name + "("
		if len(fn.Params) > 0 {
			fields := []string{}
			for _, param := range fn.Params {
				field := "$" + param.Name
				if param.Type != "" {
					field = param.Type + " " + field
				}
				fields = append(fields, field)
			}
			out += strings.Join(fields, ", ")
		}
		out += ") {\n"
		self.indent++
		out += self.Visit(fn.Body)
		self.indent--
		out += self.IndentSpace() + "}\n"
		return out
	}
	return ""
}
