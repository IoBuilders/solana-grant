//go:build gorm

// Command provider prints the Postgres DDL derived from this module's GORM
// models, so Atlas can diff it against the migration directory and emit a new
// one. The models come from eventstorepersistence.Models().
//
// It is excluded from normal builds by the gorm build tag.
package main

import (
	"fmt"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"

	"gitlab.com/iobuilders/projects/eng/naryo-go/persistence-gorm/infrastructure/eventstore/persistence"
)

func GetSchema() (string, error) {
	models := eventstorepersistence.Models()
	if len(models) == 0 {
		return "", fmt.Errorf("no models found")
	}

	schema, err := gormschema.New("postgres").Load(models...)
	if err != nil {
		return "", fmt.Errorf("failed to load schema: %w", err)
	}
	return schema, nil
}

func main() {
	schema, err := GetSchema()
	if err != nil {
		fmt.Fprintf(os.Stderr, "schema generation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(schema)
}
