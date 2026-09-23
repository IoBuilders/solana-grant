package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type EntityArgs struct {
	Entity         string
	BoundedContext string
	ExtraPath      string
	IsAggregate    bool
	IsProcess      bool   // true when Entity ends in "Process"
	FolderName     string // last path segment: ToLower(Entity) for entities, ToLower(Entity without "Process") for processes
	TypeName       string // for processes: Entity without "Process" suffix (used in base.Type); empty for entities
}

var (
	entityTmpl       *template.Template
	eventsTmpl       *template.Template
	entityRepoTmpl   *template.Template
	entityTestTmpl   *template.Template
	existsSvcTmpl    *template.Template
	existsMockTmpl   *template.Template
	persistModelTmpl *template.Template
	mapperTmpl       *template.Template
	postgresRepoTmpl *template.Template

	// Process-specific templates
	processEntityTmpl       *template.Template
	processEventsTmpl       *template.Template
	processTestTmpl         *template.Template
	processPersistModelTmpl *template.Template
	processMapperTmpl       *template.Template

	// Process domain-service templates (exists / status-check / not-exists + mocks)
	processExistsSvcTmpl       *template.Template
	processStatusCheckSvcTmpl  *template.Template
	processNotExistsSvcTmpl    *template.Template
	processExistsMockTmpl      *template.Template
	processStatusCheckMockTmpl *template.Template
	processNotExistsMockTmpl   *template.Template
)

var completedOk = false

func init() {
	titleCaser := cases.Title(language.Und)
	funcs := template.FuncMap{
		"ToLower":      strings.ToLower,
		"ToSnake":      toSnakeCase,
		"Title":        titleCaser.String,
		"ToLowerFirst": toLowerFirst,
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get working directory: %v", err))
	}
	tmplRoot := filepath.Join(cwd, "src/cmd/genentitynew")

	entityTmpl = mustLoadTemplate("entity.tmpl", tmplRoot, funcs)
	eventsTmpl = mustLoadTemplate("events.tmpl", tmplRoot, funcs)
	entityRepoTmpl = mustLoadTemplate("entity_repository.tmpl", tmplRoot, funcs)
	entityTestTmpl = mustLoadTemplate("entity_test_factory.tmpl", tmplRoot, funcs)
	existsSvcTmpl = mustLoadTemplate("entity_exists_service.tmpl", tmplRoot, funcs)
	existsMockTmpl = mustLoadTemplate("entity_exists_service_mock.tmpl", tmplRoot, funcs)
	persistModelTmpl = mustLoadTemplate("infra_persistence_model.tmpl", tmplRoot, funcs)
	mapperTmpl = mustLoadTemplate("infra_mapper.tmpl", tmplRoot, funcs)
	postgresRepoTmpl = mustLoadTemplate("infra_postgres_repository.tmpl", tmplRoot, funcs)

	processEntityTmpl = mustLoadTemplate("process_entity.tmpl", tmplRoot, funcs)
	processEventsTmpl = mustLoadTemplate("process_events.tmpl", tmplRoot, funcs)
	processTestTmpl = mustLoadTemplate("process_test_factory.tmpl", tmplRoot, funcs)
	processPersistModelTmpl = mustLoadTemplate("process_persistence_model.tmpl", tmplRoot, funcs)
	processMapperTmpl = mustLoadTemplate("process_mapper.tmpl", tmplRoot, funcs)

	processExistsSvcTmpl = mustLoadTemplate("process_exists_service.tmpl", tmplRoot, funcs)
	processStatusCheckSvcTmpl = mustLoadTemplate("process_status_check_service.tmpl", tmplRoot, funcs)
	processNotExistsSvcTmpl = mustLoadTemplate("process_not_exists_service.tmpl", tmplRoot, funcs)
	processExistsMockTmpl = mustLoadTemplate("process_exists_service_mock.tmpl", tmplRoot, funcs)
	processStatusCheckMockTmpl = mustLoadTemplate("process_status_check_service_mock.tmpl", tmplRoot, funcs)
	processNotExistsMockTmpl = mustLoadTemplate("process_not_exists_service_mock.tmpl", tmplRoot, funcs)
}

