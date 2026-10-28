import { Domain } from "../graphql/portal/globalTypes.generated";

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

export function checkIsCustomDomain(
  domains: Domain[],
  public_origin: string
): boolean {
  if (domains.length === 0) {
    return false;
  }
  const index = domains.findIndex((d) =>
    getHostFromOrigin(public_origin).includes(d.domain)
  );
  if (index < 0) {
    return false;
  }
  return domains[index].isCustom;
}
