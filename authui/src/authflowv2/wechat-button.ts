import { Controller } from "@hotwired/stimulus";

export class WechatButtonController extends Controller {
  static targets = ["openWechatAnchor", "proceedButton"];

  declare openWechatAnchorTarget: HTMLAnchorElement;
  declare proceedButtonTarget: HTMLButtonElement;

  onClickOpenWechatAnchor() {
    // Do not preventDefault nor stopPropagation.
    // We just want to changes the buttons.
    this.openWechatAnchorTarget.classList.add("hidden");
    this.proceedButtonTarget.classList.remove("hidden");
  }

  onSubmit(e: Event) {
    // It is observed that on Safari, if the form has data-turbo="false",
    // then changing the button.disabled = true is not rendered.
    // Therefore, we do a trick here.
    // We prevent the default, disable the button, and submit the form again.
    // The intention is to make the button disabled during form submission so that
    // the button cannot be clicked again.
    e.preventDefault();
    this.proceedButtonTarget.disabled = true;
    if (this.element instanceof HTMLFormElement) {
      this.element.submit();
    }
  }
}
