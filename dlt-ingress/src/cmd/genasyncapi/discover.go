package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Discovery reads the structural facts off the source so that the catalog does not have to
// restate them: which structs are events, where each one is published, and which listeners
// consume it. It reads nothing about what an event *means* — that is declared in catalog.go, and
// the domain stays free of any AsyncAPI vocabulary.

const (
	modulePath   = "dlt-ingress"
	srcMainDir   = "src/main"
	coreEventPkg = "/core/event"

	baseEvent    = "BaseEvent"
	baseDltEvent = "BaseDltEvent"
)

// discovered is one event as the code presents it.
type discovered struct {
	goType     string   // "custodykey.KeyCreatedEvent" — what %T prints and what the outbox stores
	typeName   string   // "KeyCreatedEvent"
	aggregate  string   // "custodykey" — the directory that declares it, the address's middle segment
	pos        string   // file:line, so anything reported back names a place to go
	publishers []string // derived from the publish call sites
	consumers  []string // registered listener names, from every bounded context
	cross      bool     // published on the cross bus rather than the internal one
}

func (e discovered) published() bool { return len(e.publishers) > 0 }

// typeRef identifies a declaration by the directory that declares it, which is unique where a
// package name is not: two packages named custodykey live in dltingress alone.
type typeRef struct {
	dir  string
	name string
}

type structDecl struct {
	ref     typeRef
	pkgName string
	embeds  []embeddedType
	pos     token.Position
}

// embeddedType is one embedded field of a struct. Same-package embeds carry no import path.
type embeddedType struct {
	importPath string
	name       string
}

// fileScope resolves a package qualifier, as written in one file, to the directory it refers to.
// An alias may differ from the package name (transactionDomain), and the last path segment may
// differ from both (port/event/custodykey declares package custodykeyevents), so only the import
// list can answer this.
type fileScope struct {
	dir     string
	imports map[string]string
}

// publishSite is one `<bus>.Publish(ctx, <event>)` call, kept unresolved until the whole tree is
// parsed: the argument may be built by a constructor declared in another package.
type publishSite struct {
	dir   string
	pos   token.Position
	cross bool
	arg   ast.Expr
	fn    *ast.FuncDecl
	scope *fileScope
}

type sourceIndex struct {
	pkgNames map[string]string // dir -> package name
	structs  map[typeRef]*structDecl
	ctors    map[typeRef]typeRef // constructor -> the type it returns
	sites    []publishSite
}

// scanner holds one parse of src/main. Consumers cross bounded context boundaries, and so can
// publishers, so the whole tree is indexed once and every context is discovered against it.
type scanner struct {
	index     *sourceIndex
	consumers map[string][]string
}

func newScanner(srcMain string) (*scanner, error) {
	index, err := indexSources(srcMain)
	if err != nil {
		return nil, err
	}

	consumers, err := discoverConsumers(srcMain)
	if err != nil {
		return nil, err
	}

	return &scanner{index: index, consumers: consumers}, nil
}

// discover returns the events declared under bcRoot: everything embedding event.BaseEvent or
// event.BaseDltEvent, plus anything published, which catches an event that reaches the bus
// without embedding either.
func (s *scanner) discover(bcRoot string) ([]discovered, error) {
	index := s.index
	events := map[typeRef]*discovered{}

	for ref, decl := range index.structs {
		if !within(bcRoot, ref.dir) || !embedsBaseEvent(decl.embeds) {
			continue
		}

		events[ref] = newDiscovered(decl)
	}

	for _, site := range index.sites {
		inBC := within(bcRoot, site.dir)

		ref, err := index.resolvePublishedType(site)
		if err != nil {
			// A publish site elsewhere in the repository is best-effort: it may well be a form
			// this resolver does not read, and an unrelated bounded context must not be able to
			// break this one's generation. Inside the BC it is an error, since a published event
			// missing from the catalog is the failure this file exists to prevent.
			if !inBC {
				continue
			}

			return nil, err
		}

		if !within(bcRoot, ref.dir) {
			continue
		}

		if events[ref] == nil {
			events[ref] = newDiscovered(index.structs[ref])
		}

		events[ref].publishers = appendUnique(events[ref].publishers, index.publisherLabel(bcRoot, site))
		events[ref].cross = events[ref].cross || site.cross
	}

	out := make([]discovered, 0, len(events))

	for _, ev := range events {
		ev.consumers = s.consumers[ev.goType]
		slices.Sort(ev.publishers)

		out = append(out, *ev)
	}

	slices.SortFunc(out, func(a, b discovered) int { return strings.Compare(a.goType, b.goType) })

	return out, nil
}

