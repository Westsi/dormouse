package codegen

type CodeGenerator interface {
	Generate()
	Write()
}

type VTabVar struct {
	Name string
	Type string
}
