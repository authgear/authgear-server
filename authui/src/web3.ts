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
    // https://github.com/ethers-io/ethers.js/issues/866
    return new ethers.providers.Web3Provider(provider, "any");
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
  declare _provider: Web3Provider | null;

  set provider(p: Web3Provider | null) {
    if (isProviderMetaMask(p?.provider)) {
      p!.provider.on("accountsChanged", this.onAccountChanged);
    } else if (isProviderMetaMask(this._provider?.provider) && p === null) {
      this._provider?.provider.off?.("accountsChanged", this.onAccountChanged);
    }

    this._provider = p;
  }

  get provider() {
    return this._provider;
  }

  connect() {
    getProvider(this.providerValue)
      .then((provider) => {
        if (!provider) {
          visit(`/errors/missing_web3_wallet?provider=${this.providerValue}`);
          return;
        }
        this.provider = provider;
        this._getAccount();
      })
      .catch((err) => {
        handleError(err);
      });
  }

  disconnect() {
    this.provider = null;
  }

  onAccountChanged = () => {
    this._getAccount();
  };

  async _getAccount() {
    if (!this.provider) {
      return;
    }
    this.displayedTarget.textContent = "-";

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

    // Ensure at least one account is connected if user has rejected the initial request
    await this._getAccount();

    try {
      const nonceResp = await axios("/_internals/siwe/nonce", {
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
