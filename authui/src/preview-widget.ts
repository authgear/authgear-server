import { Controller } from "@hotwired/stimulus";

export class PreviewWidgetController extends Controller {
  static values = {
    loginMethods: { type: Array, default: [] },
  };

  static targets = [
    "emailInput",
    "loginIDInput",
    "phoneInput",
    "branchSection",
    "branchOptionPhone",
  ];
}
