import { AxiosProgressEvent } from "axios";
//
// disableAllButtons remembers the disabled state of all buttons.
// It disables all button and returns a function to revert to original state.
export function disableAllButtons(): () => void {
  const buttons = document.querySelectorAll("button");
  const original: [HTMLButtonElement, boolean][] = [];
  for (let i = 0; i < buttons.length; i++) {
    const button = buttons[i];
    const disabled = button.disabled;
    const state: [HTMLButtonElement, boolean] = [button, disabled];
    button.disabled = true;
    original.push(state);
  }

  return () => {
    for (let i = 0; i < original.length; i++) {
      const [button, disabled] = original[i];
      button.disabled = disabled;
    }
  };
}

// displayButtonLoadingState remembers the content of button.
// It show loading state on button and returns a function to revert to original content.
export function displayButtonLoadingState(el: HTMLButtonElement): () => void {
  const original: string = el.innerHTML;
  el.innerHTML =
    '<span class="primary-btn__loading-icon material-icons animate-spin">progress_activity</span>';

  return () => {
    el.innerHTML = original;
  };
}

export function hideProgressBar(): void {
  const loadingProgressBar = document.getElementById("loading-progress-bar");
  if (loadingProgressBar == null) {
    return;
  }
  loadingProgressBar.style.opacity = "0";
}

export function showProgressBar(): void {
  const loadingProgressBar = document.getElementById("loading-progress-bar");
  if (loadingProgressBar == null) {
    return;
  }
  loadingProgressBar.style.opacity = "1";
}

export function progressEventHandler(progressEvent: AxiosProgressEvent): void {
  const loadingProgressBar = document.getElementById("loading-progress-bar");
  if (loadingProgressBar == null) {
    return;
  }

  if (progressEvent.total == null) {
    return;
  }

  const percentage = Math.round(
    (100 * progressEvent.loaded) / progressEvent.total
  );
  const width = Math.max(0, Math.min(100, percentage));
  loadingProgressBar.style.width = `${width}%`;
}
