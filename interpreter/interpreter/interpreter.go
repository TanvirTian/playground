package interpreter

import (
	"playground/interpreter/token"
	"playground/interpreter/parser"
	"playground/interpreter/ast"
)

type Interpreter struct {
	variables  map[string]int 
}


func NewInterpreter() *Interpreter {
	return &Interpreter{variables: make(map[string]int)}
}

func (i *Interpreter) visit(node ast.AST) int {
	switch n := node.(type) {
	case ast.BinOp:
		if n.Op.Type == token.JOG {
			return i.visit(n.Left) + i.visit(n.Right)
		} else if n.Op.Type == token.BIYOG {
			return i.visit(n.Left) - i.visit(n.Right)
		} else if n.Op.Type == token.GUN {
			return i.visit(n.Left) * i.visit(n.Right)
		} else if n.Op.Type == token.BHAG {
			return i.visit(n.Left) / i.visit(n.Right)
		}
	case ast.UnaryOp:
		if n.Op.Type == token.JOG {
			return +i.visit(n.Expression)
		} else if n.Op.Type == token.BIYOG {
			return -i.visit(n.Expression)
		}
	case ast.Num:
		return n.Value

	case ast.Var:
		val, ok := i.variables[n.Name]
		if !ok {
			panic("Variable hoyni: " + n.Name)
		}
		return val

	case ast.Assign:
		val := i.visit(n.Value)
		i.variables[n.Name] = val 
		return val 		
	}
	panic("Oooops: Kichu Ekta Missing :(")
}


func (i *Interpreter) Interpret(p *parser.Parser) int {
	tree := p.Parse()
	return i.visit(tree)
}