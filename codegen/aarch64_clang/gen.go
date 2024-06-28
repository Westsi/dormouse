package aarch64_clang

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

// generates asm for aarch64 to be compiled with clang

type AARCH64Generator struct {
	fpath            string
	out              strings.Builder
	AST              ast.Program
	VirtualStack     *util.Stack[codegen.VTabVar]
	VirtualRegisters map[StorageLoc]string
	LabelCounter     int
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
)

var Sls = []StorageLoc{X0, X1, X2, X3, X4, X5, X6, X7, X8, X9, X10, X11, X12, X13, X14, X15, X16, X17, X18, X19, X20, X21, X22, X23, X24, X25, X26, X27, X28}

var StorageLocs = []string{"x0", "x1", "x2", "x3", "x4", "x5", "x6", "x7", "x8", "x9", "x10", "x11", "x12", "x13", "x14", "x15", "x16", "x17", "x18", "x19", "x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28"}

var FNCallRegs = []StorageLoc{} // i think this is just storage locations from x0 ascending

// TODO: change registers from x86_64 to aarch64
// I THINK:
// x29 - frame pointer - rbp
// sp - stack pointer - rsp
// x0 - return value - rax
// CPSR - eflags
// https://johannst.github.io/notes/arch/arm64.html

func New(fpath string, ast *ast.Program, defs map[string]string, lc int) *AARCH64Generator {
	generator := &AARCH64Generator{
		fpath:            fpath,
		out:              strings.Builder{},
		AST:              *ast,
		VirtualStack:     util.NewStack[codegen.VTabVar](),
		VirtualRegisters: map[StorageLoc]string{},
		LabelCounter:     lc,
		Gdefs:            defs,
	}
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
	return g.LabelCounter
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
}

func (g *AARCH64Generator) GenerateFunction(f *ast.FunctionDefinition) {
	defer tracer.Untrace(tracer.Trace("GenerateFunction"))
	oldVirtStack := g.VirtualStack
	g.VirtualStack = util.NewStack[codegen.VTabVar]()
	g.VirtualRegisters = map[StorageLoc]string{}

	if f.Name.Value == "main" {
		g.out.WriteString(".globl _main\n")
	}

	g.out.WriteString("_" + f.Name.Value + ":\n")

	// setup default local stack - TODO: figure out size of stack necessary
	g.out.WriteString("sub sp, sp, #16\n")
	// TODO: does storing wzr need to go here?
	// TODO: implement passing parameters
	g.GenerateBlock(f.Body)

	g.VirtualStack = oldVirtStack
	g.VirtualRegisters = map[StorageLoc]string{}
}

func (g *AARCH64Generator) GenerateExpression(node ast.Expression) StorageLoc {
	defer tracer.Untrace(tracer.Trace("GenerateExpression"))
	switch node := node.(type) {
	// case *ast.InfixExpression:
	// 	return g.GenerateInfix(node)
	case *ast.Identifier:
		return g.GenerateIdentifier(node)
	case *ast.IntegerLiteral:
		g.out.WriteString("mov w0, " + fmt.Sprintf("%d", node.Value) + "\n") //TODO: give this same treatment as the identifier case - picking regs
		// case *ast.IfExpression:
		// 	g.GenerateIf(node)
		// case *ast.WhileExpression:
		// 	g.GenerateWhileLoop(node)
	}
	return NULLSTORAGE
}

func (g *AARCH64Generator) GenerateBlock(b *ast.BlockStatement) {
	defer tracer.Untrace(tracer.Trace("GenerateBlock"))
	for _, stmt := range b.Statements {
		fmt.Println(stmt.NType())
		switch stmt := stmt.(type) {
		case *ast.FunctionDefinition:
			g.GenerateFunction(stmt)
		case *ast.VarStatement:
			g.GenerateVarDef(stmt)
		// case *ast.VarReassignmentStatement:
		// 	g.GenerateVarReassignment(stmt)
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
	if sloc != NULLSTORAGE && sloc != X0 {
		g.out.WriteString("mov" + "x0" + StorageLocs[sloc] + "\n")
	}
	g.out.WriteString("add sp, sp, #16\n")
	g.out.WriteString("ret\n")
}

func (g *AARCH64Generator) GenerateIdentifier(i *ast.Identifier) StorageLoc {
	tracer.Trace("GenerateIdentifier")
	defer tracer.Untrace("GenerateIdentifier")

	storageLoc, err := g.GetVarStorageLoc(i.Value)
	if storageLoc == DEFINES {
		for k, v := range g.Gdefs {
			if i.Value == k {
				g.out.WriteString(v)
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
