import { Controller } from "@hotwired/stimulus";
import { PreviewableResourceController } from "./previewable-resource";
import { injectCSSAttrs } from "./cssattrs";

interface PreviewCustomisationMessage {
  theme: string;
  cssVars: Record<string, string>;
  images: Record<string, string | null>;
  translations: Record<string, string>;
}

function parsePreviewCustomisationMessage(
  message: any
): PreviewCustomisationMessage | null {
  if (message.type !== "PreviewCustomisationMessage") {
    return null;
  }
  return {
    theme: message.theme ?? "",
    cssVars: message.cssVars ?? {},
    images: message.images ?? {},
    translations: message.translations ?? {},
  };
}

export class InlinePreviewController extends Controller {
  static outlets = ["previewable-resource"];
  static values = {
    isInlinePreview: Boolean,
  };

  declare previewableResourceOutlets: PreviewableResourceController[];
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
      const currentStyle = el.style.getPropertyValue(name);
      if (currentStyle !== value) {
        el.style.setProperty(name, value);
      }
    }

    for (const [key, value] of Object.entries(
      customisationMessage.translations
    )) {
      for (const outlet of this.previewableResourceOutlets) {
        if (outlet.keyValue === key) {
          outlet?.setValue(value);
        }
      }
    }

    for (const [key, value] of Object.entries(customisationMessage.images)) {
      for (const outlet of this.previewableResourceOutlets) {
        if (outlet.keyValue === key) {
          outlet?.setValue(value);
        }
      }
    }

    el.classList.remove("dark");
    if (customisationMessage.theme === "dark") {
      el.classList.add("dark");
    }

    injectCSSAttrs(document.documentElement);
  };
}
