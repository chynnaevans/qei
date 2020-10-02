package reader

type CaseMeta struct {
	caseNum string
	title string
	court string
	originCity string
	currentCity string
	proceedingType string
	dateFiled string
	nextListing string
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
	caseNum string
	dateFiled string
	docType string
	docDesc string
	filer string
	pages int
	eDocNum string
	docUrl string
}
