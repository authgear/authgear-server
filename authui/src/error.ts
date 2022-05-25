import { AxiosResponse } from "axios";
import { session } from "@hotwired/turbo";

export function handleAxiosError(e: unknown) {
  const err = e as any;
  if (err.response != null) {
    const response: AxiosResponse = err.response;

    if (typeof response.data === "string") {
      session.renderStreamMessage(response.data);
      return;
    }

    showErrorMessage("error-message-server");
  } else if (err.request != null) {
    showErrorMessage("error-message-network");
  } else {
    // programming error
  }

  console.error(err);
}

export function showErrorMessage(id: string) {
  setErrorMessage(id, false);
}

export function hideErrorMessage(id: string) {
  setErrorMessage(id, true);
}

function setErrorMessage(id: string, hidden: boolean) {
  const errorMessageBar = document.getElementById("error-message-bar");
  if (errorMessageBar == null) {
    return;
  }
  const message = document.getElementById(id);
  if (message == null) {
    return;
  }

  if (hidden) {
    errorMessageBar.classList.add("hidden");
    message.classList.add("hidden");
  } else {
    errorMessageBar.classList.remove("hidden");
    message.classList.remove("hidden");
  }
}
