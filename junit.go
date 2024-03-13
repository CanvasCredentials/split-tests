package main

import (
	"encoding/xml"
	"github.com/bmatcuk/doublestar"
	"io"
	"os"
	"strings"
)

type junitXML struct {
	Name string  `xml:"name,attr"`
	Time float64 `xml:"time,attr"`
}

func loadJUnitXML(reader io.Reader) *junitXML {
	var junitXML junitXML

	decoder := xml.NewDecoder(reader)
	err := decoder.Decode(&junitXML)
	if err != nil {
		fatalMsg("failed to parse junit xml: %v\n", err)
	}

	return &junitXML
}

func addFileTimesFromIOReader(fileTimes map[string]float64, reader io.Reader, filename string, prefix string, postfix string) {
	junitXML := loadJUnitXML(reader)
	printMsg("using test times from JUnit report %s\n", filename)
	filePath := prefix + strings.Replace(junitXML.Name, ".", "/", -1) + postfix
	printMsg("converted test name to %s\n", filePath)
	fileTimes[filePath] = junitXML.Time
}

func getFileTimesFromJUnitXML(fileTimes map[string]float64, prefix string, postfix string) {
	if junitXMLPath != "" {
		filenames, err := doublestar.Glob(junitXMLPath)
		if err != nil {
			fatalMsg("failed to match jUnit filename pattern: %v", err)
		}
		for _, junitFilename := range filenames {
			file, err := os.Open(junitFilename)
			if err != nil {
				fatalMsg("failed to open junit xml: %v\n", err)
			}
			addFileTimesFromIOReader(fileTimes, file, junitFilename, prefix, postfix)
			file.Close()
		}
	}
}
