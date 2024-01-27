package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/westsi/dormouse/codegen/x86_64_as"
	"github.com/westsi/dormouse/lex"
	"github.com/westsi/dormouse/parse"
	"github.com/westsi/dormouse/tracer"
)

func main() {
	opts := Options{}
	isVerbose := flag.Bool("v", false, "verbose")
	isDebug := flag.Bool("d", false, "debug")
	OutFname := flag.String("o", "", "output file name")
	flag.Parse()
	opts.Verbose = *isVerbose
	opts.Debug = *isDebug
	opts.OutFname = *OutFname
	opts.Fname = flag.Arg(0)
	if opts.OutFname == "" {
		opts.OutFname = strings.Split(strings.Split(opts.Fname, ".")[0], "/")[len(strings.Split(opts.Fname, "/"))-1] + ".s"
	}
	fmt.Println(opts)
	run(opts)
}

func run(opts Options) {
	tracer.InitTrace(opts.Debug)
	reader, err := os.Open(opts.Fname)
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
	fmt.Println("Errors:")
	for _, err := range p.Errors() {
		fmt.Println(err)
	}
	if len(p.Errors()) > 0 {
		os.Exit(1)
	}
	fmt.Println(ast.String())

	cg := x86_64_as.New(opts.OutFname, ast)
	cg.Generate()
	cg.Write()
	cg.Compile()
}

type Options struct {
	Verbose  bool
	Debug    bool
	Fname    string
	OutFname string
}
