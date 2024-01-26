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
	R8
	R9
	R10
	R11
	R12
	R13
	R14
	R15
)

// var Sls = []StorageLoc{Stack, RAX, RBX, RCX, RDX, R8, R9, R10, R11, R12, R13, R14, R15}
var Sls = []StorageLoc{RAX, RBX, RCX, RDX, R8, R9, R10, R11, R12, R13, R14, R15}

var StorageLocs = []string{"%rax", "%rbx", "%rcx", "%rdx", "%r8", "%r9", "%r10", "%r11", "%r12", "%r13", "%r14", "%r15"}
