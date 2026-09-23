package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"unicode"

	cmdcore "dlt-ingress/src/cmd/core"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type EntityArgs struct {
	Action          string
	Entity          string
	BoundedContext  string
	ExtraPath       string
	IsErrorRefactor bool
}

var (
	// Domain templates
	entityTmpl, entityEventsTmpl, entityExistsTmpl                       *template.Template
	processTmpl, processEventsTmpl, processExistsTmpl, processStatusTmpl *template.Template
	// Repository templates
	entityRepoTmpl, entityPostgresRepoTmpl   *template.Template
	processRepoTmpl, processPostgresRepoTmpl *template.Template
)

var completedOk = false

func init() {
	titleCaser := cases.Title(language.Und)
	funcs := template.FuncMap{
		"ToLower":      strings.ToLower,
		"ToSnake":      toSnakeCase,
		"ToDone":       ToDone,
		"Title":        titleCaser.String,
		"ToLowerFirst": toLowerFirst,
	}

	_, fileName, _, _ := runtime.Caller(0)
	tmplRoot := filepath.Join(filepath.Dir(fileName), "")

	entityTmpl = mustLoadTemplate("entity.tmpl", tmplRoot, funcs)
	entityEventsTmpl = mustLoadTemplate("events.tmpl", tmplRoot, funcs)
	entityExistsTmpl = mustLoadTemplate("entity_exists_service.tmpl", tmplRoot, funcs)
	processTmpl = mustLoadTemplate("process.tmpl", tmplRoot, funcs)
	processEventsTmpl = mustLoadTemplate("process_events.tmpl", tmplRoot, funcs)
	processExistsTmpl = mustLoadTemplate("process_exists_service.tmpl", tmplRoot, funcs)
	processStatusTmpl = mustLoadTemplate("process_check_status_service.tmpl", tmplRoot, funcs)
	entityRepoTmpl = mustLoadTemplate("entity_repository.tmpl", tmplRoot, funcs)
	entityPostgresRepoTmpl = mustLoadTemplate("entity_postgres_repository.tmpl", tmplRoot, funcs)
	processRepoTmpl = mustLoadTemplate("process_repository.tmpl", tmplRoot, funcs)
	processPostgresRepoTmpl = mustLoadTemplate("process_postgres_repository.tmpl", tmplRoot, funcs)
}

