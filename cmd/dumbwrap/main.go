package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/angelokurtis/dumbwrap"
)

func main() {
	singlechecker.Main(dumbwrap.Analyzer)
}
