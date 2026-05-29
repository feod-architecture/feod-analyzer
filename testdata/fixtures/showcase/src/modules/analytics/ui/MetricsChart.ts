import { buildInternalDashboard } from "../model/internal";
import { DataTable } from "@/common/ui";

export const MetricsChart = DataTable([buildInternalDashboard()]);