func main() {
	starting()
	defer ending()

	// --- FLAGS ---
	fBC := flag.String("bc", "", "Bounded Context")
	fEnt := flag.String("entity", "", "Entity")
	fAction := flag.String("action", "", "Action")
	fExtra := flag.String("extra", "", "Extra Path")
	flag.Parse()

	var args EntityArgs
	if *fBC != "" && *fEnt != "" {
		args = EntityArgs{BoundedContext: *fBC, Entity: *fEnt, Action: *fAction, ExtraPath: *fExtra}
	} else {
		args = runInteractiveMode()
	}

	// --- PATH CONFIGURATION ---
	_, fileName, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(fileName), "./../../main")

	bc := strings.ToLower(args.BoundedContext)
	validateBoundedContext(bc)
	args.IsErrorRefactor = cmdcore.IsErrorRefactor(bc)
	entityLower := strings.ToLower(args.Entity)
	extra := strings.ToLower(args.ExtraPath)
	procName := strings.ToLower(args.Action + args.Entity)

	// --- STEP 1: PROTECTION (Check if process already exists) ---
	procDomainDir := filepath.Join(root, bc, "domain/process", procName)
	if _, err := os.Stat(filepath.Join(procDomainDir, procName+"_process.go")); err == nil {
		fmt.Printf("❌ ERROR: Process '%s' already exists. Aborting to prevent overwrite.\n", args.Action+args.Entity)
		os.Exit(1)
	}

	// --- STEP 2: BASE ENTITY (Domain + Repo) ---
	entityDomainDir := filepath.Join(root, bc, "domain", extra, entityLower)
	entityFile := filepath.Join(entityDomainDir, entityLower+".go")

	if _, err := os.Stat(entityFile); os.IsNotExist(err) {
		fmt.Println("🏗️  Base entity not found. Generating domain and persistence...")

		// Domain
		generateFile(entityFile, entityTmpl, args)
		generateFile(filepath.Join(entityDomainDir, "service", entityLower+"_exists_service.go"), entityExistsTmpl, args)

		// Persistence (Repository)
		entityRepoDir := filepath.Join(root, bc, "port/repository", extra, entityLower)
		entitySnake := toSnakeCase(args.Entity)
		generateFile(filepath.Join(entityRepoDir, entitySnake+"_repository.go"), entityRepoTmpl, args)
		generateFile(filepath.Join(entityRepoDir, entitySnake+"_postgres_repository.go"), entityPostgresRepoTmpl, args)
	} else {
		fmt.Printf("ℹ️  Entity '%s' already exists. Skipping base creation.\n", args.Entity)
	}

	// --- STEP 3: PROCESS (Domain + Repo + Events) ---
	if args.Action != "" {
		fmt.Printf("🚀 Generating process skeleton for '%s'...\n", args.Action+args.Entity)

		// A) Entity Events
		entityEventsFile := filepath.Join(entityDomainDir, entityLower+"_events.go")
		if _, err := os.Stat(entityEventsFile); os.IsNotExist(err) {
			generateFile(entityEventsFile, entityEventsTmpl, args)
		}

		// B) Process Domain
		generateFile(filepath.Join(procDomainDir, procName+"_process.go"), processTmpl, args)
		generateFile(filepath.Join(procDomainDir, procName+"_process_events.go"), processEventsTmpl, args)
		generateFile(filepath.Join(procDomainDir, "service", procName+"_exists_service.go"), processExistsTmpl, args)
		generateFile(filepath.Join(procDomainDir, "service", procName+"_check_status_service.go"), processStatusTmpl, args)

		// C) Process Repository
		procRepoDir := filepath.Join(root, bc, "port/repository/process", procName)
		procFilePrefix := toSnakeCase(args.Action + args.Entity)
		generateFile(filepath.Join(procRepoDir, procFilePrefix+"_process_repository.go"), processRepoTmpl, args)
		generateFile(filepath.Join(procRepoDir, procFilePrefix+"_process_postgres_repository.go"), processPostgresRepoTmpl, args)

		// D) Inject Type constant
		injectProcessType(root, bc, args.Action, args.Entity)
		// E) Inject Domain Services & repo instantiation
		injectDomainServices(root, bc, args.Action, args.Entity, args.ExtraPath)
		injectRepositories(root, bc, args.Action, args.Entity, args.ExtraPath)
	}

	cmd := exec.Command("make", "generate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		logger.Error("Error executing make generate", "error", err)
		os.Exit(1)
	}

	cmd = exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Error executing go fmt: %v", err)
	}

	completedOk = true
}

// --- HELPER FUNCTIONS ---

// validateBoundedContext aborts early if the target BC cannot be scaffolded by
// genEntity. The generator only supports LEGACY-pattern bounded contexts (those
// with domain/domainservices.go, port/repository/repositories.go and a
// domain/process Type file). New-pattern contexts (issuance, entitymgmt,
// notifications) live under internal/ and must use their own generator.
func validateBoundedContext(bc string) {
	known := false
	for _, valid := range cmdcore.BOUNDED_CONTEXTS {
		if valid == bc {
			known = true
			break
		}
	}
	if !known {
		fmt.Printf("❌ ERROR: Unknown bounded context '%s'.\n", bc)
		fmt.Printf("   Valid contexts: %s\n", strings.Join(cmdcore.BOUNDED_CONTEXTS, ", "))
		os.Exit(1)
	}

	if cmdcore.IsMigrated(bc) {
		fmt.Printf("❌ ERROR: '%s' uses the new DDD pattern (internal/) and is NOT supported by genEntity.\n", bc)
		fmt.Println("   genEntity only scaffolds LEGACY bounded contexts (e.g. registry, primarymarket, settlement, compliance, shared).")
		fmt.Println("   For new-pattern contexts use the issuancev2 generator instead.")
		os.Exit(1)
	}
}

func runInteractiveMode() EntityArgs {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter Bounded Context (e.g., Settlement): ")
	bc, _ := reader.ReadString('\n')
	fmt.Print("Enter Entity (e.g., Asset): ")
	ent, _ := reader.ReadString('\n')
	fmt.Print("Enter Action (e.g., Create): ")
	act, _ := reader.ReadString('\n')
	fmt.Print("Enter Extra Path (e.g., asset/): ")
	ext, _ := reader.ReadString('\n')

	return EntityArgs{
		BoundedContext: strings.TrimSpace(bc),
		Entity:         strings.TrimSpace(ent),
		Action:         strings.TrimSpace(act),
		ExtraPath:      strings.TrimSpace(ext),
	}
}

