import Turbolinks from "turbolinks";
import axios from "axios";
import { init } from "./core";
import { setupIntlTelInput } from "./intlTelInput";
import { setupPasswordPolicy } from "./password-policy";
import {
  clickLinkSubmitForm,
  autoSubmitForm,
  xhrSubmitForm,
  restoreForm,
} from "./form";
import { setupSelectEmptyValue, setupGenderSelect } from "./select";
import { formatDateRelative, formatInputDate } from "./date";
import { setupAccountDeletion } from "./accountdeletion";
import { setupImagePicker } from "./imagepicker";
import { setupWebsocket } from "./websocket";
import { setupModal } from "./modal";
import { setupClickToSwitch } from "./clickToSwitch";
import { setupResendButton } from "./resendButton";
import { setupPreventDoubleTap } from "./preventDoubleTap";
import { setupCopyButton } from "./copy";
import { setupMessageBar } from "./messageBar";
import { setupPasswordVisibilityToggle } from "./passwordVisibility";
// FIXME(css): Build CSS files one by one with another tool
// webpack bundles all CSS files into one bundle.

axios.defaults.withCredentials = true;

init();

window.api.onLoad(() => {
  document.body.classList.add("js");
});

window.api.onLoad(setupPasswordPolicy);

window.api.onLoad(setupIntlTelInput);

window.api.onLoad(setupSelectEmptyValue);
window.api.onLoad(setupGenderSelect);

window.api.onLoad(formatDateRelative);
window.api.onLoad(formatInputDate);

window.api.onLoad(setupAccountDeletion);

window.api.onLoad(setupImagePicker);

window.api.onLoad(setupPreventDoubleTap);

window.api.onLoad(setupWebsocket);

window.api.onLoad(setupModal);

window.api.onLoad(setupClickToSwitch);

window.api.onLoad(setupCopyButton);

window.api.onLoad(autoSubmitForm);
window.api.onLoad(clickLinkSubmitForm);
window.api.onLoad(xhrSubmitForm);
window.api.onLoad(restoreForm);

window.api.onLoad(setupResendButton);

window.api.onLoad(setupMessageBar);

window.api.onLoad(setupPasswordVisibilityToggle);
