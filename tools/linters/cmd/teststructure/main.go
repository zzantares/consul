package main

import (
	"github.com/hashicorp/consul/tools/linters/teststructure"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(teststructure.NewAnalyzer())
}
