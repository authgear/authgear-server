const ETHEREUM_ADDRESS_REGEX = /^0x[a-fA-F0-9]{40}$/;

export interface EIP681 {
  chainId: number;
  address: string;
  query?: URLSearchParams;
}

export function parseEIP681(
  uri: string,
  skipAddressCheck: boolean = false
): EIP681 {
  const url = new URL(uri);
  if (url.protocol !== "ethereum:") {
    throw new Error(`Invalid protocol: ${url.protocol}`);
  }

  const addressURI = url.pathname.split("@");

  if (addressURI.length !== 2) {
    throw new Error(`Invalid URI: ${url.pathname}`);
  }

  const address = addressURI[0];
  if (!skipAddressCheck && !ETHEREUM_ADDRESS_REGEX.test(address)) {
    throw new Error(`Invalid address: ${address}`);
  }

  const chainId = parseInt(addressURI[1], 10);
  if (chainId < 0) {
    throw new Error(`Chain ID cannot be negative: ${chainId}`);
  }

  const query =
    url.searchParams.toString() !== "" ? url.searchParams : undefined;

  return {
    chainId,
    address,
    query,
  };
}

export function createEIP681URL(
  eip681: EIP681,
  skipAddressCheck: boolean = false
): string {
  const query = eip681.query?.toString() ?? "";
  const url = `ethereum:${eip681.address}@${eip681.chainId}${
    query !== "" ? "?" + query : ""
  }`;
  // Confirm the format is correct
  parseEIP681(url, skipAddressCheck);
  return url;
}

export function explorerURL(
  eip681String: string,
  skipAddressCheck: boolean = false
): string {
  const eip681 = parseEIP681(eip681String, skipAddressCheck);

  let prefix: string;
  switch (eip681.chainId) {
    case 1:
      prefix = "https://etherscan.io/";
      break;
    case 5:
      prefix = "https://goerli.etherscan.io/";
      break;
    case 137:
      prefix = "https://polygonscan.com/";
      break;
    case 80001:
      prefix = "https://mumbai.polygonscan.com/";
      break;
    default:
      prefix = "";
  }

  return prefix;
}

export function explorerAddress(
  eip681String: string,
  skipAddressCheck: boolean = false
): string {
  const eip681 = parseEIP681(eip681String, skipAddressCheck);
  const prefix = explorerURL(eip681String, skipAddressCheck) + "address/";

  return prefix + eip681.address;
}

export function explorerTx(
  eip681String: string,
  tx: string,
  skipAddressCheck: boolean = false
): string {
  const prefix = explorerURL(eip681String, skipAddressCheck) + "tx/";

  return prefix + tx;
}

export function explorerBlock(
  eip681String: string,
  block: string,
  skipAddressCheck: boolean = false
): string {
  const prefix = explorerURL(eip681String, skipAddressCheck) + "block/";

  return prefix + block;
}

export function explorerBlocks(
  eip681String: string,
  skipAddressCheck: boolean = false
): string {
  const prefix = explorerURL(eip681String, skipAddressCheck) + "blocks/";

  return prefix;
}
