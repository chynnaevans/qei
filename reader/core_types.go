package reader

import "time"

type CaseMeta struct {
	caseNum string
	title string
	court string
	originCity string
	currentCity string
	proceedingType string
	dateFiled time.Time
	nextListing time.Time
}

type Party struct {
	caseNum string
	lastName string
	firstName string
	companyName string
	acn string
	partyRole string
	representative string
}

type Document struct {
	CaseNum string
	DateFiled time.Time
	DocType string
	DocDesc string
	Filer string
	Pages int
	EDocNum string
	DocUrl string
}
