import { CatalogPage } from "@/pages/catalog";
import { CheckoutPage } from "@/pages/checkout";
import { AdminPage } from "@/pages/admin";
import { ViewerProvider } from "@/modules/viewer";
import { createRouter } from "@/common/runtime";

export const app = createRouter([CatalogPage, CheckoutPage, AdminPage], ViewerProvider);
