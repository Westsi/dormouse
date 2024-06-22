package codegen

type CodeGenerator interface {
	Generate() int
	Write()
}

type VTabVar struct {
	Name string
	Type string
}
