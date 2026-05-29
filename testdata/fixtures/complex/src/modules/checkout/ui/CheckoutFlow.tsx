import { getCartTotal } from "@/modules/cart";
import { useViewer } from "@/modules/viewer";
import { DeliveryStep } from "../delivery";
import { PaymentStep } from "../payment";

export function CheckoutFlow() {
  return [useViewer(), getCartTotal(), DeliveryStep(), PaymentStep()].join(":");
}
