package x86_64_as

import (
	"strings"

	"github.com/westsi/dormouse/ast"
	"github.com/westsi/dormouse/codegen"
	"github.com/westsi/dormouse/util"
)

// generates asm for x86_64 to be compiled with as

type X64Generator struct {
	fpath        string
	out          strings.Builder
	AST          ast.Program
	VirtualStack *util.Stack[codegen.VTabVar]
}

func New(fpath string, ast ast.Program) *X64Generator {
	generator := &X64Generator{
		fpath:        fpath,
		out:          strings.Builder{},
		AST:          ast,
		VirtualStack: util.NewStack[codegen.VTabVar](),
	}
	generator.out.WriteString(".text\n.globl main\n")
	return generator
}

func (g *X64Generator) GetVarStackOffset(name string) int {
	for i, v := range g.VirtualStack.Elements {
		if v.Name == name {
			return g.VirtualStack.Size() - i
		}
	}
	return -1
}

func (g *X64Generator) Generate() {
	for _, stmt := range g.AST.Statements {
		switch stmt := stmt.(type) {
		case *ast.FunctionDefinition:
			g.GenerateFunction(stmt)
		}
	}
}

func (g *X64Generator) GenerateBlock(b *ast.BlockStatement) {
	// setup local stack for function
	g.out.WriteString("pushq %rbp\n")      // save old base pointer to stack
	g.out.WriteString("movq %rsp, %rbp\n") // use stack top pointer as base pointer for function

	for _, stmt := range b.Statements {
		switch stmt := stmt.(type) {
		case *ast.FunctionDefinition:
			g.GenerateFunction(stmt)
		case *ast.VarStatement:
			g.GenerateVarDef(stmt)
		case *ast.ReturnStatement:
			g.GenerateReturn(stmt)
		}
	}
}

func (g *X64Generator) GenerateFunction(f *ast.FunctionDefinition) {
	g.out.WriteString(".type " + f.Name.Literal() + ", @function\n")
	g.out.WriteString(f.Name.Literal() + ":\n")
	g.GenerateBlock(f.Body)
}

func (g *X64Generator) GenerateVarDef(v *ast.VarStatement) {
	g.VirtualStack.Push(codegen.VTabVar{Name: v.Name.Literal(), Type: v.Type.Value})
	switch v.Type.Value {
	case "int":
		g.out.WriteString("movl $" + v.Value.Literal() + ", -4(%rbp)\n") // load value into stack
	}
}

func (g *X64Generator) GenerateIdentifier(i *ast.Identifier) {

}

func (g *X64Generator) GenerateCall(c *ast.CallExpression) {

}

func (g *X64Generator) GenerateReturn(r *ast.ReturnStatement) {

}

func (g *X64Generator) GenerateExit(e *ast.Node) {

}

func (g *X64Generator) GenerateLabel(node *ast.Node) {

}
