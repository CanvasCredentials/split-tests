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
	filePath := prefix + strings.Replace(junitXML.Name, ".", "/", -1) + postfix
	fileTimes[filePath] = junitXML.Time
}

func getFileTimesFromJUnitXML(fileTimes map[string]float64, prefix string, postfix string) {
	if junitXMLPath != "" {
		filenames, err := doublestar.Glob(junitXMLPath)
		if err != nil {
			fatalMsg("failed to match jUnit filename pattern: %v", err)
		}
		numberOfReports := 0
		for _, junitFilename := range filenames {
			file, err := os.Open(junitFilename)
			if err != nil {
				fatalMsg("failed to open junit xml: %v\n", err)
			}
			addFileTimesFromIOReader(fileTimes, file, junitFilename, prefix, postfix)
			file.Close()
			numberOfReports++
		}
		printMsg("found %d JUnit XML report files\n", numberOfReports)
	}
}
