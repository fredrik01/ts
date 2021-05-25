package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

const (
	layoutDateTime = "2006-01-02 15:04:05"
	column1Width   = 19
	column2Width   = 14
	column3Width   = 14
)

var usage = `Usage: ts <command>

Example:

ts save`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Arg(0)
	runCommand(command)
}

func runCommand(command string) {
	switch command {
	case "save":
		save()
	case "show":
		show()
	default:
		flag.Usage()
		os.Exit(1)
	}
}

// Commands

func save() {
	fmt.Println("Timestamp saved")
	t := time.Now()
	fmt.Println(t.Format(layoutDateTime))
	appendToFile("timestamps", t.Format(layoutDateTime))
}

func show() {
	printHeaders()
	readFile("timestamps")
}

// Helper functions

func appendToFile(file string, data string) {
	// https://golang.org/pkg/os/#example_OpenFile_append
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	var sb strings.Builder
	sb.WriteString(data)
	sb.WriteString("\n")
	if _, err := f.Write([]byte(sb.String())); err != nil {
		f.Close() // ignore error; Write error takes precedence
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func printHeaders() {
	fmt.Printf("%-*s", column1Width, "Timestamp")
	fmt.Printf("%*s", column2Width, "Since prev")
	fmt.Printf("%*s", column3Width, "Since first")
	fmt.Printf("\n")
}

func readFile(filePath string) {
	// https://stackoverflow.com/a/36111861
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(file)

	prevLineTimeExists := false
	prevLineTime := time.Now()
	firstLineTime := time.Now()

	for scanner.Scan() {
		dateString := scanner.Text()
		lineTime, err := time.Parse(layoutDateTime, dateString)

		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("%*s", column1Width, dateString)

		// TODO: Rename to isFirst?
		if prevLineTimeExists {
			diffSincePrev := lineTime.Sub(prevLineTime)
			fmt.Printf("%*s", column2Width, diffSincePrev.String())
			fmt.Printf("%*s", column3Width, lineTime.Sub(firstLineTime).String())
		} else {
			firstLineTime = lineTime
		}

		fmt.Printf("\n")

		prevLineTimeExists = true
		prevLineTime = lineTime
	}

	if prevLineTimeExists {
		now := time.Now()
		nowString := now.Format(layoutDateTime)
		now, err = time.Parse(layoutDateTime, nowString)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%-*s", column1Width, "Now")
		fmt.Printf("%*s", column2Width, now.Sub(prevLineTime).String())
		fmt.Printf("%*s", column3Width, now.Sub(firstLineTime).String())
	}
}
