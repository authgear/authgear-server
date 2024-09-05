import "@tabler/icons/iconfont/tabler-icons.min.css";
import "cropperjs/dist/cropper.min.css";
import "intl-tel-input/build/js/utils.js";
import "intl-tel-input/build/css/intlTelInput.min.css";

import { start } from "@hotwired/turbo";
import { Application, Controller } from "@hotwired/stimulus";
import axios from "axios";
import { CopyButtonController } from "./copy";
import { PasswordVisibilityToggleController } from "./passwordVisibility";
import { PasswordPolicyController } from "./password-policy";
import { ClickToSwitchController } from "./clickToSwitch";
import { PreventDoubleTapController } from "./preventDoubleTap";
import { AccountDeletionController } from "./accountdeletion";
import { ResendButtonController } from "./resendButton";
import { MessageBarController } from "./messageBar";
import { IntlTelInputController } from "./intlTelInput";
import { SelectEmptyValueController, GenderSelectController } from "./select";
import { ImagePickerController } from "./imagepicker";
import { WebSocketController } from "./websocket";
import { AuthflowWebsocketController } from "./authflow_websocket";
import { AuthflowPollingController } from "./authflow_polling";
import {
  FormatDateRelativeController,
  FormatInputDateController,
} from "./date";
import {
  XHRSubmitFormController,
  RestoreFormController,
  RetainFormFormController,
  RetainFormInputController,
} from "./form";
import { ModalController } from "./modal";
import { BackButtonController } from "./back";
import { SimpleModalController } from "./simpleModal";
import {
  PasskeyCreationController,
  PasskeyRequestController,
  AuthflowPasskeyRequestController,
  AuthflowPasskeyCreationController,
  PasskeyAutofillController,
  AuthflowPasskeyErrorController,
} from "./passkey";
import { WalletConfirmationController, WalletIconController } from "./web3";
import { init as SentryInit } from "@sentry/browser";
import { BrowserTracing } from "@sentry/tracing";
import { LockoutController } from "./lockout";
import { MirrorButtonController } from "./mirrorbutton";
// FIXME(css): Build CSS files one by one with another tool
// webpack bundles all CSS files into one bundle.

axios.defaults.withCredentials = true;

const sentryDSN = document
  .querySelector("meta[name=x-sentry-dsn]")
  ?.getAttribute("content");
if (sentryDSN != null && sentryDSN !== "") {
  SentryInit({
    dsn: sentryDSN,
    integrations: [new BrowserTracing()],
    // Do not enable performance monitoring.
    // tracesSampleRate: 0,
  });
}

start();

const Stimulus = Application.start();

Stimulus.register("xhr-submit-form", XHRSubmitFormController);

Stimulus.register(
  "password-visibility-toggle",
  PasswordVisibilityToggleController
);
Stimulus.register("password-policy", PasswordPolicyController);
Stimulus.register("click-to-switch", ClickToSwitchController);

Stimulus.register("copy-button", CopyButtonController);

Stimulus.register("prevent-double-tap", PreventDoubleTapController);

Stimulus.register("account-deletion", AccountDeletionController);

Stimulus.register("resend-button", ResendButtonController);

Stimulus.register("message-bar", MessageBarController);

Stimulus.register("intl-tel-input", IntlTelInputController);

Stimulus.register("select-empty-value", SelectEmptyValueController);
Stimulus.register("gender-select", GenderSelectController);

Stimulus.register("image-picker", ImagePickerController);

Stimulus.register("websocket", WebSocketController);
Stimulus.register("authflow-websocket", AuthflowWebsocketController);
Stimulus.register("authflow-polling", AuthflowPollingController);

Stimulus.register("format-date-relative", FormatDateRelativeController);
Stimulus.register("format-input-date", FormatInputDateController);

Stimulus.register("restore-form", RestoreFormController);
Stimulus.register("retain-form-form", RetainFormFormController);
Stimulus.register("retain-form-input", RetainFormInputController);

Stimulus.register("modal", ModalController);
Stimulus.register("simple-modal", SimpleModalController);

Stimulus.register("back-button", BackButtonController);

Stimulus.register("passkey-creation", PasskeyCreationController);
Stimulus.register("passkey-request", PasskeyRequestController);
Stimulus.register("passkey-autofill", PasskeyAutofillController);
Stimulus.register("authflow-passkey-request", AuthflowPasskeyRequestController);
Stimulus.register(
  "authflow-passkey-creation",
  AuthflowPasskeyCreationController
);
Stimulus.register("authflow-passkey-error", AuthflowPasskeyErrorController);

Stimulus.register("web3-wallet-confirmation", WalletConfirmationController);
Stimulus.register("web3-wallet-icon", WalletIconController);

Stimulus.register("lockout", LockoutController);

Stimulus.register("mirror-button", MirrorButtonController);
