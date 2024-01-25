package x86_64_as

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/westsi/dormouse/ast"
	"github.com/westsi/dormouse/codegen"
	"github.com/westsi/dormouse/lex"
	"github.com/westsi/dormouse/tracer"
	"github.com/westsi/dormouse/util"
)

// generates asm for x86_64 to be compiled with as

type X64Generator struct {
	fpath               string
	out                 strings.Builder
	AST                 ast.Program
	VirtualStack        *util.Stack[codegen.VTabVar]
	VariableStorageLocs map[string]codegen.StorageLoc
}

func New(fpath string, ast *ast.Program) *X64Generator {
	generator := &X64Generator{
		fpath:               fpath,
		out:                 strings.Builder{},
		AST:                 *ast,
		VirtualStack:        util.NewStack[codegen.VTabVar](),
		VariableStorageLocs: map[string]codegen.StorageLoc{},
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
	out, err := exec.Command("gcc", "-o"+"out/x86_64/"+strings.Split(g.fpath, ".")[0], "out/x86_64/asm/"+g.fpath).Output()
	if err != nil {
		print(string(out))
	}
}

func (g *X64Generator) e(tok lex.LexedTok, err string) {
	panic(fmt.Sprintf("%v: %v - %s\n", tok.Pos, tok.Tok, err))
}

func (g *X64Generator) GetVarStackOffset(name string) int {
	tracer.Trace("GetVarStackOffset")
	defer tracer.Untrace("GetVarStackOffset")
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
	tracer.Trace("GetVTabVar")
	defer tracer.Untrace("GetVTabVar")
	for _, v := range g.VirtualStack.Elements {
		if v.Name == name {
			return v
		}
	}
	return codegen.VTabVar{}
}

func (g *X64Generator) Generate() {
	tracer.Trace("Generate")
	defer tracer.Untrace("Generate")
	for _, stmt := range g.AST.Statements {
		switch stmt := stmt.(type) {
		case *ast.FunctionDefinition:
			g.GenerateFunction(stmt)
		}
	}
}

func (g *X64Generator) GenerateExpression(node ast.Expression) {
	tracer.Trace("GenerateExpression")
	defer tracer.Untrace("GenerateExpression")
	switch node := node.(type) {
	case *ast.InfixExpression:
		g.GenerateInfix(node)
	case *ast.Identifier:
		g.GenerateIdentifier(node)
	case *ast.IntegerLiteral:
		g.out.WriteString("movq $" + fmt.Sprintf("%d", node.Value) + ", %rax\n")
	}
}

func (g *X64Generator) GenerateBlock(b *ast.BlockStatement) {
	tracer.Trace("GenerateBlock")
	defer tracer.Untrace("GenerateBlock")
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
	tracer.Trace("GenerateFunction")
	defer tracer.Untrace("GenerateFunction")
	g.out.WriteString(".type " + f.Name.Value + ", @function\n")
	g.out.WriteString(f.Name.Value + ":\n")
	// setup local stack for function
	g.out.WriteString("pushq %rbp\n")      // save old base pointer to stack
	g.out.WriteString("movq %rsp, %rbp\n") // use stack top pointer as base pointer for function
	g.GenerateBlock(f.Body)
}

func (g *X64Generator) GenerateVarDef(v *ast.VarStatement) {
	tracer.Trace("GenerateVarDef")
	defer tracer.Untrace("GenerateVarDef")
	// TODO: make this work with more than a simple int literal e.g. infix ops
	g.VirtualStack.Push(codegen.VTabVar{Name: v.Name.Value, Type: v.Type.Value})
	switch v.Type.Value {
	case "int":
		g.out.WriteString("pushq $" + v.Value.String() + "\n") // load value into stack
	}
}

func (g *X64Generator) GenerateIdentifier(i *ast.Identifier) {
	tracer.Trace("GenerateIdentifier")
	defer tracer.Untrace("GenerateIdentifier")

	storageLoc, err := g.VariableStorageLocs[i.Value]
	if err || storageLoc == codegen.Stack {
		offset := g.GetVarStackOffset(i.Value)
		if offset != -1 {
			g.out.WriteString("movq " + fmt.Sprintf("-%d", offset) + "(%rbp), %rax\n")
		} else {
			g.e(i.Token, "undefined variable: "+i.Value)
		}
		g.VariableStorageLocs[i.Value] = codegen.RAX
	} else {
		switch storageLoc {
		case codegen.RAX:
			g.out.WriteString("%rax")
		case codegen.RBX:
			g.out.WriteString("%rbx")
		case codegen.RCX:
			g.out.WriteString("%rcx")
		case codegen.RDX:
			g.out.WriteString("%rdx")
		}
	}
}

func (g *X64Generator) GenerateCall(c *ast.CallExpression) {
	tracer.Trace("GenerateCall")
	defer tracer.Untrace("GenerateCall")
}

func (g *X64Generator) GenerateReturn(r *ast.ReturnStatement) {
	tracer.Trace("GenerateReturn")
	defer tracer.Untrace("GenerateReturn")
	// clean up stack
	g.GenerateExpression(r.ReturnValue)
	g.out.WriteString("movq %rbp, %rsp\n")
	g.out.WriteString("popq %rbp\n")
	g.out.WriteString("ret\n")
}

func (g *X64Generator) GenerateExit(e *ast.Node) {
	tracer.Trace("GenerateExit")
	defer tracer.Untrace("GenerateExit")
	// TODO: is this correct?
	g.out.WriteString("movq $0, %rdi\n")
	g.out.WriteString("movq $60, %rax\n")
	g.out.WriteString("syscall\n")
}

func (g *X64Generator) GenerateLabel(node *ast.Node) {

}

func (g *X64Generator) GenerateInfix(node *ast.InfixExpression) {
	tracer.Trace("GenerateInfix")
	defer tracer.Untrace("GenerateInfix")
	var rightS, leftS string
	switch right := node.Right.(type) {
	case *ast.Identifier:
		g.GenerateIdentifier(right)
		rightS = codegen.StorageLocs[g.VariableStorageLocs[right.Value]]
	case *ast.IntegerLiteral:
		rightS = "$" + fmt.Sprintf("%d", right.Value)
	}

	switch left := node.Left.(type) {
	case *ast.Identifier:
		g.GenerateIdentifier(left)
		leftS = codegen.StorageLocs[g.VariableStorageLocs[left.Value]-1]
	case *ast.IntegerLiteral:
		leftS = "$" + fmt.Sprintf("%d", left.Value)
	}

	switch node.Operator {
	case "+":
		g.out.WriteString("addq ")
	case "-":
		g.out.WriteString("subq ")
	case "*":
		g.out.WriteString("imulq ")
		// TODO: implement division
	}
	g.out.WriteString(rightS + ", " + leftS + "\n")

	g.out.WriteString("\n")

}
