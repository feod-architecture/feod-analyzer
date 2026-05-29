import { formatDate } from "@/common/date";
import { DataTable } from "@/common/ui";
import { getAdminUsers } from "../model/users";

export function UsersTable() {
  return DataTable(getAdminUsers().map((user) => `${user}:${formatDate("2026-05-29")}`));
}
