import { PaymentStep } from "../payment";
import { Button } from "@/common/button";

export function CheckoutFlow() {
  return Button({ children: PaymentStep() });
}
