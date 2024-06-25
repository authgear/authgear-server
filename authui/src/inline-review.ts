import { Controller } from "@hotwired/stimulus";

export class InlinePreviewController extends Controller {
  static values = {
    isInlinePreview: Boolean,
  };

  declare isInlinePreviewValue: boolean;

  windowMessageAllowedOrigins!: string[];

  connect(): void {
    if (!this.isInlinePreviewValue) {
      return;
    }
    const windowMessageAllowedOrigins = ((): string[] => {
      const meta: HTMLMetaElement | null = document.querySelector(
        "meta[name=x-window-message-allowed-origins]"
      );
      const content = meta?.content ?? "";
      return content.split(",").map((origin) => origin.trim());
    })();
    this.windowMessageAllowedOrigins = windowMessageAllowedOrigins;
    if (windowMessageAllowedOrigins.length === 0) {
      return;
    }
    window.addEventListener("message", this.onReceiveMessage);
  }

  disconnect(): void {
    window.removeEventListener("message", this.onReceiveMessage);
  }

  onReceiveMessage = (e: MessageEvent<any>): void => {
    if (!this.windowMessageAllowedOrigins.includes(e.origin)) {
      return;
    }
  };
}
