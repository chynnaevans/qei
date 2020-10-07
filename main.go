package main

import (
	"fmt"
	"github.com/chynnaevans/qei/reader"
)

func main() {
	//url := "http://apps.courts.qld.gov.au/esearching/FileDetails.aspx?Location=BRISB&Court=SUPRE&Filenumber=6593/17"
	reader.InitApp()
	fmt.Println("------")

	//reader.StepReader(url)

}
