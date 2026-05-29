package feod

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/feod-architecture/feod-analyzer/internal/config"
	"github.com/feod-architecture/feod-analyzer/internal/report"
)

func TestAnalyzeValidFixture(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "valid")
	cfg, err := config.Load(root, "")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Analyze(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if result.Summary.Errors != 0 {
		t.Fatalf("expected no errors, got %d: %#v", result.Summary.Errors, result.Violations)
	}
	if result.Summary.Modules != 1 {
		t.Fatalf("expected one module, got %d", result.Summary.Modules)
	}
	if result.Summary.Submodules != 1 {
		t.Fatalf("expected one submodule, got %d", result.Summary.Submodules)
	}

	nodes := map[string]report.Node{}
	for _, node := range result.Nodes {
		nodes[node.ID] = node
	}

	module := nodes["module:checkout"]
	if module.Readme == nil {
		t.Fatalf("expected checkout module README")
	}
	if module.Readme.Path != "modules/checkout/README.md" {
		t.Fatalf("unexpected checkout README path: %s", module.Readme.Path)
	}
	if !strings.Contains(module.Readme.Content, "# Checkout Module") {
		t.Fatalf("expected checkout README content, got %q", module.Readme.Content)
	}

	submodule := nodes["submodule:checkout/payment"]
	if submodule.Readme == nil {
		t.Fatalf("expected checkout payment submodule README")
	}
	if submodule.Readme.Path != "modules/checkout/payment/README.md" {
		t.Fatalf("unexpected payment README path: %s", submodule.Readme.Path)
	}
	if !strings.Contains(submodule.Readme.Content, "Payment submodule") {
		t.Fatalf("expected payment README content, got %q", submodule.Readme.Content)
	}

	if page := nodes["page:checkout"]; page.Readme != nil {
		t.Fatalf("expected page README to be omitted, got %#v", page.Readme)
	}
}

func TestAnalyzeViolationsFixture(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "violations")
	cfg, err := config.Load(root, "")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Analyze(cfg)
	if err != nil {
		t.Fatal(err)
	}

	expectedRules := map[string]bool{
		RuleDirectGlobalImport:      false,
		RuleDeepImport:              false,
		RuleExternalSubmoduleImport: false,
		RulePublicAPILeak:           false,
		RuleMissingPublicAPI:        false,
		RuleCycle:                   false,
	}
	for _, violation := range result.Violations {
		if _, ok := expectedRules[violation.Rule]; ok {
			expectedRules[violation.Rule] = true
		}
	}
	for rule, found := range expectedRules {
		if !found {
			t.Fatalf("expected rule %s in violations: %#v", rule, result.Violations)
		}
	}
}

func TestAnalyzeComplexFixture(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "complex")
	cfg, err := config.Load(root, "")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Analyze(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if result.Summary.Violations != 0 {
		t.Fatalf("expected no violations, got %#v", result.Violations)
	}
	if result.Summary.Modules < 6 {
		t.Fatalf("expected complex module graph, got %d modules", result.Summary.Modules)
	}
	if result.Summary.Submodules != 2 {
		t.Fatalf("expected checkout delivery/payment submodules, got %d", result.Summary.Submodules)
	}
}

func TestAnalyzeShowcaseFixture(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "showcase")
	cfg, err := config.Load(root, "")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Analyze(cfg)
	if err != nil {
		t.Fatal(err)
	}

	expectedRules := map[string]bool{
		RuleForbiddenLevelDependency: false,
		RuleDeepImport:               false,
		RuleExternalSubmoduleImport:  false,
		RuleDirectGlobalImport:       false,
		RulePublicAPILeak:            false,
		RuleMissingPublicAPI:         false,
		RuleCycle:                    false,
	}
	for _, violation := range result.Violations {
		if _, ok := expectedRules[violation.Rule]; ok {
			expectedRules[violation.Rule] = true
		}
	}
	for rule, found := range expectedRules {
		if !found {
			t.Fatalf("expected rule %s in showcase violations: %#v", rule, result.Violations)
		}
	}
	if result.Summary.Errors == 0 || result.Summary.Warnings == 0 {
		t.Fatalf("expected mixed errors and warnings, got %d errors and %d warnings", result.Summary.Errors, result.Summary.Warnings)
	}
	if result.Summary.Modules < 10 {
		t.Fatalf("expected broad module graph, got %d modules", result.Summary.Modules)
	}
	if result.Summary.Submodules != 2 {
		t.Fatalf("expected checkout payment/delivery submodules, got %d", result.Summary.Submodules)
	}
}

func TestSubmoduleImportFromOutsideIsError(t *testing.T) {
	root := filepath.Join("..", "..", "testdata", "fixtures", "violations")
	cfg, err := config.Load(root, "")
	if err != nil {
		t.Fatal(err)
	}

	result, err := Analyze(cfg)
	if err != nil {
		t.Fatal(err)
	}

	for _, violation := range result.Violations {
		if violation.Rule == RuleExternalSubmoduleImport && violation.ImportPath == "@/modules/user/permissions" {
			return
		}
	}
	t.Fatalf("expected external submodule import violation, got %#v", result.Violations)
}
