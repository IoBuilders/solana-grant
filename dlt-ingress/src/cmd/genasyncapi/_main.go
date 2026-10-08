// Command genasyncapi emits the AsyncAPI event contract of every bounded context, combining what
// the source says with what catalog.go declares, so a committed spec cannot drift from the code.
//
//	go run ./src/cmd/genasyncapi                 write asyncapi/<bc>.yaml for every catalogued BC
//	go run ./src/cmd/genasyncapi -check          fail if a committed spec is stale (the CI gate)
//	go run ./src/cmd/genasyncapi -bc dltingress  restrict to one bounded context
//
// The structural half is read from the source: which structs are events, where each is published,
// which listeners consume it, and therefore the channels, the operations and x-orphan. The other
// half — what an event means and which stream it belongs to — is declared in catalog.go, away
// from the domain. Both modes fail when the two disagree, so an event added to the code without
// an entry in the catalog breaks the build instead of quietly missing from the spec.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitresthen/asyncapi3"
)

const (
	specDir  = "asyncapi"
	specPerm = 0o644
)

func main() {
	check := flag.Bool("check", false, "compare against the committed specs instead of writing them")
	only := flag.String("bc", "", "restrict to one bounded context")
	flag.Parse()

	err := run(*check, *only)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(check bool, only string) error {
	// One scan of src/main serves every bounded context: consumers, and sometimes producers,
	// cross context boundaries anyway.
	scanner, err := newScanner(srcMainDir)
	if err != nil {
		return err
	}

	matched := false

	for _, s := range catalog {
		if only != "" && s.bc != only {
			continue
		}

		matched = true

		out, err := scanner.generate(s)
		if err != nil {
			return err
		}

		path := filepath.Join(specDir, s.bc+".yaml")

		if check {
			err = checkStale(out, path)
			if err != nil {
				return err
			}

			fmt.Println(path + " is up to date")

			continue
		}

		err = os.WriteFile(path, out, specPerm)
		if err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		fmt.Println("wrote " + path)
	}

	if only != "" && !matched {
		return fmt.Errorf("bounded context %q has no entry in catalog.go", only)
	}

	return nil
}

// generate is the whole pipeline for one bounded context: discover, join with the catalog, render.
func (s *scanner) generate(spec spec) ([]byte, error) {
	found, err := s.discover(spec.root())
	if err != nil {
		return nil, fmt.Errorf("discovering the events of %s: %w", spec.bc, err)
	}

	messages, err := resolveMessages(spec, found)
	if err != nil {
		return nil, err
	}

	doc, err := buildDocument(spec, messages)
	if err != nil {
		return nil, fmt.Errorf("building %s: %w", spec.bc, err)
	}

	out, err := render(doc)
	if err != nil {
		return nil, fmt.Errorf("rendering %s: %w", spec.bc, err)
	}

	return out, nil
}

func buildDocument(spec spec, messages []message) (*document, error) {
	components, schemas, exts, err := buildMessages(messages)
	if err != nil {
		return nil, err
	}

	doc := asyncapi3.New(spec.title, spec.version)
	doc.Info.Description = spec.description
	doc.DefaultContentType = "application/json"
	doc.Servers = buildServers()
	doc.Channels = buildChannels(messages)
	doc.Operations = buildOperations(messages)
	doc.Components = &asyncapi3.Components{Schemas: schemas, Messages: components}

	return &document{header: generatedHeader, doc: doc, exts: exts}, nil
}