func main() {
	starting()
	defer ending()

	fBC := flag.String("bc", "", "Bounded Context (e.g. Issuance)")
	fEnt := flag.String("entity", "", "Entity name PascalCase (e.g. Token)")
	fExtra := flag.String("extra", "", "Extra subpath within domain (e.g. asset/)")
	fAggregate := flag.Bool("aggregate", false, "Is aggregate (generates repository + infra layer)")
	flag.Parse()

	var args EntityArgs
	if *fBC != "" && *fEnt != "" {
		args = EntityArgs{BoundedContext: *fBC, Entity: *fEnt, ExtraPath: *fExtra, IsAggregate: *fAggregate}
	} else {
		args = runInteractiveMode()
	}

	// Derive process-specific fields
	entityLower := strings.ToLower(args.Entity)
	if strings.HasSuffix(entityLower, "process") {
		args.IsProcess = true
		if args.ExtraPath == "" {
			args.ExtraPath = "process/"
		} else if !strings.HasPrefix(strings.ToLower(args.ExtraPath), "process/") {
			fmt.Printf("ERROR: Process entities must live under process/ (got extra=%q). All *Process aggregates belong in domain/process/.\n", args.ExtraPath)
			os.Exit(1)
		}
		args.FolderName = strings.TrimSuffix(entityLower, "process")
		args.TypeName = strings.TrimSuffix(args.Entity, "Process")
	} else {
		args.FolderName = entityLower
		args.TypeName = ""
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("failed to get working directory: %v\n", err)
		os.Exit(1)
	}
	root := filepath.Join(cwd, "src/main")

	bc := strings.ToLower(args.BoundedContext)
	extra := strings.ToLower(args.ExtraPath)
	entitySnake := toSnakeCase(args.Entity) // file names use snake_case; package names keep entityLower

	// --- Protection: abort if entity already exists ---
	domainDir := filepath.Join(root, bc, "internal/domain", extra, args.FolderName)
	entityFile := filepath.Join(domainDir, entitySnake+".go")
	if _, err := os.Stat(entityFile); err == nil {
		fmt.Printf("ERROR: Entity '%s' already exists at %s. Aborting to prevent overwrite.\n", args.Entity, entityFile)
		os.Exit(1)
	}

	fmt.Printf("Generating new-architecture entity '%s' for bounded context '%s'...\n\n", args.Entity, args.BoundedContext)

	// --- Domain layer ---
	if args.IsProcess {
		generateFile(entityFile, processEntityTmpl, args)
		generateFile(filepath.Join(domainDir, entitySnake+"_test_factory.go"), processTestTmpl, args)
		// Processes emit the lifecycle events (Started/Ordered/TransitTo*), not the
		// generic entity DLT/domain event pair.
		generateFile(filepath.Join(domainDir, entitySnake+"_events.go"), processEventsTmpl, args)
		// Register the process Type constant in <bc>_process.go — without this the
		// generated process references an undefined Type and won't compile.
		injectProcessType(root, bc, args.TypeName)
	} else {
		generateFile(entityFile, entityTmpl, args)
		generateFile(filepath.Join(domainDir, entitySnake+"_test_factory.go"), entityTestTmpl, args)
		generateFile(filepath.Join(domainDir, entitySnake+"_events.go"), eventsTmpl, args)
	}

	if args.IsAggregate {
		generateFile(filepath.Join(domainDir, entitySnake+"_repository.go"), entityRepoTmpl, args)

		// --- Domain services ---
		serviceDir := filepath.Join(domainDir, "service")
		mockDir := filepath.Join(serviceDir, "mock")
		if args.IsProcess {
			// Processes get exists / status-check / not-exists services + mocks.
			generateFile(filepath.Join(serviceDir, "service_"+entitySnake+"_exists.go"), processExistsSvcTmpl, args)
			generateFile(filepath.Join(serviceDir, "service_"+entitySnake+"_status_check.go"), processStatusCheckSvcTmpl, args)
			generateFile(filepath.Join(serviceDir, "service_"+entitySnake+"_not_exists.go"), processNotExistsSvcTmpl, args)
			generateFile(filepath.Join(mockDir, "service_"+entitySnake+"_exists_mock.go"), processExistsMockTmpl, args)
			generateFile(filepath.Join(mockDir, "service_"+entitySnake+"_status_check_mock.go"), processStatusCheckMockTmpl, args)
			generateFile(filepath.Join(mockDir, "service_"+entitySnake+"_not_exists_mock.go"), processNotExistsMockTmpl, args)
		} else {
			generateFile(filepath.Join(serviceDir, "service_"+entitySnake+"_exists.go"), existsSvcTmpl, args)
			generateFile(filepath.Join(mockDir, "service_"+entitySnake+"_exists_mock.go"), existsMockTmpl, args)
		}

		// --- Infra layer ---
		infraDir := filepath.Join(root, bc, "internal/infra/repository", extra, args.FolderName)

		if args.IsProcess {
			generateFile(filepath.Join(infraDir, entitySnake+"_persistence_model.go"), processPersistModelTmpl, args)
			generateFile(filepath.Join(infraDir, entitySnake+"_mapper.go"), processMapperTmpl, args)
		} else {
			generateFile(filepath.Join(infraDir, entitySnake+"_persistence_model.go"), persistModelTmpl, args)
			generateFile(filepath.Join(infraDir, entitySnake+"_mapper.go"), mapperTmpl, args)
		}
		generateFile(filepath.Join(infraDir, entitySnake+"_postgres_repository.go"), postgresRepoTmpl, args)
		// NOTE: _commonrepo_gen.go and _repository_mock_gen.go are NOT templated here.
		// They are produced by the genrepositories generator (run below), which scans
		// the AST and owns those "// Code generated ... DO NOT EDIT" files.

		// --- Wire-up injections ---
		injectRepositories(root, bc, args.Entity, extra, args.FolderName)
		injectDomainServices(root, bc, args.Entity, extra, args.FolderName, args.IsProcess)

		// --- Generate repository boilerplate (_commonrepo_gen / _repository_mock_gen) ---
		runGenRepositories(root)
	}

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing go fmt: %v\n", err)
		os.Exit(1)
	}

	completedOk = true
}

