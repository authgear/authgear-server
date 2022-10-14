import BigNumber from "bignumber.js";

export function truncateAddress(address: string): string {
  const delimiter = "...";

  // No need to trim if it's not long enough
  if (address.length < 6) {
    return address;
  }

  // We only show the first 6 and last 4 characters of the address
  return address.slice(0, 6) + delimiter + address.slice(address.length - 4);
}

export function parseHexstring(hex: string): string {
  const bn = BigNumber(hex);

  if (bn.isNaN()) {
    return "";
  }

  return bn.toString(10);
}

export function convertToHexstring(dec: string): string {
  const bn = BigNumber(dec);

  if (bn.isNaN()) {
    return "";
  }

  return "0x" + bn.toString(16);
}
