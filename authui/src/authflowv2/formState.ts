import { Controller } from "@hotwired/stimulus";

export class FormStateController extends Controller {
  static targets = ["input", "submit"];

  declare readonly inputTargets: HTMLInputElement[];
  declare readonly submitTarget: HTMLButtonElement;

  private initialValues: Record<string, any> = {};

  connect() {
    this.initialValues = this.getValues();
    this.onUpdate();
    this.subscribeInputChange();
  }

  disconnect() {
    this.unsubscribeInputsChange();
  }

  inputTargetConnected = (input: HTMLInputElement) => {
    this.onUpdate();
    input.addEventListener("input", this.onUpdate);
  };

  inputTargetDisconnected = (input: HTMLInputElement) => {
    input.removeEventListener("input", this.onUpdate);
  };

  private getValues() {
    const values: Record<string, any> = {};
    this.inputTargets.forEach((input) => {
      switch (input.type) {
        case "checkbox":
          values[input.name] = input.checked;
          break;
        case "radio":
          if (input.checked) {
            values[input.name] = input.value;
          }
          break;
        default:
          values[input.name] = input.value;
      }
    });
    return values;
  }

  private subscribeInputChange = () => {
    this.inputTargets.forEach((input) =>
      input.addEventListener("input", this.onUpdate)
    );
  };

  private unsubscribeInputsChange = () => {
    this.inputTargets.forEach((input) =>
      input.removeEventListener("input", this.onUpdate)
    );
  };

  private onUpdate = () => {
    const values = this.getValues();
    const changed = Object.keys(values).some(
      (key) => this.initialValues[key] !== values[key]
    );

    this.submitTarget.disabled = !changed;
  };
}
