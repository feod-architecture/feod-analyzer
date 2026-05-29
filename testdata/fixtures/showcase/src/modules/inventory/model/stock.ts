import { productCatalog } from "@/modules/catalog";
import { logger } from "@/common/logger/lib/logger";

export const stockRows = productCatalog.map((name) => logger(name));
