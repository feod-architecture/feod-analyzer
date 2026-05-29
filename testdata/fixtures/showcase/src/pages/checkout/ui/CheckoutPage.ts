import { CartSummary } from "@/modules/cart";
import { CheckoutFlow } from "@/modules/checkout";
import { BillingPanel } from "@/modules/billing";
import { PageShell } from "@/common/ui";

export const CheckoutPage = PageShell("Checkout", [CartSummary, CheckoutFlow, BillingPanel]);
