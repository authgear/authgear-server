import { Controller } from "@hotwired/stimulus";
import { handleAxiosError, showErrorMessage } from "./messageBar";
import jazzicon from "@metamask/jazzicon";
import axios from "axios";
import { SiweMessage } from "siwe";
import { visit } from "@hotwired/turbo";
import detectEthereumProvider from "@metamask/detect-provider";
import { ethers } from "ethers";

enum WalletProvider {
  MetaMask = "metamask",
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

function truncateAddress(address: string): string {
  return address.slice(0, 6) + "..." + address.slice(address.length - 4);
}

interface MetamaskProvider extends ethers.providers.ExternalProvider {
  isMetaMask: true;
  on: (eventName: string, callback: () => void) => void;
  off: (eventName: string, callback: () => void) => void;
}

interface Web3Provider extends Omit<ethers.providers.Web3Provider, "provider"> {
  provider: ethers.providers.ExternalProvider | MetamaskProvider;
}

function isProviderMetaMask(
  provider?: ethers.providers.ExternalProvider | MetamaskProvider
): provider is MetamaskProvider {
  return provider?.isMetaMask === true;
}

async function getProvider(type: string): Promise<Web3Provider | null> {
  const provider = (await detectEthereumProvider({
    mustBeMetaMask: type === WalletProvider.MetaMask,
  })) as ethers.providers.ExternalProvider | null;

  if (provider) {
    return new ethers.providers.Web3Provider(provider);
  }

  return null;
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
function parseWalletError(err: unknown): string | null {
  if (isMetaMaskError(err)) {
    switch (err.code) {
      // User rejection, no need to show error message
      case 4001:
        return null;
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

  if (parsedErrorId) {
    showErrorMessage(parsedErrorId);
  }
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
    getProvider(this.providerValue)
      .then((provider) => {
        this.provider = provider;
      })
      .catch((err) => {
        handleError(err);
      });
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

    try {
      // Ensure wallet is connected
      await this.provider.send("eth_requestAccounts", []);
      visit(`/confirm_web3_account?provider=${this.providerValue}`);
    } catch (err) {
      handleError(err);
    }
  }
}

export class WalletIconController extends Controller {
  static targets = ["iconContainer"];
  static values = {
    address: String,
    size: Number,
  };

  declare iconContainerTarget: HTMLDivElement;

  declare sizeValue: number;
  declare addressValue: string;

  generateIcon(): SVGElement | null {
    // Metamask uses 8 characters from the address as seed
    const addr = this.addressValue.slice(2, 10);
    const seed = parseInt(addr, 16);

    const icon = jazzicon(this.sizeValue, seed);

    const child = icon.firstChild;
    if (child instanceof SVGElement) {
      child.style.borderRadius = "50%";
      return child;
    }

    return null;
  }

  sizeValueChanged() {
    this.renderIcon();
  }

  addressValueChanged() {
    this.renderIcon();
  }

  onAddressUpdate({ detail: { address } }: { detail: { address: string } }) {
    this.addressValue = address;
  }

  renderIcon() {
    const icon = this.generateIcon();

    if (icon) {
      // Clear previous icons if exists
      this.iconContainerTarget.innerHTML = "";
      this.iconContainerTarget.appendChild(icon);
    }
  }
}

export class WalletConfirmationController extends Controller {
  static targets = ["button", "displayed", "message", "signature", "submit"];
  static values = {
    provider: String,
  };

  declare buttonTarget: HTMLButtonElement;
  declare displayedTarget: HTMLSpanElement;
  declare messageTarget: HTMLInputElement;
  declare signatureTarget: HTMLInputElement;
  declare submitTarget: HTMLButtonElement;

  declare providerValue: string;
  declare provider: Web3Provider | null;

  connect() {
    getProvider(this.providerValue)
      .then((provider) => {
        if (!provider) {
          visit(`/missing_web3_wallet?provider=${this.providerValue}`);
          return;
        }
        this.provider = provider;
        this._getAccount();
      })
      .catch((err) => {
        handleError(err);
      });

    if (isProviderMetaMask(this.provider?.provider)) {
      this.provider!.provider.on("accountsChanged", () => this._getAccount());
    }
  }

  disconnect() {
    if (isProviderMetaMask(this.provider?.provider)) {
      this.provider!.provider.off("accountsChanged", () => this._getAccount());
    }
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

    this.dispatch("addressUpdate", { detail: { address: account } });
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
