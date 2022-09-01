import { describe, it, expect } from "@jest/globals";
import { createNetworkIdURL, NetworkId, parseNetworkId } from "./networkId";

describe("NetworkId", () => {
  it("parses network id", () => {
    function test(uri: string, expected: NetworkId) {
      const networkId = parseNetworkId(uri);

      expect(networkId).toEqual(expected);
    }

    test("ethereum:0x0@1", {
      blockchain: "ethereum",
      network: "1",
    });
  });

  it("generate network id url", () => {
    function test(networkId: NetworkId, expected: string) {
      const url = createNetworkIdURL(networkId);

      expect(url).toEqual(expected);
    }

    test(
      {
        blockchain: "ethereum",
        network: "1",
      },
      "ethereum:0x0@1"
    );
  });
});
