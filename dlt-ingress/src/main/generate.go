// genbaseprocessnew is disabled for issuance: the get-process endpoint is now backed by
// a hand-written AppService (getprocessservice) that resolves base processes first and
// falls back to parent processes. Re-enable once the generator supports that shape.
// //go:generate go run ../cmd/genbaseprocessnew/main.go
//
//go:generate go run ../cmd/gencommandname/main.go
//go:generate go run ../cmd/gencrosscommandname/main.go
//go:generate go run ../cmd/genrepositoriesold/main.go
//go:generate go run ../cmd/genrepositories/main.go
//go:generate go run ../cmd/genbaseprocess/main.go
//go:generate go run ../cmd/gencallbacks/main.go
//go:generate go run ../cmd/gencallbacksold/main.go
//go:generate go run ../cmd/genqueryname/main.go
//go:generate go run ../cmd/gencrossqueryname/main.go
//go:generate go run ../cmd/gendltadapters/main.go
//go:generate go run ../cmd/genlistener/main.go
//go:generate go run ../cmd/genhttpendpoint/main.go
package main
