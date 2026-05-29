import { orderModel } from "@/modules/order";

export function normalizeUser(id: string) {
  return `${id}:${String(orderModel)}`;
}
