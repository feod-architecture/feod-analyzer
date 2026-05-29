import { formatMoney } from "@/common/money";
import type { ProductId } from "../model/types";

export function ProductGrid() {
  const id: ProductId = "sku-1";
  return `product:${id}:${formatMoney(1200)}`;
}
