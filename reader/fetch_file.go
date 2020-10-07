package reader

import (
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"
)

func StepReader(url string) (docs []Document) {
	urlPattern := `^http:\/\/apps\.courts\.qld\.gov\.au\/esearching\/FileDetails\.aspx\?Location=[A-Z]{5}&Court=[A-Z]{5}&Filenumber=[0-9]+\/[0-9]+$`
	if matches, _ := regexp.MatchString(urlPattern, url); !matches {
		log.Println("cannot read invalid url")
		return
	}
	resp, err := http.Get(url)
	//TODO: can I test this?
	if err != nil {
		fmt.Printf("error fetching webpage: %v", err)
		return
	}

	tokenizer := html.NewTokenizer(resp.Body)

	meta, files := extractData(tokenizer)
	resp.Body.Close()

	//TODO: delete print statements
	fmt.Println(meta)
	fmt.Println(len(files))

	return files
}

func extractData(tokenizer *html.Tokenizer) (caseMeta CaseMeta, files []Document) {
	currTag := ""
	timeLayout := "02/01/2006"
	for {
		tt := tokenizer.Next()

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
	}
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
						doc.DateFiled = filedTime
					}
					fieldCount--
				case tt == html.TextToken && fieldCount == 4:
					doc.DocType = t.Data
					fieldCount--
				case tt == html.TextToken && fieldCount == 3:
					doc.DocDesc = t.Data
					fieldCount--
				case tt == html.TextToken && fieldCount == 2:
					doc.Filer = t.Data
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

					docUrl := getTokenField(t, "href")
					doc.DocUrl = generateDocUrl(docUrl)
					pattern, err := regexp.Compile(`edocsno=([0-9]+)`)
					if err != nil || doc.DocUrl == "" {
						continue
					}
					if docNum := pattern.FindStringSubmatch(doc.DocUrl); len(docNum) >= 1 {
						doc.EDocNum = docNum[1]
					}

				}

				if fieldCount == 0 {
					if doc.EDocNum != "" {
						doc.CaseNum = caseMeta.caseNum
						docs = append(docs, doc)
					}
					break
				}
			}
		}

	}
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
func pageHasDocs(body []byte) bool {
	hasDoc, err := regexp.Match("edocsno", body)
	if err != nil {
		log.Println("error checking if page has docs")
	}
	return hasDoc
}

// Check if page exists
func fileExists(resp http.Response) bool {
	body, err := ioutil.ReadAll(resp.Body)
	invalidFile, err := regexp.Match("No such file found", body)
	if err != nil {
		fmt.Println("error checking page validity has docs")
	}

	return !invalidFile
}

// Generate doc URL
func generateDocUrl(suffix string) string {
	return "http://apps.courts.qld.gov.au/esearching/" + suffix
}