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

	caseMeta := extractMeta(resp)
	fmt.Println(caseMeta)
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

// Extract metadata from FileDetails page
func extractMeta(body string) (meta CaseMeta) {
	caseNum, err := extractField(body, "filenumber")
	title, err := extractField(body, "filename")
	court, err := extractField(body, "court")
	originCity, err := extractField(body, "originatinglocation")
	currentCity, err := extractField(body, "currentlocation")
	proceedingType, err := extractField(body, "proceedingtype")
	dateFiled, err := extractField(body, "datefiled")
	nextListing, err := extractField(body, "bookingdate")

	meta = CaseMeta{
		caseNum: caseNum,
		title: title,
		court: court,
		originCity: originCity,
		currentCity: currentCity,
		proceedingType: proceedingType,
		dateFiled: dateFiled,
		nextListing: nextListing,
	}

	if err != nil {
		println("one or more case metadata fields missing")
	}
	return
}

// Extract field from FileDetails page
func extractField(body string, field string) (value string, err error){
	pattern := `<span id="ctl00_ContentPlaceHolder1_` + field + `">(.+)</span>`
	r, err := regexp.Compile(pattern)
	if err != nil {
		return
	}

	if len(r.FindStringSubmatch(body)) < 2 {
		err = errors.New("field not found")
		return
	}

	value = r.FindStringSubmatch(body)[1]
	return
}