import { Controller } from "@hotwired/stimulus";

enum LoginMethod {
  Email = "Email",
  Phone = "Phone",
  Username = "username",
}

interface PreviewWidgetViewModel {
  loginInput: "email" | "username" | "phone" | "none";
  branches: LoginMethod[];
}

export class PreviewWidgetController extends Controller {
  static values = {
    loginMethods: { type: Array, default: [] },
  };

  declare loginMethodsValue: LoginMethod[];

  static targets = [
    "emailInput",
    "usernameInput",
    "phoneInput",
    "loginIDSection",
    "branchSection",
    "branchOptionUsername",
    "branchOptionPhone",
    "noLoginMethodsError",
  ];

  declare emailInputTarget: HTMLElement;
  declare usernameInputTarget: HTMLElement;
  declare phoneInputTarget: HTMLElement;
  declare loginIDSectionTarget: HTMLElement;
  declare branchSectionTarget: HTMLElement;
  declare branchOptionUsernameTarget: HTMLElement;
  declare branchOptionPhoneTarget: HTMLElement;
  declare noLoginMethodsErrorTarget: HTMLElement;

  connect() {
    this.loginMethodsValueChanged();
  }

  loginMethodsValueChanged() {
    const loginMethodsSet = new Set(this.loginMethodsValue);
    const loginInput = loginMethodsSet.has(LoginMethod.Email)
      ? "email"
      : loginMethodsSet.has(LoginMethod.Username)
      ? "username"
      : loginMethodsSet.has(LoginMethod.Phone)
      ? "phone"
      : "none";

    const remainingMethods = new Set(this.loginMethodsValue);
    switch (loginInput) {
      case "email":
        remainingMethods.delete(LoginMethod.Email);
        break;
      case "username":
        remainingMethods.delete(LoginMethod.Username);
        break;
      case "phone":
        remainingMethods.delete(LoginMethod.Phone);
        break;
      default:
        break;
    }

    const viewModel: PreviewWidgetViewModel = {
      loginInput,
      branches: this.loginMethodsValue.filter((method) =>
        remainingMethods.has(method)
      ),
    };

    // Hide irrelevant elements
    this.updateElements(viewModel);
  }

  private updateElements(vm: PreviewWidgetViewModel) {
    showElementIf(this.emailInputTarget, vm.loginInput === "email");
    showElementIf(this.usernameInputTarget, vm.loginInput === "username");
    showElementIf(this.phoneInputTarget, vm.loginInput === "phone");
    showElementIf(this.loginIDSectionTarget, vm.loginInput !== "none");
    showElementIf(this.branchSectionTarget, vm.branches.length > 0);
    showElementIf(
      this.branchOptionUsernameTarget,
      vm.branches.includes(LoginMethod.Username)
    );
    showElementIf(
      this.branchOptionPhoneTarget,
      vm.branches.includes(LoginMethod.Phone)
    );
    showElementIf(
      this.noLoginMethodsErrorTarget,
      vm.loginInput === "none" && vm.branches.length === 0
    );
  }
}

function showElementIf(el: HTMLElement, condition: boolean) {
  if (condition) {
    el.classList.remove("hidden");
  } else {
    el.classList.add("hidden");
  }
}