func newDiscovered(decl *structDecl) *discovered {
	return &discovered{
		goType:    decl.pkgName + "." + decl.ref.name,
		typeName:  decl.ref.name,
		aggregate: filepath.Base(decl.ref.dir),
		pos:       fmt.Sprintf("%s:%d", decl.pos.Filename, decl.pos.Line),
	}
}

func embedsBaseEvent(embeds []embeddedType) bool {
	for _, embed := range embeds {
		if !strings.HasSuffix(embed.importPath, coreEventPkg) {
			continue
		}

		if embed.name == baseEvent || embed.name == baseDltEvent {
			return true
		}
	}

	return false
}

// publisherLabel names the producer the way the layout already does: app/command/createkey is the
// createkey command handler, port/listener/.../keycreated is the keycreated listener. A publisher
// in another bounded context keeps its path, since that is the interesting part.
func (index *sourceIndex) publisherLabel(bcRoot string, site publishSite) string {
	rel, err := filepath.Rel(bcRoot, site.dir)
	if err != nil || strings.HasPrefix(rel, "..") {
		return strings.TrimPrefix(site.dir, srcMainDir+string(filepath.Separator))
	}

	base := filepath.Base(rel)

	switch {
	case strings.Contains(rel, "command"):
		return base + " command handler"
	case strings.Contains(rel, "listener"):
		return base + " listener"
	default:
		return rel
	}
}

func within(root, dir string) bool {
	rel, err := filepath.Rel(root, dir)

	return err == nil && !strings.HasPrefix(rel, "..")
}

// indexSources parses every non-test file under root once, collecting what resolution needs:
// package names, struct declarations with what they embed, constructors and publish call sites.
func indexSources(root string) (*sourceIndex, error) {
	index := &sourceIndex{
		pkgNames: map[string]string{},
		structs:  map[typeRef]*structDecl{},
		ctors:    map[typeRef]typeRef{},
	}

	fset := token.NewFileSet()

	files := map[string][]*ast.File{}
	scopes := map[*ast.File]*fileScope{}

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Comments and object resolution are the expensive halves of parsing, and nothing here
		// reads either.
		parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		dir := filepath.Dir(path)
		index.pkgNames[dir] = parsed.Name.Name
		files[dir] = append(files[dir], parsed)

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Scopes are built only once every package name is known: an import written without an alias
	// is referred to by the package's own name, which need not match its last path segment.
	for dir, parsedFiles := range files {
		for _, parsed := range parsedFiles {
			scopes[parsed] = newFileScope(dir, parsed, index.pkgNames)
		}
	}

	// Declarations before publish sites, so that a publish site parsed in one package can be
	// resolved against a constructor declared in another.
	for dir, parsedFiles := range files {
		for _, parsed := range parsedFiles {
			index.collectDeclarations(dir, parsed, scopes[parsed], fset)
		}
	}

	for dir, parsedFiles := range files {
		for _, parsed := range parsedFiles {
			index.collectPublishSites(dir, parsed, scopes[parsed], fset)
		}
	}

	return index, nil
}

func newFileScope(dir string, parsed *ast.File, pkgNames map[string]string) *fileScope {
	scope := &fileScope{dir: dir, imports: map[string]string{}}

	for _, spec := range parsed.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil || !strings.HasPrefix(path, modulePath+"/") {
			continue
		}

		imported := strings.TrimPrefix(path, modulePath+"/")

		// The package's own name, which is what an unaliased import is referred to by; a directory
		// outside the scanned tree falls back to its path segment.
		qualifier, known := pkgNames[imported]
		if !known {
			qualifier = filepath.Base(imported)
		}

		if spec.Name != nil {
			qualifier = spec.Name.Name
		}

		scope.imports[qualifier] = imported
	}

	return scope
}

func (scope *fileScope) dirOf(qualifier string) (string, bool) {
	dir, ok := scope.imports[qualifier]

	return dir, ok
}

func (index *sourceIndex) collectDeclarations(
	dir string,
	parsed *ast.File,
	scope *fileScope,
	fset *token.FileSet,
) {
	for _, decl := range parsed.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			index.collectStructs(dir, decl, scope, fset)
		case *ast.FuncDecl:
			index.collectConstructor(dir, decl, scope)
		}
	}
}

func (index *sourceIndex) collectStructs(
	dir string,
	decl *ast.GenDecl,
	scope *fileScope,
	fset *token.FileSet,
) {
	if decl.Tok != token.TYPE {
		return
	}

	for _, spec := range decl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		var embeds []embeddedType

		for _, field := range structType.Fields.List {
			if len(field.Names) > 0 {
				continue
			}

			embed, ok := embeddedRef(field.Type, scope)
			if ok {
				embeds = append(embeds, embed)
			}
		}

		ref := typeRef{dir: dir, name: typeSpec.Name.Name}
		index.structs[ref] = &structDecl{
			ref:     ref,
			pkgName: index.pkgNames[dir],
			embeds:  embeds,
			pos:     fset.Position(typeSpec.Pos()),
		}
	}
}

