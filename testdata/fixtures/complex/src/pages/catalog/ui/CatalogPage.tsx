import { ProductGrid } from "@/modules/catalog";
import { AddToCartButton, CartBadge } from "@/modules/cart";
import { PageShell } from "@/common/ui";

export function CatalogPage() {
  return PageShell({
    children: [ProductGrid(), AddToCartButton("sku-1"), CartBadge()],
  });
}
