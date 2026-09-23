package main

import (
	"bytes"
	cmdcore "dlt-ingress/src/cmd/core"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var tmpl, tmplTest *template.Template
var completedOk = false

func init() {
	var err error
	titleCaser := cases.Title(language.Und)

	funcs := template.FuncMap{
		"EntityName":    entityName,
		"EntityPackage": entityPackage,
		"Title":         titleCaser.String,
		"ToShort":       ToShort,
	}
	tmpl, err = template.New("method.tmpl").Funcs(funcs).ParseFiles("./../cmd/genrepositories/method.tmpl")
	if err != nil {
		fmt.Println("❌ Failed to load template:", err)
		os.Exit(1)
	}
	tmplTest, err = template.New("mock.tmpl").Funcs(funcs).ParseFiles("./../cmd/genrepositories/mock.tmpl")
	if err != nil {
		fmt.Println("❌ Failed to load template:", err)
		os.Exit(1)
	}

}

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("🛢️ Starting repositories generator")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("───────────────────────────────")

	if completedOk {
		fmt.Println("✅ Repositories generation finished successfully")
	} else {
		fmt.Println("❌ Repositories generation failed")
	}
	fmt.Println("───────────────────────────────")
	fmt.Println()
	fmt.Println()
}

func main() {
	starting()
	defer ending()

	rootBase := "./../main/"
	anyError := false

	for _, bc := range cmdcore.BOUNDED_CONTEXTS {
		if !cmdcore.IsMigrated(bc) {
			fmt.Printf("-- Skipping %s: Handled by old generator\n", bc)
			continue
		}

		fmt.Printf(">> Processing Bounded Context: %s\n", bc)

		bcPath := filepath.Join(rootBase, bc)

		if _, err := os.Stat(bcPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(bcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			return processPackage(bc, path)
		})

		if err != nil {
			fmt.Printf("? Error processing BC %s: %v\n", bc, err)
			anyError = true
		}
	}

	if !anyError {
		completedOk = true
	}
}

func processPackage(bc, dir string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for pkgName, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				gen, ok := decl.(*ast.GenDecl)
				if !ok || gen.Tok != token.TYPE {
					continue
				}

				for _, spec := range gen.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || !strings.HasSuffix(typeSpec.Name.Name, "Repository") {
						continue
					}

					interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
					if !ok {
						continue
					}

					entityName, ok := extractEntityFromBaseRepo(interfaceType)
					if !ok {
						continue
					}

					var relationName string
					if gen.Doc != nil {
						for _, comment := range gen.Doc.List {
							if strings.Contains(comment.Text, "+gen:rel=") {
								relationName = strings.TrimSpace(strings.Split(comment.Text, "+gen:rel=")[1])
							}
						}
					}

					absPath, _ := filepath.Abs(fset.File(file.Pos()).Name())
					domainImport := buildDomainImportFromPath(absPath)

					infraDir := strings.Replace(dir, "domain", "infra/repository", 1)

					return findAndGenerate(bc, infraDir, pkgName, typeSpec.Name.Name, domainImport, entityName, relationName)
				}
			}
		}
	}
	return nil
}

func extractEntityFromBaseRepo(iface *ast.InterfaceType) (string, bool) {
	for _, method := range iface.Methods.List {
		idxExpr, ok := method.Type.(*ast.IndexListExpr)
		if !ok {
			continue
		}

		if len(idxExpr.Indices) == 2 {
			if ident, ok := idxExpr.Indices[1].(*ast.Ident); ok {
				return ident.Name, true
			}
		}
	}
	return "", false
}

func buildDomainImportFromPath(path string) string {
	idx := strings.Index(path, "main/")
	if idx == -1 {
		return ""
	}

	importPath := path[idx:]
	importPath = filepath.Dir(importPath)
	return "dlt-ingress/src/" + filepath.ToSlash(importPath)
}

func findAndGenerate(bc, infraDir, pkgName, repoName, domainImport, entityName, relation string) error {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, infraDir, nil, 0)
	if err != nil {
		return nil
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			fileName := fset.File(file.Pos()).Name()
			if !strings.Contains(fileName, "postgres_") {
				continue
			}

			fullEntityPath := strings.ToLower(entityName) + "." + entityName

			err := generateCode(bc, fileName, pkg.Name, "Postgres"+repoName, domainImport, fullEntityPath, relation)
			if err != nil {
				return err
			}
			return generateMock(bc, fileName, pkg.Name, "Postgres"+repoName, domainImport, fullEntityPath)
		}
	}
	return nil
}

