export type Severity = "info" | "warning" | "error";
export type EdgeStatus = "allowed" | "warning" | "error";
export type NodeKind = "level" | "module" | "submodule" | "commonEntity" | "page" | "file";

export type PublicAPIInfo = {
  hasIndex: boolean;
  indexPath?: string;
  status: string;
  exports?: string[];
  starExports?: string[];
  exposedSubmodules?: string[];
};

export type ReadmeInfo = {
  path: string;
  content: string;
};

export type ReportNode = {
  id: string;
  kind: NodeKind;
  name: string;
  level: string;
  path: string;
  parentId?: string;
  publicApi: PublicAPIInfo;
  fileCount: number;
  readme?: ReadmeInfo;
};

export type ImportUsage = {
  file: string;
  line: number;
  importPath: string;
  resolvedPath?: string;
  kind: string;
  typeOnly: boolean;
};

export type ReportEdge = {
  id: string;
  source: string;
  target: string;
  imports: ImportUsage[];
  status: EdgeStatus;
  ruleIds?: string[];
};

export type Violation = {
  rule: string;
  severity: Severity;
  file?: string;
  line?: number;
  from?: string;
  to?: string;
  importPath?: string;
  message: string;
  suggestion?: string;
};

export type FileReport = {
  path: string;
  nodeId?: string;
  imports?: ImportUsage[];
};

export type ReportSummary = {
  files: number;
  nodes: number;
  edges: number;
  errors: number;
  warnings: number;
  infos: number;
  violations: number;
  modules: number;
  submodules: number;
  pages: number;
  commonItems: number;
};

export type ReportMeta = {
  tool: string;
  version: string;
  rootDir: string;
  srcDir: string;
  generated: string;
  schema: string;
  durationMs: number;
};

export type FeodReport = {
  meta: ReportMeta;
  summary: ReportSummary;
  nodes: ReportNode[];
  edges: ReportEdge[];
  violations: Violation[];
  files: FileReport[];
};
