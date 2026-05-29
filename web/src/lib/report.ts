import type { FeodReport } from "../types";

export async function loadReport(): Promise<FeodReport> {
  const response = await fetch("./feod-report.json", { cache: "no-store" });
  if (!response.ok) {
    throw new Error(`Unable to load feod-report.json: ${response.status}`);
  }
  return response.json() as Promise<FeodReport>;
}

export function formatDate(value: string) {
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}

export function severityLabel(value: string) {
  switch (value) {
    case "error":
      return "error";
    case "warning":
      return "warning";
    default:
      return "info";
  }
}
