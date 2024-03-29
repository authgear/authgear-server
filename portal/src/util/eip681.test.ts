import { describe, it, expect } from "@jest/globals";
import {
  createEIP681URL,
  EIP681,
  explorerAddress,
  explorerBlock,
  explorerBlocks,
  explorerTx,
  explorerURL,
  parseEIP681,
} from "./eip681";

describe("EIP681", () => {
  it("parses eip681 success with address check", () => {
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

    test(
      "ethereum:0x71c7656ec7ab88b098defb751b7401b5f6d8976f@23821?token_ids=0x1&token_ids=0x2",
      {
        chainId: 23821,
        address: "0x71c7656ec7ab88b098defb751b7401b5f6d8976f",
        query: new URLSearchParams([
          ["token_ids", "0x1"],
          ["token_ids", "0x2"],
        ]),
      }
    );
  });

  it("parses eip681 without address check", () => {
    function test(uri: string, expected: EIP681) {
      const eip681 = parseEIP681(uri, true);

      expect(eip681).toEqual(expected);
    }

    test("ethereum:0x0@1", {
      chainId: 1,
      address: "0x0",
    });

    test("ethereum:0x0@1231", {
      chainId: 1231,
      address: "0x0",
    });

    test("ethereum:0x0@23821", {
      chainId: 23821,
      address: "0x0",
    });
  });

  it("create eip681 url with address check", () => {
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

  it("create eip681 url without address check", () => {
    function test(eip681: EIP681, expected: string) {
      const url = createEIP681URL(eip681, true);

      expect(url).toEqual(expected);
    }

    test(
      {
        chainId: 1,
        address: "0x0",
      },
      "ethereum:0x0@1"
    );

    test(
      {
        chainId: 1231,
        address: "0x0",
      },
      "ethereum:0x0@1231"
    );

    test(
      {
        chainId: 23821,
        address: "0x0",
      },
      "ethereum:0x0@23821"
    );
  });

  it("create etherscan url with address check", () => {
    function test(uri: string, expected: string) {
      const url = explorerURL(uri);

      expect(url).toEqual(expected);
    }

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
      "https://etherscan.io/"
    );

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@5",
      "https://goerli.etherscan.io/"
    );

    test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1234", "");
  });

  it("create etherscan url without address check", () => {
    function test(uri: string, expected: string) {
      const url = explorerURL(uri, true);

      expect(url).toEqual(expected);
    }

    test("ethereum:0x0@1", "https://etherscan.io/");

    test("ethereum:0x0@5", "https://goerli.etherscan.io/");

    test("ethereum:0x0@1234", "");
  });

  it("create etherscan address url with address check", () => {
    function test(uri: string, expected: string) {
      const url = explorerAddress(uri);

      expect(url).toEqual(expected);
    }

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
      "https://etherscan.io/address/0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"
    );

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@5",
      "https://goerli.etherscan.io/address/0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"
    );

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1234",
      "address/0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d"
    );
  });

  it("create etherscan address url without address check", () => {
    function test(uri: string, expected: string) {
      const url = explorerAddress(uri, true);

      expect(url).toEqual(expected);
    }

    test("ethereum:0x0@1", "https://etherscan.io/address/0x0");

    test("ethereum:0x0@5", "https://goerli.etherscan.io/address/0x0");

    test("ethereum:0x0@1234", "address/0x0");
  });

  it("create etherscan transaction hash url with address check", () => {
    function test(uri: string, tx: string, expected: string) {
      const url = explorerTx(uri, tx);

      expect(url).toEqual(expected);
    }

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
      "0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3",
      "https://etherscan.io/tx/0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3"
    );

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@5",
      "0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3",
      "https://goerli.etherscan.io/tx/0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3"
    );

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1234",
      "0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3",
      "tx/0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3"
    );
  });

  it("create etherscan transaction hash url without address check", () => {
    function test(uri: string, tx: string, expected: string) {
      const url = explorerTx(uri, tx, true);

      expect(url).toEqual(expected);
    }

    test(
      "ethereum:0x0@1",
      "0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3",
      "https://etherscan.io/tx/0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3"
    );

    test(
      "ethereum:0x0@5",
      "0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3",
      "https://goerli.etherscan.io/tx/0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3"
    );

    test(
      "ethereum:0x0@1234",
      "0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3",
      "tx/0x13aec95326ee5d0ba5d09d814746a7580a7d2a8bf5fb907b6fa398112353abd3"
    );
  });

  it("create etherscan block url with address check", () => {
    function test(uri: string, block: string, expected: string) {
      const url = explorerBlock(uri, block);

      expect(url).toEqual(expected);
    }

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
      "1234",
      "https://etherscan.io/block/1234"
    );

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@5",
      "1234",
      "https://goerli.etherscan.io/block/1234"
    );

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1234",
      "1234",
      "block/1234"
    );
  });

  it("create etherscan block url without address check", () => {
    function test(uri: string, block: string, expected: string) {
      const url = explorerBlock(uri, block, true);

      expect(url).toEqual(expected);
    }

    test("ethereum:0x0@1", "1234", "https://etherscan.io/block/1234");

    test("ethereum:0x0@5", "1234", "https://goerli.etherscan.io/block/1234");

    test("ethereum:0x0@1234", "1234", "block/1234");
  });

  it("create etherscan blocks url with address check", () => {
    function test(uri: string, expected: string) {
      const url = explorerBlocks(uri);

      expect(url).toEqual(expected);
    }

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1",
      "https://etherscan.io/blocks/"
    );

    test(
      "ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@5",
      "https://goerli.etherscan.io/blocks/"
    );

    test("ethereum:0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d@1234", "blocks/");
  });
});
