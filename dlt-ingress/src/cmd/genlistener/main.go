// Command genlistener generates the RegistrableListener interface methods
// (EventType/Name/Listen) for every event listener in the bounded contexts listed in
// cmdcore.HANDLER_GEN_BCS, so they don't have to be hand-written. For each package that declares
// `type Listener` with a `handle(ctx, *Event) error` method, it writes a
// <prefix>_listener_gen.go next to it.
//
// The event type and key are derived from the handler's parameter, so they cannot drift.
// Run it like the other generators, from src/main/:
//
//	go run ./../cmd/genlistener
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"dlt-ingress/src/cmd/core"
)

var tmpl *template.Template
var completedOk = false

func init() {
	var err error
	tmpl, err = template.ParseFiles("./../cmd/genlistener/method.tmpl")
	if err != nil {
		fmt.Println("❌ Failed to load template:", err)
		os.Exit(1)
	}
}

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("👂 Starting listener generator")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("───────────────────────────────")
	if completedOk {
		fmt.Println("✅ Listener generation finished successfully")
	} else {
		fmt.Println("❌ Listener generation failed")
	}
	fmt.Println("───────────────────────────────")
}

func main() {
	starting()
	defer ending()

	dirs := map[string]bool{}
	if args := os.Args[1:]; len(args) > 0 {
		// Optional: restrict to the given directories (handy for testing a single listener).
		for _, d := range args {
			dirs[d] = true
		}
	} else {
		for _, bc := range cmdcore.HANDLER_GEN_BCS {
			root := "./../main/" + bc + "/"
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				dirs[filepath.Dir(path)] = true
				return nil
			})
			if err != nil {
				fmt.Println("❌ Error:", err)
				os.Exit(1)
			}
		}
	}

	for dir := range dirs {
		if perr := processDir(dir); perr != nil {
			fmt.Println("❌ Error:", perr)
			os.Exit(1)
		}
	}
	completedOk = true
}

func recvIsListener(fd *ast.FuncDecl) bool {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return false
	}
	t := fd.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	id, ok := t.(*ast.Ident)
	return ok && id.Name == "Listener"
}

func processDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	fset := token.NewFileSet()

	var pkg, listenerFile string
	hasListener := false
	var handleFile *ast.File
	var handleDecl *ast.FuncDecl

	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, "_gen.go") {
			continue
		}
		path := filepath.Join(dir, name)
		f, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return perr
		}
		pkg = f.Name.Name

		for _, d := range f.Decls {
			switch decl := d.(type) {
			case *ast.GenDecl:
				if decl.Tok != token.TYPE {
					continue
				}
				for _, s := range decl.Specs {
					ts, ok := s.(*ast.TypeSpec)
					if ok && ts.Name.Name == "Listener" {
						if _, ok := ts.Type.(*ast.StructType); ok {
							hasListener = true
						}
					}
				}
			case *ast.FuncDecl:
				if recvIsListener(decl) && decl.Name.Name == "Listen" {
					handleDecl = decl
					handleFile = f
					listenerFile = path
				}
			}
		}
	}

	// Only generate for real listeners: a Listener struct with a handle method.
	if !hasListener || handleDecl == nil {
		return nil
	}

	eventType, eventImport, err := extractEvent(handleDecl, handleFile, dir)
	if err != nil {
		// Not every Listener/Listen pair fits the pointer-event convention (e.g. a generic
		// catch-all listener declared as Listen(ctx, event.Event)) — skip instead of aborting
		// the whole run, since that's a valid pattern this generator just can't derive from.
		fmt.Printf("⏭️  Skipping %s: %v\n", dir, err)
		return nil
	}

	prefix := genPrefix(listenerFile)
	genPath := filepath.Join(dir, prefix+"_listener_gen.go")
	fmt.Println("⚙️  Generating:", genPath)

	var buf bytes.Buffer
	if terr := tmpl.Execute(&buf, struct {
		Package     string
		EventType   string
		EventImport string
	}{Package: pkg, EventType: eventType, EventImport: eventImport}); terr != nil {
		return terr
	}
	formatted, ferr := format.Source(buf.Bytes())
	if ferr != nil {
		return fmt.Errorf("%s: gofmt: %w", genPath, ferr)
	}
	return os.WriteFile(genPath, formatted, 0o644)
}

// extractEvent reads the event type from handle's second parameter (ctx, *Event) and
// resolves the import line to reference it from the generated file.
func extractEvent(handle *ast.FuncDecl, file *ast.File, dir string) (eventType, eventImport string, err error) {
	params := handle.Type.Params.List
	if len(params) < 2 {
		return "", "", fmt.Errorf("handle must have (ctx, *Event) parameters")
	}
	// The event is the last parameter; its type is *Event.
	star, ok := params[len(params)-1].Type.(*ast.StarExpr)
	if !ok {
		return "", "", fmt.Errorf("handle event parameter must be a pointer")
	}
	switch t := star.X.(type) {
	case *ast.SelectorExpr: // pkg.Event
		qualifier, ok := t.X.(*ast.Ident)
		if !ok {
			return "", "", fmt.Errorf("unexpected event qualifier")
		}
		eventType = qualifier.Name + "." + t.Sel.Name
		eventImport, err = resolveImport(file, qualifier.Name)
		return eventType, eventImport, err
	case *ast.Ident: // Event declared in the listener's own package (no import)
		return t.Name, "", nil
	default:
		return "", "", fmt.Errorf("unsupported event type expression")
	}
}

// resolveImport returns the listener file's import for the given qualifier, preserving its
// original form (alias or not) so the generated file references the event identically.
func resolveImport(file *ast.File, qualifier string) (string, error) {
	for _, imp := range file.Imports {
		path, _ := strconv.Unquote(imp.Path.Value)
		var matched bool
		if imp.Name != nil {
			matched = imp.Name.Name == qualifier
		} else {
			matched = declaredPackageName(path) == qualifier
		}
		if !matched {
			continue
		}
		if imp.Name != nil {
			return imp.Name.Name + " " + strconv.Quote(path), nil
		}
		return strconv.Quote(path), nil
	}
	return "", fmt.Errorf("import for qualifier %q not found", qualifier)
}

// declaredPackageName returns the `package X` name for an import path.
// For local (dlt-ingress) imports it reads the package clause directly; for external
// packages it asks `go list` so names that differ from the path segment (e.g.
// event/api → apievent) are resolved correctly.
func declaredPackageName(importPath string) string {
	const localPrefix = "dlt-ingress/src/main/"
	if !strings.HasPrefix(importPath, localPrefix) {
		out, err := exec.Command("go", "list", "-f", "{{.Name}}", importPath).Output()
		if err == nil {
			if name := strings.TrimSpace(string(out)); name != "" {
				return name
			}
		}
		return filepath.Base(importPath)
	}
	pkgDir := "./" + strings.TrimPrefix(importPath, localPrefix)
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return filepath.Base(importPath)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, perr := parser.ParseFile(fset, filepath.Join(pkgDir, e.Name()), nil, parser.PackageClauseOnly)
		if perr == nil {
			return f.Name.Name
		}
	}
	return filepath.Base(importPath)
}

// genPrefix turns ".../document_removed_listener.go" into "document_removed".
func genPrefix(listenerFile string) string {
	base := filepath.Base(listenerFile)
	if strings.HasSuffix(base, "_listener.go") {
		return strings.TrimSuffix(base, "_listener.go")
	}
	return strings.TrimSuffix(base, ".go")
}
