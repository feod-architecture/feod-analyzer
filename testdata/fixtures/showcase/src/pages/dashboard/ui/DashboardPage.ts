import { MetricsChart } from "@/modules/analytics";
import { NotificationBell } from "@/modules/notifications";
import { PageShell } from "@/common/ui";
import { buildInternalDashboard } from "@/modules/analytics/model/internal";

export const DashboardPage = PageShell("Dashboard", [MetricsChart, NotificationBell, buildInternalDashboard()]);
