import { Button } from "@/common/ui";
import { getSession } from "../model/session";

export const LoginButton = Button(String(Boolean(getSession())));
