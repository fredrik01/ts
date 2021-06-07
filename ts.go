package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/fredrik01/ts/src/storage"
	"github.com/fredrik01/ts/src/timezone"
)

const (
	version              = "0.8.1"
	layoutDateTime       = "2006-01-02 15:04:05"
	minNameColumnWidth   = 6
	timestampColumnWidth = 19
	prevDiffColumnWidth  = 11
	firstDiffColumnWidth = 11
	nowDiffColumnWidth   = 11
	nameHeader           = "Name"
	timestampHeader      = "Timestamp"
	prevDiffHeader       = "D.prev"
	firstDiffHeader      = "D.first"
	nowDiffHeader        = "D.now"
)

type nameAndDate struct {
	name string
	date time.Time
}

type tsConfig struct {
	prevDiff  bool
	firstDiff bool
	nowDiff   bool
}

var config = tsConfig{
	prevDiff:  false,
	firstDiff: false,
	nowDiff:   false,
}

var usage = `Usage: ts [command] [flags] [arguments]

  Commands:
    add		Add timestamp to default stopwatch or to a named one (ts save mystopwatch)

    show	Show default stopwatch timestamps or a named one (ts show mystopwatch)
    		-all		Print all stopwatches
    		-diff-prev	Diff all rows against previous row in the list
    		-diff-first	Diff all rows against first row
    		-diff-now	Diff all rows against current time
    		-combine	Show all or some stopwatches in a sorted list. Additional arguments can be used to only keep some stopwatches in the list (ts show -combine mystop)
    		-combine-exact	Use exact matching for additional arguments when combining

    reset	Reset default stopwatch or a named one (ts reset mystopwatch)
    		-all		Reset all stopwatches

    rename	Rename a stopwatch (ts rename oldname newname)

    edit	Edit a stopwatch using the editor in your $EDITOR environment variable. Timestamps are stored in UTC.

    list	List stopwatches

    timezone	Set timezone (ts timezone "America/New_York")
    		-reset		Reset previous timezone settings and use the local timezone

    version	Print version
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)

	showCmd := flag.NewFlagSet("show", flag.ExitOnError)
	showAllFlag := showCmd.Bool("all", false, "Show all")
	showPrevDiffFlag := showCmd.Bool("diff-prev", false, "Diff all rows against previous row in the list")
	showFirstDiffFlag := showCmd.Bool("diff-first", false, "Diff all rows against first row")
	showNowDiffFlag := showCmd.Bool("diff-now", false, "Diff all rows against current time")
	showCombineFlag := showCmd.Bool("combine", false, "combine")
	showCombineExactFlag := showCmd.Bool("combine-exact", false, "combine-exact")

	resetCmd := flag.NewFlagSet("reset", flag.ExitOnError)
	resetAllFlag := resetCmd.Bool("all", false, "all")

	renameCmd := flag.NewFlagSet("rename", flag.ExitOnError)

	editCmd := flag.NewFlagSet("edit", flag.ExitOnError)

	timezoneCmd := flag.NewFlagSet("timezone", flag.ExitOnError)
	timezoneResetFlag := timezoneCmd.Bool("reset", false, "reset")

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	storage.SetupStorage()

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		add(nameOrDefault(addCmd.Arg(0)))
	case "show":
		showCmd.Parse(os.Args[2:])

		config.prevDiff = *showPrevDiffFlag
		config.firstDiff = *showFirstDiffFlag
		config.nowDiff = *showNowDiffFlag

		if *showAllFlag {
			all()
		} else if *showCombineExactFlag {
			combine(showCmd.Args(), true)
		} else if *showCombineFlag {
			combine(showCmd.Args(), false)
		} else {
			show(nameOrDefault(showCmd.Arg(0)))
		}
	case "reset":
		resetCmd.Parse(os.Args[2:])
		if *resetAllFlag {
			resetAll()
		} else {
			reset(nameOrDefault(resetCmd.Arg(0)))
		}
	case "list":
		list()
	case "rename":
		renameCmd.Parse(os.Args[2:])
		rename(renameCmd.Arg(0), renameCmd.Arg(1))
	case "edit":
		editCmd.Parse(os.Args[2:])
		edit(nameOrDefault(editCmd.Arg(0)))
	case "timezone":
		timezoneCmd.Parse(os.Args[2:])
		if *timezoneResetFlag {
			timezone.DeleteTimezoneFileIfExists()
		} else {
			timezone.SetTimezone(timezoneCmd.Arg(0))
		}
	case "version":
		fmt.Println(version)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

// Return default name if the name variable is empty
func nameOrDefault(name string) string {
	if name == "" {
		return "default"
	}
	return name
}

// Commands

func add(name string) {
	fmt.Println("Timestamp added")
	fmt.Printf(name)
	fmt.Printf(": ")
	t := time.Now().UTC()
	fmt.Println(timezone.InTimezone(t).Format(layoutDateTime))
	appendToFile(storage.GetFilePath(name), t.Format(layoutDateTime))
}

func show(name string) {
	filePath := storage.GetFilePath(name)
	if _, err := os.Stat(filePath); err == nil {
		printHeaders()
		timestamps := readFile(filePath)
		printTimestamps(timestamps)
	} else {
		fmt.Println("This stopwatch is not running")
	}
}

func combine(arguments []string, exactMatch bool) {
	var allTimestamps []nameAndDate
	for _, filename := range timezone.GetTimestampFiles() {
		name := getNameFromFilename(filename)
		if len(arguments) > 0 {
			if exactMatch && !containsExact(arguments, name) {
				continue
			} else if !exactMatch && !contains(arguments, name) {
				continue
			}
		}
		filePath := storage.GetFilePathForFilename(filename)
		timestamps := readFile(filePath)
		nameAndDates := convertToNameAndDateSlice(name, timestamps)
		allTimestamps = append(allTimestamps, nameAndDates...)
	}
	sort.Slice(allTimestamps, func(i, j int) bool {
		return allTimestamps[i].date.Before(allTimestamps[j].date)
	})
	printHeadersNamed(allTimestamps)
	printNameAndDates(allTimestamps)
}

func all() {
	for _, filename := range timezone.GetTimestampFiles() {
		name := getNameFromFilename(filename)
		fmt.Println(name)
		fmt.Println("-------------------")
		show(name)
		fmt.Println()
	}
}

func reset(name string) {
	filename := storage.GetFilePath(name)
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

func rename(oldName string, newName string) {
	oldPath := storage.GetFilePath(oldName)
	if !fileExists(oldPath) {
		fmt.Println("This stopwatch does not exist")
		return
	}

	newPath := storage.GetFilePath(newName)
	if fileExists(newPath) {
		fmt.Println("This stopwatch already exists")
		return
	}

	os.Rename(oldPath, newPath)
	fmt.Println("Done")
}

func edit(name string) {
	path := storage.GetFilePath(name)
	cmd := exec.Command(os.Getenv("EDITOR"), path)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func resetAll() {
	fmt.Printf("Reset all? (y/n) ")
	ok := askForConfirmation()

	if ok {
		for _, filename := range timezone.GetTimestampFiles() {
			if strings.Contains(filename, ".timestamps-") {
				path := storage.GetFilePathForFilename(filename)
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

func list() {
	for _, filename := range timezone.GetTimestampFiles() {
		if strings.Contains(filename, ".timestamps-") {
			fmt.Println(getNameFromFilename(filename))
		}
	}
}

// Helper functions

func containsExact(values []string, str string) bool {
	for _, value := range values {
		if value == str {
			return true
		}
	}
	return false
}

func contains(values []string, str string) bool {
	for _, value := range values {
		if strings.Contains(str, value) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func remove(name string) {
	path := storage.GetFilePath(name)
	if _, err := os.Stat(path); err == nil {
		e := os.Remove(path)
		if e != nil {
			log.Fatal(e)
		}
	}
}

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

// TODO: Merge this and printHeadersNamed
func printHeaders() {
	fmt.Printf("%-*s", timestampColumnWidth, timestampHeader)
	if config.prevDiff {
		fmt.Printf("%*s", prevDiffColumnWidth, prevDiffHeader)
	}
	if config.firstDiff {
		fmt.Printf("%*s", firstDiffColumnWidth, firstDiffHeader)
	}
	if config.nowDiff {
		fmt.Printf("%*s", nowDiffColumnWidth, nowDiffHeader)
	}
	fmt.Println()
}

func printHeadersNamed(timestamps []nameAndDate) {
	column1WidthNamed := getNameColumnLength(timestamps)
	fmt.Printf("%-*s", column1WidthNamed, nameHeader)
	fmt.Printf("%-*s", timestampColumnWidth, timestampHeader)
	if config.prevDiff {
		fmt.Printf("%*s", prevDiffColumnWidth, prevDiffHeader)
	}
	if config.firstDiff {
		fmt.Printf("%*s", firstDiffColumnWidth, firstDiffHeader)
	}
	if config.nowDiff {
		fmt.Printf("%*s", nowDiffColumnWidth, nowDiffHeader)
	}
	fmt.Println()
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

func nowTime() time.Time {
	now := time.Now().UTC()
	nowString := now.Format(layoutDateTime)
	now, err := time.Parse(layoutDateTime, nowString)
	if err != nil {
		fmt.Println(err)
	}
	return now
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

func getNameColumnLength(timestamps []nameAndDate) int {
	column1Length := getLengthOfLongestName(timestamps) + 2
	if minNameColumnWidth > column1Length {
		column1Length = minNameColumnWidth
	}
	return column1Length
}

func printTimestamps(timestamps []time.Time) {
	prevLineTimeExists := false
	prevLineTime := time.Now()
	firstLineTime := time.Now()
	now := timezone.InTimezone(nowTime())

	for _, lineTime := range timestamps {
		lineTime = timezone.InTimezone(lineTime)
		dateString := lineTime.Format(layoutDateTime)
		fmt.Printf("%*s", timestampColumnWidth, dateString)

		// TODO: Rename to isFirst?
		if prevLineTimeExists {
			diffSincePrev := lineTime.Sub(prevLineTime)
			if config.prevDiff {
				fmt.Printf("%*s", prevDiffColumnWidth, diffSincePrev.String())
			}
			if config.firstDiff {
				fmt.Printf("%*s", firstDiffColumnWidth, lineTime.Sub(firstLineTime).String())
			}
		} else {
			firstLineTime = lineTime
			if config.prevDiff {
				fmt.Printf("%*s", prevDiffColumnWidth, "")
			}
			if config.firstDiff {
				fmt.Printf("%*s", firstDiffColumnWidth, "")
			}
		}

		if config.nowDiff {
			fmt.Printf("%*s", nowDiffColumnWidth, now.Sub(lineTime).String())
		}

		fmt.Printf("\n")

		prevLineTimeExists = true
		prevLineTime = lineTime
	}

	if prevLineTimeExists {
		if config.prevDiff || config.firstDiff {
			fmt.Printf("%-*s", timestampColumnWidth, "Now")
			if config.prevDiff {
				fmt.Printf("%*s", prevDiffColumnWidth, now.Sub(prevLineTime).String())
			}
			if config.firstDiff {
				fmt.Printf("%*s", firstDiffColumnWidth, now.Sub(firstLineTime).String())
			}
			fmt.Println()
		}
	}
}

func printNameAndDates(timestamps []nameAndDate) {
	column1WidthNamed := getNameColumnLength(timestamps)
	prevLineTimeExists := false
	var prevLineTime nameAndDate
	var firstLineTime nameAndDate
	now := timezone.InTimezone(nowTime())

	for _, lineTime := range timestamps {
		lineTime.date = timezone.InTimezone(lineTime.date)
		dateString := lineTime.date.Format(layoutDateTime)
		fmt.Printf("%-*s", column1WidthNamed, lineTime.name)
		fmt.Printf("%*s", timestampColumnWidth, dateString)

		// TODO: Rename to isFirst?
		if prevLineTimeExists {
			diffSincePrev := lineTime.date.Sub(prevLineTime.date)
			if config.prevDiff {
				fmt.Printf("%*s", prevDiffColumnWidth, diffSincePrev.String())
			}
			if config.firstDiff {
				fmt.Printf("%*s", firstDiffColumnWidth, lineTime.date.Sub(firstLineTime.date).String())
			}
		} else {
			firstLineTime = lineTime
			if config.prevDiff {
				fmt.Printf("%*s", prevDiffColumnWidth, "")
			}
			if config.firstDiff {
				fmt.Printf("%*s", firstDiffColumnWidth, "")
			}
		}

		if config.nowDiff {
			fmt.Printf("%*s", nowDiffColumnWidth, now.Sub(lineTime.date).String())
		}

		fmt.Printf("\n")

		prevLineTimeExists = true
		prevLineTime = lineTime
	}

	if prevLineTimeExists {
		if config.firstDiff || config.prevDiff {
			fmt.Printf("%*s", column1WidthNamed, "")
			fmt.Printf("%-*s", timestampColumnWidth, "Now")
			if config.prevDiff {
				fmt.Printf("%*s", prevDiffColumnWidth, now.Sub(prevLineTime.date).String())
			}
			if config.firstDiff {
				fmt.Printf("%*s", firstDiffColumnWidth, now.Sub(firstLineTime.date).String())
			}
			fmt.Println()
		}
	}
}
