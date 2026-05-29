export function PageShell(props: { children: unknown[] }) {
  return props.children.join("\n");
}
