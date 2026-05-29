import { getProfile } from "@/modules/profile";

export function getSession() {
  return { user: getProfile() };
}
