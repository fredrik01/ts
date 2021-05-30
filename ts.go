package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	version              = "0.1.3"
	storageFolder        = ".config/ts"
	layoutDateTime       = "2006-01-02 15:04:05"
	column1Width         = 19
	column2Width         = 12
	column3Width         = 13
	column1MinWidthNamed = 6
	column2WidthNamed    = 19
	column3WidthNamed    = 12
	column4WidthNamed    = 13
)

type nameAndDate struct {
	name string
	date time.Time
}

var usage = `Usage: ts [command] [argument]

  Commands:
    add			Add timestamp to default stopwatch or to a named one (ts save mystopwatch)
    show		Show default stopwatch timestamps or a named one (ts show mystopwatch)
    reset		Reset default stopwatch or a named one (ts reset mystopwatch)
    reset-all		Reset all stopwatches
    list		List stopwatches
    combine		Show all stopwatches in one sorted list
	  version		Print version
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		return
	}

	command := flag.Arg(0)
	name := flag.Arg(1)
	if name == "" {
		name = "default"
	}
	setupStorage()
	runCommand(command, name)
}

func setupStorage() {
	storagePath := getStoragePath()
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		os.Mkdir(storagePath, 0700)
	}
}

func getStoragePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	var storage strings.Builder
	storage.WriteString(home)
	storage.WriteString("/")
	storage.WriteString(storageFolder)
	return storage.String()
}

func runCommand(command string, name string) {
	switch command {
	case "add":
		add(name)
	case "show":
		show(name)
	case "combine":
		combine()
	case "reset":
		reset(name)
	case "reset-all":
		resetAll()
	case "list":
		list(name)
	case "version":
		fmt.Println(version)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

// Commands

func add(name string) {
	fmt.Println("Timestamp added")
	fmt.Printf(name)
	fmt.Printf(": ")
	t := time.Now()
	fmt.Println(t.Local().Format(layoutDateTime))
	appendToFile(getFilePath(name), t.Local().Format(layoutDateTime))
}

func show(name string) {
	filePath := getFilePath(name)
	if _, err := os.Stat(filePath); err == nil {
		printHeaders()
		timestamps := readFile(filePath)
		printTimestamps(timestamps)
	} else {
		fmt.Println("This stopwatch is not running")
	}
}

func combine() {
	var allTimestamps []nameAndDate
	for _, filename := range getTimestampFiles() {
		filePath := getFilePathForFilename(filename)
		timestamps := readFile(filePath)
		nameAndDates := convertToNameAndDateSlice(getNameFromFilename(filename), timestamps)
		allTimestamps = append(allTimestamps, nameAndDates...)
	}
	sort.Slice(allTimestamps, func(i, j int) bool {
		return allTimestamps[i].date.Before(allTimestamps[j].date)
	})
	printHeadersNamed(allTimestamps)
	printNameAndDates(allTimestamps)
}

func reset(name string) {
	filename := getFilePath(name)
	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("Reset ")
		fmt.Printf(name)
		fmt.Printf("? (y/n) ")
		ok := askForConfirmation()

		if ok {
			e := os.Remove(filename)
			if e != nil {
				log.Fatal(e)
			}
			fmt.Println("Done")
		} else {
			fmt.Println("Aborted")
		}
	} else {
		fmt.Println("This stopwatch is not running")
	}
}

func resetAll() {
	fmt.Printf("Reset all? (y/n) ")
	ok := askForConfirmation()

	if ok {
		for _, filename := range getTimestampFiles() {
			if strings.Contains(filename, ".timestamps-") {
				path := getFilePathForFilename(filename)
				e := os.Remove(path)
				if e != nil {
					log.Fatal(e)
				}
			}
		}
		fmt.Println("Done")
	} else {
		fmt.Println("Aborted")
	}
}

func list(name string) {
	for _, filename := range getTimestampFiles() {
		if strings.Contains(filename, ".timestamps-") {
			fmt.Println(getNameFromFilename(filename))
		}
	}
}

// Helper functions

func convertToNameAndDateSlice(name string, timestamps []time.Time) []nameAndDate {
	var nameAndDates []nameAndDate
	for _, dateTime := range timestamps {
		nameAndDates = append(nameAndDates, nameAndDate{name: name, date: dateTime})
	}
	return nameAndDates
}

func getNameFromFilename(filename string) string {
	return filename[12:]
}

func getTimestampFiles() []string {
	files, err := ioutil.ReadDir(getStoragePath())
	if err != nil {
		log.Fatal(err)
	}

	var timestampFiles []string
	for _, file := range files {
		filename := file.Name()
		if strings.Contains(filename, ".timestamps-") {
			timestampFiles = append(timestampFiles, filename)
		}
	}
	return timestampFiles
}

func askForConfirmation() bool {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

func getFilePathForFilename(filename string) string {
	var sb strings.Builder
	sb.WriteString(getStoragePath())
	sb.WriteString("/")
	sb.WriteString(filename)
	return sb.String()
}

func getFilePath(name string) string {
	var sb strings.Builder
	sb.WriteString(getStoragePath())
	sb.WriteString("/")
	sb.WriteString(".timestamps-")
	sb.WriteString(name)
	return sb.String()
}

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

func printHeadersNamed(timestamps []nameAndDate) {
	column1WidthNamed := getColumn1WidthNamedLength(timestamps)
	fmt.Printf("%-*s", column1WidthNamed, "Name")
	fmt.Printf("%-*s", column2WidthNamed, "Timestamp")
	fmt.Printf("%*s", column3WidthNamed, "Since prev")
	fmt.Printf("%*s", column4WidthNamed, "Since first")
	fmt.Printf("\n")
}

func readFile(filePath string) []time.Time {
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
	var timestamps []time.Time

	for scanner.Scan() {
		dateString := scanner.Text()
		dateTime, err := time.Parse(layoutDateTime, dateString)

		if err != nil {
			fmt.Println(err)
		}
		timestamps = append(timestamps, dateTime)
	}
	return timestamps
}

func printTimestamps(timestamps []time.Time) {
	prevLineTimeExists := false
	prevLineTime := time.Now()
	firstLineTime := time.Now()

	for _, lineTime := range timestamps {
		dateString := lineTime.Format(layoutDateTime)
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
		now, err := time.Parse(layoutDateTime, nowString)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%-*s", column1Width, "Now")
		fmt.Printf("%*s", column2Width, now.Sub(prevLineTime).String())
		fmt.Printf("%*s", column3Width, now.Sub(firstLineTime).String())
	}
}

func getLengthOfLongestName(timestamps []nameAndDate) int {
	longest := 0
	for _, timestamp := range timestamps {
		if len(timestamp.name) > longest {
			longest = len(timestamp.name)
		}
	}
	return longest
}

func getColumn1WidthNamedLength(timestamps []nameAndDate) int {
	column1Length := getLengthOfLongestName(timestamps) + 2
	if column1MinWidthNamed > column1Length {
		column1Length = column1MinWidthNamed
	}
	return column1Length
}

func printNameAndDates(timestamps []nameAndDate) {
	column1WidthNamed := getColumn1WidthNamedLength(timestamps)
	prevLineTimeExists := false
	var prevLineTime nameAndDate
	var firstLineTime nameAndDate

	for _, lineTime := range timestamps {
		dateString := lineTime.date.Format(layoutDateTime)
		fmt.Printf("%-*s", column1WidthNamed, lineTime.name)
		fmt.Printf("%*s", column2WidthNamed, dateString)

		// TODO: Rename to isFirst?
		if prevLineTimeExists {
			diffSincePrev := lineTime.date.Sub(prevLineTime.date)
			fmt.Printf("%*s", column3WidthNamed, diffSincePrev.String())
			fmt.Printf("%*s", column4WidthNamed, lineTime.date.Sub(firstLineTime.date).String())
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
		now, err := time.Parse(layoutDateTime, nowString)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%*s", column1WidthNamed, "")
		fmt.Printf("%-*s", column2WidthNamed, "Now")
		fmt.Printf("%*s", column3WidthNamed, now.Sub(prevLineTime.date).String())
		fmt.Printf("%*s", column4WidthNamed, now.Sub(firstLineTime.date).String())
	}
}
