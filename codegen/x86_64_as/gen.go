package x86_64_as

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/westsi/dormouse/ast"
	"github.com/westsi/dormouse/codegen"
	"github.com/westsi/dormouse/lex"
	"github.com/westsi/dormouse/util"
)

// generates asm for x86_64 to be compiled with as

type X64Generator struct {
	fpath        string
	out          strings.Builder
	AST          ast.Program
	VirtualStack *util.Stack[codegen.VTabVar]
}

func New(fpath string, ast *ast.Program) *X64Generator {
	generator := &X64Generator{
		fpath:        fpath,
		out:          strings.Builder{},
		AST:          *ast,
		VirtualStack: util.NewStack[codegen.VTabVar](),
	}
	generator.out.WriteString(".text\n.globl main\n")
	os.MkdirAll("out/x86_64", os.ModePerm)
	os.MkdirAll("out/x86_64/asm", os.ModePerm)
	return generator
}

func (g *X64Generator) Write() {
	f, err := os.Create("out/x86_64/asm/" + g.fpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(g.out.String())
	if err != nil {
		panic(err)
	}
}

func (g *X64Generator) Compile() {
	cmd := exec.Command("gcc", "-o"+"out/x86_64/"+strings.Split(g.fpath, ".")[0], "out/x86_64/asm/"+g.fpath)
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

func (g *X64Generator) e(tok lex.LexedTok, err string) {
	panic(fmt.Sprintf("%v: %v - %s\n", tok.Pos, tok.Tok, err))
}

func (g *X64Generator) GetVarStackOffset(name string) int {
	// get offset from bottom of stack in terms of indices and multiply by sizes to get bits
	sizeBelow := 0
	for i, v := range g.VirtualStack.Elements {
		switch v.Type {
		case "int":
			sizeBelow += 8
		}

		if v.Name == name {
			return (i + 1) * sizeBelow
		}
	}
	return -1
}

func (g *X64Generator) GetVTabVar(name string) codegen.VTabVar {
	for _, v := range g.VirtualStack.Elements {
		if v.Name == name {
			return v
		}
	}
	return codegen.VTabVar{}
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
	g.out.WriteString(".type " + f.Name.Value + ", @function\n")
	g.out.WriteString(f.Name.Value + ":\n")
	// setup local stack for function
	g.out.WriteString("pushq %rbp\n")      // save old base pointer to stack
	g.out.WriteString("movq %rsp, %rbp\n") // use stack top pointer as base pointer for function
	g.GenerateBlock(f.Body)
}

func (g *X64Generator) GenerateVarDef(v *ast.VarStatement) {
	g.VirtualStack.Push(codegen.VTabVar{Name: v.Name.Value, Type: v.Type.Value})
	switch v.Type.Value {
	case "int":
		g.out.WriteString("pushq $" + v.Value.String() + "\n") // load value into stack
	}
}

func (g *X64Generator) GenerateIdentifier(i *ast.Identifier) {
	// TODO: fix
	offset := g.GetVarStackOffset(i.Value)
	if offset != -1 {
		g.out.WriteString("movq " + fmt.Sprintf("-%d", offset) + "(%rbp), %rax\n")
	} else {
		g.e(i.Token, "undefined variable: "+i.Value)
	}
}

func (g *X64Generator) GenerateCall(c *ast.CallExpression) {

}

func (g *X64Generator) GenerateReturn(r *ast.ReturnStatement) {
	// clean up stack
	switch val := r.ReturnValue.(type) {
	case *ast.Identifier:
		offset := g.GetVarStackOffset(val.Value)
		if offset != -1 {
			g.out.WriteString("movq " + fmt.Sprintf("-%d", offset) + "(%rbp), %rax\n")
		} else {
			g.e(val.Token, "undefined variable: "+val.Value)
		}
	case *ast.IntegerLiteral:
		g.out.WriteString("movq $" + fmt.Sprintf("%d", val.Value) + ", %rax\n")
	}
	g.out.WriteString("movq %rbp, %rsp\n")
	g.out.WriteString("popq %rbp\n")
	g.out.WriteString("ret\n")

}

func (g *X64Generator) GenerateExit(e *ast.Node) {

}

func (g *X64Generator) GenerateLabel(node *ast.Node) {

}
