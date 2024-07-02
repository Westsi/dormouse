package parse

import "github.com/westsi/dormouse/lex"

const (
	_ = iota
	LOWEST
	EQUALS
	LESSGREATER
	LOGICAL
	BITWISE
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[lex.Token]int{
	lex.EQUALS:    EQUALS,
	lex.NOTEQUALS: EQUALS,
	lex.LT:        LESSGREATER,
	lex.GT:        LESSGREATER,
	lex.AND:       LOGICAL,
	lex.OR:        LOGICAL,
	lex.NOT:       LOGICAL,
	lex.BWXOR:     BITWISE,
	lex.BWAND:     BITWISE,
	lex.BWNOT:     BITWISE,
	lex.BWOR:      BITWISE,
	lex.ADD:       SUM,
	lex.SUB:       SUM,
	lex.MUL:       PRODUCT,
	lex.DIV:       PRODUCT,
	lex.LPAREN:    CALL,
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekTok.Tok]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curTok.Tok]; ok {
		return p
	}
	return LOWEST
}
