import { createEIP681URL, parseEIP681 } from "./eip681";

export interface ContractID {
  blockchain: string;
  network: string;
  address: string;
  query?: URLSearchParams;
}

export function parseContractID(
  url: string,
  skipAddressCheck: boolean = false
): ContractID {
  const curl = new URL(url);

  const protocol = curl.protocol.replace(":", "");

  switch (protocol) {
    case "ethereum": {
      const eip681 = parseEIP681(url, skipAddressCheck);

      return {
        blockchain: "ethereum",
        network: eip681.chainId.toString(),
        address: eip681.address,
        query: eip681.query,
      };
    }

    default:
      throw new Error(`Unknown protocol: ${protocol}`);
  }
}

export function createContractIDURL(contractId: ContractID): string {
  switch (contractId.blockchain) {
    case "ethereum":
      return createEIP681URL({
        chainId: parseInt(contractId.network, 10),
        address: contractId.address,
        query: contractId.query,
      });
    default:
      throw new Error(`Unknown blockchain: ${contractId.blockchain}`);
  }
}
