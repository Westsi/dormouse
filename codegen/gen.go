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
	Stack StorageLoc = iota
	RAX
	RBX
	RCX
	RDX
)

var StorageLocs = []string{"%rax", "%rbx", "%rcx", "%rdx"}
