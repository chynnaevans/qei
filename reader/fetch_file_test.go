package reader

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func Test_fetchPage_working(t *testing.T) {
	output, _ := fetchPage("http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17")
	fileRead, _ := ioutil.ReadFile("testpage.txt")
	expected := _stripWhitespace(string(fileRead))

	assert.Equal(t, expected[:100], _stripWhitespace(output)[:100])
}

func Test_fetchPage_invalid(t *testing.T) {
	output, err := fetchPage("")
	want := errors.New("cannot read invalid url")
	assert.Equal(t, want, err)
	assert.Equal(t, "", output)
}

func Test_fetchPage_errors (t *testing.T) {
	type resp struct {
		response string
		err error
	}
	tests := []struct {
		name string
		url string
		response resp
	}{
		{ "Empty URL", "" , resp{"",errors.New("cannot read invalid url")}},
		{"Wrong Page", "google.com", resp{"",errors.New("cannot read invalid url")}},
		{"Invalid Page", "aaaaaahttp://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17", resp{"",errors.New("cannot read invalid url")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := fetchPage(tt.url); got != tt.response.response || errors.Is(tt.response.err, err) {
				t.Errorf("fetchPage = %v, want %v", got, tt.response.response)
				t.Errorf("fetchPage = %v, want %v", err, tt.response.err)
			}
		})
	}
}

func Test_isValidPage(t *testing.T) {
	tests := []struct {
		name string
		body string
		want bool
	}{
		{ "Valid Page", "edocsno=1234", true},
		{"Invalid Page", "yeet", false},
		{"Empty Page", "", false},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pageHasDocs(tt.body); got != tt.want {
				t.Errorf("pageHasDocs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Reader(t *testing.T) {
	tests := []struct {
		name string
		url string
		want error
	}{
		{ "Invalid URL", "", errors.New("Get \"\": unsupported protocol scheme \"\"")},
		{"No Docs", "http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/18", errors.New("page has no docs")},
		{"Invalid Page", "http://google.com", errors.New("cannot read invalid url")},
		{"Valid Page", "http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17", (*os.PathError)(nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Reader(tt.url); errors.Is(tt.want, got) {
				t.Errorf("Reader() = %v, want %v", got, tt.want)
			}
		})
	}
}



func _stripWhitespace(str string) string {
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\r", "")
	str = strings.ReplaceAll(str, "\t", "")
	return str
}