package reader

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
)

func Test_fetchPage_working(t *testing.T) {
	output, _ := fetchPage("http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17")
	fileRead, _ := ioutil.ReadFile("testpage.txt")
	expected := _stripWhitespace(string(fileRead))

	assert.Equal(t, expected[:100], _stripWhitespace(output)[:100])
}

func Test_fetchPage_error(t *testing.T) {
	output, err := fetchPage("")

	assert.Error(t, err)
	assert.Equal(t, "", output)
}

func _stripWhitespace(str string) string {
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, "\n", "")
	str = strings.ReplaceAll(str, "\r", "")
	str = strings.ReplaceAll(str, "\t", "")
	return str
}