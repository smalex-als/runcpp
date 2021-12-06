package main

import (
	"strings"
	"testing"
)

func TestCompile(t *testing.T) {
	t.Log("TestCompile successful")

	filename := "./testdata/5.longest-palindromic-substring.cpp"
	app := NewApp()
	res, err := app.Run(filename)
	if err != nil || !res.Compiled {
		t.Errorf("compile failed %s", "?")
	} else {
		for _, line := range res.Lines {
			t.Log(strings.TrimSpace(line.Value))
		}
	}
}

func TestCompileFailed(t *testing.T) {
	t.Log("TestCompile failed")

	filename := "./testdata/139.word-break.cpp"
	app := NewApp()
	res, err := app.Run(filename)
	if err == nil || res.Compiled {
		t.Error("expected compile failed")
	}
	for _, line := range res.Lines {
		t.Log(line.Value)
		if line.Error != nil {
			t.Log(line.Error)
		}
	}
}
