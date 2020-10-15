package reader

import (
	"context"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func ScanYear(startDoc int, year string){
	var docBatch []Document
	docs := make(chan Document)
	ctx := context.Background()
	client := InitApp(ctx)
	//TODO: change back doc number
	go EvaluatePages(startDoc, year, docs)

	for doc := range docs {
		if len(docBatch) < 500 {
			docBatch = append(docBatch, doc)
		} else {
			WriteBulkDocs(ctx, client, docBatch)
			docBatch = []Document{}
		}
	}

	WriteBulkDocs(ctx, client, docBatch)
}

func StepReader(url string) (docs []Document) {
	urlPattern := `^http:\/\/apps\.courts\.qld\.gov\.au\/esearching\/FileDetails\.aspx\?Location=[A-Z]{5}&Court=[A-Z]{5}&Filenumber=[0-9]+\/[0-9]+$`
	if matches, _ := regexp.MatchString(urlPattern, url); !matches {
		log.Println("cannot read invalid url")
		return
	}
	resp, err := http.Get(url)
	//TODO: can I test this?
	if err != nil {
		log.Printf("error fetching webpage: %v", err)
		return
	}
	//TODO: can I do a regex on the bytes here instead of casting as str? 
	tokenizer := html.NewTokenizer(resp.Body)

	_, files := extractData(tokenizer)
	resp.Body.Close()

	//TODO: delete print statements
	//fmt.Println(meta)
	//fmt.Println(len(files))

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

// Generate doc URL
func generateDocUrl(suffix string) string {
	return "http://apps.courts.qld.gov.au/esearching/" + suffix
}

func IsValidPage(url string) (isValid bool, isFinished bool) {	urlPattern := `^http:\/\/apps\.courts\.qld\.gov\.au\/esearching\/FileDetails\.aspx\?Location=[A-Z]{5}&Court=[A-Z]{5}&Filenumber=[0-9]+\/[0-9]+$`
	isValid = false
	isFinished = false
	if matches, _ := regexp.MatchString(urlPattern, url); !matches {
		log.Println("cannot read invalid url")
		return
	}
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("error fetching url %v", err,)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)

	//TODO: can I test this?
	if err != nil {
		log.Printf("error fetching webpage in verification: %v", err)
		return
	}

	if match, _ := regexp.Match("No such file found", body); match {
		isFinished = true
	} else if match, _ := regexp.Match("edocsno", body); match {
		isValid = true
	}

	resp.Body.Close()
	return
}

func EvaluatePages(docNumber int, year string, docs chan Document) {
	validCounter := 0
	base := `http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=`
	invalidCount := 0
	for ;;docNumber++ {
		if isValid, isFinished := IsValidPage(base + strconv.Itoa(docNumber) + `/` + year); isValid {
			validCounter++
			invalidCount = 0

			fileDocs := StepReader(base + strconv.Itoa(docNumber) + `/` + year)
			for _, doc := range fileDocs {
				docs <- doc
			}

		} else if isFinished && invalidCount < 100 {
			invalidCount++
		} else if isFinished && invalidCount >= 100 {
			log.Printf("Found end of doc, %d scanned & %d found", docNumber, validCounter)
			break
		} else if docNumber % 200 == 0 {
			log.Printf("Docs scanned: %d", docNumber)
		}
	}

	close(docs)
	return
}