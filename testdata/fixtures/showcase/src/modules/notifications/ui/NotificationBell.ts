import { env } from "@/global/env";
import { Button } from "@/common/ui";

export const NotificationBell = Button(env.notifyEndpoint);
