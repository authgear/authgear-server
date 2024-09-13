export function getOriginFromDomain(domain: string): string {
  // assume domain has no scheme
  // use https scheme
  return `https://${domain}`;
}

export function getHostFromOrigin(urlOrigin: string): string {
  try {
    return new URL(urlOrigin).host;
  } catch (_: unknown) {
    return "";
  }
}
