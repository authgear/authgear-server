import { Controller } from "@hotwired/stimulus";
import { TranslatedMessageController } from "./translated-message";

interface PreviewCustomisationMessage {
  cssVars: Record<string, string>;
  translations: Record<string, string>;
}

function parsePreviewCustomisationMessage(
  message: any
): PreviewCustomisationMessage | null {
  if (message.type !== "PreviewCustomisationMessage") {
    return null;
  }
  return {
    cssVars: message.cssVars ?? {},
    translations: message.translations ?? {},
  };
}

export class InlinePreviewController extends Controller {
  static outlets = ["translated-message"];
  static values = {
    isInlinePreview: Boolean,
  };

  declare translatedMessageOutlets: TranslatedMessageController[];
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
    const customisationMessage = parsePreviewCustomisationMessage(e.data);
    if (customisationMessage == null) {
      return;
    }
    const el = this.element as HTMLElement;
    for (const [name, value] of Object.entries(customisationMessage.cssVars)) {
      el.style.setProperty(name, value);
    }

    const keyToTranslatedMessageOutlet: Partial<
      Record<string, TranslatedMessageController>
    > = {};
    this.translatedMessageOutlets.forEach((outlet) => {
      keyToTranslatedMessageOutlet[outlet.keyValue] = outlet;
    });
    for (const [key, value] of Object.entries(
      customisationMessage.translations
    )) {
      const outlet = keyToTranslatedMessageOutlet[key];
      outlet?.setMessage(value);
    }
  };
}
