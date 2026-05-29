import { formatDate } from "@/common/date";
import { formatMoney } from "@/common/money";
import { DataTable } from "@/common/ui";
import { getOrders } from "../model/orders";

export function OrdersTable() {
  return DataTable(getOrders().map((order) => `${order.id}:${formatDate(order.date)}:${formatMoney(order.total)}`));
}
