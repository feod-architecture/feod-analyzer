import { PageShell } from "@/common/ui";
import { getProfile } from "../model/profile";

export const ProfileCard = PageShell("Profile", getProfile());
