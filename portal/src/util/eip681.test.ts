import { describe, it, expect } from "@jest/globals";
import { createEIP681URL, EIP681, parseEIP681 } from "./eip681";

describe("EIP681", () => {
  it("parses eip681", () => {
    function test(uri: string, expected: EIP681) {
      const eip681 = parseEIP681(uri);

      expect(eip681).toEqual(expected);
    }

    test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1", {
      chainId: 1,
      address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
    });

    test("ethereum:0xdc0479cc5bba033b3e7de9f178607150b3abce1f@1231", {
      chainId: 1231,
      address: "0xdc0479cc5bba033b3e7de9f178607150b3abce1f",
    });

    test("ethereum:0x71c7656ec7ab88b098defb751b7401b5f6d8976f@23821", {
      chainId: 23821,
      address: "0x71c7656ec7ab88b098defb751b7401b5f6d8976f",
    });
  });

  it("create eip681 url", () => {
    function test(eip681: EIP681, expected: string) {
      const url = createEIP681URL(eip681);

      expect(url).toEqual(expected);
    }

    test(
      {
        chainId: 1,
        address: "0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d",
      },
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1"
    );

    test(
      {
        chainId: 1231,
        address: "0xdc0479cc5bba033b3e7de9f178607150b3abce1f",
      },
      "ethereum:0xdc0479cc5bba033b3e7de9f178607150b3abce1f@1231"
    );

    test(
      {
        chainId: 23821,
        address: "0x71c7656ec7ab88b098defb751b7401b5f6d8976f",
      },
      "ethereum:0x71c7656ec7ab88b098defb751b7401b5f6d8976f@23821"
    );
  });
});
