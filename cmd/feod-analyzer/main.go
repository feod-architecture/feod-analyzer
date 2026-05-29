package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/feod-architecture/feod-analyzer/internal/config"
	"github.com/feod-architecture/feod-analyzer/internal/exporter"
	"github.com/feod-architecture/feod-analyzer/internal/feod"
	"github.com/feod-architecture/feod-analyzer/internal/report"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		printUsage()
		return 0
	}

	switch args[0] {
	case "analyze":
		return runAnalyze(args[1:])
	case "version":
		fmt.Println(feod.Version)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		printUsage()
		return 2
	}
}

func runAnalyze(args []string) int {
	args = normalizeAnalyzeArgs(args)

	flags := flag.NewFlagSet("analyze", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	configPath := flags.String("config", "", "Path to feod-analyzer.yml")
	outDir := flags.String("out", "", "Output directory")
	formatsValue := flags.String("formats", "", "Comma-separated formats: html,json")
	serve := flags.Bool("serve", false, "Serve generated HTML report")
	port := flags.Int("port", 3123, "Report server port")
	failOn := flags.String("fail-on", "never", "Fail policy: error, warning, never")

	if err := flags.Parse(args); err != nil {
		return 2
	}

	root := "."
	if flags.NArg() > 0 {
		root = flags.Arg(0)
	}

	cfg, err := config.Load(root, *configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		return 2
	}
	if *outDir != "" {
		cfg.OutputDir = *outDir
	}

	formats := cfg.OutputFormats
	if *formatsValue != "" {
		formats = parseFormats(*formatsValue)
	}
	formats, err = validateFormats(formats)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid --formats value: %v\n", err)
		return 2
	}

	result, err := feod.Analyze(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Analyze error: %v\n", err)
		return 2
	}

	if err := exporter.Export(result, cfg, formats); err != nil {
		fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
		return 2
	}

	printSummary(result, cfg.OutputAbs())

	if *serve {
		if err := exporter.Serve(cfg.OutputAbs(), *port); err != nil {
			fmt.Fprintf(os.Stderr, "Serve error: %v\n", err)
			return 2
		}
	}

	switch *failOn {
	case "never":
		return 0
	case "warning":
		if result.Summary.Errors > 0 || result.Summary.Warnings > 0 {
			return 1
		}
		return 0
	case "error":
		if result.Summary.Errors > 0 {
			return 1
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Invalid --fail-on value: %s\n", *failOn)
		return 2
	}
}

func normalizeAnalyzeArgs(args []string) []string {
	valueFlags := map[string]bool{
		"--config":  true,
		"--out":     true,
		"--formats": true,
		"--port":    true,
		"--fail-on": true,
	}
	flags := []string{}
	positionals := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positionals = append(positionals, arg)
			continue
		}

		flags = append(flags, arg)
		name := arg
		if equal := strings.Index(arg, "="); equal >= 0 {
			name = arg[:equal]
		}
		if valueFlags[name] && !strings.Contains(arg, "=") && i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}

	return append(flags, positionals...)
}

func parseFormats(value string) []string {
	parts := strings.Split(value, ",")
	formats := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			formats = append(formats, part)
		}
	}
	return formats
}

func validateFormats(formats []string) ([]string, error) {
	allowed := map[string]bool{
		"html": true,
		"json": true,
	}
	seen := map[string]bool{}
	normalized := make([]string, 0, len(formats))
	for _, format := range formats {
		format = strings.ToLower(strings.TrimSpace(format))
		if format == "" {
			continue
		}
		if !allowed[format] {
			return nil, fmt.Errorf("%s (allowed: html,json)", format)
		}
		if !seen[format] {
			seen[format] = true
			normalized = append(normalized, format)
		}
	}
	if len(normalized) == 0 {
		return []string{"html", "json"}, nil
	}
	return normalized, nil
}

func printSummary(result *report.Report, outputDir string) {
	fmt.Printf("FEOD Analyzer: %d files, %d nodes, %d edges\n", result.Summary.Files, result.Summary.Nodes, result.Summary.Edges)
	fmt.Printf("Violations: %d errors, %d warnings\n", result.Summary.Errors, result.Summary.Warnings)
	fmt.Printf("Output: %s\n", outputDir)
}

func printUsage() {
	fmt.Println(`FEOD Analyzer

Usage:
  feod-analyzer analyze [path] [flags]
  feod-analyzer version

Flags:
  --config <file>          Path to feod-analyzer.yml
  --out <dir>              Output directory
  --formats html,json      Output formats
  --serve                  Serve generated HTML report
  --port 3123              Report server port
  --fail-on never          error | warning | never`)
}
