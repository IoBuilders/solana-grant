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
	"strings"
	"text/template"
	"unicode"

	"gitlab.com/iobuilders/projects/eng/iob-core/iob-go-core/v4/src/main/core/logger"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type TemplateArgs struct {
	Action         string
	Entity         string
	Name           string
	BoundedContext string
	Url            string
	HttpMethod     string
	ExtraPath      string
	DLTEventName   string
}

var args = TemplateArgs{}

// Templates
var (
	adapterStartEndpointTmpl *template.Template
	requestStartEndpointTmpl *template.Template
	adapterDltEndpointTmpl   *template.Template
	startCommandTmpl         *template.Template
	startHandlerTmpl         *template.Template
	orderListenerTmpl        *template.Template
	actionListenerTmpl       *template.Template
	ttorderedListenerTmpl    *template.Template
	ttfinishedListenerTmpl   *template.Template
	orderServiceTmpl         *template.Template
	orderServiceModelTmpl    *template.Template
	startServiceTmpl         *template.Template
	startServiceModelTmpl    *template.Template
	orderCommandTmpl         *template.Template
	orderHandlerTmpl         *template.Template
	actionCommandTmpl        *template.Template
	actionHandlerTmpl        *template.Template
	converterCommandTmpl     *template.Template
	converterHandlerTmpl     *template.Template
	ttorderedCommandTmpl     *template.Template
	ttorderedHandlerTmpl     *template.Template
	ttfinishedCommandTmpl    *template.Template
	ttfinishedHandlerTmpl    *template.Template
)

var completedOk = false

func init() {
	titleCaser := cases.Title(language.Und)

	funcs := template.FuncMap{
		"ToLower":      strings.ToLower,
		"ToSnake":      toSnakeCase,
		"ToNatural":    toNatural,
		"ToDone":       ToDone,
		"Title":        titleCaser.String,
		"ToLowerFirst": toLowerFirst,
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
	} else {
		fmt.Printf("Current working directory: %s\n", cwd)
	}

	tmplRoot := cwd + "/src/cmd/genprocessnew/"

	adapterDltEndpointTmpl = mustLoadTemplate("dlt_http_adapter.tmpl", tmplRoot+"endpoint/dlt", funcs)
	adapterStartEndpointTmpl = mustLoadTemplate("start_http_adapter.tmpl", tmplRoot+"endpoint/start", funcs)
	requestStartEndpointTmpl = mustLoadTemplate("start_process_request.tmpl", tmplRoot+"endpoint/start", funcs)
	startCommandTmpl = mustLoadTemplate("start_process_command.tmpl", tmplRoot+"command/start", funcs)
	startHandlerTmpl = mustLoadTemplate("start_process_handler.tmpl", tmplRoot+"command/start", funcs)
	orderServiceTmpl = mustLoadTemplate("order_service.tmpl", tmplRoot+"service/order", funcs)
	orderServiceModelTmpl = mustLoadTemplate("order_model.tmpl", tmplRoot+"service/order", funcs)
	startServiceTmpl = mustLoadTemplate("start_service.tmpl", tmplRoot+"service/start", funcs)
	startServiceModelTmpl = mustLoadTemplate("start_model.tmpl", tmplRoot+"service/start", funcs)
	orderCommandTmpl = mustLoadTemplate("order_command.tmpl", tmplRoot+"command/order", funcs)
	orderHandlerTmpl = mustLoadTemplate("order_handler.tmpl", tmplRoot+"command/order", funcs)
	actionCommandTmpl = mustLoadTemplate("action_command.tmpl", tmplRoot+"command/action", funcs)
	actionHandlerTmpl = mustLoadTemplate("action_handler.tmpl", tmplRoot+"command/action", funcs)
	converterCommandTmpl = mustLoadTemplate("converter_command.tmpl", tmplRoot+"command/converter", funcs)
	converterHandlerTmpl = mustLoadTemplate("converter_handler.tmpl", tmplRoot+"command/converter", funcs)
	ttorderedCommandTmpl = mustLoadTemplate("ttordered_command.tmpl", tmplRoot+"command/ttordered", funcs)
	ttorderedHandlerTmpl = mustLoadTemplate("ttordered_handler.tmpl", tmplRoot+"command/ttordered", funcs)
	ttfinishedCommandTmpl = mustLoadTemplate("ttfinished_command.tmpl", tmplRoot+"command/ttfinished", funcs)
	ttfinishedHandlerTmpl = mustLoadTemplate("ttfinished_handler.tmpl", tmplRoot+"command/ttfinished", funcs)
	orderListenerTmpl = mustLoadTemplate("order_listener.tmpl", tmplRoot+"listener", funcs)
	actionListenerTmpl = mustLoadTemplate("action_listener.tmpl", tmplRoot+"listener", funcs)
	ttorderedListenerTmpl = mustLoadTemplate("ttordered_listener.tmpl", tmplRoot+"listener", funcs)
	ttfinishedListenerTmpl = mustLoadTemplate("ttfinished_listener.tmpl", tmplRoot+"listener", funcs)
}

