package main

import (
	"os"
	"strings"
)

func main() {
	app := NewApp()
	if len(os.Args) <= 1 {
		panic("Expected filename")
	}
	for i := 1; i < len(os.Args); i++ {
		filename := os.Args[i]
		if strings.HasSuffix(filename, ".cpp") {
			res, err := app.Run(filename)
			res.PrintAll(os.Stdout)
			if err != nil {
				os.Exit(1)
			}
		} else {
			res, err := app.RunDir(filename)
			res.PrintAll(os.Stdout)
			if err != nil {
				os.Exit(1)
			}
		}
	}
}
