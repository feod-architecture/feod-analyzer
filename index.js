#!/usr/bin/env node

import { spawnSync } from "node:child_process";
import { existsSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const packageDir = dirname(fileURLToPath(import.meta.url));
const binaryName = process.platform === "win32" ? "feod-analyzer.exe" : "feod-analyzer";
const candidates = [
  join(packageDir, "bin", binaryName),
  join(packageDir, "bin", "feod-analyzer"),
];

const binary = candidates.find((candidate) => existsSync(candidate));

if (!binary) {
  console.error("feod-analyzer binary not found. Run `bun run build:cli` first.");
  process.exit(2);
}

const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
  env: {
    ...process.env,
    FEOD_ANALYZER_WEB_DIST: process.env.FEOD_ANALYZER_WEB_DIST || join(packageDir, "web", "dist"),
  },
});

if (result.error) {
  console.error(result.error.message);
  process.exit(2);
}

process.exit(result.status ?? 0);
