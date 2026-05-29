import type { Violation } from "../types";

export type Locale = "ru" | "en";
export type CountUnit = "edge" | "error" | "file" | "import" | "issue" | "node" | "warning";

export const defaultLocale: Locale = "ru";
export const localeStorageKey = "feod-report-locale";
export const localeOptions = [
  { value: "ru", label: "RU" },
  { value: "en", label: "EN" },
] as const satisfies readonly { value: Locale; label: string }[];

const ru = {
  closeInspector: "Закрыть информацию о блоке",
  dependencies: "Зависимости",
  dependencyGraph: "Граф зависимостей",
  dependents: "Зависимые",
  graphScrollToIssues: "Прокрутить к ошибкам и предупреждениям",
  graphScrollToTop: "Прокрутить наверх",
  incomingShort: "вход",
  issuesTitle: "Ошибки и предупреждения",
  languageSwitcher: "Выбор языка отчёта",
  loadingDescription: "Загружаю feod-report.json...",
  nodeInspectorLabel: "Информация о блоке",
  missingReportAction: "Сгенерируйте отчёт командой feod-analyzer analyze --formats html,json",
  missingReportDescription: "JSON-отчёт не найден.",
  noIncomingDependencies: "Нет входящих зависимостей.",
  noOutgoingDependencies: "Нет исходящих зависимостей.",
  noSubmodules: "Сабмодули не найдены.",
  noViolations: "Нарушений не найдено.",
  outgoingShort: "выход",
  page: "Страница",
  paginationLabel: "Пагинация нарушений",
  parentModule: "Родительский модуль",
  readmeMissing: "README.md не найден.",
  reportLoadFailed: "Не удалось загрузить feod-report.json",
  reportOpenErrorTitle: "Не удалось открыть отчёт",
  scrollTopTitle: "Наверх",
  scrollToIssuesTitle: "К ошибкам и предупреждениям",
  submodules: "Сабмодули",
  next: "Вперёд",
  previous: "Назад",
} as const;

const en: Record<keyof typeof ru, string> = {
  closeInspector: "Close block details",
  dependencies: "Dependencies",
  dependencyGraph: "Dependency graph",
  dependents: "Dependents",
  graphScrollToIssues: "Scroll to errors and warnings",
  graphScrollToTop: "Scroll to top",
  incomingShort: "in",
  issuesTitle: "Errors and warnings",
  languageSwitcher: "Report language",
  loadingDescription: "Loading feod-report.json...",
  nodeInspectorLabel: "Block details",
  missingReportAction: "Generate a report with feod-analyzer analyze --formats html,json",
  missingReportDescription: "JSON report is missing.",
  noIncomingDependencies: "No incoming dependencies.",
  noOutgoingDependencies: "No outgoing dependencies.",
  noSubmodules: "No submodules found.",
  noViolations: "No violations found.",
  outgoingShort: "out",
  page: "Page",
  paginationLabel: "Violations pagination",
  parentModule: "Parent module",
  readmeMissing: "README.md was not found.",
  reportLoadFailed: "Unable to load feod-report.json",
  reportOpenErrorTitle: "Unable to open report",
  scrollTopTitle: "Top",
  scrollToIssuesTitle: "To errors and warnings",
  submodules: "Submodules",
  next: "Next",
  previous: "Prev",
};

const messages = { ru, en } as const;

const countUnits: Record<Locale, Record<CountUnit, Partial<Record<Intl.LDMLPluralRule, string>> & { other: string }>> = {
  ru: {
    edge: { one: "связь", few: "связи", many: "связей", other: "связи" },
    error: { one: "ошибка", few: "ошибки", many: "ошибок", other: "ошибки" },
    file: { one: "файл", few: "файла", many: "файлов", other: "файла" },
    import: { one: "импорт", few: "импорта", many: "импортов", other: "импорта" },
    issue: { one: "проблема", few: "проблемы", many: "проблем", other: "проблемы" },
    node: { one: "узел", few: "узла", many: "узлов", other: "узла" },
    warning: { one: "предупреждение", few: "предупреждения", many: "предупреждений", other: "предупреждения" },
  },
  en: {
    edge: { one: "edge", other: "edges" },
    error: { one: "error", other: "errors" },
    file: { one: "file", other: "files" },
    import: { one: "import", other: "imports" },
    issue: { one: "issue", other: "issues" },
    node: { one: "node", other: "nodes" },
    warning: { one: "warning", other: "warnings" },
  },
};

const localeTags: Record<Locale, string> = {
  ru: "ru-RU",
  en: "en-US",
};

const pluralRules = {
  ru: new Intl.PluralRules(localeTags.ru),
  en: new Intl.PluralRules(localeTags.en),
};

const numberFormatters = {
  ru: new Intl.NumberFormat(localeTags.ru),
  en: new Intl.NumberFormat(localeTags.en),
};

const severityLabels: Record<Locale, Record<string, string>> = {
  ru: {
    error: "ошибка",
    info: "инфо",
    warning: "предупреждение",
  },
  en: {
    error: "error",
    info: "info",
    warning: "warning",
  },
};

const statusLabels: Record<Locale, Record<string, string>> = {
  ru: {
    allowed: "разрешено",
    error: "ошибка",
    warning: "предупреждение",
  },
  en: {
    allowed: "allowed",
    error: "error",
    warning: "warning",
  },
};

