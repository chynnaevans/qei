package main

import "github.com/chynnaevans/qei/reader"

func main() {
	url := "http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17"
	reader.Reader(url)
}
