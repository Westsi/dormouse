package aarch64_clang

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/westsi/dormouse/ast"
	"github.com/westsi/dormouse/codegen"
	"github.com/westsi/dormouse/lex"
	"github.com/westsi/dormouse/tracer"
	"github.com/westsi/dormouse/util"
)

// generates asm for aarch64 to be compiled with clang

type AARCH64Generator struct {
	fpath            string
	out              strings.Builder
	data             strings.Builder
	AST              ast.Program
	VirtualStack     *util.Armstack[codegen.VTabVar]
	VirtualRegisters map[StorageLoc]string
	ConditionCounter int
	StringCounter    int
	Gdefs            map[string]string
}

type StorageLoc int

const (
	// Stack StorageLoc = iota
	X0 StorageLoc = iota
	X1
	X2
	X3
	X4
	X5
	X6
	X7
	X8
	X9
	X10
	X11
	X12
	X13
	X14
	X15
	X16
	X17
	X18 // WAS commented out - why?
	X19
	X20
	X21
	X22
	X23
	X24
	X25
	X26
	X27
	X28 // TODO: check if any of these are reserved/have other uses
	NULLSTORAGE
	DEFINES
	DATASECT
)

var Sls = []StorageLoc{X0, X1, X2, X3, X4, X5, X6, X7, X8, X9, X10, X11, X12, X13, X14, X15, X16, X17, X18, X19, X20, X21, X22, X23, X24, X25, X26, X27, X28}

var StorageLocs = []string{"x0", "x1", "x2", "x3", "x4", "x5", "x6", "x7", "x8", "x9", "x10", "x11", "x12", "x13", "x14", "x15", "x16", "x17", "x18", "x19", "x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28"}

var FNCallRegs = []StorageLoc{X0, X1, X2, X3, X4, X5, X6, X7}

// https://johannst.github.io/notes/arch/arm64.html

func New(fpath string, ast *ast.Program, defs map[string]string, cc int) *AARCH64Generator {
	generator := &AARCH64Generator{
		fpath:            fpath,
		out:              strings.Builder{},
		data:             strings.Builder{},
		AST:              *ast,
		VirtualStack:     util.NewAStack[codegen.VTabVar](32),
		VirtualRegisters: map[StorageLoc]string{},
		ConditionCounter: cc,
		StringCounter:    0,
		Gdefs:            defs,
	}
	generator.out.WriteString(".text\n")
	generator.data.WriteString(".data\n")
	os.MkdirAll("out/aarch64", os.ModePerm)
	os.MkdirAll("out/aarch64/asm", os.ModePerm)
	return generator
}

func (g *AARCH64Generator) Generate() int {
	defer tracer.Untrace(tracer.Trace("Generate"))
	for _, stmt := range g.AST.Statements {
		switch stmt := stmt.(type) {
		case *ast.FunctionDefinition:
			g.GenerateFunction(stmt)
		}
	}
	return g.ConditionCounter
}

func (g *AARCH64Generator) e(tok lex.LexedTok, err string) {
	fmt.Printf("%v: %v - %s\n", tok.Pos, tok.Tok, err)
	os.Exit(1)
}

