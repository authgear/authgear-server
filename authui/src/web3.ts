import { Controller } from "@hotwired/stimulus";
import { handleAxiosError, showErrorMessage } from "./messageBar";
import jazzicon from "@metamask/jazzicon";
import { ethers } from "ethers";
import axios from "axios";
import { SiweMessage } from "siwe";
import { visit } from "@hotwired/turbo";

enum WalletProvider {
  MetaMask = "metamask",
}

function metamaskIsAvailable(): boolean {
  return (
    typeof window.ethereum !== "undefined" &&
    window.ethereum?.isMetaMask === true
  );
}

function deserializeNonceResponse(data: any): SIWENonce {
  return {
    nonce: data.nonce,
    expireAt: new Date(data.expire_at),
  };
}

function createSIWEMessage(
  address: string,
  chainId: number,
  nonce: string,
  expiry: string
): string {
  const message = new SiweMessage({
    domain: window.location.host,
    address,
    uri: window.location.origin,
    version: "1",
    chainId,
    nonce,
    expirationTime: expiry,
  });

  return message.prepareMessage();
}

function checkProviderIsAvailable(provider: string): boolean {
  switch (provider) {
    case WalletProvider.MetaMask:
    default:
      return metamaskIsAvailable();
  }
}

function getProvider(provider: string): ethers.providers.Web3Provider | null {
  if (!checkProviderIsAvailable(provider)) {
    return null;
  }

  switch (provider) {
    case WalletProvider.MetaMask:
    default:
      return new ethers.providers.Web3Provider(window.ethereum!);
  }
}

function truncateAddress(address: string): string {
  return address.slice(0, 6) + "..." + address.slice(address.length - 4);
}

function generateIcon(address: string, diameter: number): SVGElement | null {
  // Metamask uses 8 characters from the address as seed
  const addr = address.slice(2, 10);
  const seed = parseInt(addr, 16);

  const icon = jazzicon(diameter, seed);

  const child = icon.firstChild;

  const svg = child as SVGElement | null;

  if (svg === null) {
    return svg;
  }

  svg.style.borderRadius = "50%";
  return svg;
}

interface MetaMaskError {
  code: number;
  message: string;
}
function isMetaMaskError(err: unknown): err is MetaMaskError {
  return (
    typeof err === "object" && err !== null && "code" in err && "message" in err
  );
}
function parseWalletError(err: unknown): string {
  if (isMetaMaskError(err)) {
    switch (err.code) {
      // User rejection
      case 4001:
        return "error-message-metamask-user-rejected";
      // Unauthorized
      case 4100:
        return "error-message-metamask-unauthorized";
      // Request method not supported
      case 4200:
        return "error-message-metamask-unsupported-method";
      // Disconnected from chains
      case 4900:
      case 4901:
        return "error-message-metamask-disconnected";
      default:
        return "error-message-failed-to-connect-wallet";
    }
  }
  return "error-message-failed-to-connect-wallet";
}

function handleError(err: unknown) {
  console.error(err);

  const parsedErrorId = parseWalletError(err);

  showErrorMessage(parsedErrorId);
  return;
}

export class WalletConnectionController extends Controller {
  static targets = ["button"];
  static values = {
    provider: String,
  };

  declare buttonTarget: HTMLButtonElement;

  declare providerValue: string;
  declare provider: ethers.providers.Web3Provider | null;

  connect() {
    this.provider = getProvider(this.providerValue);
  }

  connectWallet(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    this._connectWallet();
  }

  async _connectWallet() {
    if (!this.provider) {
      visit(`/missing_web3_wallet?provider=${this.providerValue}`);
      return;
    }

    // Ensure wallet is connected
    await this.provider.send("eth_requestAccounts", []);

    visit(`/confirm_web3_account?provider=${this.providerValue}`);
  }
}

export class WalletConfirmationController extends Controller {
  static targets = [
    "button",
    "icon",
    "displayed",
    "message",
    "signature",
    "submit",
  ];
  static values = {
    provider: String,
  };

  declare buttonTarget: HTMLButtonElement;
  declare displayedTarget: HTMLSpanElement;
  declare iconTarget: HTMLDivElement;
  declare messageTarget: HTMLInputElement;
  declare signatureTarget: HTMLInputElement;
  declare submitTarget: HTMLButtonElement;

  declare providerValue: string;
  declare provider: ethers.providers.Web3Provider | null;

  connect() {
    this.provider = getProvider(this.providerValue);

    if (!this.provider) {
      visit(`/missing_web3_wallet?provider=${this.providerValue}`);
      return;
    }

    this._getAccount();
  }

  async _getAccount() {
    if (!this.provider) {
      return;
    }

    await this.provider.send("eth_requestAccounts", []);

    // Get account from the signer to ensure the requested account is the correct one
    const signer = this.provider.getSigner();
    const account = await signer.getAddress();

    this.displayedTarget.textContent = truncateAddress(account);

    const icon = generateIcon(account, 20);
    if (icon) {
      // Clear previous icons if exists
      this.iconTarget.innerHTML = "";
      this.iconTarget.appendChild(icon);
    }
  }

  reconnect(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();

    this.disconnectAccount().then(() => {
      this._getAccount();
    });
  }

  async disconnectAccount() {
    if (!this.provider) {
      return;
    }

    try {
      // Requesting this to revoke account permission
      await this.provider.send("wallet_requestPermissions", [
        {
          eth_accounts: {},
        },
      ]);
    } catch (err: unknown) {
      handleError(err);
    }
  }

  async performSIWE() {
    if (!this.provider) {
      return;
    }

    try {
      const nonceResp = await axios("/siwe/nonce", {
        method: "get",
      });

      const nonce = deserializeNonceResponse(nonceResp.data.result);

      const signer = this.provider.getSigner();

      const address = await signer.getAddress();
      const chainId = await signer.getChainId();

      const siweMessage = createSIWEMessage(
        address,
        chainId,
        nonce.nonce,
        nonce.expireAt.toISOString()
      );

      const signature = await signer.signMessage(siweMessage);

      this.messageTarget.value = siweMessage;
      this.signatureTarget.value = signature;

      this.submitTarget.click();
    } catch (e: unknown) {
      handleAxiosError(e);
    }
  }
}
