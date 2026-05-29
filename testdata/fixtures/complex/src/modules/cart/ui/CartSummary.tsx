import { formatMoney } from "@/common/money";
import { getCartTotal } from "../model/cart";

export function CartSummary() {
  return formatMoney(getCartTotal());
}
