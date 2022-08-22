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
import {
  FormatDateRelativeController,
  FormatInputDateController,
} from "./date";
import { TransferClickController } from "./click";
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
  PasskeyAutofillController,
} from "./passkey";
// FIXME(css): Build CSS files one by one with another tool
// webpack bundles all CSS files into one bundle.

axios.defaults.withCredentials = true;

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

Stimulus.register("format-date-relative", FormatDateRelativeController);
Stimulus.register("format-input-date", FormatInputDateController);

Stimulus.register("transfer-click", TransferClickController);

Stimulus.register("restore-form", RestoreFormController);
Stimulus.register("retain-form-form", RetainFormFormController);
Stimulus.register("retain-form-input", RetainFormInputController);

Stimulus.register("modal", ModalController);
Stimulus.register("simple-modal", SimpleModalController);

Stimulus.register("back-button", BackButtonController);

Stimulus.register("passkey-creation", PasskeyCreationController);
Stimulus.register("passkey-request", PasskeyRequestController);
Stimulus.register("passkey-autofill", PasskeyAutofillController);