// --- Interactive mode ---

func runInteractiveMode() EntityArgs {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Bounded Context (e.g. Issuance): ")
	bc, _ := reader.ReadString('\n')
	fmt.Print("Enter Entity name PascalCase (e.g. Token): ")
	ent, _ := reader.ReadString('\n')
	fmt.Print("Enter Extra Path (e.g. asset/ — leave blank if none): ")
	extra, _ := reader.ReadString('\n')
	fmt.Print("Is this an aggregate? Generates repository + infra layer (y/N): ")
	aggregateStr, _ := reader.ReadString('\n')
	aggregateStr = strings.TrimSpace(strings.ToLower(aggregateStr))

	return EntityArgs{
		BoundedContext: strings.TrimSpace(bc),
		Entity:         strings.TrimSpace(ent),
		ExtraPath:      strings.TrimSpace(extra),
		IsAggregate:    aggregateStr == "y" || aggregateStr == "yes",
	}
}

// --- Wire-up: <bc>_process.go (process Type constant) ---

func injectProcessType(root, bc, typeName string) {
	// New-pattern process Type files live at internal/domain/process/<bc>_process.go
	// (e.g. issuance_process.go). Try the no-underscore variant too for safety.
	processDir := filepath.Join(root, bc, "internal/domain/process")
	candidates := []string{
		filepath.Join(processDir, bc+"_process.go"),
		filepath.Join(processDir, bc+"process.go"),
	}

	filePath := ""
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			filePath = candidate
			break
		}
	}
	if filePath == "" {
		fmt.Printf("ERROR: Process Type file not found in %s (tried %s_process.go and %sprocess.go). Create it before adding the first process.\n", processDir, bc, bc)
		os.Exit(1)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("ERROR: Could not read %s: %v\n", filePath, err)
		os.Exit(1)
	}

	sContent := string(content)
	if strings.Contains(sContent, typeName+" Type") {
		return
	}

	typeValue := strings.ToUpper(toSnakeCase(typeName))
	newLine := fmt.Sprintf("\t%s Type = \"%s\"\n", typeName, typeValue)

	// Insert INSIDE the `const ( ... )` block (locate its closing ')'), not the
	// file's last ')' — some Type files may have functions after the const block.
	constIdx := strings.Index(sContent, "const (")
	if constIdx == -1 {
		fmt.Printf("ERROR: 'const (' block not found in %s\n", filePath)
		os.Exit(1)
	}
	relCloseIdx := strings.Index(sContent[constIdx:], ")")
	if relCloseIdx == -1 {
		fmt.Printf("ERROR: closing ')' for const block not found in %s\n", filePath)
		os.Exit(1)
	}
	insertAt := constIdx + relCloseIdx

	newContent := sContent[:insertAt] + newLine + sContent[insertAt:]
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		fmt.Printf("ERROR: writing %s: %v\n", filePath, err)
		os.Exit(1)
	}
	fmt.Printf("OK Constant %s added to %s\n", typeName, filePath)
}

