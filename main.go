package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type JobDescription struct {
	wg *sync.WaitGroup
}

func main() {
	start := time.Now()
	app := NewApp()
	if len(os.Args) <= 1 {
		panic("Expected filename")
	}
	//var wg sync.WaitGroup
	jobs := make(chan string, 100)
	results := make(chan string, 100)
	for i := 0; i < 8; i++ {
		go worker(app, jobs, results)
	}
	numJobs := 0
	for i := 1; i < len(os.Args); i++ {
		filename := os.Args[i]
		if strings.HasSuffix(filename, ".cpp") {
			if i+1 < len(os.Args) && strings.HasSuffix(os.Args[i+1], ".in") {
				_, err := app.RunOne(filename, os.Args[i+1])
				if err != nil {
					os.Exit(1)
				}
				i++
			} else {
				res, err := app.Run(filename)
				res.PrintAll(os.Stdout)
				if err != nil {
					os.Exit(1)
				}
			}
		} else {
			numJobs++
			jobs <- filename
		}
	}
	close(jobs)
	for i := 0; i < numJobs; i++ {
		fmt.Println(<-results)
	}
	if numJobs > 0 {
		duration := time.Since(start)
		fmt.Println("elapsed time " + strconv.FormatInt(duration.Milliseconds(), 10) + "ms")
	}
}

func worker(app *App, jobs <-chan string, results chan<- string) {
	for j := range jobs {
		res, err := app.RunDir(j)
		buf := bytes.NewBufferString("")
		if err != nil {
			buf.WriteString("Error for ")
			buf.WriteString(j)
			buf.WriteString(err.Error())
		}
		if res != nil {
			res.PrintAll(buf)
			if err != nil {
				os.Exit(1)
			}
		} else {
			buf.WriteString("Empty result for ")
			buf.WriteString(j)
		}
		results <- strings.TrimSpace(buf.String())
	}
}
