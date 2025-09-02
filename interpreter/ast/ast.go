package ast  

import (
	"playground/interpreter/token"
)

type AST interface{}

type BinOp struct {
	Left  AST
	Op    token.Token
	Right AST
}

type Num struct {
	Token token.Token
	Value int
}

type UnaryOp struct {
	Token      token.Token 
	Op 	       token.Token 
	Expression AST
}

type Var struct {
	Name  string 
	Token token.Token 
}

type Assign struct {
	Name  string 
	Value AST
}