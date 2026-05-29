import { DashboardPage } from "@/pages/dashboard";
import { CheckoutPage } from "@/pages/checkout";
import { AdminPage } from "@/pages/admin";
import { ProfilePage } from "@/pages/profile";
import { ReportsPage } from "@/pages/reports";
import { createRouter } from "@/common/runtime";

export const app = createRouter([DashboardPage, CheckoutPage, AdminPage, ProfilePage, ReportsPage]);
