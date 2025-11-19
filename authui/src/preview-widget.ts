import { Controller } from "@hotwired/stimulus";

export enum LoginMethod {
  Email = "Email",
  Phone = "Phone",
  Username = "Username",
  Google = "Google",
  Apple = "Apple",
  Facebook = "Facebook",
  Github = "Github",
  LinkedIn = "LinkedIn",
  MicrosoftEntraID = "MicrosoftEntraID",
  MicrosoftADFS = "MicrosoftADFS",
  MicrosoftAzureADB2C = "MicrosoftAzureADB2C",
  WechatWeb = "WechatWeb",
  WechatMobile = "WechatMobile",
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
    "branchOptionGoogle",
    "branchOptionApple",
    "branchOptionFacebook",
    "branchOptionGithub",
    "branchOptionLinkedin",
    "branchOptionAzureadv2",
    "branchOptionAdfs",
    "branchOptionAzureadb2c",
    "branchOptionWechat",
    "noLoginMethodsError",
  ];

  declare emailInputTarget: HTMLElement;
  declare usernameInputTarget: HTMLElement;
  declare phoneInputTarget: HTMLElement;
  declare loginIDSectionTarget: HTMLElement;
  declare branchSectionTarget: HTMLElement;
  declare branchOptionUsernameTarget: HTMLElement;
  declare branchOptionPhoneTarget: HTMLElement;
  declare branchOptionGoogleTarget: HTMLElement;
  declare branchOptionAppleTarget: HTMLElement;
  declare branchOptionFacebookTarget: HTMLElement;
  declare branchOptionGithubTarget: HTMLElement;
  declare branchOptionLinkedinTarget: HTMLElement;
  declare branchOptionAzureadv2Target: HTMLElement;
  declare branchOptionAdfsTarget: HTMLElement;
  declare branchOptionAzureadb2cTarget: HTMLElement;
  declare branchOptionWechatTarget: HTMLElement;
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
      this.branchOptionGoogleTarget,
      vm.branches.includes(LoginMethod.Google)
    );
    showElementIf(
      this.branchOptionAppleTarget,
      vm.branches.includes(LoginMethod.Apple)
    );
    showElementIf(
      this.branchOptionFacebookTarget,
      vm.branches.includes(LoginMethod.Facebook)
    );
    showElementIf(
      this.branchOptionGithubTarget,
      vm.branches.includes(LoginMethod.Github)
    );
    showElementIf(
      this.branchOptionLinkedinTarget,
      vm.branches.includes(LoginMethod.LinkedIn)
    );
    showElementIf(
      this.branchOptionAzureadv2Target,
      vm.branches.includes(LoginMethod.MicrosoftEntraID)
    );
    showElementIf(
      this.branchOptionAdfsTarget,
      vm.branches.includes(LoginMethod.MicrosoftADFS)
    );
    showElementIf(
      this.branchOptionAzureadb2cTarget,
      vm.branches.includes(LoginMethod.MicrosoftAzureADB2C)
    );
    showElementIf(
      this.branchOptionWechatTarget,
      vm.branches.includes(LoginMethod.WechatWeb) ||
        vm.branches.includes(LoginMethod.WechatMobile)
    );
    showElementIf(
      this.noLoginMethodsErrorTarget,
      vm.loginInput === "none" && vm.branches.length === 0
    );

    if (vm.branches.length > 0 && vm.loginInput === "none") {
      this.branchSectionTarget.classList.add(
        "preview-widget__branch-section--branch-only"
      );
    } else {
      this.branchSectionTarget.classList.remove(
        "preview-widget__branch-section--branch-only"
      );
    }
  }
}

function showElementIf(el: HTMLElement, condition: boolean) {
  if (condition) {
    el.classList.remove("hidden");
  } else {
    el.classList.add("hidden");
  }
}
