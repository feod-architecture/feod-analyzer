import { CartSummary } from "@/modules/cart";
import { CheckoutFlow } from "@/modules/checkout";
import { ViewerMenu } from "@/modules/viewer";
import { PageShell } from "@/common/ui";

export function CheckoutPage() {
  return PageShell({
    children: [ViewerMenu(), CartSummary(), CheckoutFlow()],
  });
}