func mustLoadTemplate(name, path string, funcs template.FuncMap) *template.Template {
	tmpl, err := template.New(name).Funcs(funcs).ParseFiles(path + "/" + name)
	if err != nil {
		panic(fmt.Sprintf("Failed to load template %s: %v", path+"/"+name, err))
	}
	return tmpl
}

func starting() {
	fmt.Println("──────────────────────────────────────")
	fmt.Println("  Starting Genprocessnew generator")
	fmt.Println("  (New-architecture: DDD + Hexagonal)")
	fmt.Println("──────────────────────────────────────")
}

func ending() {
	fmt.Println("───────────────────────────────")
	if completedOk {
		fmt.Println("OK Genprocessnew generation finished successfully")
	} else {
		fmt.Println("   Genprocessnew generation failed")
	}
	fmt.Println("───────────────────────────────")
}

func main() {
	starting()
	defer ending()

	fAction := flag.String("action", "", "Process action")
	fEntity := flag.String("entity", "", "Entity name")
	fBC := flag.String("bc", "", "Bounded Context")
	fUrl := flag.String("url", "", "Endpoint URL")
	fMethod := flag.String("method", "", "HTTP Method")
	fExtra := flag.String("extra", "", "Extra Path (optional)")
	fDlt := flag.String("dlt", "", "DLT Event name")
	flag.Parse()

	if *fAction != "" && *fEntity != "" {
		args = TemplateArgs{
			Action:         *fAction,
			Entity:         *fEntity,
			BoundedContext: *fBC,
			Url:            *fUrl,
			HttpMethod:     *fMethod,
			ExtraPath:      *fExtra,
			DLTEventName:   *fDlt,
		}
		if args.BoundedContext == "" || args.Url == "" || args.HttpMethod == "" || args.DLTEventName == "" {
			logger.Error("Missing mandatory parameters. Required: -bc, -url, -method, -dlt")
			os.Exit(1)
		}
	} else {
		args = runInteractiveMode()
	}

	args.Name = args.Action + args.Entity
	fmt.Printf("\nProcessing arguments:\n %+v\n\n", args)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}
	root := filepath.Join(cwd, "src/main")

	if checkDltAdapterCodeAlreadyGenerated(root) != nil {
		os.Exit(1)
	}

	if err := executeGeneration(root); err != nil {
		logger.Error("Generation failed", "error", err)
		os.Exit(1)
	}

	// Run code generators. `make generate` covers legacy BCs (//go:generate
	// directives). New-pattern BCs (internal/) have no //go:generate, so the
	// CommandName methods the generated commands rely on are produced by the
	// standalone gencommandname generator instead — run it too.
	runGenerator(exec.Command("make", "generate"), cwd, "make generate")
	runGenerator(exec.Command("go", "run", "../cmd/gencommandname/main.go"), root, "gencommandname")

	cmd := exec.Command("go", "fmt", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error executing go fmt: %v", err)
	}

	completedOk = true
}

