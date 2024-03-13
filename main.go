package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar"
)

var useJUnitXML bool
var junitXMLPath string
var testFilePattern = ""
var testFilePrefix = ""
var testFilePostfix = ""
var excludeFilePattern = ""
var splitIndex int
var splitTotal int

func printMsg(msg string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Fprint(os.Stderr, msg)
	} else {
		fmt.Fprintf(os.Stderr, msg, args...)
	}
}

func fatalMsg(msg string, args ...interface{}) {
	printMsg(msg, args...)
	os.Exit(1)
}

func removeDeletedFiles(fileTimes map[string]float64, currentFileSet map[string]bool) {
	for file := range fileTimes {
		if !currentFileSet[file] {
			delete(fileTimes, file)
		}
	}
}

func addNewFiles(fileTimes map[string]float64, currentFileSet map[string]bool) {
	averageFileTime := 0.0
	if len(fileTimes) > 0 {
		for _, time := range fileTimes {
			averageFileTime += time
		}
		averageFileTime /= float64(len(fileTimes))
	} else {
		averageFileTime = 1.0
	}

	for file := range currentFileSet {
		if _, isSet := fileTimes[file]; isSet {
			continue
		}
		if useJUnitXML {
			printMsg("missing file time for %s\n", file)
		}
		fileTimes[file] = averageFileTime
	}
}

func parseFlags() {
	flag.StringVar(&testFilePattern, "glob", "spec/**/*_spec.rb", "Glob pattern to find test files. Make sure to single-quote to avoid shell expansion.")
	flag.StringVar(&testFilePrefix, "prefix", "", "Enables to specify a prefix to align naming of the test files with the name in the JUnit report.")
	flag.StringVar(&testFilePostfix, "postfix", "", "Enables to specify a prefix to align naming of the test files with the name in the JUnit report.")
	flag.StringVar(&excludeFilePattern, "exclude-glob", "", "Glob pattern to exclude test files. Make sure to single-quote.")

	flag.IntVar(&splitIndex, "split-index", -1, "This test container's index")
	flag.IntVar(&splitTotal, "split-total", -1, "Total number of containers")

	flag.StringVar(&junitXMLPath, "junit-path", "", "Path to a JUnit XML report (leave empty to read from stdin; use glob pattern to load multiple files)")

	var showHelp bool
	flag.BoolVar(&showHelp, "help", false, "Show this help text")

	flag.Parse()

	if showHelp {
		printMsg("Splits test files into containers of even duration\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if splitTotal == 0 || splitIndex < 0 || splitIndex > splitTotal {
		fatalMsg("-split-index and -split-total are missing or invalid\n")
	}
	if junitXMLPath == "" {
		fatalMsg("-junit-path is missing\n")
	}
}

func main() {
	parseFlags()
	printMsg("prefix: %s, postfix: %s", testFilePrefix, testFilePostfix)

	// We are not using filepath.Glob,
	// because it doesn't support '**' (to match all files in all nested directories)
	currentFiles, err := doublestar.Glob(testFilePattern)
	if err != nil {
		printMsg("failed to enumerate current file set: %v", err)
		os.Exit(1)
	}
	currentFileSet := make(map[string]bool)
	for _, file := range currentFiles {
		currentFileSet[file] = true
	}

	if excludeFilePattern != "" {
		excludedFiles, err := doublestar.Glob(excludeFilePattern)
		if err != nil {
			printMsg("failed to enumerate excluded file set: %v", err)
			os.Exit(1)
		}
		for _, file := range excludedFiles {
			delete(currentFileSet, file)
		}
	}

	fileTimes := make(map[string]float64)
	getFileTimesFromJUnitXML(fileTimes, testFilePrefix, testFilePostfix)

	removeDeletedFiles(fileTimes, currentFileSet)
	addNewFiles(fileTimes, currentFileSet)

	buckets, bucketTimes := splitFiles(fileTimes, splitTotal)
	printMsg("expected test time: %0.1fs\n", bucketTimes[splitIndex])

	fmt.Println(strings.Join(buckets[splitIndex], " "))
}
