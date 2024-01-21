package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/westsi/dormouse/codegen/x86_64_as"
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

	outFname := strings.Split(strings.Split(fname, ".")[0], "/")[len(strings.Split(fname, "/"))-1] + ".s"

	fmt.Println(outFname)

	cg := x86_64_as.New(outFname, ast)
	cg.Generate()
	cg.Write()
	cg.Compile()
}