// runGenerator runs a code-generation command from the given working directory,
// streaming its output. Aborts on failure.
func runGenerator(cmd *exec.Cmd, dir, label string) {
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Error("Error running "+label, "error", err)
		os.Exit(1)
	}
}

func runInteractiveMode() TemplateArgs {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Action (Example: Delete): ")
	action, _ := reader.ReadString('\n')
	action = strings.TrimSpace(action)
	if action == "" {
		logger.Error("Action cannot be empty")
		os.Exit(1)
	}

	fmt.Print("Enter Entity (Example: Document): ")
	entity, _ := reader.ReadString('\n')
	entity = strings.TrimSpace(entity)
	if entity == "" {
		logger.Error("Entity cannot be empty")
		os.Exit(1)
	}

	fmt.Print("Enter Bounded Context (Example: Issuance): ")
	bc, _ := reader.ReadString('\n')
	bc = strings.TrimSpace(bc)
	if bc == "" {
		logger.Error("Bounded Context cannot be empty")
		os.Exit(1)
	}

	fmt.Print("Enter URL (Example: /issuance/assets/{assetId}): ")
	url, _ := reader.ReadString('\n')
	url = strings.TrimSpace(url)
	if url == "" {
		logger.Error("URL cannot be empty")
		os.Exit(1)
	}

	fmt.Print("Enter HTTP Method (Example: DELETE): ")
	method, _ := reader.ReadString('\n')
	method = strings.TrimSpace(method)
	if method == "" {
		logger.Error("HTTP Method cannot be empty")
		os.Exit(1)
	}

	fmt.Print("Enter Extra Path (Example: asset/ — leave blank if none): ")
	extra, _ := reader.ReadString('\n')
	extra = strings.TrimSpace(extra)

	fmt.Print("Enter DLT Event Name (Example: DocumentRemoved): ")
	dlt, _ := reader.ReadString('\n')
	dlt = strings.TrimSpace(dlt)
	if dlt == "" {
		logger.Error("DLT Event Name cannot be empty")
		os.Exit(1)
	}

	return TemplateArgs{
		Action:         action,
		Entity:         entity,
		BoundedContext: bc,
		Url:            url,
		HttpMethod:     method,
		ExtraPath:      extra,
		DLTEventName:   dlt,
	}
}

// --- Generation orchestration ---

func executeGeneration(root string) error {
	if err := generateDltAdapterCode(root, adapterDltEndpointTmpl); err != nil {
		return err
	}
	if err := generateStartAdapterCode(root, adapterStartEndpointTmpl, requestStartEndpointTmpl); err != nil {
		return err
	}
	if err := generateStartServiceCode(root, startServiceModelTmpl, startServiceTmpl); err != nil {
		return err
	}
	if err := generateStartCommandCode(root, startCommandTmpl, startHandlerTmpl); err != nil {
		return err
	}
	if err := generateListenerCode(root, "order", orderListenerTmpl); err != nil {
		return err
	}
	if err := generateActionListenerCode(root, actionListenerTmpl); err != nil {
		return err
	}
	if err := generateListenerCode(root, "transittoordered", ttorderedListenerTmpl); err != nil {
		return err
	}
	if err := generateListenerCode(root, "transittofinished", ttfinishedListenerTmpl); err != nil {
		return err
	}
	if err := generateOrderServiceCode(root, orderServiceModelTmpl, orderServiceTmpl); err != nil {
		return err
	}
	if err := generateOrderCommandCode(root, orderCommandTmpl, orderHandlerTmpl); err != nil {
		return err
	}
	if err := generateActionCommandCode(root, actionCommandTmpl, actionHandlerTmpl); err != nil {
		return err
	}
	if err := generateConverterCommandCode(root, converterCommandTmpl, converterHandlerTmpl); err != nil {
		return err
	}
	if err := generateConverterCommandTransit(root, "ordered", ttorderedCommandTmpl, ttorderedHandlerTmpl); err != nil {
		return err
	}
	if err := generateConverterCommandTransit(root, "finished", ttfinishedCommandTmpl, ttfinishedHandlerTmpl); err != nil {
		return err
	}
	return nil
}