// embeddedRef reads an embedded field as an import path plus a type name. Anything outside the
// module keeps its qualifier as the path, which is all embedsBaseEvent needs: the iob-go-core
// event package is always imported as "event".
func embeddedRef(expr ast.Expr, scope *fileScope) (embeddedType, bool) {
	expr = deref(expr)

	switch expr := expr.(type) {
	case *ast.Ident:
		return embeddedType{name: expr.Name}, true
	case *ast.SelectorExpr:
		qualifier, ok := expr.X.(*ast.Ident)
		if !ok {
			return embeddedType{}, false
		}

		path, ok := scope.dirOf(qualifier.Name)
		if !ok {
			path = qualifier.Name
			if qualifier.Name == "event" {
				path = coreEventPkg
			}
		}

		return embeddedType{importPath: path, name: expr.Sel.Name}, true
	default:
		return embeddedType{}, false
	}
}

// collectConstructor records `func NewXEvent(...) *XEvent`, the other way a published value is
// spelled at the call site.
func (index *sourceIndex) collectConstructor(dir string, decl *ast.FuncDecl, scope *fileScope) {
	if decl.Recv != nil || decl.Type.Results == nil || len(decl.Type.Results.List) == 0 {
		return
	}

	ref, ok := index.resolveTypeExpr(decl.Type.Results.List[0].Type, scope)
	if !ok {
		return
	}

	index.ctors[typeRef{dir: dir, name: decl.Name.Name}] = ref
}

func (index *sourceIndex) collectPublishSites(
	dir string,
	parsed *ast.File,
	scope *fileScope,
	fset *token.FileSet,
) {
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			receiver, ok := busReceiver(call)
			if !ok {
				return true
			}

			index.sites = append(index.sites, publishSite{
				dir:   dir,
				pos:   fset.Position(call.Pos()),
				cross: strings.Contains(strings.ToLower(receiver), "cross"),
				arg:   call.Args[1],
				fn:    fn,
				scope: scope,
			})

			return true
		})
	}
}

// busReceiver matches `<something>bus.Publish(ctx, event)` and returns the receiver's name. The
// name is the only signal available without type checking, and it separates the event buses from
// the domain methods that also happen to be called Publish.
func busReceiver(call *ast.CallExpr) (string, bool) {
	method, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || method.Sel.Name != "Publish" || len(call.Args) != 2 {
		return "", false
	}

	var receiver string

	switch expr := method.X.(type) {
	case *ast.Ident:
		receiver = expr.Name
	case *ast.SelectorExpr:
		receiver = expr.Sel.Name
	default:
		return "", false
	}

	if !strings.Contains(strings.ToLower(receiver), "bus") {
		return "", false
	}

	return receiver, true
}

// resolvePublishedType answers what type was handed to Publish.
func (index *sourceIndex) resolvePublishedType(site publishSite) (typeRef, error) {
	ref, ok := index.resolveValueExpr(site.arg, site.fn, site.scope)
	if !ok {
		return typeRef{}, fmt.Errorf(
			"%s: cannot tell which event is published here — publish a composite literal, a"+
				" declared variable or a constructor result, or teach resolveValueExpr this form",
			site.pos,
		)
	}

	if _, known := index.structs[ref]; !known {
		return typeRef{}, fmt.Errorf("%s: published type %s is not a struct of src/main", site.pos, ref.name)
	}

	return ref, nil
}

// resolveValueExpr resolves the expression passed to Publish to the struct it builds: a literal,
// a constructor call, or a local variable holding either.
func (index *sourceIndex) resolveValueExpr(expr ast.Expr, fn *ast.FuncDecl, scope *fileScope) (typeRef, bool) {
	switch expr := deref(expr).(type) {
	case *ast.CompositeLit:
		return index.resolveTypeExpr(expr.Type, scope)
	case *ast.CallExpr:
		return index.resolveCtorCall(expr, scope)
	case *ast.Ident:
		return index.resolveLocalVar(expr.Name, fn, scope)
	default:
		return typeRef{}, false
	}
}

func (index *sourceIndex) resolveCtorCall(call *ast.CallExpr, scope *fileScope) (typeRef, bool) {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		ref, ok := index.ctors[typeRef{dir: scope.dir, name: fun.Name}]

		return ref, ok
	case *ast.SelectorExpr:
		qualifier, ok := fun.X.(*ast.Ident)
		if !ok {
			return typeRef{}, false
		}

		dir, ok := scope.dirOf(qualifier.Name)
		if !ok {
			return typeRef{}, false
		}

		ref, ok := index.ctors[typeRef{dir: dir, name: fun.Sel.Name}]

		return ref, ok
	default:
		return typeRef{}, false
	}
}

