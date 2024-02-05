package aarch64_clang

import (
	"os"
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
	x0 StorageLoc = iota
	x1
	x2
	x3
	x4
	x5
	x6
	x7
	x8
	x9
	x10
	x11
	x12
	x13
	x14
	x15
	x16
	x17
	x18
	x19
	x20
	x21
	x22
	x23
	x24
	x25
	x26
	x27
	x28 // TODO: check if any of these are reserved/have other uses
	NULLSTORAGE
)

var Sls = []StorageLoc{}

var StorageLocs = []string{}

var FNCallRegs = []StorageLoc{}

// TODO: change registers from x86_64 to aarch64
// I THINK:
// x29 - frame pointer - rsp
// sp - stack pointer - rbp
// x30 - link register
// pc - program counter
// https://johannst.github.io/notes/arch/arm64.html

func New(fpath string, ast *ast.Program) *AARCH64Generator {
	generator := &AARCH64Generator{
		fpath:            fpath,
		out:              strings.Builder{},
		AST:              *ast,
		VirtualStack:     util.NewStack[codegen.VTabVar](),
		VirtualRegisters: map[StorageLoc]string{},
		LabelCounter:     0,
	}
	os.MkdirAll("out/aarch64", os.ModePerm)
	os.MkdirAll("out/aarch64/asm", os.ModePerm)
	return generator
}

func (g *AARCH64Generator) Generate() {
}

func (g *AARCH64Generator) Write() {
}
