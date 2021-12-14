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
	if runResult != nil && len(runResult.Lines) > 0 {
		for _, item := range runResult.Lines {
			writer.Write([]byte(item.Value))
			writer.Write([]byte("\n"))
			if item.Error != nil {
				writer.Write([]byte(item.Error.Error()))
			}
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
	cntUnique := 0
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".cpp") {
			cntUnique++
		}
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".cpp") {
			res, err := app.Run(filepath.Join(dir, entry.Name()))
			line := ""
			if cntUnique > 1 {
				line += entry.Name() + " "
			} else {
				line += dir + " "
			}
			if err != nil {
				line += RED + "FAILED" + NC
			} else {
				if res.TestsPassed > 0 {
					line += GREEN + "PASSED" + NC + " " + strconv.Itoa(res.TestsPassed) + " "
				}
				if res.TestsFailed > 0 {
					line += RED + "FAILED" + NC + " " + strconv.Itoa(res.TestsFailed) + " "
				}
			}
			result.Append(line)
		}
	}
	return result, nil
}

func (app *App) RunOne(filename string, input string) (*RunResult, error) {
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
	res, err := app.executeCode(binaryFilename, input)
	if err != nil {
		log.Println("Execution failed:", err)
		log.Println(res)
	}
	fmt.Println(res)
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
	entries, err := os.ReadDir(dir)
	if err != nil {
		return result, err
	}
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".in") {
			continue
		}
		input := filepath.Join(dir, entry.Name())
		start := time.Now()
		res, err := app.executeCode(binaryFilename, input)
		if err != nil {
			log.Println("Execution failed:", err)
			log.Println(res)
			continue
		}
		duration := time.Since(start)
		elaTime := strconv.FormatInt(duration.Milliseconds(), 10) + "ms"
		name := strings.TrimSuffix(entry.Name(), ".in")
		output := filepath.Join(dir, name+".out")
		if _, err := os.Stat(output); err == nil {
			if app.ValidateOutput(res, output) {
				result.Append(GREEN + "PASSED" + NC + " " + name + " " + elaTime)
				result.TestsPassed++
				// dumpString("ex", a1)
			} else {
				result.Append(RED + "FAILED" + NC + " " + name + " " + elaTime)
				result.TestsFailed++
				result.Append(res)
			}
		} else {
			result.Append(res)
			result.Append("elapsed time " + elaTime)
		}
		// fmt.Println(res)
	}
	return result, nil
}

func dumpString(prefix string, a []string) {
	for i := 0; i < len(a); i++ {
		fmt.Println(prefix, ">", a[i])
	}
}

func cleanOutput(s string) []string {
	res := make([]string, 0)
	a := strings.Split(s, "\n")
	for i := 0; i < len(a); i++ {
		str := strings.TrimSpace(a[i])
		if len(str) != 0 {
			res = append(res, str)
		}
	}
	return res
}
func compareString(s1, s2 string) bool {
	a1 := cleanOutput(s1)
	a2 := cleanOutput(s2)
	if len(a1) != len(a2) {
		return false
	}
	for i := 0; i < len(a1); i++ {
		if a1[i] != a2[i] {
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
	args := fmt.Sprintf("-std=c++17 -O2 -DLOCAL -Wall %s -o %s", filename, binaryFilename)
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
