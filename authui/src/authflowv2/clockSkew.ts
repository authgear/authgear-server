import { Controller } from "@hotwired/stimulus";
import { visit } from "@hotwired/turbo";

// Leaves for the clock skew page when the device clock is outside the skew the
// server accepts for DPoP proofs, because the token request would fail anyway.
export class ClockSkewController extends Controller {
  static values = {
    serverTime: Number,
    maxAhead: Number,
    maxBehind: Number,
    redirectUrl: String,
  };

  declare readonly serverTimeValue: number;
  declare readonly maxAheadValue: number;
  declare readonly maxBehindValue: number;
  declare readonly redirectUrlValue: string;

  connect(): void {
    // serverTime is only meaningful for the response it came with.
    // Skip Turbo previews and restored snapshots, which carry this attribute.
    if (
      document.documentElement.hasAttribute("data-turbo-preview") ||
      this.element.hasAttribute("data-clock-skew-checked")
    ) {
      return;
    }
    this.element.setAttribute("data-clock-skew-checked", "");

    const skew = Date.now() - this.serverTimeValue;
    if (skew > this.maxAheadValue || skew < -this.maxBehindValue) {
      visit(this.redirectUrlValue, { action: "replace" });
    }
  }
}
