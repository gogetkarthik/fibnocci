package main

import (
	"fmt"
	"github.com/fibonacci/pkg/fibonacci/cmd"
	"os"
)

func main() {
	fmt.Println(cmd.NewFibonacci("fibonacci").Run(os.Args))
}