func (g *AARCH64Generator) Write() {
	f, err := os.Create("out/aarch64/asm/" + g.fpath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(g.out.String())
	if err != nil {
		panic(err)
	}
	_, err = f.WriteString(g.data.String())
	if err != nil {
		panic(err)
	}
}

func (g *AARCH64Generator) GenerateFunction(f *ast.FunctionDefinition) {
	defer tracer.Untrace(tracer.Trace("GenerateFunction"))
	oldVirtStack := g.VirtualStack
	g.VirtualStack = util.NewAStack[codegen.VTabVar](32)
	g.VirtualRegisters = map[StorageLoc]string{}

	if f.Name.Value == "main" {
		g.out.WriteString(".globl _main\n")
	}

	g.out.WriteString("_" + f.Name.Value + ":\n")

	// setup default local stack - TODO: figure out size of stack necessary
	g.out.WriteString("sub sp, sp, #32\n")
	// TODO: does storing wzr need to go here?
	// TODO: implement passing parameters
	g.GenerateBlock(f.Body)

	g.VirtualStack = oldVirtStack
	g.VirtualRegisters = map[StorageLoc]string{}
}

func (g *AARCH64Generator) GenerateExpression(node ast.Expression) StorageLoc {
	defer tracer.Untrace(tracer.Trace("GenerateExpression"))
	switch node := node.(type) {
	case *ast.InfixExpression:
		return g.GenerateInfix(node)
	case *ast.Identifier:
		return g.GenerateIdentifier(node)
	case *ast.IntegerLiteral:
		// g.out.WriteString("mov x0, " + fmt.Sprintf("#%d", node.Value) + "\n") //TODO: give this same treatment as the identifier case - picking regs
		// return X0
		return g.GenerateIntegerLiteral(node)
	case *ast.StringLiteral:
		return g.GenerateStringLiteral(node)
	case *ast.IfExpression:
		g.GenerateIf(node)
	case *ast.WhileExpression:
		g.GenerateWhileLoop(node)
	case *ast.CallExpression:
		g.GenerateCall(node)
		return X0
	}
	return NULLSTORAGE
}

func (g *AARCH64Generator) GenerateBlock(b *ast.BlockStatement) {
	defer tracer.Untrace(tracer.Trace("GenerateBlock"))
	for _, stmt := range b.Statements {
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

func (g *AARCH64Generator) GenerateReturn(r *ast.ReturnStatement) {
	defer tracer.Untrace(tracer.Trace("GenerateReturn"))
	// clean up stack
	sloc := g.GenerateExpression(r.ReturnValue)
	if sloc != NULLSTORAGE && sloc != X0 && sloc != DATASECT {
		g.out.WriteString("mov " + "x0, " + StorageLocs[sloc] + "\n")
	}
	g.out.WriteString("add sp, sp, #32\n")
	g.out.WriteString("ret\n")
}

func (g *AARCH64Generator) GenerateIdentifier(i *ast.Identifier) StorageLoc {
	tracer.Trace("GenerateIdentifier")
	defer tracer.Untrace("GenerateIdentifier")

	storageLoc, err := g.GetVarStorageLoc(i.Value)
	if storageLoc == DEFINES {
		for k, v := range g.Gdefs {
			if i.Value == k {
				val, err := strconv.ParseInt(v, 0, 64)
				if err != nil {
					fmt.Println("error in parsing integer in define")
				}
				return g.GenerateIntegerLiteral(&ast.IntegerLiteral{Token: i.Token, Value: val})
			}
		}
	}

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

func (g *AARCH64Generator) GetVarStorageLoc(name string) (StorageLoc, error) {
	tracer.Trace("GetVarStorageLoc")
	defer tracer.Untrace("GetVarStorageLoc")
	for k, v := range g.VirtualRegisters {
		if v == name {
			return k, nil
		}
	}
	for k := range g.Gdefs {
		if k == name {
			return DEFINES, nil
		}
	}
	return NULLSTORAGE, fmt.Errorf("undefined variable: %s", name)
}

func (g *AARCH64Generator) GetVarStackOffset(name string) int {
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

func (g *AARCH64Generator) LoadIdentFromStack(i *ast.Identifier, offset int) StorageLoc {
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
	g.out.WriteString("ldr " + StorageLocs[reg] + ", [sp, " + fmt.Sprintf("#%d", offset) + "]\n")
	return reg
}

func (g *AARCH64Generator) GetNextEmptyStackLoc() int {
	for i := 8; i < g.VirtualStack.Size(); i += 8 {
		if (g.VirtualStack.Get(i) == codegen.VTabVar{}) {
			return i
		}
	}
	fmt.Println("Stack full!")
	os.Exit(1)
	return -1
}

func (g *AARCH64Generator) GenerateVarDef(v *ast.VarStatement) {
	tracer.Trace("GenerateVarDef")
	defer tracer.Untrace("GenerateVarDef")
	sloc := g.GenerateExpression(v.Value.(*ast.ExpressionStatement).Expression)
	if sloc == NULLSTORAGE {
		fmt.Println("\033[31mPROBLEM PANICCCCCCC\033[0m")
	}

	stackloc := g.GetNextEmptyStackLoc()
	g.VirtualStack.Set(codegen.VTabVar{Name: v.Name.Value, Type: v.Type.Value}, stackloc)

	switch v.Type.Value {
	case "int":
		g.out.WriteString("str " + StorageLocs[sloc] + ", [sp, #" + fmt.Sprintf("%d", stackloc) + "]\n")
	}
}

func (g *AARCH64Generator) GenerateInfix(node *ast.InfixExpression) StorageLoc {
	tracer.Trace("GenerateInfix")
	defer tracer.Untrace("GenerateInfix")
	leftS, rightS, destLoc := g.GetInfixOperands(node)

	switch node.Operator {
	case "+":
		g.out.WriteString("add ")
	case "-":
		g.out.WriteString("sub ")
	case "*":
		g.out.WriteString("mul ")
	case "/":
		g.out.WriteString("sdiv ")
	case "^":
		g.out.WriteString("eor ")
	case "&":
		g.out.WriteString("and ")
	case "|":
		g.out.WriteString("orr ")
	}
	// fmt.Println(leftS, rightS, destLoc)
	g.out.WriteString(StorageLocs[destLoc] + ", " + leftS + ", " + rightS + "\n")
	return destLoc
}

func (g *AARCH64Generator) GetInfixOperands(node *ast.InfixExpression) (string, string, StorageLoc) {
	tracer.Trace("GetInfixOperands")
	defer tracer.Untrace("GetInfixOperands")
	var leftS, rightS string
	var destLoc StorageLoc
	switch left := node.Left.(type) {
	case *ast.Identifier:
		leftS = StorageLocs[g.GenerateIdentifier(left)]
	case *ast.InfixExpression:
		leftS = StorageLocs[g.GenerateInfix(left)]
	case *ast.CallExpression:
		g.GenerateCall(left)
		leftS = StorageLocs[X0]
	case *ast.IntegerLiteral:
		leftS = StorageLocs[g.GenerateIntegerLiteral(left)]
	}

	switch right := node.Right.(type) {
	case *ast.Identifier:
		rightS = StorageLocs[g.GenerateIdentifier(right)]
	case *ast.IntegerLiteral:
		// rightS = "#" + fmt.Sprintf("%d", right.Value)
		rightS = StorageLocs[g.GenerateIntegerLiteral(right)]
	case *ast.InfixExpression:
		rightS = StorageLocs[g.GenerateInfix(right)]
	case *ast.CallExpression:
		g.GenerateCall(right)
		rightS = StorageLocs[X0]
	}

	for _, v := range Sls {
		_, ok := g.VirtualRegisters[v]
		if !ok {
			g.VirtualRegisters[v] = "TEMP"
			destLoc = v
			break
		}
	}
	return leftS, rightS, destLoc
}

func (g *AARCH64Generator) GenerateIntegerLiteral(il *ast.IntegerLiteral) StorageLoc {
	defer tracer.Untrace(tracer.Trace("GenerateIntegerLiteral"))
	var sloc StorageLoc
	for _, v := range Sls {
		_, ok := g.VirtualRegisters[v]
		if !ok {
			g.VirtualRegisters[v] = "TEMP"
			sloc = v
			break
		}
	}

	g.out.WriteString("mov " + StorageLocs[sloc] + ", " + fmt.Sprintf("#%d", il.Value) + "\n")

	return sloc
}

func (g *AARCH64Generator) GenerateStringLiteral(sl *ast.StringLiteral) StorageLoc {
	defer tracer.Untrace(tracer.Trace("GenerateStringLiteral"))
	g.data.WriteString("string" + strconv.Itoa(g.StringCounter) + ":\n")
	g.StringCounter++
	g.data.WriteString(".asciz \"" + sl.Value + "\"\n")
	return DATASECT
}

func (g *AARCH64Generator) GenerateIf(i *ast.IfExpression) {
	tracer.Trace("GenerateIf")
	defer tracer.Untrace("GenerateIf")
	// check if condition is true
	// to do this, check what the comparative expr is and generate the corresponding jump instruction
	separator := i.Condition.(*ast.InfixExpression).Operator
	leftS, rightS, _ := g.GetInfixOperands(i.Condition.(*ast.InfixExpression))

	// cmp reg, val
	// cset reg, operator
	// tbnz reg, bit number, true label
	// jump to false case label
	// true case code
	// jump to end of true case section

	// e.g.
	// cmp x8, #2
	// cset x8, gt		; set x8 to 1 if flags show le
	// tbnz x8, #0, LBB1; if bit #0 of reg x8 is not zero (hence is true), jump to LBB1
	// b LBB2
	// LBB1:
	// mov x9, #0
	// b LBB3
	// LBB2:
	// mov x9, #5
	// b LBB3
	// LBB3:
	// ...
	trueLabel := fmt.Sprintf("LBBif%dtrue", g.ConditionCounter)
	falseLabel := fmt.Sprintf("LBBif%dfalse", g.ConditionCounter)
	endLabel := fmt.Sprintf("LBBif%dend", g.ConditionCounter)
	g.ConditionCounter++

	g.out.WriteString("cmp " + leftS + ", " + rightS + "\n")
	g.out.WriteString("cset x8, ")
	switch separator {
	case "==":
		g.out.WriteString("eq\n")
	case "!=":
		g.out.WriteString("ne\n")
	case "<":
		g.out.WriteString("lt\n")
	case ">":
		g.out.WriteString("gt\n")
	case "<=":
		g.out.WriteString("le\n")
	case ">=":
		g.out.WriteString("ge\n")
	}
	g.out.WriteString("tbnz x8, #0, " + trueLabel + "\n")
	g.out.WriteString("b " + falseLabel + "\n")
	g.out.WriteString(trueLabel + ":\n")
	g.GenerateBlock(i.Consequence)
	g.out.WriteString("b " + endLabel + "\n")
	g.out.WriteString(falseLabel + ":\n")
	if i.Alternative != nil {
		g.GenerateBlock(i.Alternative)
	}
	g.out.WriteString("b " + endLabel + "\n")
	g.out.WriteString(endLabel + ":\n")
}

func (g *AARCH64Generator) GenerateWhileLoop(w *ast.WhileExpression) {
	defer tracer.Untrace(tracer.Trace("GenerateWhileLoop"))
	// 	b       LBB0_1          ; jump to LBB0_1
	// COMPAR: LBB0_1:                                 ; =>This Inner Loop Header: Depth=1
	// 	ldr     w8, [sp, #8]    ; load a to w8
	// 	subs    w8, w8, #5      ; subtract 5 from w8 and set flags
	// 	cset    w8, ge          ; set w8 to 1 if w8-5 is greater than or equal to 0, and 0 otherwise
	// 	tbnz    w8, #0, LBB0_3  ; test if bit 0 of w8 is 0, and if not exit loop (jump to LBB0_3)
	// 	b       LBB0_2
	// BODY: LBB0_2:                                 ;   in Loop: Header=BB0_1 Depth=1
	// 	ldr     w8, [sp, #4]    ; load s into w8
	// 	add     w8, w8, #1      ; increment s
	// 	str     w8, [sp, #4]    ; store s into sp#4
	// 	ldr     w8, [sp, #8]    ; load a into w8
	// 	add     w8, w8, #1      ; increment a
	// 	str     w8, [sp, #8]    ; store a into sp#8
	// 	b       LBB0_1          ; branch to checking of loop
	// END: LBB0_3:

	comparLabel := fmt.Sprintf("LBBwhile%dcompar", g.ConditionCounter)
	bodyLabel := fmt.Sprintf("LBBwhile%dbody", g.ConditionCounter)
	endLabel := fmt.Sprintf("LBBwhile%dend", g.ConditionCounter)
	g.out.WriteString("b " + comparLabel + "\n")
	g.out.WriteString(comparLabel + ":\n")
	g.GenerateComparisonCheck(w.Condition.(*ast.InfixExpression), bodyLabel, endLabel)
	g.out.WriteString(bodyLabel + ":\n")
	g.GenerateBlock(w.Body)
	g.out.WriteString("b " + comparLabel + "\n")
	g.out.WriteString(endLabel + ":\n")
}

func (g *AARCH64Generator) GenerateComparisonCheck(c *ast.InfixExpression, trueLab, falseLab string) {
	defer tracer.Untrace(tracer.Trace("GenerateComparisonCheck"))
	leftS, rightS, _ := g.GetInfixOperands(c)
	g.out.WriteString("cmp " + leftS + ", " + rightS + "\n")
	g.out.WriteString("cset x8, ")
	switch c.Operator {
	case "==":
		g.out.WriteString("eq\n")
	case "!=":
		g.out.WriteString("ne\n")
	case "<":
		g.out.WriteString("lt\n")
	case ">":
		g.out.WriteString("gt\n")
	case "<=":
		g.out.WriteString("le\n")
	case ">=":
		g.out.WriteString("ge\n")
	}
	g.out.WriteString("tbnz x8, #0, " + trueLab + "\n")
	g.out.WriteString("b " + falseLab + "\n")
	g.VirtualRegisters = map[StorageLoc]string{}
}

// TODO: nested ifs!!

func (g *AARCH64Generator) GenerateVarReassignment(v *ast.VarReassignmentStatement) {
	tracer.Trace("GenerateVarReassignment")
	defer tracer.Untrace("GenerateVarReassignment")
	// find the variable's location in the stack
	offset := g.GetVarStackOffset(v.Name.Value)
	// update it with the new value
	switch val := v.Value.(type) {
	case *ast.IntegerLiteral:
		sloc := g.GenerateIntegerLiteral(val)
		g.out.WriteString("str " + StorageLocs[sloc] + ", " + fmt.Sprintf("[sp, #%d]", offset) + "\n")
	case *ast.Identifier:
		g.out.WriteString("str " + StorageLocs[g.GenerateIdentifier(val)] + ", " + fmt.Sprintf("[sp, #%d]", offset) + "\n")
	case *ast.CallExpression:
		g.GenerateCall(val)
		g.out.WriteString("str " + "x0" + ", " + fmt.Sprintf("[sp, #%d]", offset) + "\n")
	case *ast.InfixExpression:
		sloc := g.GenerateInfix(v.Value.(*ast.InfixExpression))
		g.out.WriteString("str " + StorageLocs[sloc] + ", " + fmt.Sprintf("[sp, #%d]", offset) + "\n")
	}
	// remove the old value from any registers
	sloc, _ := g.GetVarStorageLoc(v.Name.Value)
	if sloc != NULLSTORAGE && sloc != DATASECT {
		g.out.WriteString("mov " + StorageLocs[sloc] + ", #0\n")
		delete(g.VirtualRegisters, sloc)
	}
}

func (g *AARCH64Generator) GenerateCall(c *ast.CallExpression) {
	tracer.Trace("GenerateCall")
	defer tracer.Untrace("GenerateCall")
	for i, arg := range c.Arguments {
		sloc := g.GenerateExpression(arg)
		if sloc != NULLSTORAGE && sloc != DATASECT {
			g.out.WriteString("mov " + StorageLocs[sloc] + ", " + StorageLocs[FNCallRegs[i]] + "\n")
		}
	}
	// save x29 (frame pointer) and x30 (link register, holds return address) to stack before calling and potentially overwriting them
	g.out.WriteString("stp x29, x30, [sp, #16]\n")
	g.out.WriteString("add x29, sp, #16\n")
	g.out.WriteString("bl _" + c.Function.Value + "\n")
	g.out.WriteString("ldp x29, x30, [sp, #16]\n")

	for sl := range g.VirtualRegisters {
		delete(g.VirtualRegisters, sl)
	}
	g.VirtualRegisters[X0] = "CALLRET"
}
