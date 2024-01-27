package tracer

import (
	"fmt"
	"strings"
)

var tracing bool

var traceLevel int = 0

const traceIdentPlaceholder string = "\t"

func identLevel() string {
	return strings.Repeat(traceIdentPlaceholder, traceLevel-1)
}

func tracePrint(fs string) {
	fmt.Printf("%s%s\n", identLevel(), fs)
}

func incIdent() { traceLevel = traceLevel + 1 }
func decIdent() { traceLevel = traceLevel - 1 }

func Trace(msg string) string {
	if !tracing {
		return msg
	}
	incIdent()
	tracePrint("BEGIN " + msg)
	return msg
}

func Untrace(msg string) {
	if !tracing {
		return
	}
	tracePrint("END " + msg)
	decIdent()
}

func InitTrace(isDebug bool) {
	tracing = isDebug
}