const nodeKindLabels: Record<Locale, Record<string, string>> = {
  ru: {
    commonEntity: "common",
    file: "файл",
    level: "уровень",
    module: "модуль",
    page: "страница",
    submodule: "сабмодуль",
  },
  en: {
    commonEntity: "common",
    file: "file",
    level: "level",
    module: "module",
    page: "page",
    submodule: "submodule",
  },
};

export function isLocale(value: string | null | undefined): value is Locale {
  return value === "ru" || value === "en";
}

export function getMessages(locale: Locale) {
  return messages[locale];
}

export function readStoredLocale(): Locale {
  if (typeof window === "undefined") {
    return defaultLocale;
  }
  try {
    const value = window.localStorage.getItem(localeStorageKey);
    return isLocale(value) ? value : defaultLocale;
  } catch {
    return defaultLocale;
  }
}

export function writeStoredLocale(locale: Locale) {
  if (typeof window === "undefined") {
    return;
  }
  try {
    window.localStorage.setItem(localeStorageKey, locale);
  } catch {
    return;
  }
}

export function formatCount(locale: Locale, value: number, unit: CountUnit) {
  const category = pluralRules[locale].select(value);
  const forms = countUnits[locale][unit];
  return `${numberFormatters[locale].format(value)} ${forms[category] ?? forms.other}`;
}

export function formatMoreCount(locale: Locale, value: number, unit: CountUnit) {
  if (locale === "ru") {
    return `+ещё ${formatCount(locale, value, unit)}`;
  }
  return `+${formatCount(locale, value, unit)} more`;
}

export function formatDate(value: string, locale: Locale) {
  return new Intl.DateTimeFormat(localeTags[locale], {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function severityLabel(locale: Locale, value: string) {
  return severityLabels[locale][value] ?? value;
}

export function statusLabel(locale: Locale, value: string) {
  return statusLabels[locale][value] ?? value;
}

export function nodeKindLabel(locale: Locale, value: string) {
  return nodeKindLabels[locale][value] ?? value;
}

export function violationMessage(locale: Locale, violation: Violation) {
  if (locale === "ru") {
    return violation.message;
  }

  switch (violation.rule) {
    case "cycle":
      return `Dependency cycle detected between ${entityRef(violation.from)} and ${entityRef(violation.to)}.`;
    case "deep-import":
      return "The import bypasses the FEOD entity public API and depends on its internal structure.";
    case "direct-global-import":
      return "Application FEOD code must not import global directly.";
    case "external-submodule-import":
      return "External code imports a submodule directly, bypassing the parent module public API.";
    case "forbidden-level-dependency":
      return `Forbidden FEOD dependency: ${entityRef(violation.from)} imports ${entityRef(violation.to)}.`;
    case "missing-public-api":
      return `${violation.file ?? entityRef(violation.from)} does not have an explicit root index.ts public API.`;
    case "public-api-leak":
      return "Public API uses export * from internal structure.";
    case "read-file":
      return `Unable to read file${readFileReason(violation.message)}`;
    default:
      return violation.message;
  }
}

export function violationSuggestion(locale: Locale, violation: Violation) {
  if (!violation.suggestion || locale === "ru") {
    return violation.suggestion;
  }

  switch (violation.rule) {
    case "cycle":
      return "Break the cycle through a public API, move shared code to common, or revisit module responsibilities.";
    case "deep-import":
      return `Use the public API: "${publicImportPath(violation.to)}".`;
    case "direct-global-import":
      return "Connect global through an entrypoint, runtime bootstrap, or bundler configuration.";
    case "external-submodule-import":
      return `Export the required contract from ${ownerIndexPath(violation.to)} and import from "${ownerImportPath(violation.to)}".`;
    case "forbidden-level-dependency":
      return "Check the FEOD import matrix and move the contract to an allowed level.";
    case "missing-public-api":
      return "Add a root index.ts and export only the supported contract.";
    case "public-api-leak":
      return "Replace export * with an explicit list of supported exports.";
    default:
      return violation.suggestion;
  }
}

function entityRef(value?: string) {
  return value || "unknown entity";
}

function readFileReason(message: string) {
  const separator = message.indexOf(":");
  return separator >= 0 ? `: ${message.slice(separator + 1).trim()}` : ".";
}

function ownerIndexPath(nodeId?: string) {
  return `${ownerPath(nodeId)}/index.ts`;
}

function ownerImportPath(nodeId?: string) {
  return `@/${ownerPath(nodeId)}`;
}

function ownerPath(nodeId?: string) {
  const parsed = parseNodeId(nodeId);
  if (parsed.kind === "submodule") {
    const [moduleName] = parsed.name.split("/");
    return `modules/${moduleName || parsed.name}`;
  }
  return publicPath(nodeId);
}

function publicImportPath(nodeId?: string) {
  return `@/${publicPath(nodeId)}`;
}

function publicPath(nodeId?: string) {
  const parsed = parseNodeId(nodeId);
  switch (parsed.kind) {
    case "common":
      return `common/${parsed.name}`;
    case "module":
      return `modules/${parsed.name}`;
    case "page":
      return `pages/${parsed.name}`;
    case "submodule":
      return `modules/${parsed.name}`;
    default:
      return parsed.name || entityRef(nodeId);
  }
}

function parseNodeId(nodeId?: string) {
  const value = nodeId ?? "";
  const separator = value.indexOf(":");
  if (separator < 0) {
    return { kind: "", name: value };
  }
  return {
    kind: value.slice(0, separator),
    name: value.slice(separator + 1),
  };
}