// --- Wire-up: repositories.go ---

func injectRepositories(root, bc, entity, extra, folderName string) {
	filePath := filepath.Join(root, bc, "internal/infra/repository/repositories.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("ERROR: Could not read %s: %v. Is '%s' a refactored (internal/) BC?\n", filePath, err, bc)
		os.Exit(1)
	}

	entityLower := strings.ToLower(entity)
	extraLower := strings.ToLower(extra)

	domainPkg := entityLower
	repoPkgAlias := entityLower + "repo"
	domainImport := fmt.Sprintf("\t\"dlt-ingress/src/main/%s/internal/domain/%s%s\"", bc, extraLower, folderName)
	repoImport := fmt.Sprintf("\t\"dlt-ingress/src/main/%s/internal/infra/repository/%s%s\"", bc, extraLower, folderName)

	sContent := string(content)

	// Add imports if missing
	if !strings.Contains(sContent, domainImport) {
		idx := strings.Index(sContent, ")")
		if idx != -1 {
			sContent = sContent[:idx] + domainImport + "\n" + repoImport + "\n" + sContent[idx:]
		}
	}

	// Detect the struct/return names by pattern instead of hardcoding BC names.
	// The struct is "<Something>Repositories" (e.g. IssuanceRepositories,
	// EntityMgmtRepositories) — the exact casing comes from the file, not from
	// titlecasing the lowercased bc (which produced "Entitymgmt" and missed).
	structOpen := regexp.MustCompile(`type \w+Repositories struct \{`).FindString(sContent)
	if structOpen == "" {
		fmt.Printf("ERROR: no 'type <X>Repositories struct {' found in %s\n", filePath)
		os.Exit(1)
	}
	returnOpen := regexp.MustCompile(`return &\w+Repositories\{`).FindString(sContent)
	if returnOpen == "" {
		fmt.Printf("ERROR: no 'return &<X>Repositories{' found in %s\n", filePath)
		os.Exit(1)
	}

	fieldName := entity + "Repo"
	if !strings.Contains(sContent, fieldName) {
		structLine := fmt.Sprintf("\n\t%s %s.Repository", fieldName, domainPkg)
		idx := strings.Index(sContent, structOpen) + len(structOpen)
		sContent = sContent[:idx] + structLine + sContent[idx:]
	}

	if !strings.Contains(sContent, fieldName+":") {
		setupLine := fmt.Sprintf("\n\t\t%s: %s.NewPostgres(db, tm),", fieldName, repoPkgAlias)
		idx := strings.Index(sContent, returnOpen) + len(returnOpen)
		sContent = sContent[:idx] + setupLine + sContent[idx:]
	}

	if err := os.WriteFile(filePath, []byte(sContent), 0644); err != nil {
		fmt.Printf("ERROR: writing %s: %v\n", filePath, err)
		os.Exit(1)
	}
	fmt.Printf("OK Repositories updated in %s\n", filePath)
}

// --- Wire-up: domainservices.go ---

