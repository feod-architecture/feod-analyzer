import { UserPermissionsPanel } from "@/modules/user/permissions";
import { normalizeUser } from "@/modules/user/lib/normalizeUser";

export function UserPage() {
  return UserPermissionsPanel(normalizeUser("id"));
}
