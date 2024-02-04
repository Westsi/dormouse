package x86_64_as

import (
	"fmt"
	"os"
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
	VirtualRegisters map[StorageLoc]string
	LabelCounter     int
}

type StorageLoc int

const (
	// Stack StorageLoc = iota
	RAX StorageLoc = iota
	RCX
	RDX
	RDI
	RSI
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
	NULLSTORAGE
)

var Sls = []StorageLoc{RAX, RCX, RDX, RDI, RSI, R8, R9, R10, R11, R12, R13, R14, R15}

var StorageLocs = []string{"%rax", "%rcx", "%rdx", "%rdi", "%rsi", "%r8", "%r9", "%r10", "%r11", "%r12", "%r13", "%r14", "%r15"}

var FNCallRegs = []StorageLoc{RDI, RSI, RDX, RCX, R8, R9}

func New(fpath string, ast *ast.Program) *X64Generator {
	generator := &X64Generator{
		fpath:            fpath,
		out:              strings.Builder{},
		AST:              *ast,
		VirtualStack:     util.NewStack[codegen.VTabVar](),
		VirtualRegisters: map[StorageLoc]string{},
		LabelCounter:     0,
	}
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

func (g *X64Generator) e(tok lex.LexedTok, err string) {
	fmt.Printf("%v: %v - %s\n", tok.Pos, tok.Tok, err)
	os.Exit(1)
}

func (g *X64Generator) GetVarStackOffset(name string) int {
	tracer.Trace("GetVarStackOffset")
	defer tracer.Untrace("GetVarStackOffset")
	// get offset from bottom of stack in terms of indices and multiply by sizes to get bits
	sizeBelow := 0
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

func (g *X64Generator) GetVarStorageLoc(name string) (StorageLoc, error) {
	tracer.Trace("GetVarStorageLoc")
	defer tracer.Untrace("GetVarStorageLoc")
	for k, v := range g.VirtualRegisters {
		if v == name {
			return k, nil
		}
	}
	return NULLSTORAGE, fmt.Errorf("undefined variable: %s", name)
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

func (g *X64Generator) GenerateExpression(node ast.Expression) StorageLoc {
	tracer.Trace("GenerateExpression")
	defer tracer.Untrace("GenerateExpression")
	switch node := node.(type) {
	case *ast.InfixExpression:
		return g.GenerateInfix(node)
	case *ast.Identifier:
		return g.GenerateIdentifier(node)
	case *ast.IntegerLiteral:
		g.out.WriteString("movq $" + fmt.Sprintf("%d", node.Value) + ", %rax\n") //TODO: give this same treatment as the identifier case - picking regs
	case *ast.IfExpression:
		g.GenerateIf(node)
	case *ast.WhileExpression:
		g.GenerateWhileLoop(node)
	}
	return NULLSTORAGE
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
		case *ast.VarReassignmentStatement:
			g.GenerateVarReassignment(stmt)
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
	// save old virtual stack but assume all registers other than rsp, rbp are clobbered
	oldVirtStack := g.VirtualStack
	g.VirtualStack = util.NewStack[codegen.VTabVar]()
	g.VirtualRegisters = map[StorageLoc]string{}

	if f.Name.Value == "main" {
		g.out.WriteString(".text\n.globl main\n")
	}

	g.out.WriteString(".type " + f.Name.Value + ", @function\n")
	g.out.WriteString(f.Name.Value + ":\n")
	// setup local stack for function
	g.out.WriteString("pushq %rbp\n")      // save old base pointer to stack
	g.out.WriteString("movq %rsp, %rbp\n") // use stack top pointer as base pointer for function
	// move params to stack and set virtual stack
	for i, param := range f.Parameters {
		g.out.WriteString("pushq " + StorageLocs[FNCallRegs[i]] + "\n")
		g.VirtualStack.Push(codegen.VTabVar{Name: param.Name.Value, Type: param.Type.Value})
	}
	g.GenerateBlock(f.Body)
	// restore old virtual stack
	g.VirtualStack = oldVirtStack
	g.VirtualRegisters = map[StorageLoc]string{}
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

func (g *X64Generator) GenerateIdentifier(i *ast.Identifier) StorageLoc {
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

func (g *X64Generator) LoadIdentFromStack(i *ast.Identifier, offset int) StorageLoc {
	tracer.Trace("LoadIdentFromStack")
	defer tracer.Untrace("LoadIdentFromStack")
	var reg StorageLoc
	for _, v := range Sls {
		_, ok := g.VirtualRegisters[v]
		if !ok {
			g.VirtualRegisters[v] = i.Value
			reg = v
			break
		}
	}
	g.out.WriteString("movq " + fmt.Sprintf("-%d", offset) + "(%rbp), " + StorageLocs[reg] + "\n")
	return reg
}

func (g *X64Generator) GenerateCall(c *ast.CallExpression) {
	tracer.Trace("GenerateCall")
	defer tracer.Untrace("GenerateCall")
	for i, arg := range c.Arguments {
		sloc := g.GenerateExpression(arg)
		if sloc != NULLSTORAGE {
			g.out.WriteString("movq " + StorageLocs[sloc] + ", " + StorageLocs[FNCallRegs[i]] + "\n")
		}
	}
	g.out.WriteString("call " + c.Function.Value + "\n")
	// g.out.WriteString("addq $" + fmt.Sprintf("%d", len(c.Arguments)*8) + ", %rsp\n")
}

func (g *X64Generator) GenerateReturn(r *ast.ReturnStatement) {
	tracer.Trace("GenerateReturn")
	defer tracer.Untrace("GenerateReturn")
	// clean up stack
	sloc := g.GenerateExpression(r.ReturnValue)
	if sloc != NULLSTORAGE && sloc != RAX {
		g.out.WriteString("movq " + StorageLocs[sloc] + ", %rax\n")
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

func (g *X64Generator) GenerateInfix(node *ast.InfixExpression) StorageLoc {
	tracer.Trace("GenerateInfix")
	defer tracer.Untrace("GenerateInfix")
	leftS, rightS, destLoc := g.GetInfixOperands(node)

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
	return destLoc
}

func (g *X64Generator) GetInfixOperands(node *ast.InfixExpression) (string, string, StorageLoc) {
	tracer.Trace("GetInfixOperands")
	defer tracer.Untrace("GetInfixOperands")
	var leftS, rightS string
	var leftLoc StorageLoc
	switch left := node.Left.(type) {
	case *ast.Identifier:
		leftLoc = g.GenerateIdentifier(left)
		leftS = StorageLocs[leftLoc]
	case *ast.IntegerLiteral:
		leftS = "$" + fmt.Sprintf("%d", left.Value)
	}

	switch right := node.Right.(type) {
	case *ast.Identifier:
		rightS = StorageLocs[g.GenerateIdentifier(right)]
	case *ast.IntegerLiteral:
		rightS = "$" + fmt.Sprintf("%d", right.Value)
	}
	return leftS, rightS, leftLoc
}

func (g *X64Generator) GenerateIf(i *ast.IfExpression) {
	tracer.Trace("GenerateIf")
	defer tracer.Untrace("GenerateIf")
	// check if condition is true
	// to do this, check what the comparative expr is and generate the corresponding jump instruction
	separator := i.Condition.(*ast.InfixExpression).Operator
	leftS, rightS, _ := g.GetInfixOperands(i.Condition.(*ast.InfixExpression))

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
	case "<=":
		g.out.WriteString("jle " + predictedTrueLabel + "\n")
	case ">=":
		g.out.WriteString("jge " + predictedTrueLabel + "\n")
	}
	g.GenerateBlock(i.Alternative)
	g.out.WriteString("jmp " + predictedEndLabel + "\n")
	g.GenerateLabel()
	g.GenerateBlock(i.Consequence)
	g.GenerateLabel()

}

func (g *X64Generator) GenerateVarReassignment(v *ast.VarReassignmentStatement) {
	tracer.Trace("GenerateVarReassignment")
	defer tracer.Untrace("GenerateVarReassignment")
	// find the variable's location in the stack
	offset := g.GetVarStackOffset(v.Name.Value)
	// update it with the new value
	switch v.Value.(type) {
	case *ast.IntegerLiteral:
		g.out.WriteString("movq $" + fmt.Sprintf("%d", v.Value.(*ast.IntegerLiteral).Value) + ", " + fmt.Sprintf("-%d(%%rbp)", offset) + "\n")
	case *ast.Identifier:
		g.out.WriteString("movq " + StorageLocs[g.GenerateIdentifier(v.Value.(*ast.Identifier))] + ", " + fmt.Sprintf("-%d(%%rbp)", offset) + "\n")
	case *ast.CallExpression:
		g.GenerateCall(v.Value.(*ast.CallExpression))
		g.out.WriteString("movq " + "%rax" + ", " + fmt.Sprintf("-%d(%%rbp)", offset) + "\n")
	case *ast.InfixExpression:
		sloc := g.GenerateInfix(v.Value.(*ast.InfixExpression))
		g.out.WriteString("movq " + StorageLocs[sloc] + ", " + fmt.Sprintf("-%d(%%rbp)", offset) + "\n")
	}
	// remove the old value from any registers
	sloc, _ := g.GetVarStorageLoc(v.Name.Value)
	if sloc != NULLSTORAGE {
		g.out.WriteString("movq $0, " + StorageLocs[sloc] + "\n")
		delete(g.VirtualRegisters, sloc)
	}
}

func (g *X64Generator) GenerateWhileLoop(w *ast.WhileExpression) {
	tracer.Trace("GenerateWhileLoop")
	defer tracer.Untrace("GenerateWhileLoop")
	// generate jump to condition check
	// generate body
	// generate condition check
	// e.g.
	// jmp .L1
	// .L2:
	// ...
	// .L1:
	// cmpq $2, %rax
	// jle .L2
	// ...

	predictedBodyLabel := fmt.Sprintf(".L%d", g.LabelCounter)
	predictedConditionLabel := fmt.Sprintf(".L%d", g.LabelCounter+1)

	g.out.WriteString("jmp " + predictedConditionLabel + "\n")

	g.GenerateLabel()
	g.GenerateBlock(w.Body)

	g.GenerateLabel()
	separator := w.Condition.(*ast.InfixExpression).Operator
	leftS, rightS, _ := g.GetInfixOperands(w.Condition.(*ast.InfixExpression))
	g.out.WriteString("cmpq " + rightS + ", " + leftS + "\n")
	switch separator {
	case "==":
		g.out.WriteString("je " + predictedBodyLabel + "\n")
	case "!=":
		g.out.WriteString("jne " + predictedBodyLabel + "\n")
	case "<":
		g.out.WriteString("jl " + predictedBodyLabel + "\n")
	case ">":
		g.out.WriteString("jg " + predictedBodyLabel + "\n")
	case "<=":
		g.out.WriteString("jle " + predictedBodyLabel + "\n")
	case ">=":
		g.out.WriteString("jge " + predictedBodyLabel + "\n")
	}

}

// TODO: add support in parser and generator for while loops and then for loops
