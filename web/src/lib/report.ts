import type { FeodReport } from "../types";

export async function loadReport(): Promise<FeodReport> {
  const response = await fetch("./feod-report.json", { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`HTTP ${response.status}`);
  }
  return response.json() as Promise<FeodReport>;
}
