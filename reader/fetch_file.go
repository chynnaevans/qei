package reader

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
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

func StepReader(url string) {
	urlPattern := `^http:\/\/apps\.courts\.qld\.gov\.au\/esearching\/FileDetails\.aspx\?Location=[A-Z]{5}&Court=[A-Z]{5}&Filenumber=[0-9]+\/[0-9]+$`
	if matches, _ := regexp.MatchString(urlPattern, url); !matches {
		//err := errors.New("cannot read invalid url")
		return
	}

	resp, err := http.Get(url)

	//TODO: can I test this?
	if err != nil {
		fmt.Printf("error fetching webpage: %v", err)
		return
	}

	tokenizer := html.NewTokenizer(resp.Body)
	//TODO: add test for error here

	meta, files := extractData(tokenizer)
	resp.Body.Close()
	fmt.Println(meta)
	fmt.Println(files)
}

func extractData(tokenizer *html.Tokenizer) (caseMeta CaseMeta, files []Document) {
	currTag := ""
	timeLayout := "02/01/2006"
	for {
		tt := tokenizer.Next()

		//TODO: implement docs
		switch {
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			t := tokenizer.Token()

			if t.Data == "span" {
				currTag = getTokenField(t, "id")
			} else if t.Data == "table" && getTokenField(t, "id") == metaFieldName("DocumentGrid") {

				files = readDocs(tokenizer, caseMeta)
				return
			}
		case tt == html.EndTagToken:
			currTag = ""
		case tt == html.TextToken && currTag == metaFieldName("filenumber"):
			caseMeta.caseNum = tokenizer.Token().Data
		case tt == html.TextToken && currTag == metaFieldName("filename"):
			caseMeta.title = tokenizer.Token().Data
		case tt == html.TextToken && currTag == metaFieldName("court"):
			caseMeta.court = tokenizer.Token().Data
		case tt == html.TextToken && currTag == metaFieldName("originatinglocation"):
			caseMeta.originCity = tokenizer.Token().Data
		case tt == html.TextToken && currTag == metaFieldName("currentlocation"):
			caseMeta.currentCity = tokenizer.Token().Data
		case tt == html.TextToken && currTag == metaFieldName("proceedingtype"):
			caseMeta.proceedingType = tokenizer.Token().Data
		case tt == html.TextToken && currTag == metaFieldName("datefiled"):
			if filedTime, err := time.Parse(timeLayout, tokenizer.Token().Data); err != nil {
				continue
			} else {
				caseMeta.dateFiled = filedTime
			}
		case tt == html.TextToken && currTag == metaFieldName("bookingdate"):
			if nextTime, err := time.Parse(timeLayout, tokenizer.Token().Data); err != nil {
				continue
			} else {
				caseMeta.nextListing = nextTime
			}

		}

		//TODO: implement docs
		//if isMetaComplete(caseMeta) {
		//	return
		//}
	}
	return
}

func readDocs(tokenizer *html.Tokenizer, caseMeta CaseMeta) (docs []Document) {
	for {
		timeLayout := "02/01/2006"

		tt := tokenizer.Next()
		t := tokenizer.Token()

		if tt == html.EndTagToken && t.Data == "table" {
			return
		} else if tt == html.StartTagToken && t.Data == "tr" && getTokenField(t, "class") != "GridHeader" {
			fieldCount := 6
			doc := Document{}
			for {
				if tt == html.StartTagToken && t.Data != "tr" {
					break
				}
				tt = tokenizer.Next()
				t = tokenizer.Token()
			}
			for {
				tt = tokenizer.Next()
				t = tokenizer.Token()
				if tt == html.EndTagToken && t.Data == "tr" {
					break
				}
				//TODO: should change fieldCount to enums
				switch {
				case tt == html.TextToken && fieldCount == 6:
					fieldCount--
				case tt == html.TextToken && fieldCount == 5:
					if filedTime, err := time.Parse(timeLayout, t.Data); err != nil {
						continue
					} else {
						doc.dateFiled = filedTime
					}
					fieldCount--
				case tt == html.TextToken && fieldCount == 4:
					doc.docType = t.Data
					fieldCount--
				case tt == html.TextToken && fieldCount == 3:
					doc.docDesc = t.Data
					fieldCount--
				case tt == html.TextToken && fieldCount == 2:
					doc.filer = t.Data
					fieldCount--
				case tt == html.StartTagToken && fieldCount == 1:
					fieldCount--
					for {
						if (tt == html.StartTagToken && t.Data == "a") || (tt == html.EndTagToken && t.Data == "td") {
							break
						}
						tt = tokenizer.Next()
						t = tokenizer.Token()
					}
					if tt != html.StartTagToken || t.Data != "a"  {
						continue
					}

					doc.docUrl = getTokenField(t, "href")
					pattern, err := regexp.Compile(`edocsno\=([0-9]+)`)
					if err != nil || doc.docUrl == "" {
						continue
					}
					if docNum := pattern.FindStringSubmatch(doc.docUrl); len(docNum) >= 1 {
						doc.eDocNum = docNum[1]
					}

				}

				if fieldCount == 0 {
					doc.caseNum = caseMeta.caseNum
					docs = append(docs, doc)
					break
				}
			}
		}

	}
	return
}

func getTokenField(token html.Token, field string) (value string) {
	for _, a := range token.Attr {
		if a.Key == field {
			return a.Val
		}
	}

	return
}

func metaFieldName(name string) (fullName string) {
	return "ctl00_ContentPlaceHolder1_" + name
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