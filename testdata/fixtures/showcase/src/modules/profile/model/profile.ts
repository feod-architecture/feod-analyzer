import { getSession } from "@/modules/auth";

export function getProfile() {
  return { id: "usr_1", session: Boolean(getSession) };
}
