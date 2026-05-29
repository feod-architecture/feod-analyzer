import { orderList } from "@/modules/orders";
import { formatDate } from "@/common/date";

export function buildInternalDashboard() {
  return { totalOrders: orderList.length, generatedAt: formatDate(new Date()) };
}
