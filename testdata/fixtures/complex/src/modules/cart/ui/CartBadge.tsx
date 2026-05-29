import { getCartTotal } from "../model/cart";

export function CartBadge() {
  return `cart:${getCartTotal()}`;
}
