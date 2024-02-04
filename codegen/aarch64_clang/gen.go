package aarch64_clang

import (
	"strings"

	"github.com/westsi/dormouse/ast"
	"github.com/westsi/dormouse/codegen"
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

// TODO: change registers from x86_64 to aarch64
// x29 - frame pointer
// rsp -
