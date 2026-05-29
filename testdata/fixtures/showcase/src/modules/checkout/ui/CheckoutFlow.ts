import { CartSummary } from "@/modules/cart";
import { collectPayment } from "../payment";
import { scheduleDelivery } from "../delivery";

export const CheckoutFlow = [CartSummary, collectPayment(), scheduleDelivery()];
