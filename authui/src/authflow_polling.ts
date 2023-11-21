import { Controller } from "@hotwired/stimulus";
import { visit } from "@hotwired/turbo";
import axios from "axios";

function refreshPage() {
  let url = window.location.pathname;
  if (window.location.search !== "") {
    url += window.location.search;
  }
  if (window.location.hash !== "") {
    url += window.location.hash;
  }
  visit(url, { action: "replace" });
}

const TIMEOUT = 3000;

export class AuthflowPollingController extends Controller {
  static values = {
    statetoken: String,
  };

  declare statetokenValue: string;

  setTimeoutHandle: number | null = null;

  connect() {
    this.poll();
  }

  disconnect() {
    this.cancel();
  }

  poll() {
    let recur = true;

    this.setTimeoutHandle = window.setTimeout(async () => {
      try {
        const response = await axios("/api/v1/authentication_flows/states", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          data: {
            state_token: this.statetokenValue,
          },
        });
        console.info("authflow_polling poll success");

        if (response.data?.result?.action?.data?.can_check === true) {
          recur = false;
          refreshPage();
        }
      } catch (e) {
        console.error("authflow_polling error", e);
      } finally {
        if (recur) {
          this.poll();
        } else {
          this.cancel();
        }
      }
    }, TIMEOUT);
  }

  cancel() {
    if (this.setTimeoutHandle != null) {
      window.clearTimeout(this.setTimeoutHandle);
    }
  }
}
