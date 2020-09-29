package reader

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func Reader() {
	url := "http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17"

	resp, err := fetchPage(url)
	if err != nil {
		return
	} else if !strings.Contains(resp, "edocsno") {
		fmt.Println("This page has no docs")
		return
	}

	fmt.Println("yeet")

	//caseMeta := CaseMeta{}

	//TODO: extract meta data about the case
	//TODO: write case parties to table
	//TODO: extract each file and write to table with metadata
	return
}

// Extract HTML from URL
func fetchPage(url string) (response string, err error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("error fetching webpage: %v", err)
		return
	}

	readBuffer, err := ioutil.ReadAll(resp.Body)
	response = string(readBuffer)

	resp.Body.Close()

	return
}