func injectProcessType(root, bc, action, entity string) {
	// The process Type file name is NOT consistent across legacy BCs:
	// most use "<bc>process.go" (registry, primarymarket, settlement, shared)
	// while compliance uses "<bc>_process.go". Try both before giving up.
	processDir := filepath.Join(root, bc, "domain/process")
	candidates := []string{
		filepath.Join(processDir, bc+"process.go"),
		filepath.Join(processDir, bc+"_process.go"),
	}

	filePath := ""
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			filePath = candidate
			break
		}
	}
	if filePath == "" {
		fmt.Printf("❌ ERROR: Process Type file not found in %s (tried %sprocess.go and %s_process.go). Is '%s' a legacy BC?\n", processDir, bc, bc, bc)
		os.Exit(1)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ ERROR: Could not read %s: %v\n", filePath, err)
		os.Exit(1)
	}

	typeName := action + entity
	typeValue := strings.ToUpper(toSnakeCase(action + entity))
	newLine := fmt.Sprintf("\t%s Type = \"%s\"\n", typeName, typeValue)

	sContent := string(content)
	if strings.Contains(sContent, typeName+" Type") {
		return
	}

	// Insert INSIDE the `const ( ... )` block: locate its closing ')'.
	// Using the file's last ')' is wrong — some Type files have functions
	// (IsValid, var validType) after the const block.
	constIdx := strings.Index(sContent, "const (")
	if constIdx == -1 {
		fmt.Printf("❌ ERROR: 'const (' block not found in %s\n", filePath)
		os.Exit(1)
	}
	relCloseIdx := strings.Index(sContent[constIdx:], ")")
	if relCloseIdx == -1 {
		fmt.Printf("❌ ERROR: closing ')' for const block not found in %s\n", filePath)
		os.Exit(1)
	}
	insertAt := constIdx + relCloseIdx

	newContent := sContent[:insertAt] + newLine + sContent[insertAt:]
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		fmt.Printf("❌ ERROR: writing %s: %v\n", filePath, err)
		os.Exit(1)
	}
	fmt.Printf("✅ Constant %s added to %s\n", typeName, filePath)
}

