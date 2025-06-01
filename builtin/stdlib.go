package builtin

import (
	"fmt"
	"os"

	"github.com/westsi/dormouse/lex"
)

func HandleStdlib(prevImps []string, imp string) (*lex.Lexer, []string, map[string]string) {
	for _, i := range prevImps {
		if i == imp {
			return nil, []string{}, make(map[string]string)
		}
	}
	wd, _ := os.Getwd()
	reader, err := os.Open(wd + "/builtin/" + imp + ".dor")
	if err != nil {
		fmt.Println("Imported file not found:", wd+imp+".dor")
		os.Exit(1)
	}

	l := lex.NewLexer(reader)
	_, imported, defined := l.Lex()
	d := make(map[string]string)
	for k, v := range defined {
		if val, ok := d[k]; ok {
			// define already exists, warn of overwriting but allow it
			fmt.Printf("WARNING: %s has already been @defined as %s. It is being overwritten.", k, val)
		}
		d[k] = v
	}

	fmt.Println(imported)
	return l, []string{}, d
}
