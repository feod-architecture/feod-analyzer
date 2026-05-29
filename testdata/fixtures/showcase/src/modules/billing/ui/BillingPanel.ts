import { collectPayment } from "@/modules/checkout/payment";
import { orderList } from "@/modules/orders";
import { formatMoney } from "@/common/money";

export const BillingPanel = [collectPayment(), orderList.length, formatMoney(100)];