// resolveLocalVar walks the enclosing function for the declaration of name, covering the two
// forms this repository uses: `evt := pkg.Event{...}` and `var evt *pkg.Event` filled in later.
func (index *sourceIndex) resolveLocalVar(name string, fn *ast.FuncDecl, scope *fileScope) (typeRef, bool) {
	if fn == nil || fn.Body == nil {
		return typeRef{}, false
	}

	var (
		ref   typeRef
		found bool
	)

	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if found {
			return false
		}

		switch node := node.(type) {
		case *ast.AssignStmt:
			for i, lhs := range node.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if !ok || ident.Name != name {
					continue
				}

				// x, err := f() assigns from a single call; x := lit assigns positionally.
				rhs := node.Rhs[0]
				if len(node.Rhs) == len(node.Lhs) {
					rhs = node.Rhs[i]
				}

				ref, found = index.resolveValueExpr(rhs, fn, scope)
			}
		case *ast.ValueSpec:
			if node.Type == nil || !slices.ContainsFunc(node.Names, func(n *ast.Ident) bool { return n.Name == name }) {
				return true
			}

			ref, found = index.resolveTypeExpr(node.Type, scope)
		}

		return !found
	})

	return ref, found
}

// resolveTypeExpr turns a type as written — Event, *Event, pkg.Event, *pkg.Event — into the
// directory-qualified reference the index is keyed by.
func (index *sourceIndex) resolveTypeExpr(expr ast.Expr, scope *fileScope) (typeRef, bool) {
	switch expr := deref(expr).(type) {
	case *ast.Ident:
		return typeRef{dir: scope.dir, name: expr.Name}, true
	case *ast.SelectorExpr:
		qualifier, ok := expr.X.(*ast.Ident)
		if !ok {
			return typeRef{}, false
		}

		dir, ok := scope.dirOf(qualifier.Name)
		if !ok {
			return typeRef{}, false
		}

		return typeRef{dir: dir, name: expr.Sel.Name}, true
	default:
		return typeRef{}, false
	}
}

func deref(expr ast.Expr) ast.Expr {
	for {
		switch inner := expr.(type) {
		case *ast.StarExpr:
			expr = inner.X
		case *ast.UnaryExpr:
			if inner.Op != token.AND {
				return expr
			}

			expr = inner.X
		case *ast.ParenExpr:
			expr = inner.X
		default:
			return expr
		}
	}
}

// discoverConsumers maps an event type to the listeners registered for it, across every bounded
// context: a consumer of a dltingress event may well live in issuance. The EventType/Name pair is
// read from the genlistener output, which is the same string the runtime registry keys on.
func discoverConsumers(srcMain string) (map[string][]string, error) {
	wired, err := wiredListenerDirs(srcMain)
	if err != nil {
		return nil, err
	}

	consumers := map[string][]string{}
	fset := token.NewFileSet()

	err = filepath.WalkDir(srcMain, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() || !strings.HasSuffix(path, "_listener_gen.go") {
			return nil
		}

		if !wired[filepath.Dir(path)] {
			return nil
		}

		parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		eventType := returnedString(parsed, "EventType")
		name := returnedString(parsed, "Name")

		if eventType == "" || name == "" {
			return fmt.Errorf("%s: expected generated EventType and Name methods returning a literal", path)
		}

		consumers[eventType] = appendUnique(consumers[eventType], name)

		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, names := range consumers {
		slices.Sort(names)
	}

	return consumers, nil
}

// wiredListenerDirs is the set of directories imported by a config/eventlisteners.go. A listener
// package that nothing imports there is not registered, so its event has no consumer.
func wiredListenerDirs(srcMain string) (map[string]bool, error) {
	dirs := map[string]bool{}
	fset := token.NewFileSet()

	entries, err := os.ReadDir(srcMain)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(srcMain, entry.Name(), "config", "eventlisteners.go")

		parsed, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		for _, dir := range newFileScope(filepath.Dir(path), parsed, nil).imports {
			dirs[dir] = true
		}
	}

	return dirs, nil
}

func returnedString(parsed *ast.File, method string) string {
	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name.Name != method || fn.Body == nil || len(fn.Body.List) != 1 {
			continue
		}

		ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			continue
		}

		literal, ok := ret.Results[0].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			continue
		}

		value, err := strconv.Unquote(literal.Value)
		if err != nil {
			continue
		}

		return value
	}

	return ""
}

func appendUnique(values []string, value string) []string {
	if slices.Contains(values, value) {
		return values
	}

	return append(values, value)
}
