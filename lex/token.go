package lex

import "fmt"

type Token int

const (
	EOF Token = iota
	ILLEGAL
	IDENT
	// language keywords
	VAR
	IF
	ELSE
	FOR
	WHILE
	RETURN
	BREAK
	CONTINUE
	AS
	TYPEDEF
	// end of language keywords
	// compiler directives
	IMPORT
	DEFINE
	EXTERN
	IFNDEF
	IFDEF
	ENDIF
	UNDEF
	// end of compiler directives
	TYPE
	ASSIGN
	ADD
	MUL
	SUB
	DIV
	MOD
	LPAREN
	RPAREN
	LSQRBRAC
	RSQRBRAC
	BLOCKSTART
	BLOCKEND
	INTLITERAL
	STRINGLITERAL
	NEWLINE
	AND
	NOT
	OR
	BWOR
	BWAND
	BWNOT
	BWXOR
	GT
	LT
	TRUE
	FALSE
	NOTEQUALS
	EQUALS
	COMMA
)

var tokens = []string{
	EOF:           "EOF",
	ILLEGAL:       "ILLEGAL",
	IDENT:         "IDENT",
	IF:            "IF",
	ELSE:          "ELSE",
	FOR:           "FOR",
	WHILE:         "WHILE",
	RETURN:        "RETURN",
	BREAK:         "BREAK",
	CONTINUE:      "CONTINUE",
	AS:            "AS",
	TYPEDEF:       "TYPEDEF",
	IMPORT:        "IMPORT",
	DEFINE:        "DEFINE",
	EXTERN:        "EXTERN",
	IFNDEF:        "IFNDEF",
	IFDEF:         "IFDEF",
	ENDIF:         "ENDIF",
	UNDEF:         "UNDEF",
	TYPE:          "TYPE",
	ASSIGN:        "ASSIGN",
	ADD:           "ADD",
	MUL:           "MUL",
	SUB:           "SUB",
	DIV:           "DIV",
	MOD:           "MOD",
	LPAREN:        "LPAREN",
	RPAREN:        "RPAREN",
	LSQRBRAC:      "LSQRBRAC",
	RSQRBRAC:      "RSQRBRAC",
	BLOCKSTART:    "BLOCKSTART",
	BLOCKEND:      "BLOCKEND",
	INTLITERAL:    "INTLITERAL",
	STRINGLITERAL: "STRINGLITERAL",
	NEWLINE:       "NEWLINE",
	AND:           "AND",
	NOT:           "NOT",
	OR:            "OR",
	BWOR:          "BWOR",
	BWAND:         "BWAND",
	BWNOT:         "BWNOT",
	BWXOR:         "BWXOR",
	GT:            "GT",
	LT:            "LT",
	TRUE:          "TRUE",
	FALSE:         "FALSE",
	NOTEQUALS:     "NOTEQUALS",
	EQUALS:        "EQUALS",
	COMMA:         "COMMA",
}
var keywords = []string{
	"if",
	"else",
	"for",
	"while",
	"return",
	"break",
	"continue",
	"as",
	"typedef",
}

var kwmap = map[string]Token{
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"while":    WHILE,
	"return":   RETURN,
	"break":    BREAK,
	"continue": CONTINUE,
	"as":       AS,
	"typedef":  TYPEDEF,
}

var types = []string{
	"string",
	"int",
	"float",
	"double",
	"bool",
	"void",
}

func (t Token) String() string {
	return tokens[t]
}

type LexedTok struct {
	Pos Position
	Tok Token
	Val string
}

func NewLexedTok(pos Position, tok Token, val string) LexedTok {
	return LexedTok{
		Pos: pos,
		Tok: tok,
		Val: val,
	}
}

func (lt *LexedTok) String() string {
	return fmt.Sprint(lt.Pos) + " " + lt.Tok.String() + " " + lt.Val
}

var datatypes = map[Token]string{
	INTLITERAL:    "int",
	STRINGLITERAL: "string",
	TRUE:          "bool",
	FALSE:         "bool",
}
