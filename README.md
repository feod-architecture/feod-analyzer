# FEOD Analyzer

CLI-анализатор FEOD-архитектуры с JSON-отчётом и статическим React/shadcn-style HTML report.

## Быстрый запуск

```bash
bun install
bun run build
./bin/feod-analyzer analyze ./testdata/fixtures/violations --out ./dist --formats html,json
```

Более сложный корректный пример:

```bash
./bin/feod-analyzer analyze ./testdata/fixtures/complex --out ./dist/complex-report --formats html,json --serve --fail-on error
```

Showcase-пример для визуальной отладки графа, где есть и корректные, и проблемные модули:

```bash
./bin/feod-analyzer analyze ./testdata/fixtures/showcase --out ./dist/showcase-report --formats html,json --serve --fail-on never
```

## CLI

```bash
feod-analyzer analyze [path] \
  --config feod-analyzer.yml \
  --out ./dist \
  --formats html,json \
  --serve \
  --port 3123 \
  --fail-on error
```

Exit codes:

- `0` - анализ успешен;
- `1` - найдены нарушения, соответствующие `--fail-on`;
- `2` - ошибка конфигурации, чтения проекта или экспорта.

## Конфигурация

Поддерживаемые имена файлов:

- `feod-analyzer.yml`
- `feod-analyzer.yaml`
- `.feod-analyzer.yml`
- `.feod-analyzer.yaml`

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

## FEOD rules

Анализатор проверяет:

- запрещённые зависимости между FEOD-уровнями;
- deep imports во внутренности чужих модулей, страниц и common-сущностей;
- внешние импорты подмодулей вроде `@/modules/checkout/payment`;
- прямой импорт `global`;
- отсутствие root `index.ts`;
- `export *` leaks в public API;
- циклы между FEOD-сущностями.
