import { createEIP681URL, parseEIP681 } from "./eip681";

export const ALL_SUPPORTED_NETWORKS: NetworkID[] = [
  {
    blockchain: "ethereum",
    network: "1",
  },
  {
    blockchain: "ethereum",
    network: "5",
  },
  {
    blockchain: "ethereum",
    network: "137",
  },
  {
    blockchain: "ethereum",
    network: "80001",
  },
];
export interface NetworkID {
  blockchain: string;
  network: string;
}

export function parseNetworkID(url: string): NetworkID {
  const curl = new URL(url);

  const protocol = curl.protocol.replace(":", "");

  switch (protocol) {
    case "ethereum": {
      const eip681 = parseEIP681(url, true);

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

export function createNetworkIDURL(networkID: NetworkID): string {
  switch (networkID.blockchain) {
    case "ethereum":
      return createEIP681URL(
        {
          chainId: parseInt(networkID.network, 10),
          address: "0x0",
        },
        true
      );
    default:
      throw new Error(`Unknown blockchain: ${networkID.blockchain}`);
  }
}

export function getNetworkNameID(networkID: NetworkID): string {
  switch (networkID.blockchain) {
    case "ethereum":
      switch (networkID.network) {
        case "1":
          return "NetworkId.ethereum-mainnet";
        case "5":
          return "NetworkId.ethereum-goerli";
        case "137":
          return "NetworkId.polygon-mainnet";
        case "80001":
          return "NetworkId.polygon-mumbai";
        default:
          throw new Error(`Unsupported chain id: ${networkID.network}`);
      }
    default:
      throw new Error(`Unsupported blockchain: ${networkID.blockchain}`);
  }
}
