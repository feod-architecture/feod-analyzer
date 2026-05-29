import { MetricsChart } from "@/modules/analytics";
import { OrdersTable } from "@/modules/orders";
import { formatDate } from "@/common/date";
import { PageShell } from "@/common/ui";

export const ReportsPage = PageShell("Reports", [MetricsChart, OrdersTable, formatDate(new Date())]);
