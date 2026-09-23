package getcustodykey

type Query struct {
	DltAccountId string
}

type Response struct {
	KeyType         string
	Status          string
	DltAccountId    string
	Dlt             string
	ExternalId      string
	CustodyProvider string
}
