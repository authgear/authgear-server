import { Controller } from "@hotwired/stimulus";
import { PasswordPolicyController } from "./password-policy";

export class NewPasswordFieldController extends Controller {
  static values = {
    inputErrorClass: { type: String, default: "input--error" },
    confirmPasswordErrorMessage: { type: String },
  };
  static targets = [
    "newPasswordInput",
    "confirmPasswordInput",
    "confirmPasswordError",
  ];

  declare inputErrorClassValue: string;
  declare confirmPasswordErrorMessageValue: string;
  declare newPasswordInputTarget: HTMLInputElement;
  declare confirmPasswordInputTarget: HTMLInputElement;
  declare confirmPasswordErrorTarget: HTMLElement;

  newPasswordInputTargetConnected(el: HTMLInputElement) {
    el.addEventListener("blur", this.handlePasswordInputBlur);
  }

  newPasswordInputTargetDisconnected(el: HTMLInputElement) {
    el.removeEventListener("blur", this.handlePasswordInputBlur);
  }

  confirmPasswordInputTargetConnected(el: HTMLInputElement) {
    el.addEventListener("blur", this.handleConfirmPasswordInputBlur);
  }

  confirmPasswordInputTargetDisconnected(el: HTMLInputElement) {
    el.removeEventListener("blur", this.handleConfirmPasswordInputBlur);
  }

  handlePasswordInputBlur = () => {
    if (this.isNewPasswordInputEmpty()) {
      return;
    }
    const violated = this.newPasswordInputTarget.getAttribute(
      PasswordPolicyController.ATTR_POLICY_VIOLATED
    );
    if (violated != null && violated !== "") {
      this.newPasswordInputTarget.classList.add(this.inputErrorClassValue);
    }
    if (this.isConfirmPasswordInputEmpty()) {
      return;
    }
    if (!this.isConfirmPasswordCorrect()) {
      this.confirmPasswordInputTarget.classList.add(this.inputErrorClassValue);
      this.confirmPasswordErrorTarget.classList.remove("hidden");
      this.confirmPasswordErrorTarget.innerHTML =
        this.confirmPasswordErrorMessageValue;
    } else {
      this.confirmPasswordInputTarget.classList.remove(
        this.inputErrorClassValue
      );
      this.confirmPasswordErrorTarget.classList.add("hidden");
    }
  };

  handleConfirmPasswordInputBlur = () => {
    if (this.isConfirmPasswordInputEmpty()) {
      return;
    }
    this.handlePasswordInputBlur();
  };

  private isNewPasswordInputEmpty(): boolean {
    return this.newPasswordInputTarget.value === "";
  }
  private isConfirmPasswordInputEmpty(): boolean {
    return this.confirmPasswordInputTarget.value === "";
  }

  private isConfirmPasswordCorrect(): boolean {
    return (
      this.newPasswordInputTarget.value ===
      this.confirmPasswordInputTarget.value
    );
  }
}
