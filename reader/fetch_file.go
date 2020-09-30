package reader

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func Reader(url string) (err error) {
	//url = "http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17"

	resp, err := fetchPage(url)
	if err != nil {
		return
	} else if !isValidPage(resp) {
		err = errors.New("page has no docs")
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
	urlPattern := `^http:\/\/apps\.courts\.qld\.gov\.au\/esearching\/FileDetails\.aspx\?Location=[A-Z]{5}&Court=[A-Z]{5}&Filenumber=[0-9]+\/[0-9]+$`
	if matches, _ := regexp.MatchString(urlPattern, url); !matches {
		err = errors.New("cannot read invalid url")
		return
	}

	resp, err := http.Get(url)

	//TODO: can I test this?
	if err != nil {
		fmt.Printf("error fetching webpage: %v", err)
		return
	}

	readBuffer, err := ioutil.ReadAll(resp.Body)
	response = string(readBuffer)

	//TODO: add test for error here
	err = resp.Body.Close()

	return
}

// Check if page has court docs
func isValidPage(body string) bool {
	return strings.Contains(body, "edocsno")
}