func executeAndWriteTemplates(tmplMap map[string]*template.Template, data interface{}) error {
	for path, tmpl := range tmplMap {
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(path), err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("executing template for %s: %w", path, err)
		}
		if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("writing file %s: %w", path, err)
		}
		fmt.Printf("Generated: %s\n", path)
	}
	return nil
}

// --- Path builders (new architecture: internal/ hierarchy) ---

func bcInternal(root string) string {
	return filepath.Join(root, strings.ToLower(args.BoundedContext), "internal")
}

func generateStartAdapterCode(root string, tmplAdapter, tmplRequest *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "infra/api/process/start"+strings.ToLower(args.Name))
	genPathAdapter := filepath.Join(genDir, "start_"+toSnakeCase(args.Name)+"_http_adapter.go")
	genPathRequest := filepath.Join(genDir, "start_"+toSnakeCase(args.Name)+"_request.go")

	fmt.Printf("Generating: %s, %s\n", genPathAdapter, genPathRequest)

	templates := map[string]*template.Template{
		genPathAdapter: tmplAdapter,
		genPathRequest: tmplRequest,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateStartServiceCode(root string, tmplModel, tmplService *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "app/service/process/"+strings.ToLower(args.Name), "start")
	genPathModel := filepath.Join(genDir, "start_"+toSnakeCase(args.Name)+"_model.go")
	genPathService := filepath.Join(genDir, "start_"+toSnakeCase(args.Name)+"_service.go")

	fmt.Printf("Generating: %s, %s\n", genPathModel, genPathService)

	templates := map[string]*template.Template{
		genPathModel:   tmplModel,
		genPathService: tmplService,
	}
	return executeAndWriteTemplates(templates, args)
}

func checkDltAdapterCodeAlreadyGenerated(root string) error {
	genDir := filepath.Join(bcInternal(root), "infra/api/dlt/"+strings.ToLower(args.DLTEventName))
	genPathAdapter := filepath.Join(genDir, strings.ToLower(toSnakeCase(args.DLTEventName))+"_http_adapter.go")
	if _, err := os.Stat(genPathAdapter); err == nil {
		fmt.Println("Warn: It appears the generation process has already been executed before. Do you want to overwrite all files? (y/N):")
		reader := bufio.NewReader(os.Stdin)
		resp, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("error reading user confirmation: %w", err)
		}
		resp = strings.TrimSpace(strings.ToLower(resp))
		if resp != "y" && resp != "yes" {
			fmt.Println("Generation aborted by user. Files were not overwritten.")
			return fmt.Errorf("operation canceled by user")
		}
	}
	return nil
}

