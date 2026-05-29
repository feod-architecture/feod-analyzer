export function createRouter(routes: unknown[], provider: (children: unknown) => unknown) {
  return provider(routes);
}
