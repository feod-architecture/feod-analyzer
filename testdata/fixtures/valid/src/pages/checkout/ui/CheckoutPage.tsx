import { CheckoutFlow } from "@/modules/checkout";
import { Button } from "@/common/button";

export function CheckoutPage() {
  return Button({ children: CheckoutFlow() });
}
