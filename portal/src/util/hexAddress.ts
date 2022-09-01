export function truncateAddress(address: string): string {
  const delimiter = "...";

  // We only show the first 6 and last 4 characters of the address
  return address.slice(0, 6) + delimiter + address.slice(address.length - 4);
}
