import { Button } from "@/common/ui";
import type { ProductId } from "@/modules/catalog";

export function AddToCartButton(productId: ProductId) {
  return Button({ children: `add:${productId}` });
}
