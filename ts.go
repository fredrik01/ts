package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/fredrik01/ts/src/csv"
	"github.com/fredrik01/ts/src/storage"
	"github.com/fredrik01/ts/src/timezone"
)

const (
	version              = "0.9.0"
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

type record struct {
	name      string
	timestamp time.Time
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

    show	Show all or some stopwatches in a sorted list. Additional arguments can be used to only keep some stopwatches in the list (ts show mystopwatch)
    		-split		Print all stopwatches separately
    		-diff-prev	Diff all rows against previous row in the list
    		-diff-first	Diff all rows against first row
    		-diff-now	Diff all rows against current time
    		-exact	Use exact matching for additional arguments

    reset	Reset default stopwatch or a named one (ts reset mystopwatch)
    		-all		Reset all stopwatches

    rename	Rename a stopwatch (ts rename oldname newname)

    edit	Edit save file in your $EDITOR. Timestamps are stored in UTC.

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
	showSplitCmd := showCmd.Bool("split", false, "Split by name")
	showPrevDiffFlag := showCmd.Bool("diff-prev", false, "Diff all rows against previous row in the list")
	showFirstDiffFlag := showCmd.Bool("diff-first", false, "Diff all rows against first row")
	showNowDiffFlag := showCmd.Bool("diff-now", false, "Diff all rows against current time")
	showExactFlag := showCmd.Bool("exact", false, "exact")

	resetCmd := flag.NewFlagSet("reset", flag.ExitOnError)
	resetAllFlag := resetCmd.Bool("all", false, "all")

	renameCmd := flag.NewFlagSet("rename", flag.ExitOnError)

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

		if *showSplitCmd {
			split()
		} else if *showExactFlag {
			records := getRecords()
			records = keepMatchingRecords(records, showCmd.Args(), true)
			show(records)
		} else {
			records := getRecords()
			records = keepMatchingRecords(records, showCmd.Args(), false)
			show(records)
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
		edit()
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

func add(name string) {
	fmt.Println("Timestamp added")
	fmt.Printf(name)
	fmt.Printf(": ")
	t := time.Now().UTC()
	fmt.Println(timezone.InTimezone(t).Format(layoutDateTime))
	records := []record{{timestamp: t, name: name}}
	saveRecords(records)
}

func saveRecords(records []record) {
	for _, record := range records {
		newLine := []string{record.timestamp.Format(layoutDateTime), record.name}
		csv.AppendToFile(storage.GetSaveFilePath(), newLine)
	}
}

func getRecords() []record {
	saveFile := storage.GetSaveFilePath()
	if storage.FileExist(saveFile) {
		lines := csv.Read(saveFile)
		return convertToRecordSlice(lines)
	}
	return []record{}
}

func getUniqueNames(records []record) []string {
	var uniqueNames []string
	unique := make(map[string]bool)
	for _, record := range records {
		if !unique[record.name] {
			unique[record.name] = true
			uniqueNames = append(uniqueNames, record.name)
		}
	}
	return uniqueNames
}

func nameExists(records []record, name string) bool {
	exists := false
	for _, record := range records {
		if record.name == name {
			exists = true
			break
		}
	}
	return exists
}

func keepMatchingRecords(records []record, arguments []string, exactMatch bool) []record {
	if len(arguments) > 0 {
		var filtered []record
		for _, record := range records {
			if exactMatch && containsExact(arguments, record.name) {
				filtered = append(filtered, record)
			} else if !exactMatch && contains(arguments, record.name) {
				filtered = append(filtered, record)
			}
		}
		records = filtered
	}
	return records
}

func removeMatchingRecords(records []record, arguments []string, exactMatch bool) []record {
	if len(arguments) > 0 {
		var filtered []record
		for _, record := range records {
			if exactMatch && !containsExact(arguments, record.name) {
				filtered = append(filtered, record)
			} else if !exactMatch && !contains(arguments, record.name) {
				filtered = append(filtered, record)
			}
		}
		records = filtered
	}
	return records
}

func show(records []record) {
	if len(records) == 0 {
		fmt.Println("No records found")
	} else {
		sort.Slice(records, func(i, j int) bool {
			return records[i].timestamp.Before(records[j].timestamp)
		})
		printHeaders(records)
		printRecords(records)
	}
}

func split() {
	allRecords := getRecords()
	names := getUniqueNames(allRecords)
	for _, name := range names {
		records := keepMatchingRecords(allRecords, []string{name}, true)
		show(records)
		fmt.Println()
	}
}

func reset(name string) {
	records := getRecords()
	if !nameExists(records, name) {
		fmt.Printf("Name \"")
		fmt.Printf(name)
		fmt.Printf("\" was not found\n")
		os.Exit(0)
	}

	fmt.Printf("Reset ")
	fmt.Printf(name)
	fmt.Printf("? (y/n) ")
	ok := askForConfirmation()

	if ok {
		// Get records to keep
		recordsToKeep := removeMatchingRecords(records, []string{name}, true)
		// Remove save file
		storage.Delete(storage.GetSaveFilePath())
		// Create new save file with records to keep
		saveRecords(recordsToKeep)
		fmt.Println("Done")
	} else {
		fmt.Println("Aborted")
	}
}

func rename(oldName string, newName string) {
	records := getRecords()

	// Check old name
	if !nameExists(records, oldName) {
		fmt.Printf(oldName)
		fmt.Printf(" does not exist\n")
		os.Exit(0)
	}

	// Check new name
	if nameExists(records, newName) {
		fmt.Printf(newName)
		fmt.Printf(" already exists\n")
		os.Exit(0)
	}

	for i, record := range records {
		if record.name == oldName {
			records[i].name = newName
		}
	}

	storage.Delete(storage.GetSaveFilePath())
	saveRecords(records)

	fmt.Println("Done")
}

func edit() {
	path := storage.GetSaveFilePath()
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
		storage.Delete(storage.GetSaveFilePath())
		fmt.Println("Done")
	} else {
		fmt.Println("Aborted")
	}
}

func list() {
	names := getUniqueNames(getRecords())
	for _, name := range names {
		fmt.Println(name)
	}
}

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

func convertToRecordSlice(lines [][]string) []record {
	var records []record
	for _, line := range lines {
		timestamp, err := time.Parse(layoutDateTime, line[0])
		if err != nil {
			fmt.Println(err)
		}
		records = append(records, record{timestamp: timestamp, name: line[1]})
	}
	return records
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

func printHeaders(timestamps []record) {
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

func nowTime() time.Time {
	now := time.Now().UTC()
	nowString := now.Format(layoutDateTime)
	now, err := time.Parse(layoutDateTime, nowString)
	if err != nil {
		fmt.Println(err)
	}
	return now
}

func getLengthOfLongestName(timestamps []record) int {
	longest := 0
	for _, record := range timestamps {
		if len(record.name) > longest {
			longest = len(record.name)
		}
	}
	return longest
}

func getNameColumnLength(timestamps []record) int {
	column1Length := getLengthOfLongestName(timestamps) + 2
	if minNameColumnWidth > column1Length {
		column1Length = minNameColumnWidth
	}
	return column1Length
}

func printRecords(timestamps []record) {
	column1WidthNamed := getNameColumnLength(timestamps)
	prevLineTimeExists := false
	var prevLineTime record
	var firstLineTime record
	now := timezone.InTimezone(nowTime())

	for _, lineTime := range timestamps {
		lineTime.timestamp = timezone.InTimezone(lineTime.timestamp)
		dateString := lineTime.timestamp.Format(layoutDateTime)
		fmt.Printf("%-*s", column1WidthNamed, lineTime.name)
		fmt.Printf("%*s", timestampColumnWidth, dateString)

		if prevLineTimeExists {
			diffSincePrev := lineTime.timestamp.Sub(prevLineTime.timestamp)
			if config.prevDiff {
				fmt.Printf("%*s", prevDiffColumnWidth, diffSincePrev.String())
			}
			if config.firstDiff {
				fmt.Printf("%*s", firstDiffColumnWidth, lineTime.timestamp.Sub(firstLineTime.timestamp).String())
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
			fmt.Printf("%*s", nowDiffColumnWidth, now.Sub(lineTime.timestamp).String())
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
				fmt.Printf("%*s", prevDiffColumnWidth, now.Sub(prevLineTime.timestamp).String())
			}
			if config.firstDiff {
				fmt.Printf("%*s", firstDiffColumnWidth, now.Sub(firstLineTime.timestamp).String())
			}
			fmt.Println()
		}
	}
}
