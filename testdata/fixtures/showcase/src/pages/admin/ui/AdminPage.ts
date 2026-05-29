import { StockTable } from "@/modules/inventory";
import { OrdersTable } from "@/modules/orders";
import { LoginButton } from "@/modules/auth";
import { PageShell } from "@/common/ui";

export const AdminPage = PageShell("Admin", [StockTable, OrdersTable, LoginButton]);
