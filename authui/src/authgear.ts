import Turbolinks from "turbolinks";
import { Application } from "@hotwired/stimulus";
import axios from "axios";
import { init } from "./core";
import { setupIntlTelInput } from "./intlTelInput";
import {
  clickLinkSubmitForm,
  autoSubmitForm,
  xhrSubmitForm,
  restoreForm,
} from "./form";
import { setupSelectEmptyValue, setupGenderSelect } from "./select";
import { formatDateRelative, formatInputDate } from "./date";
import { setupImagePicker } from "./imagepicker";
import { setupWebsocket } from "./websocket";
import { setupModal } from "./modal";
import { CopyButtonController } from "./copy";
import { PasswordVisibilityToggleController } from "./passwordVisibility";
import { PasswordPolicyController } from "./password-policy";
import { ClickToSwitchController } from "./clickToSwitch";
import { PreventDoubleTapController } from "./preventDoubleTap";
import { AccountDelectionController } from "./accountdeletion";
import { ResendButtonController } from "./resendButton";
import { MessageBarController } from "./messageBar";
// FIXME(css): Build CSS files one by one with another tool
// webpack bundles all CSS files into one bundle.

axios.defaults.withCredentials = true;

init();

const Stimulus = Application.start();
Stimulus.register(
  "password-visibility-toggle",
  PasswordVisibilityToggleController
);
Stimulus.register("password-policy", PasswordPolicyController);
Stimulus.register("click-to-switch", ClickToSwitchController);

Stimulus.register("copy-button", CopyButtonController);

Stimulus.register("prevent-double-tap", PreventDoubleTapController);

Stimulus.register("account-delection", AccountDelectionController);

Stimulus.register("resend-button", ResendButtonController);

Stimulus.register("message-bar", MessageBarController);

window.api.onLoad(() => {
  document.body.classList.add("js");
});

window.api.onLoad(setupIntlTelInput);

window.api.onLoad(setupSelectEmptyValue);
window.api.onLoad(setupGenderSelect);

window.api.onLoad(formatDateRelative);
window.api.onLoad(formatInputDate);

window.api.onLoad(setupImagePicker);

window.api.onLoad(setupWebsocket);

window.api.onLoad(setupModal);

window.api.onLoad(autoSubmitForm);
window.api.onLoad(clickLinkSubmitForm);
window.api.onLoad(xhrSubmitForm);
window.api.onLoad(restoreForm);
