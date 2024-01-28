package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
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
	opts.BaseDir = strings.Join(strings.Split(opts.Fname, "/")[0:len(strings.Split(opts.Fname, "/"))-1], "/") + "/"
	opts.Fname = strings.Split((strings.Split(opts.Fname, "/")[len(strings.Split(opts.Fname, "/"))-1]), ".")[0]
	fmt.Println(opts)
	run(opts)
}

func run(opts Options) {
	tracer.InitTrace(opts.Debug)

	lexers, _ := ResolveImports([]string{}, opts.BaseDir, opts.Fname)
	var asmNames []string

	for _, lexer := range lexers {
		asmNames = append(asmNames, strings.Split((strings.Split(lexer.GetRdrFname(), "/")[len(strings.Split(lexer.GetRdrFname(), "/"))-1]), ".")[0]+".s")
		fmt.Println("Compiling", lexer.GetRdrFname())
		Compile(lexer)
	}
	CompileAll(opts, asmNames)
}

func CompileAll(opts Options, fnames []string) {
	outf, _ := os.Create("out/x86_64/asm/___concat.s")
	for _, fname := range fnames {
		rdr, err := os.Open("out/x86_64/asm/" + fname)
		if err != nil {
			panic(err)
		}
		defer rdr.Close()
		_, err = io.Copy(outf, rdr)
		if err != nil {
			panic(err)
		}
	}

	out, err := exec.Command("gcc", "-o"+"out/x86_64/"+strings.Split(opts.Fname, ".")[0], "out/x86_64/asm/___concat.s").Output()
	if err != nil {
		print(string(out))
	}
}

func Compile(lexer *lex.Lexer) {
	tokens, _ := lexer.Lex()
	fname := strings.Split((strings.Split(lexer.GetRdrFname(), "/")[len(strings.Split(lexer.GetRdrFname(), "/"))-1]), ".")[0]
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

	cg := x86_64_as.New(fname+".s", ast)
	cg.Generate()
	cg.Write()

}

func ResolveImports(prevImps []string, baseDir, imp string) ([]*lex.Lexer, []string) {
	var lexers []*lex.Lexer
	for _, i := range prevImps {
		if i == imp {
			return lexers, prevImps
		}
	}
	reader, err := os.Open(baseDir + imp + ".dor")
	if err != nil {
		fmt.Println("Imported file not found:", baseDir+imp+".dor")
		os.Exit(1)
	}

	lexer := lex.NewLexer(reader)
	lexers = append(lexers, lexer)
	_, imported := lexer.Lex()
	fmt.Println("Imported:", imported)
	prevImps = append(prevImps, imp)
	for _, imp := range imported {
		l, pi := ResolveImports(prevImps, baseDir, imp)
		lexers = append(lexers, l...)
		prevImps = append(prevImps, pi...)
	}
	return lexers, prevImps
}

type Options struct {
	Verbose  bool
	Debug    bool
	Fname    string
	BaseDir  string
	OutFname string
}
