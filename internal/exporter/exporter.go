package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/feod-architecture/feod-analyzer/internal/config"
	"github.com/feod-architecture/feod-analyzer/internal/report"
)

const ReportJSONName = "feod-report.json"

func Export(result *report.Report, cfg *config.Config, formats []string) error {
	outputDir := cfg.OutputAbs()
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	needsJSON := contains(formats, "json") || contains(formats, "html")
	if needsJSON {
		if err := ExportJSON(result, outputDir); err != nil {
			return err
		}
	}

	if contains(formats, "html") {
		if err := ExportHTML(outputDir); err != nil {
			return err
		}
	}

	return nil
}

func ExportJSON(result *report.Report, outputDir string) error {
	path := filepath.Join(outputDir, ReportJSONName)
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create JSON report: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("write JSON report: %w", err)
	}
	return nil
}

func ExportHTML(outputDir string) error {
	webDist := findWebDist()
	if webDist == "" {
		return writeFallbackHTML(outputDir)
	}
	return copyDir(webDist, outputDir)
}

func Serve(outputDir string, port int) error {
	if port == 0 {
		port = 3123
	}

	server := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           http.FileServer(http.Dir(outputDir)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/", port)
	go func() {
		_ = openBrowser(url)
	}()

	fmt.Printf("FEOD Analyzer report: %s\n", url)
	return server.ListenAndServe()
}

func findWebDist() string {
	candidates := []string{}
	if env := os.Getenv("FEOD_ANALYZER_WEB_DIST"); env != "" {
		candidates = append(candidates, env)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(cwd, "web", "dist"))
	}
	if executable, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(executableDir, "..", "web", "dist"),
			filepath.Join(executableDir, "web", "dist"),
		)
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if entry.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		return copyFile(path, target)
	})
}

func copyFile(src string, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	return err
}

func writeFallbackHTML(outputDir string) error {
	html := `<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>FEOD Analyzer</title>
  <style>
    body{font-family:ui-sans-serif,system-ui,sans-serif;margin:0;background:#f8fafc;color:#0f172a}
    main{max-width:960px;margin:48px auto;padding:24px;background:white;border:1px solid #e2e8f0;border-radius:8px}
    code{background:#f1f5f9;padding:2px 6px;border-radius:4px}
  </style>
</head>
<body>
  <main>
    <h1>FEOD Analyzer</h1>
    <p>React report assets are not built. Run <code>bun run build:web</code> and execute the analyzer again.</p>
    <p>JSON report: <a href="./feod-report.json">feod-report.json</a></p>
  </main>
</body>
</html>`
	return os.WriteFile(filepath.Join(outputDir, "index.html"), []byte(html), 0644)
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item), value) {
			return true
		}
	}
	return false
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
