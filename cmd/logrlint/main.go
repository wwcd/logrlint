package main

import (
	"github.com/wwcd/logrlint"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(logrlint.Analyzer)
}
