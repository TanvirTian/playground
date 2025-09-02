package parser 

import (
	"playground/interpreter/token"
	"playground/interpreter/lexer"
	"playground/interpreter/ast"
)


type Parser struct {
	lexer        *lexer.Lexer
	currentToken token.Token
}

func NewParser(lexer *lexer.Lexer) *Parser {
	return &Parser{lexer: lexer, currentToken: lexer.GetNextToken()}
}

func (p *Parser) Eat(tokenType string) {
	if p.currentToken.Type == tokenType {
		p.currentToken = p.lexer.GetNextToken()
	} else {
		panic("Syntax Bhul: Abar Check Korun :(")
	}
}


func (p *Parser) Statement() ast.AST {
	if p.currentToken.Type == token.DHORO {
		p.Eat(token.DHORO)
		varName := p.currentToken.Value.(string)
		p.Eat(token.IDENT)
		p.Eat(token.ASSIGN)
		value := p.Expression()
		return ast.Assign{Name: varName, Value: value}
	}
	return p.Expression()
}


func (p *Parser) Factor() ast.AST {
    tok := p.currentToken

    if tok.Type == token.IDENT {
    	p.Eat(token.IDENT)
    	if p.currentToken.Type == token.ASSIGN {
    		p.Eat(token.ASSIGN)
    		return ast.Assign{Name: tok.Value.(string), Value: p.Expression()}
    	}

    	return ast.Var{Name: tok.Value.(string), Token: tok}
    }

    if tok.Type == token.JOG {
        p.Eat(token.JOG)
        node := ast.UnaryOp{Token: tok, Op: tok, Expression: p.Factor()}
        return node 
    } else if tok.Type == token.BIYOG {
        p.Eat(token.BIYOG)
        node := ast.UnaryOp{Token: tok, Op: tok, Expression: p.Factor()}
        return node 
    } else if tok.Type == token.INTEGER {
    	p.Eat(token.INTEGER)
    	return ast.Num{Token: tok, Value: tok.Value.(int)}
    } else if tok.Type == token.LPAREN {
    	p.Eat(token.LPAREN)
    	node := p.Expression()
    	p.Eat(token.RPAREN)
    	return node 
    }
  
    panic("Bhul Syntax: Bujhte partesi na ekhane ki ache :(")
}


func (p *Parser) term() ast.AST {
	node := p.Factor()
	for p.currentToken.Type == token.GUN || p.currentToken.Type == token.BHAG {
		tok := p.currentToken
		if tok.Type == token.GUN {
			p.Eat(token.GUN)
		} else if tok.Type == token.BHAG {
			p.Eat(token.BHAG)
		}
		node = ast.BinOp{Left: node, Op: tok, Right: p.Factor()}
	}
	return node
}

func (p *Parser) Expression() ast.AST {
	node := p.term()
	for p.currentToken.Type == token.JOG || p.currentToken.Type == token.BIYOG {
		tok := p.currentToken
		if tok.Type == token.JOG {
			p.Eat(token.JOG)
		} else if tok.Type == token.BIYOG {
			p.Eat(token.BIYOG)
		}
		node = ast.BinOp{Left: node, Op: tok, Right: p.term()}
	}
	return node
}


func (p *Parser) Parse() ast.AST {
	return p.Statement()
}