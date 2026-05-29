import { ProfileCard } from "@/modules/profile";
import { LoginButton } from "@/modules/auth";
import { PageShell } from "@/common/ui";

export const ProfilePage = PageShell("Profile", [ProfileCard, LoginButton]);
