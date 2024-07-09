package ssa

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/westsi/dormouse/ast"
)

type SSAGen struct {
	fpath string
	out   strings.Builder
	AST   ast.Program
	Gdefs map[string]string
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

func New(fpath string, ast *ast.Program, defs map[string]string) *SSAGen {
	generator := &SSAGen{
		fpath: fpath,
		out:   strings.Builder{},
		AST:   *ast,
		Gdefs: defs,
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

func (s *SSAGen) ProcessFunction(f *ast.FunctionDefinition) {
	s.ProcessBlock(f.Body)
}

func (s *SSAGen) ProcessBlock(b *ast.BlockStatement) {
	for _, stmt := range b.Statements {
		switch stmt := stmt.(type) {
		case *ast.VarStatement:
			s.ProcessVarStatement(stmt)
		}
	}
}

func (s *SSAGen) ProcessVarStatement(v *ast.VarStatement) {
	dependents := s.GetDefinitionDependents(v.Value)
	fmt.Println(dependents)
}

func (s *SSAGen) GetDefinitionDependents(e ast.Expression) []Dependent {
	var deps []Dependent
	switch et := e.(type) {
	case *ast.InfixExpression:
		deps = append(deps, s.GetDefinitionDependents(et.Left)...)
		deps = append(deps, s.GetDefinitionDependents(et.Right)...)
	case *ast.IntegerLiteral:
		dep := Dependent{
			Type:  CONST,
			Value: strconv.Itoa(int(et.Value)),
		}
		deps = append(deps, dep)
	case *ast.Identifier:
		dep := Dependent{
			Type:  VAR,
			Value: et.Value,
		}
		deps = append(deps, dep)
	case *ast.ExpressionStatement:
		deps = append(deps, s.GetDefinitionDependents(et.Expression)...)
	default:
		fmt.Printf("%T\n", e)
	}
	return deps
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
