import { productCatalog } from "@/modules/catalog";
import { formatMoney } from "@/common/money";
import { DataTable } from "@/common/ui";
import { cartItems } from "../model/cart";

export const CartSummary = DataTable(cartItems.map((name) => ({ name, inCatalog: productCatalog.includes(name), total: formatMoney(10) })));