func injectDomainServices(root, bc, entity, extra, folderName string, isProcess bool) {
	filePath := filepath.Join(root, bc, "internal/domain/domainservices.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("ERROR: Could not read %s: %v. Is '%s' a refactored (internal/) BC?\n", filePath, err, bc)
		os.Exit(1)
	}

	entityLower := strings.ToLower(entity)
	extraLower := strings.ToLower(extra)

	svcPkgAlias := "service" + entityLower
	svcImport := fmt.Sprintf("\t\"dlt-ingress/src/main/%s/internal/domain/%s%s/service\"", bc, extraLower, folderName)
	repoImport := fmt.Sprintf("\t\"dlt-ingress/src/main/%s/internal/infra/repository\"", bc)

	sContent := string(content)

	// Add imports if missing
	if !strings.Contains(sContent, svcImport) {
		idx := strings.Index(sContent, ")")
		if idx != -1 {
			sContent = sContent[:idx] + svcImport + "\n" + repoImport + "\n" + sContent[idx:]
		}
	}

	// A process registers three services (exists / not-exists / status-check);
	// a plain entity registers only exists.
	var structLines, setupLines string
	if isProcess {
		structLines = fmt.Sprintf("\n\t%sExists %s.ExistsInterface\n\t%sNotExists %s.NotExistsInterface\n\t%sStatusCheck %s.StatusCheckInterface",
			entity, svcPkgAlias, entity, svcPkgAlias, entity, svcPkgAlias)
		setupLines = fmt.Sprintf("\n\t\t%sExists: %s.NewExists(repositories.%sRepo),\n\t\t%sNotExists: %s.NewNotExists(repositories.%sRepo),\n\t\t%sStatusCheck: %s.NewStatusCheck(repositories.%sRepo),",
			entity, svcPkgAlias, entity, entity, svcPkgAlias, entity, entity, svcPkgAlias, entity)
	} else {
		structLines = fmt.Sprintf("\n\t%sExists %s.ExistsInterface", entity, svcPkgAlias)
		setupLines = fmt.Sprintf("\n\t\t%sExists: %s.NewExists(repositories.%sRepo),", entity, svcPkgAlias, entity)
	}

	if !strings.Contains(sContent, entity+"Exists ") {
		structOpen := "type DomainServices struct {"
		idx := strings.Index(sContent, structOpen)
		if idx == -1 {
			fmt.Printf("ERROR: 'type DomainServices struct {' not found in %s\n", filePath)
			os.Exit(1)
		}
		insertAt := idx + len(structOpen)
		sContent = sContent[:insertAt] + structLines + sContent[insertAt:]
	}

	if !strings.Contains(sContent, entity+"Exists:") {
		returnOpen := "return &DomainServices{"
		idx := strings.Index(sContent, returnOpen)
		if idx == -1 {
			fmt.Printf("ERROR: 'return &DomainServices{' not found in %s\n", filePath)
			os.Exit(1)
		}
		insertAt := idx + len(returnOpen)
		sContent = sContent[:insertAt] + setupLines + sContent[insertAt:]
	}

	if err := os.WriteFile(filePath, []byte(sContent), 0644); err != nil {
		fmt.Printf("ERROR: writing %s: %v\n", filePath, err)
		os.Exit(1)
	}
	fmt.Printf("OK DomainServices updated in %s\n", filePath)
}

// --- Repository code generation ---

// runGenRepositories invokes the genrepositories generator, which OWNS the
// "// Code generated ... DO NOT EDIT" files (_commonrepo_gen.go and
// _repository_mock_gen.go) by scanning the AST. These must NOT be templated by
// hand. genrepositories' paths are relative to src/main, so it runs with that
// directory as cwd.
func runGenRepositories(srcMain string) {
	cmd := exec.Command("go", "run", "../cmd/genrepositories/main.go")
	cmd.Dir = srcMain
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running genrepositories: %v\n", err)
		os.Exit(1)
	}
}

// --- File generation helpers ---

func generateFile(path string, tmpl *template.Template, args EntityArgs) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		fmt.Printf("Error creating directories for %s: %v\n", path, err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, args); err != nil {
		fmt.Printf("Error executing template for %s: %v\n", path, err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		fmt.Printf("Error writing file %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Printf("Generated: %s\n", path)
}

func mustLoadTemplate(name, path string, funcs template.FuncMap) *template.Template {
	fullPath := filepath.Join(path, name)
	tmpl, err := template.New(name).Funcs(funcs).ParseFiles(fullPath)
	if err != nil {
		panic(fmt.Sprintf("Critical failure loading template %s: %v", fullPath, err))
	}
	return tmpl
}

// --- Banner helpers ---

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("  Starting Genentitynew Generator")
	fmt.Println("  (New-architecture: DDD + Hexagonal)")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("──────────────────────────────────────")
	if completedOk {
		fmt.Println("OK Genentitynew generation finished successfully")
	} else {
		fmt.Println("   Genentitynew generation failed")
	}
	fmt.Println("──────────────────────────────────────")
}

// --- String helpers ---

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func toLowerFirst(str string) string {
	if str == "" {
		return ""
	}
	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
