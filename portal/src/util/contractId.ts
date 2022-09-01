import { createEIP681URL, parseEIP681 } from "./eip681";

export interface ContractId {
  blockchain: string;
  network: string;
  address: string;
}

export function parseContractId(url: string): ContractId {
  const curl = new URL(url);

  const protocol = curl.protocol.replace(":", "");

  switch (protocol) {
    case "ethereum": {
      const eip681 = parseEIP681(url);

      return {
        blockchain: "ethereum",
        network: eip681.chainId.toString(),
        address: eip681.address,
      };
    }

    default:
      throw new Error(`Unknown protocol: ${protocol}`);
  }
}

export function createContractIdURL(contractId: ContractId): string {
  switch (contractId.blockchain) {
    case "ethereum":
      return createEIP681URL({
        chainId: parseInt(contractId.network, 10),
        address: contractId.address,
      });
    default:
      throw new Error(`Unknown blockchain: ${contractId.blockchain}`);
  }
}
