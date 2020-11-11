import Turbolinks from "turbolinks";

declare global {
  interface Window {
    api: API;
  }
}

export interface API {
  onLoad(handler: () => void | (() => void)): void;
}

export function init() {
  let isLoaded = false;
  const onLoadHandlers: Function[] = [];
  const onLoadDisposers: Function[] = [];

  window.api = {
    onLoad(handler) {
      onLoadHandlers.push(handler);
      if (isLoaded) {
        const disposer = handler();
        if (disposer) {
          onLoadDisposers.push(disposer);
        }
      }
    }
  };

  Turbolinks.start();
  document.addEventListener("turbolinks:load", () => {
    isLoaded = true;

    for (const disposer of onLoadDisposers) {
      disposer();
    }
    onLoadDisposers.length = 0;

    for (const handler of onLoadHandlers) {
      const disposer = handler();
      if (disposer) {
        onLoadDisposers.push(disposer);
      }
    }
  });
}
