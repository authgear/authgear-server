import { createEIP681URL, parseEIP681 } from "./eip681";

export const ALL_SUPPORTED_NETWORKS: NetworkId[] = [
  {
    blockchain: "ethereum",
    network: "1",
  },
];
export interface NetworkId {
  blockchain: string;
  network: string;
}

export function parseNetworkId(url: string): NetworkId {
  const curl = new URL(url);

  const protocol = curl.protocol.replace(":", "");

  switch (protocol) {
    case "ethereum": {
      const eip681 = parseEIP681(url);

      if (eip681.address !== "0x0") {
        throw new Error(`Unknown network Id: ${url}`);
      }

      return {
        blockchain: "ethereum",
        network: eip681.chainId.toString(),
      };
    }

    default:
      throw new Error(`Unknown protocol: ${protocol}`);
  }
}

export function createNetworkIdURL(networkId: NetworkId): string {
  switch (networkId.blockchain) {
    case "ethereum":
      return createEIP681URL({
        chainId: parseInt(networkId.network, 10),
        address: "0x0",
      });
    default:
      throw new Error(`Unknown blockchain: ${networkId.blockchain}`);
  }
}

export function getNetworkNameId(networkId: NetworkId): string {
  switch (networkId.blockchain) {
    case "ethereum":
      switch (networkId.network) {
        case "1":
          return "NetworkId.ethereum-mainnet";
        default:
          throw new Error(`Unsupported chain id: ${networkId.network}`);
      }
    default:
      throw new Error(`Unsupported blockchain: ${networkId.blockchain}`);
  }
}
