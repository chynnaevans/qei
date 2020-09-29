package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type CaseMeta struct {
	caseNum string
	title string
	court string
	originCity string
	currentCity string
	proceedingType string
	dateFiled time.Time
	nextListing time.Time
}

type Party struct {
	caseNum string
	lastName string
	firstName string
	companyName string
	acn string
	partyRole string
	representative string
}

type Document struct {
	caseNum string
	dateFiled string
	docType string
	docDesc string
	filer string
	pages int
	eDocNum string
	docUrl string
}

func main() {
	url := "http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17"

	resp := fetchPage(url)
	if resp == "" {
		fmt.Println("This page has no docs")
		return
	}

	fmt.Println("yeet")

	//TODO: extract meta data about the case
	//TODO: write case parties to table
	//TODO: extract each file and write to table with metadata
	return
}

// Fetch HTML from webpage. Return empty string if not contains docs
func fetchPage(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("error fetching webpage: ", err)
	}

	buf := new(strings.Builder)
	readSize, err := io.Copy(buf, resp.Body)
	if err != nil {
		log.Fatal("error copying reader, size: ", readSize)
	}

	resp.Body.Close()

	if !strings.Contains(buf.String(), "edocsno") {
		return ""
	}

	return buf.String()
}