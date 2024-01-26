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
	fpath            string
	out              strings.Builder
	AST              ast.Program
	VirtualStack     *util.Stack[codegen.VTabVar]
	VirtualRegisters map[codegen.StorageLoc]string
	LabelCounter     int
}

func New(fpath string, ast *ast.Program) *X64Generator {
	generator := &X64Generator{
		fpath:            fpath,
		out:              strings.Builder{},
		AST:              *ast,
		VirtualStack:     util.NewStack[codegen.VTabVar](),
		VirtualRegisters: map[codegen.StorageLoc]string{},
		LabelCounter:     0,
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
	fmt.Println(g.VirtualStack.Elements)
	for _, v := range g.VirtualStack.Elements {
		switch v.Type {
		case "int":
			sizeBelow += 8
		}

		if v.Name == name {
			return sizeBelow
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

func (g *X64Generator) GetVarStorageLoc(name string) (codegen.StorageLoc, error) {
	tracer.Trace("GetVarStorageLoc")
	defer tracer.Untrace("GetVarStorageLoc")
	for k, v := range g.VirtualRegisters {
		if v == name {
			return k, nil
		}
	}
	return codegen.RAX, fmt.Errorf("undefined variable: %s", name)
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

func (g *X64Generator) GenerateExpression(node ast.Expression) codegen.StorageLoc {
	tracer.Trace("GenerateExpression")
	defer tracer.Untrace("GenerateExpression")
	switch node := node.(type) {
	case *ast.InfixExpression:
		g.GenerateInfix(node)
	case *ast.Identifier:
		return g.GenerateIdentifier(node)
	case *ast.IntegerLiteral:
		g.out.WriteString("movq $" + fmt.Sprintf("%d", node.Value) + ", %rax\n") //TODO: give this same treatment as the identifier case - picking regs
	case *ast.IfExpression:
		g.GenerateIf(node)
	}
	return codegen.NULLSTORAGE
}

func (g *X64Generator) GenerateBlock(b *ast.BlockStatement) {
	tracer.Trace("GenerateBlock")
	defer tracer.Untrace("GenerateBlock")
	for _, stmt := range b.Statements {
		fmt.Println(stmt.NType())
		switch stmt := stmt.(type) {
		case *ast.FunctionDefinition:
			g.GenerateFunction(stmt)
		case *ast.VarStatement:
			g.GenerateVarDef(stmt)
		case *ast.ReturnStatement:
			g.GenerateReturn(stmt)
		case *ast.ExpressionStatement:
			g.GenerateExpression(stmt.Expression)
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
		// TODO: this only works with infix ops in the def because of gcc optimizations afaik.
		g.out.WriteString("pushq $" + v.Value.String() + "\n") // load value into stack
	}
}

func (g *X64Generator) GenerateIdentifier(i *ast.Identifier) codegen.StorageLoc {
	tracer.Trace("GenerateIdentifier")
	defer tracer.Untrace("GenerateIdentifier")

	storageLoc, err := g.GetVarStorageLoc(i.Value)
	if err != nil {
		offset := g.GetVarStackOffset(i.Value)
		if offset != -1 {
			return g.LoadIdentFromStack(i, offset)
		} else {
			g.e(i.Token, "undefined variable: "+i.Value)
		}
	}
	return storageLoc
}

func (g *X64Generator) LoadIdentFromStack(i *ast.Identifier, offset int) codegen.StorageLoc {
	tracer.Trace("LoadIdentFromStack")
	defer tracer.Untrace("LoadIdentFromStack")
	var reg codegen.StorageLoc
	for _, v := range codegen.Sls {
		_, ok := g.VirtualRegisters[v]
		if !ok {
			g.VirtualRegisters[v] = i.Value
			fmt.Println(g.VirtualRegisters)
			reg = v
			break
		}
	}
	g.out.WriteString("movq " + fmt.Sprintf("-%d", offset) + "(%rbp), " + codegen.StorageLocs[reg] + "\n")
	return reg
}

func (g *X64Generator) GenerateCall(c *ast.CallExpression) {
	tracer.Trace("GenerateCall")
	defer tracer.Untrace("GenerateCall")
}

func (g *X64Generator) GenerateReturn(r *ast.ReturnStatement) {
	tracer.Trace("GenerateReturn")
	defer tracer.Untrace("GenerateReturn")
	// clean up stack
	fmt.Printf("returning %s\n", r.ReturnValue.String())
	sloc := g.GenerateExpression(r.ReturnValue)
	if sloc != codegen.NULLSTORAGE && sloc != codegen.RAX {
		g.out.WriteString("movq " + codegen.StorageLocs[sloc] + ", %rax\n")
	}
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

func (g *X64Generator) GenerateLabel() string {
	tracer.Trace("GenerateLabel")
	defer tracer.Untrace("GenerateLabel")
	g.out.WriteString(fmt.Sprintf(".L%d:\n", g.LabelCounter))
	g.LabelCounter++
	return fmt.Sprintf(".L%d", g.LabelCounter-1)
}

func (g *X64Generator) GenerateInfix(node *ast.InfixExpression) {
	tracer.Trace("GenerateInfix")
	defer tracer.Untrace("GenerateInfix")

	leftS, rightS := g.GetInfixOperands(node)

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

func (g *X64Generator) GetInfixOperands(node *ast.InfixExpression) (string, string) {
	tracer.Trace("GetInfixOperands")
	defer tracer.Untrace("GetInfixOperands")
	fmt.Println(node.Left)
	var leftS, rightS string
	switch left := node.Left.(type) {
	case *ast.Identifier:
		g.GenerateIdentifier(left)
		leftS = codegen.StorageLocs[g.GenerateIdentifier(left)]
	case *ast.IntegerLiteral:
		leftS = "$" + fmt.Sprintf("%d", left.Value)
	}

	switch right := node.Right.(type) {
	case *ast.Identifier:
		rightS = codegen.StorageLocs[g.GenerateIdentifier(right)]
	case *ast.IntegerLiteral:
		rightS = "$" + fmt.Sprintf("%d", right.Value)
	}
	return leftS, rightS
}

func (g *X64Generator) GenerateIf(i *ast.IfExpression) {
	tracer.Trace("GenerateIf")
	defer tracer.Untrace("GenerateIf")
	// check if condition is true
	// to do this, check what the comparative expr is and generate the corresponding jump instruction
	separator := i.Condition.(*ast.InfixExpression).Operator
	leftS, rightS := g.GetInfixOperands(i.Condition.(*ast.InfixExpression))

	// cmpl left, right
	// jump to true case label
	// false case code
	// jump to end of true case section

	// e.g.
	// cmpq -4(%rbp), %rax
	// je .L1
	// movq $1, %rax
	// jmp .L2
	// .L1:
	// movq $0, %rax
	// .L2:
	// ...

	predictedTrueLabel := fmt.Sprintf(".L%d", g.LabelCounter)
	predictedEndLabel := fmt.Sprintf(".L%d", g.LabelCounter+1)

	g.out.WriteString("cmpq " + rightS + ", " + leftS + "\n")
	switch separator {
	case "==":
		g.out.WriteString("je " + predictedTrueLabel + "\n")
	case "!=":
		g.out.WriteString("jne " + predictedTrueLabel + "\n")
	case "<":
		g.out.WriteString("jl " + predictedTrueLabel + "\n")
	case ">":
		g.out.WriteString("jg " + predictedTrueLabel + "\n")
	}
	g.GenerateBlock(i.Alternative)
	g.out.WriteString("jmp " + predictedEndLabel + "\n")
	g.GenerateLabel()
	g.GenerateBlock(i.Consequence)
	g.GenerateLabel()

}