func generateCode(bc, origPath, packageName, typeName, domainImport, entityNamePackage, relationName string) error {
	prefix := extractPrefix(origPath)
	genName := prefix + "commonrepo_gen.go"
	genPath := filepath.Join(filepath.Dir(origPath), genName)
	domainErrorImport := getDomainErrorsPath(bc, domainImport)

	imports := []string{
		"context",
		"fmt",
		"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/db",
		"errors",
		"github.com/google/uuid",
		"gorm.io/gorm",
		"gorm.io/gorm/clause",
		"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination",
		strings.Replace(domainImport, "\"", "", -1),
	}

	if domainErrorImport != "" {
		imports = append(imports, domainErrorImport)
	}

	sort.Strings(imports)
	entityName := strings.Split(entityNamePackage, ".")[1]

	var notFoundDomainError string
	if cmdcore.IsErrorRefactor(bc) {
		notFoundDomainError = fmt.Sprintf("domainerrors.NewEntityNotFoundDomainError(\"%s\", id)", entityName)
	} else {
		notFoundDomainError = fmt.Sprintf("domainerrors.NewEntityNotFoundError(\"%s\", id, nil)", entityName)
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, struct {
		Package             string
		Repository          string
		DomainImport        string
		Entity              string
		EntityName          string
		Imports             []string
		Relation            string
		NotFoundDomainError string
	}{
		Package:             packageName,
		Repository:          typeName,
		DomainImport:        domainImport,
		Entity:              entityNamePackage,
		EntityName:          entityName,
		Imports:             imports,
		Relation:            relationName,
		NotFoundDomainError: notFoundDomainError,
	})

	if err != nil {
		return err
	}
	return os.WriteFile(genPath, buf.Bytes(), 0644)
}

func generateMock(bc, origPath, packageName, typeName, domainImport, entityNamePackage string) error {
	prefix := extractPrefix(origPath)
	mockDir := filepath.Join(filepath.Dir(origPath), "mock")
	genName := prefix + "repository_mock_gen.go"
	genPath := filepath.Join(mockDir, genName)

	if err := os.MkdirAll(mockDir, 0755); err != nil {
		return err
	}

	mockPackageName := "mock" + packageName
	domainImport = strings.Replace(domainImport, "\"", "", -1)
	fmt.Println("?? Generating Mock:", genPath)

	imports := []string{
		"\"context\"",
		"\"github.com/google/uuid\"",
		"\"github.com/stretchr/testify/mock\"",
		"\"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/api/pagination\"",
		"\"" + domainImport + "\"",
	}
	sort.Strings(imports)
	entityName := strings.Split(entityNamePackage, ".")[1]

	var buf bytes.Buffer
	err := tmplTest.Execute(&buf, struct {
		Package      string
		Repository   string
		DomainImport string
		Entity       string
		EntityName   string
		Imports      []string
	}{
		Package:      mockPackageName,
		Repository:   typeName,
		DomainImport: domainImport,
		Entity:       entityNamePackage,
		EntityName:   entityName,
		Imports:      imports,
	})

	if err != nil {
		return err
	}
	return os.WriteFile(genPath, buf.Bytes(), 0644)
}

func getDomainErrorsPath(bc, importPath string) string {
	if cmdcore.IsErrorRefactor(bc) {
		return "gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/domain/domainerrors"
	} else {
		path := strings.Trim(importPath, "\" ")
		re := regexp.MustCompile(`(dlt-ingress/src/main/)([^/]+)/((internal/)?domain)`)
		matches := re.FindStringSubmatch(path)

		if len(matches) < 4 {
			return ""
		}

		prefix := matches[1]
		boundedContext := matches[2]
		domainPart := matches[3]

		return fmt.Sprintf("%s%s/%s/domainerrors", prefix, boundedContext, domainPart)
	}
}

func entityName(entity string) string {
	parts := strings.SplitN(entity, ".", 2)
	if len(parts) > 0 {
		return strings.ToLower(parts[1])
	}
	return ""
}

func entityPackage(entity string) string {
	if entity == "baseprocess.BaseProcess" {
		return "process"
	}
	parts := strings.SplitN(entity, ".", 2)
	if len(parts) > 0 {
		return strings.ToLower(parts[0])
	}
	return ""
}

func extractPrefix(path string) string {
	const suffix = "postgres_repository.go"
	base := filepath.Base(path)
	if !strings.HasSuffix(base, suffix) {
		return ""
	}
	prefix := base[:len(base)-len(suffix)]
	if prefix == "" {
		return ""
	}
	return prefix
}

func ToShort(name string) string {
	if len(name) == 0 {
		return ""
	}

	n := strings.ToLower(name)
	fmt.Println("extracting short name for", name, "=>", n)

	// ⚠️ explicit override
	overrides := map[string]string{
		"asset": "ast", // avoid "ass"
		"user":  "usr", // avoid "use"
	}

	for prefix, value := range overrides {
		if strings.HasPrefix(n, prefix) {
			return value
		}
	}

	res := string(n[0])

	// remove vowels
	for i := 1; i < len(n) && len(res) < 3; i++ {
		c := n[i]
		if c != 'a' && c != 'e' && c != 'i' && c != 'o' && c != 'u' {
			res += string(c)
		}
	}

	// autocomplete
	for i := 1; i < len(n) && len(res) < 3; i++ {
		res += string(n[i])
	}

	// fallback
	if len(res) < 3 && len(n) >= 3 {
		return n[:3]
	}

	if len(res) > 3 {
		return res[:3]
	}

	return res
}
