#!/usr/bin/env node

import { spawnSync } from "node:child_process";

const script = process.argv[2];

if (!script) {
  console.error("Usage: node scripts/run-web-script.mjs <script>");
  process.exit(2);
}

const userAgent = process.env.npm_config_user_agent || "";
const execPath = process.env.npm_execpath || "";
const packageManager = resolvePackageManager(userAgent, execPath);
const result = spawnSync(packageManager.command, [...packageManager.args, script], {
  stdio: "inherit",
  shell: process.platform === "win32",
});

if (result.error) {
  console.error(result.error.message);
  process.exit(2);
}

process.exit(result.status ?? 0);

function resolvePackageManager(userAgent, execPath) {
  if (userAgent.startsWith("pnpm/") || execPath.includes("pnpm")) {
    return { command: "pnpm", args: ["--dir", "web", "run"] };
  }

  if (userAgent.startsWith("bun/") || execPath.includes("bun")) {
    return { command: "bun", args: ["run", "--cwd", "web"] };
  }

  return { command: "npm", args: ["--prefix", "web", "run"] };
}
