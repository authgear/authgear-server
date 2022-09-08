import { describe, it, expect } from "@jest/globals";
import { createNetworkIDURL, NetworkID, parseNetworkID } from "./networkId";

describe("NetworkId", () => {
  it("parses network id", () => {
    function test(uri: string, expected: NetworkID) {
      const networkId = parseNetworkID(uri);

      expect(networkId).toEqual(expected);
    }

    test("ethereum:0x0@1", {
      blockchain: "ethereum",
      network: "1",
    });
  });

  it("generate network id url", () => {
    function test(networkId: NetworkID, expected: string) {
      const url = createNetworkIDURL(networkId);

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
