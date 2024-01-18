package main

import (
	"fmt"
	"os"

	"github.com/westsi/dormouse/lex"
	"github.com/westsi/dormouse/parse"
)

func main() {
	fname := os.Args[1]
	reader, err := os.Open(fname)
	if err != nil {
		panic(err)
	}

	lexer := lex.NewLexer(reader)
	tokens := lexer.Lex()

	for _, tok := range tokens {
		fmt.Println(tok.String())
	}

	p := parse.New(tokens)
	ast := p.Parse()
	fmt.Printf("Errors: %s\n", p.Errors())
	fmt.Println(ast.String())
}