func injectDomainServices(root, bc, action, entity, extra string) {
	filePath := filepath.Join(root, bc, "domain/domainservices.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ ERROR: Could not read %s: %v. Is '%s' a legacy BC?\n", filePath, err, bc)
		os.Exit(1)
	}

	sContent := string(content)
	entityLower := strings.ToLower(entity)
	procName := strings.ToLower(action + entity)
	extraPath := strings.ToLower(extra)

	procPkg := "service" + procName + "process"
	entPkg := "service" + entityLower

	procImport := fmt.Sprintf("\t\"dlt-ingress/src/main/%s/domain/process/%s/service\"", bc, procName)
	entImport := fmt.Sprintf("\t\"dlt-ingress/src/main/%s/domain/%s%s/service\"", bc, extraPath, entityLower)

	if !strings.Contains(sContent, procImport) {
		idx := strings.Index(sContent, ")")
		sContent = sContent[:idx] + procImport + "\n" + entImport + "\n" + sContent[idx:]
	}

	if !strings.Contains(sContent, action+entity+"ProcessExists") {
		structLine := fmt.Sprintf("\n\t%sProcessExists %s.Exists\n\t%sProcessStatusCheck %s.StatusCheck\n\t%sExists %s.Exists",
			action+entity, procPkg, action+entity, procPkg, entity, entPkg)

		openingBrace := strings.Index(sContent, "type DomainServices struct {") + len("type DomainServices struct {")
		sContent = sContent[:openingBrace] + structLine + sContent[openingBrace:]
	}

	if !strings.Contains(sContent, action+entity+"ProcessExists:") {
		setupLine := fmt.Sprintf("\n\t\t%sProcessExists: *%s.NewExists(repositories.%sProcessRepo),\n\t\t%sProcessStatusCheck: *%s.NewStatusCheck(repositories.%sProcessRepo),\n\t\t%sExists: *%s.NewExists(repositories.%sRepo),",
			action+entity, procPkg, action+entity, action+entity, procPkg, action+entity, entity, entPkg, entity)

		openingBrace := strings.Index(sContent, "return &DomainServices{") + len("return &DomainServices{")
		sContent = sContent[:openingBrace] + setupLine + sContent[openingBrace:]
	}

	_ = os.WriteFile(filePath, []byte(sContent), 0644)
	fmt.Printf("✅ DomainServices updated in %s\n", filePath)
}
func injectRepositories(root, bc, action, entity, extra string) {
	filePath := filepath.Join(root, bc, "port/repository/repositories.go")
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("❌ ERROR: Could not read %s: %v. Is '%s' a legacy BC?\n", filePath, err, bc)
		os.Exit(1)
	}

	sContent := string(content)
	entityLower := strings.ToLower(entity)
	procName := strings.ToLower(action + entity)
	extraPath := strings.ToLower(extra)

	procRepoPkg := procName + "processrepo"
	entRepoPkg := entityLower + "repo"
	bcTitle := strings.ToUpper(bc[:1]) + bc[1:]

	procRepoImport := fmt.Sprintf("\t\"dlt-ingress/src/main/%s/port/repository/process/%s\"", bc, procName)
	entRepoImport := fmt.Sprintf("\t\"dlt-ingress/src/main/%s/port/repository/%s%s\"", bc, extraPath, entityLower)

	if !strings.Contains(sContent, procRepoImport) {
		idx := strings.Index(sContent, ")")
		sContent = sContent[:idx] + procRepoImport + "\n" + entRepoImport + "\n" + sContent[idx:]
	}

	if !strings.Contains(sContent, action+entity+"ProcessRepo") {
		structLine := fmt.Sprintf("\n\t%sProcessRepo %s.Repository\n\t%sRepo %s.Repository",
			action+entity, procRepoPkg, entity, entRepoPkg)

		marker := "type " + bcTitle + "Repositories struct {"
		openingBrace := strings.Index(sContent, marker) + len(marker)
		sContent = sContent[:openingBrace] + structLine + sContent[openingBrace:]
	}

	if !strings.Contains(sContent, action+entity+"ProcessRepo:") {
		setupLine := fmt.Sprintf("\n\t\t%sProcessRepo: %s.NewPostgres(gormDB, tm),\n\t\t%sRepo: %s.NewPostgres(gormDB, tm),",
			action+entity, procRepoPkg, entity, entRepoPkg)

		marker := "return &" + bcTitle + "Repositories{"
		openingBrace := strings.Index(sContent, marker) + len(marker)
		sContent = sContent[:openingBrace] + setupLine + sContent[openingBrace:]
	}

	_ = os.WriteFile(filePath, []byte(sContent), 0644)
	fmt.Printf("✅ Repositories updated in %s\n", filePath)
}
func generateFile(path string, tmpl *template.Template, args EntityArgs) {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		fmt.Printf("❌ Error creating directories for %s: %v\n", path, err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, args); err != nil {
		fmt.Printf("❌ Error executing template for %s: %v\n", path, err)
		os.Exit(1)
	}

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		fmt.Printf("❌ Error writing file %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Printf("✅ Generated: %s\n", path)
}

func mustLoadTemplate(name, path string, funcs template.FuncMap) *template.Template {
	fullPath := filepath.Join(path, name)
	tmpl, err := template.New(name).Funcs(funcs).ParseFiles(fullPath)
	if err != nil {
		panic(fmt.Sprintf("❌ Critical failure loading template %s: %v", fullPath, err))
	}
	return tmpl
}

func starting() {
	fmt.Println("--------------------------------------")
	fmt.Println("🚀 Starting Genentity Generator")
	fmt.Println("--------------------------------------")
}

func ending() {
	fmt.Println("--------------------------------------")
	if completedOk {
		fmt.Println("✅ Genentity generation finished successfully")
	} else {
		fmt.Println("❌ Genentity generation failed")
	}
	fmt.Println("--------------------------------------")
}

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

func ToDone(action string) string {
	if action == "" {
		return ""
	}
	lower := strings.ToLower(action)
	if strings.HasSuffix(lower, "e") {
		return action + "d"
	}
	return action + "ed"
}

func toLowerFirst(str string) string {
	if str == "" {
		return ""
	}
	runes := []rune(str)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
