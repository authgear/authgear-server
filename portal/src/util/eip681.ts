const ETHEREUM_ADDRESS_REGEX = /^0x[a-fA-F0-9]{40}$/;

export interface EIP681 {
  chainId: number;
  address: string;
}

export function parseEIP681(
  url: string,
  skipAddressCheck: boolean = false
): EIP681 {
  const protocolURI = url.split(":");

  if (protocolURI.length !== 2) {
    throw new Error(`Invalid URI: ${url}`);
  }

  if (protocolURI[0] !== "ethereum") {
    throw new Error(`Invalid protocol: ${protocolURI[0]}`);
  }

  const addressURI = protocolURI[1].split("@");

  if (addressURI.length !== 2) {
    throw new Error(`Invalid URI: ${url}`);
  }

  const address = addressURI[0];
  const regex = new RegExp(ETHEREUM_ADDRESS_REGEX);
  if (!skipAddressCheck && !regex.test(address)) {
    throw new Error(`Invalid address: ${address}`);
  }

  const chainId = parseInt(addressURI[1], 10);
  if (chainId < 0) {
    throw new Error(`Chain ID cannot be negative: ${chainId}`);
  }

  return {
    chainId,
    address,
  };
}

export function createEIP681URL(
  eip681: EIP681,
  skipAddressCheck: boolean = false
): string {
  const url = `ethereum:${eip681.address}@${eip681.chainId}`;
  // Confirm the format is correct
  parseEIP681(url, skipAddressCheck);
  return url;
}
