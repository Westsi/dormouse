package main

/*
TODO:
bring new label improvements to x86_64 for nested loops

*/

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/westsi/dormouse/builtin"
	"github.com/westsi/dormouse/codegen"
	"github.com/westsi/dormouse/codegen/aarch64_clang"
	"github.com/westsi/dormouse/codegen/x86_64_as"
	"github.com/westsi/dormouse/lex"
	"github.com/westsi/dormouse/parse"
	"github.com/westsi/dormouse/tracer"
)

var globalDefines = make(map[string]string)
var condcnt int = 0

func main() {
	opts := Options{}
	isVerbose := flag.Bool("v", false, "verbose")
	isDebug := flag.Bool("d", false, "debug")
	OutFname := flag.String("o", "", "output file name")
	targetArch := flag.String("a", "x86_64", "target architecture")
	flag.Parse()
	opts.Verbose = *isVerbose
	opts.Debug = *isDebug
	opts.OutFname = *OutFname
	opts.TargetArch = *targetArch
	opts.Fname = flag.Arg(0)
	if opts.OutFname == "" {
		opts.OutFname = strings.Split(strings.Split(opts.Fname, ".")[0], "/")[len(strings.Split(opts.Fname, "/"))-1] + ".s"
	}
	opts.BaseDir = strings.Join(strings.Split(opts.Fname, "/")[0:len(strings.Split(opts.Fname, "/"))-1], "/") + "/"
	opts.Fname = strings.Split((strings.Split(opts.Fname, "/")[len(strings.Split(opts.Fname, "/"))-1]), ".")[0]
	if opts.Debug {
		fmt.Println(opts)
	}
	run(opts)
}

func run(opts Options) {
	tracer.InitTrace(opts.Debug)

	lexers, _ := ResolveImports([]string{}, opts.BaseDir, opts.Fname)
	var asmNames []string

	for _, lexer := range lexers {
		if lexer == nil {
			continue
		}
		asmNames = append(asmNames, strings.Split((strings.Split(lexer.GetRdrFname(), "/")[len(strings.Split(lexer.GetRdrFname(), "/"))-1]), ".")[0]+".s")
		fmt.Println("Compiling", lexer.GetRdrFname())
		Compile(&opts, lexer)
	}
	CompileAll(opts, asmNames)
}

func CompileAll(opts Options, fnames []string) {
	outf, _ := os.Create("out/" + opts.TargetArch + "/asm/___concat.s")
	for _, fname := range fnames {
		rdr, err := os.Open("out/" + opts.TargetArch + "/asm/" + fname)
		if err != nil {
			panic(err)
		}
		defer rdr.Close()
		_, err = io.Copy(outf, rdr)
		if err != nil {
			panic(err)
		}
	}

	out, err := exec.Command("gcc", "-o"+"out/"+opts.TargetArch+"/"+strings.Split(opts.Fname, ".")[0], "out/"+opts.TargetArch+"/asm/___concat.s").Output()
	if err != nil {
		print(string(out))
	}
}

func Compile(opts *Options, lexer *lex.Lexer) {
	tokens, _, _ := lexer.Lex()
	fname := strings.Split((strings.Split(lexer.GetRdrFname(), "/")[len(strings.Split(lexer.GetRdrFname(), "/"))-1]), ".")[0]
	p := parse.New(tokens)
	ast := p.Parse()
	if len(p.Errors()) > 0 {
		fmt.Println("Errors:")
		for _, err := range p.Errors() {
			fmt.Println(err)
		}
		os.Exit(1)
	}

	// ssag := ssa.New(fname+".dssa", ast, globalDefines)
	// ssag.Generate()
	// ssag.Write()
	// os.Exit(0)
	fmt.Println(ast.String())
	var cg codegen.CodeGenerator
	switch opts.TargetArch {
	case "x86_64":
		cg = x86_64_as.New(fname+".s", ast, globalDefines, condcnt)
	case "aarch64":
		cg = aarch64_clang.New(fname+".s", ast, globalDefines, condcnt)
	}
	condcnt = cg.Generate()
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
	_, imported, defined := lexer.Lex()
	for k, v := range defined {
		if val, ok := globalDefines[k]; ok {
			// define already exists, warn of overwriting but allow it
			fmt.Printf("WARNING: %s has already been @defined as %s. It is being overwritten.", k, val)
		}
		globalDefines[k] = v
	}
	// fmt.Println("Imported:", imported)
	prevImps = append(prevImps, imp)
	for _, imp := range imported {
		if strings.HasPrefix(imp, "dor.") {
			l, pi, d := builtin.HandleStdlib(prevImps, strings.Replace(imp, "dor.", "", 1))
			for k, v := range d {
				if val, ok := globalDefines[k]; ok {
					// define already exists, warn of overwriting but allow it
					fmt.Printf("WARNING: %s has already been @defined as %s. It is being overwritten.", k, val)
				}
				globalDefines[k] = v
			}
			lexers = append(lexers, l)
			prevImps = append(prevImps, pi...)
			continue
		}
		l, pi := ResolveImports(prevImps, baseDir, imp)
		lexers = append(lexers, l...)
		prevImps = append(prevImps, pi...)
	}
	return lexers, prevImps
}

type Options struct {
	Verbose    bool
	Debug      bool
	Fname      string
	BaseDir    string
	OutFname   string
	TargetArch string
}
