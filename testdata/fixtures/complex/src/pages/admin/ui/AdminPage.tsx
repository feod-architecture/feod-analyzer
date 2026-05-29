import { UsersTable } from "@/modules/admin-users";
import { OrdersTable } from "@/modules/orders";
import { ViewerMenu } from "@/modules/viewer";
import { DataTable, PageShell } from "@/common/ui";

export function AdminPage() {
  return PageShell({
    children: [ViewerMenu(), DataTable([UsersTable(), OrdersTable()])],
  });
}
