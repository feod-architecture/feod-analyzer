# FEOD Analyzer

FEOD Analyzer is a CLI for checking FEOD architecture in frontend projects. It scans source imports, classifies FEOD entities, reports architectural violations, and exports both machine-readable JSON and a static HTML report.

The tool supports the canonical FEOD levels: `app`, `pages`, `modules`, `common`, and `global`.

## Installation

Use the package as a CLI:

```bash
bunx @feod-architecture/analyzer analyze ./src --out ./dist/feod --formats html,json
npx @feod-architecture/analyzer analyze ./src --out ./dist/feod --formats html,json
pnpm dlx @feod-architecture/analyzer analyze ./src --out ./dist/feod --formats html,json
```

For local development from the repository:

```bash
bun install
bun run build
npm install
npm run build
pnpm install
pnpm run build
./bin/feod-analyzer analyze ./testdata/fixtures/showcase --out ./dist/showcase-report --formats html,json
```

## CLI

```bash
feod-analyzer analyze [path] \
  --config feod-analyzer.yml \
  --out ./dist/feod \
  --formats html,json \
  --serve \
  --port 3123 \
  --fail-on error
```

Options:

| Option | Default | Description |
| --- | --- | --- |
| `[path]` | `.` | Project root to analyze. |
| `--config` | auto-discovery | Path to a YAML config file. |
| `--out` | config value or `dist/feod` | Output directory for generated reports. |
| `--formats` | `html,json` | Comma-separated output formats. Supported values: `html`, `json`. |
| `--serve` | `false` | Serve the generated report after analysis. |
| `--port` | `3123` | Local report server port. |
| `--fail-on` | `error` | Exit with code `1` on `error`, `warning`, or `never`. |

Exit codes:

- `0` - analysis completed and did not hit the selected `--fail-on` threshold.
- `1` - violations matched the selected `--fail-on` threshold.
- `2` - configuration, project reading, analysis, or export error.

## Configuration

Supported config filenames:

- `feod-analyzer.yml`
- `feod-analyzer.yaml`
- `.feod-analyzer.yml`
- `.feod-analyzer.yaml`

Example:

```yaml
srcDir: src
outputDir: dist/feod
outputFormats: [html, json]
excludeDirs: [node_modules, .git, dist, build, coverage]
aliases:
  "@": src
tsconfig: tsconfig.json
levels: [app, pages, modules, common, global]
segments: [ui, model, api, lib, config, types, test, tests]
submodules:
  enabled: true
  maxDepth: 2
ignoreRules: []
```

`aliases` maps import aliases to source-relative directories. `segments` marks internal folders that should not be treated as FEOD entities. `submodules.maxDepth` controls how many nested module folders can be treated as submodules before the remaining path is considered internal structure.

## Reports

The JSON report contains:

- `meta` - tool version, analyzed root, source directory, schema version, and duration.
- `summary` - files, nodes, edges, violations, modules, submodules, pages, and common entities.
- `nodes` - FEOD levels, modules, submodules, pages, common entities, and files.
- `edges` - import relationships with status and source import usages.
- `violations` - rule id, severity, file, line, message, and suggestion.
- `files` - source files and extracted imports.

The HTML report is a static React report. It can be generated into any output directory and opened directly or served with `--serve`.

## Checked Rules

The analyzer checks:

- forbidden dependencies between FEOD levels;
- deep imports into another module, page, or common entity internals;
- external imports of submodules such as `@/modules/checkout/payment`;
- direct imports from `global`;
- missing root `index.ts` public API files;
- `export *` leaks in public API files;
- cycles between FEOD entities.

The analyzer supports code review and CI. It does not replace architectural decisions: entity boundaries, public API size, and justified exceptions still need human review.

## Development

The repository supports Bun, npm, and pnpm for local build and test commands. Go 1.22 and Node.js 20 or newer are required.

```bash
bun install
bun run test
bun run build
npm install
npm run test
npm run build
pnpm install
pnpm run test
pnpm run build
```

Useful commands:

```bash
bun run test:go
bun run test:web
bun run build:cli
bun run build:web
npm run test:go
npm run test:web
npm run build:cli
npm run build:web
pnpm run test:go
pnpm run test:web
pnpm run build:cli
pnpm run build:web
```

Smoke test:

```bash
./bin/feod-analyzer analyze ./testdata/fixtures/showcase --out ./dist/showcase-report --formats html,json --fail-on never
```

Package dry run:

```bash
bun pm pack --dry-run
npm pack --dry-run
pnpm pack --dry-run
```

The package includes the Node launcher, the compiled Go binary, the static web report, README, and LICENSE.

## Release

Before publishing a release, use one package manager consistently for the command sequence:

1. Run `bun run test`, `npm run test`, or `pnpm run test`.
2. Run `bun run build`, `npm run build`, or `pnpm run build`.
3. Run the smoke test against `testdata/fixtures/showcase`.
4. Run `bun pm pack --dry-run`, `npm pack --dry-run`, or `pnpm pack --dry-run` and check the file list.
5. Publish only after the package contains `index.js`, `bin/feod-analyzer`, `web/dist`, `README.md`, and `LICENSE`.

## License

MIT.
