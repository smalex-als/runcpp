package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (app *App) ValidateOutput(res, output string) bool {
	data, err := os.ReadFile(output)
	if err != nil {
		panic(fmt.Sprintf("Failed to read output file: %V", err))
	}
	return compareString(string(data), res)
}

type RunResultLine struct {
	Value string
	Error error
}

type RunResult struct {
	Lines       []RunResultLine
	Compiled    bool
	TestsPassed int
	TestsFailed int
}

func (runResult *RunResult) Append(value string) {
	runResult.Lines = append(runResult.Lines, RunResultLine{Value: value})
}

func (runResult *RunResult) AppendError(err error, value string) {
	runResult.Lines = append(runResult.Lines, RunResultLine{Value: value, Error: err})
}

func (runResult *RunResult) PrintAll(writer io.Writer) {
	for _, item := range runResult.Lines {
		writer.Write([]byte(item.Value))
		writer.Write([]byte("\n"))
		if item.Error != nil {
			writer.Write([]byte(item.Error.Error()))
		}
	}
}

const GREEN = "\033[0;32m"
const RED = "\033[0;31m"
const NC = "\033[0m"

func (app *App) RunDir(dir string) (*RunResult, error) {
	result := &RunResult{}
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, errors.New("should be specified existing dir")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".cpp") {
			res, err := app.Run(filepath.Join(dir, entry.Name()))
			if err != nil {
				result.Append(RED + "FAILED" + NC + " " + entry.Name())
			} else {
				result.Append(GREEN + "PASSED" + NC + " " + entry.Name() + " Tests " + strconv.Itoa(res.TestsPassed))
			}
		}
	}
	return result, nil
}

func (app *App) Run(filename string) (*RunResult, error) {
	dir := filepath.Dir(filename)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	uid := strconv.Itoa(r.Intn(1_000_000_000))
	binaryFilename := filepath.Join(os.TempDir(), "test"+uid)
	result := &RunResult{}
	defer os.Remove(binaryFilename)
	err := app.CompileSource(filename, binaryFilename)
	if err != nil {
		result.AppendError(err, fmt.Sprintf("%scompile failed%s: %s", RED, NC, filename))
		return result, err
	}
	result.Compiled = true
	result.Append(fmt.Sprint(filename, " compiled"))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".in") {
			continue
		}
		input := filepath.Join(dir, entry.Name())
		result.Append(fmt.Sprint("run ", filepath.Base(filename), " ", entry.Name()))
		res, err := app.executeCode(binaryFilename, input)
		if err != nil {
			log.Println("Execution failed:", err)
			log.Println(res)
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".in")
		output := filepath.Join(dir, name+".out")
		if _, err := os.Stat(output); err == nil {
			if app.ValidateOutput(res, output) {
				result.Append(GREEN + "PASSED" + NC)
				result.TestsPassed++
			} else {
				result.Append(RED + "FAILED" + NC)
				result.TestsFailed++
			}
		}
	}
	return result, nil
}

func dumpString(prefix string, a []string) {
	for i := 0; i < len(a); i++ {
		fmt.Println(prefix, ">", a[i])
	}
}

func compareString(s1, s2 string) bool {
	a1 := strings.Split(s1, "\n")
	a2 := strings.Split(s2, "\n")
	for i := 0; i < len(a1); i++ {
		a1[i] = strings.TrimSpace(a1[i])
	}
	for i := 0; i < len(a2); i++ {
		a2[i] = strings.TrimSpace(a2[i])
	}
	if len(a1) != len(a2) {
		dumpString("ex", a1)
		dumpString("real", a2)
		return false
	}
	for i := 0; i < len(a1); i++ {
		if a1[i] != a2[i] {
			dumpString("ex", a1)
			dumpString("real", a2)
			return false
		}
	}
	return true
}

type CompileError struct {
	Message string
}

func (e *CompileError) Error() string {
	return fmt.Sprint(e.Message)
}

func (app *App) CompileSource(filename string, binaryFilename string) error {
	args := fmt.Sprintf("-std=c++17 -O2 -Wall %s -o %s", filename, binaryFilename)
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, "g++", strings.Split(args, " ")...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return &CompileError{Message: string(out)}
	}
	return nil
}

func (app *App) executeCode(binaryFilename string, filenameInput string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5000*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, binaryFilename)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		file, err := os.Open(filenameInput)
		defer file.Close()
		if err != nil {
			return
		}
		io.Copy(stdin, file)
	}()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
