import { Controller } from "@hotwired/stimulus";
import { showErrorMessage } from "./messageBar";
import jazzicon from "@metamask/jazzicon";
import { ethers } from "ethers";

enum WalletProvider {
  MetaMask = "metamask",
}

function metamaskIsAvailable(): boolean {
  return (
    typeof window.ethereum !== "undefined" &&
    window.ethereum?.isMetaMask === true
  );
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
      return new ethers.providers.Web3Provider(window.ethereum);
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

function handleError(err: unknown) {
  console.error(err);

  showErrorMessage("error-message-failed-to-connect-wallet");
  return;
}

export class WalletConnectionController extends Controller {
  static targets = ["button", "confirm"];
  static values = {
    provider: String,
  };

  declare buttonTarget: HTMLButtonElement;
  declare confirmTarget: HTMLAnchorElement;

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
      return;
    }

    // Ensure wallet is connected
    await this.provider.send("eth_requestAccounts", []);

    this.confirmTarget.href = `${this.confirmTarget.href}?provider=${this.providerValue}`;
    this.confirmTarget.click();
  }
}

export class WalletConfirmationController extends Controller {
  static targets = ["button", "icon", "displayed", "address", "network"];
  static values = {
    provider: String,
  };

  declare buttonTarget: HTMLButtonElement;
  declare displayedTarget: HTMLSpanElement;
  declare addressTarget: HTMLInputElement;
  declare networkTarget: HTMLInputElement;
  declare iconTarget: HTMLDivElement;

  declare providerValue: string;
  declare provider: ethers.providers.Web3Provider | null;

  connect() {
    this.provider = getProvider(this.providerValue);

    this._getAccount();
  }

  async _getAccount() {
    if (!this.provider) {
      return;
    }

    const account = await this.provider.send("eth_requestAccounts", []);
    const network = await this.provider.getNetwork();

    this.addressTarget.value = account;
    this.networkTarget.value = network.chainId.toString();
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
}
