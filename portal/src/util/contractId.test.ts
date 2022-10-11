import { describe, it, expect } from "@jest/globals";
import { parseContractID, ContractID, createContractIDURL } from "./contractId";

describe("ContractID", () => {
  it("parses contract id", () => {
    function test(uri: string, expected: ContractID) {
      const contractId = parseContractID(uri);

      expect(contractId).toEqual(expected);
    }

    test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1", {
      blockchain: "ethereum",
      network: "1",
      address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
    });

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1?token_ids=0x1&token_ids=0x2",
      {
        blockchain: "ethereum",
        network: "1",
        address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
        query: new URLSearchParams([
          ["token_ids", "0x1"],
          ["token_ids", "0x2"],
        ]),
      }
    );
  });

  it("generate contract id url", () => {
    function test(contractId: ContractID, expected: string) {
      const url = createContractIDURL(contractId);

      expect(url).toEqual(expected);
    }

    test(
      {
        blockchain: "ethereum",
        network: "1",
        address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
      },
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1"
    );

    test(
      {
        blockchain: "ethereum",
        network: "1",
        address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
        query: new URLSearchParams([
          ["token_ids", "0x1"],
          ["token_ids", "0x2"],
        ]),
      },
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1?token_ids=0x1&token_ids=0x2"
    );
  });
});