func generateDltAdapterCode(root string, tmplAdapter *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "infra/api/dlt/"+strings.ToLower(args.DLTEventName))
	genPathAdapter := filepath.Join(genDir, strings.ToLower(toSnakeCase(args.DLTEventName))+"_http_adapter.go")

	fmt.Printf("Generating: %s\n", genPathAdapter)

	templates := map[string]*template.Template{
		genPathAdapter: tmplAdapter,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateStartCommandCode(root string, tmplCommand, tmplHandler *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "app/command/process/"+strings.ToLower(args.Name), "start")
	genPathCommand := filepath.Join(genDir, "start_"+toSnakeCase(args.Name)+"_process_command.go")
	genPathHandler := filepath.Join(genDir, "start_"+toSnakeCase(args.Name)+"_process_handler.go")

	fmt.Printf("Generating: %s, %s\n", genPathCommand, genPathHandler)

	templates := map[string]*template.Template{
		genPathCommand: tmplCommand,
		genPathHandler: tmplHandler,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateListenerCode(root, listener string, tmpl *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "infra/listener/process/"+strings.ToLower(args.Name), listener)
	genPath := filepath.Join(genDir, listener+"_"+toSnakeCase(args.Name)+"_listener.go")

	fmt.Printf("Generating: %s\n", genPath)

	templates := map[string]*template.Template{
		genPath: tmpl,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateActionListenerCode(root string, tmpl *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "infra/listener", strings.ToLower(args.Entity+ToDone(args.Action)))
	genPath := filepath.Join(genDir, strings.ToLower(args.Entity+"_"+ToDone(args.Action))+"_listener.go")

	fmt.Printf("Generating: %s\n", genPath)

	templates := map[string]*template.Template{
		genPath: tmpl,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateOrderServiceCode(root string, tmplModel, tmplService *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "app/service/process/"+strings.ToLower(args.Name), "order")
	genPathModel := filepath.Join(genDir, "order_"+toSnakeCase(args.Name)+"_model.go")
	genPathService := filepath.Join(genDir, "order_"+toSnakeCase(args.Name)+"_service.go")

	fmt.Printf("Generating: %s, %s\n", genPathModel, genPathService)

	templates := map[string]*template.Template{
		genPathModel:   tmplModel,
		genPathService: tmplService,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateOrderCommandCode(root string, tmplCommand, tmplHandler *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "app/command/"+strings.ToLower(args.ExtraPath+args.Entity), strings.ToLower(args.Action), "order")
	genPathCommand := filepath.Join(genDir, "order_"+toSnakeCase(args.Name)+"_command.go")
	genPathHandler := filepath.Join(genDir, "order_"+toSnakeCase(args.Name)+"_handler.go")

	fmt.Printf("Generating: %s, %s\n", genPathCommand, genPathHandler)

	templates := map[string]*template.Template{
		genPathCommand: tmplCommand,
		genPathHandler: tmplHandler,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateActionCommandCode(root string, tmplCommand, tmplHandler *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "app/command/"+strings.ToLower(args.ExtraPath+args.Entity), strings.ToLower(args.Action))
	genPathCommand := filepath.Join(genDir, toSnakeCase(args.Name)+"_command.go")
	genPathHandler := filepath.Join(genDir, toSnakeCase(args.Name)+"_handler.go")

	fmt.Printf("Generating: %s, %s\n", genPathCommand, genPathHandler)

	templates := map[string]*template.Template{
		genPathCommand: tmplCommand,
		genPathHandler: tmplHandler,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateConverterCommandCode(root string, tmplCommand, tmplHandler *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "app/command/surikata/converter/"+strings.ToLower(args.DLTEventName))
	genPathCommand := filepath.Join(genDir, toSnakeCase(args.DLTEventName)+"_converter_command.go")
	genPathHandler := filepath.Join(genDir, toSnakeCase(args.DLTEventName)+"_converter_handler.go")

	fmt.Printf("Generating: %s, %s\n", genPathCommand, genPathHandler)

	templates := map[string]*template.Template{
		genPathCommand: tmplCommand,
		genPathHandler: tmplHandler,
	}
	return executeAndWriteTemplates(templates, args)
}

func generateConverterCommandTransit(root, transit string, tmplCommand, tmplHandler *template.Template) error {
	genDir := filepath.Join(bcInternal(root), "app/command/process/"+strings.ToLower(args.Name), "transitto"+transit)
	genPathCommand := filepath.Join(genDir, "tt"+transit+"_"+toSnakeCase(args.Name)+"_process_command.go")
	genPathHandler := filepath.Join(genDir, "tt"+transit+"_"+toSnakeCase(args.Name)+"_process_handler.go")

	fmt.Printf("Generating: %s, %s\n", genPathCommand, genPathHandler)

	templates := map[string]*template.Template{
		genPathCommand: tmplCommand,
		genPathHandler: tmplHandler,
	}
	return executeAndWriteTemplates(templates, args)
}

// --- Helpers ---

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

func toNatural(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) && i > 0 {
			result = append(result, ' ')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func ToDone(action string) string {
	if len(action) == 0 {
		return ""
	}
	if action[len(action)-1] == 'e' {
		return action + "d"
	}
	if len(action) > 1 && action[len(action)-1] == 'y' {
		c := action[len(action)-2]
		if c != 'a' && c != 'e' && c != 'i' && c != 'o' && c != 'u' {
			return action[:len(action)-1] + "ied"
		}
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
