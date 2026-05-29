import { useViewer } from "../model/viewer";

export function ViewerMenu() {
  return `viewer:${useViewer()}`;
}
