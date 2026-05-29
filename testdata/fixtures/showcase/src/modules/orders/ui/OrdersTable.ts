import { formatDate } from "@/common/date";
import { formatMoney } from "@/common/money";
import { DataTable } from "@/common/ui";
import { orderList } from "../model/orders";

export const OrdersTable = DataTable(orderList.map((order) => ({ ...order, date: formatDate(new Date()), total: formatMoney(30) })));
