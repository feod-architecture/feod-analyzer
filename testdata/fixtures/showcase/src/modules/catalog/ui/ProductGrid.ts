import { formatMoney } from "@/common/money";
import { DataTable } from "@/common/ui";
import { productCatalog } from "../model/products";

export const ProductGrid = DataTable(productCatalog.map((name, index) => ({ name, price: formatMoney(index + 1) })));
