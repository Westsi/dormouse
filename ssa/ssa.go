package ssa

import (
	"os"
	"strings"

	"github.com/westsi/dormouse/ast"
)

type SSAGen struct {
	fpath   string
	out     strings.Builder
	AST     ast.Program
	Gdefs   map[string]string
	vcounts map[string]int
}

type DType int

const (
	VAR DType = iota
	CONST
)

type Dependent struct {
	Type  DType
	Value string // holds int val itoa or var name
}

func (s *SSAGen) ow(st string) {
	s.out.WriteString(st)
}

func New(fpath string, ast *ast.Program, defs map[string]string) *SSAGen {
	generator := &SSAGen{
		fpath:   fpath,
		out:     strings.Builder{},
		AST:     *ast,
		Gdefs:   defs,
		vcounts: make(map[string]int),
	}
	os.MkdirAll("out/ssa", os.ModePerm)
	return generator
}

func (s *SSAGen) Generate() {
	for _, stmt := range s.AST.Statements {
		switch stmt := stmt.(type) {
		case *ast.FunctionDefinition:
			s.ProcessFunction(stmt)
		}
	}
}

func (s *SSAGen) Write() {
	f, err := os.Create("out/ssa/" + s.fpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(s.out.String())
	if err != nil {
		panic(err)
	}
}

/*
Var statement steps
- check what its set to
- figure out whether that depends on any other variables
- generate new variable with value


Reassignment steps
- same as above except create new indexed value

int x = 6
int y = x + 2
int z = y + x
x = z + 1

v1 = 6
v2 = v1 + 2
v3 = v2 + v1
v4 = v3 + 1

.
.
.
eventually simplifies down to
v4 = 15
*/
