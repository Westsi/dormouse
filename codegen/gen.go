package codegen

import (
	"github.com/westsi/dormouse/ast"
)

type CodeGenerator interface {
	Generate()
	GenerateBlock(b *ast.BlockStatement)
	GenerateFunction(f *ast.FunctionDefinition)
	GenerateVarDef(v *ast.VarStatement)
	GenerateIdentifier(i *ast.Identifier)
	GenerateCall(c *ast.CallExpression)
	GenerateReturn(r *ast.ReturnStatement)
	GenerateExit(e *ast.Node)
	GenerateLabel(node *ast.Node)
	GenerateInfix(node *ast.Node)
	GenerateIf(i *ast.IfExpression)
}

type VTabVar struct {
	Name string
	Type string
}

type StorageLoc int

const (
	// Stack StorageLoc = iota
	RAX StorageLoc = iota
	RBX
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

var Sls = []StorageLoc{RAX, RBX, RCX, RDX, RDI, RSI, R8, R9, R10, R11, R12, R13, R14, R15}

var StorageLocs = []string{"%rax", "%rbx", "%rcx", "%rdx", "%rdi", "%rsi", "%r8", "%r9", "%r10", "%r11", "%r12", "%r13", "%r14", "%r15"}

var FNCallRegs = []StorageLoc{RDI, RSI, RDX, RCX, R8, R9}
