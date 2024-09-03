import { start } from "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";
import axios from "axios";
import { init as SentryInit } from "@sentry/browser";
import { BrowserTracing } from "@sentry/tracing";
import {
  RestoreFormController,
  RetainFormFormController,
  RetainFormInputController,
} from "./form";
import { TurboFormController } from "./authflowv2/turboForm";
import { LoadingController } from "./authflowv2/loading";
import { PreventDoubleTapController } from "./preventDoubleTap";
import { LockoutController } from "./lockout";
import { FormatDateRelativeController } from "./date";
import { injectCSSAttrs } from "./cssattrs";
import { ResendButtonController } from "./resendButton";
import { OtpInputController } from "./authflowv2/otpInput";
import { PasswordVisibilityToggleController } from "./passwordVisibility";
import { PasswordPolicyController } from "./authflowv2/password-policy";
import { PasswordStrengthMeterController } from "./authflowv2/password-strength-meter";
import { PhoneInputController } from "./authflowv2/phoneInput";
import { CustomSelectController } from "./authflowv2/customSelect";
import { CountdownController } from "./countdown";
import { TextFieldController } from "./authflowv2/text-field";
import { OverlayController } from "./authflowv2/overlay";
import { CopyButtonController } from "./copy";
import { AuthflowWebsocketController } from "./authflow_websocket";
import { AuthflowPollingController } from "./authflow_polling";
import {
  AuthflowPasskeyRequestController,
  AuthflowPasskeyCreationController,
  AuthflowV2PasskeyErrorController,
} from "./passkey";
import { NewPasswordFieldController } from "./authflowv2/new-password-field";
import { AlertMessageController } from "./authflowv2/alert-message";
import { DismissKeyboardOnScrollController } from "./authflowv2/dismissKeyboard";
import { BodyScrollLockController } from "./authflowv2/bodyScrollLock";
import { ClickToSwitchController } from "./clickToSwitch";
import { InlinePreviewController } from "./inline-preview";
import { PreviewableResourceController } from "./previewable-resource";
import { CloudflareTurnstileController } from "./authflowv2/botprotection/cloudflareTurnstile";
import { RecaptchaV2Controller } from "./authflowv2/botprotection/recaptchav2";
import { BotProtectionTokenInputController } from "./authflowv2/botprotection/botProtectionTokenInput";
import { BotProtectionStandalonePageSubmitBtnController } from "./authflowv2/botprotection/botProtectionStandalonePageSubmitBtn";
import { BotProtectionController } from "./authflowv2/botprotection/botProtection";
import { BotProtectionDialogController } from "./authflowv2/botprotection/botProtectionDialog";
import { DialogController } from "./authflowv2/dialog";
import { BotProtectionStandalonePageController } from "./authflowv2/botprotection/botProtectionStandalonePage";
import { ImagePickerController } from "./imagepicker";
import { LocaleInputController } from "./authflowv2/localeInput";
import { TimezoneInputController } from "./authflowv2/timezoneInput";
import { CountryInputController } from "./authflowv2/countryInput";

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

Stimulus.register("turbo-form", TurboFormController);
Stimulus.register("restore-form", RestoreFormController);
Stimulus.register("retain-form-form", RetainFormFormController);
Stimulus.register("retain-form-input", RetainFormInputController);

Stimulus.register("prevent-double-tap", PreventDoubleTapController);

Stimulus.register("lockout", LockoutController);

Stimulus.register("format-date-relative", FormatDateRelativeController);
Stimulus.register("format-date-relative", FormatDateRelativeController);
Stimulus.register(
  "password-visibility-toggle",
  PasswordVisibilityToggleController
);

Stimulus.register("otp-input", OtpInputController);
Stimulus.register("resend-button", ResendButtonController);
Stimulus.register("password-policy", PasswordPolicyController);
Stimulus.register("password-strength-meter", PasswordStrengthMeterController);
Stimulus.register("custom-select", CustomSelectController);
Stimulus.register("phone-input", PhoneInputController);
Stimulus.register("locale-input", LocaleInputController);
Stimulus.register("countdown", CountdownController);
Stimulus.register("copy-button", CopyButtonController);
Stimulus.register("image-picker", ImagePickerController);

Stimulus.register("text-field", TextFieldController);
Stimulus.register("dialog", DialogController);
Stimulus.register("overlay", OverlayController);
Stimulus.register("loading", LoadingController);
Stimulus.register("new-password-field", NewPasswordFieldController);

Stimulus.register("authflow-websocket", AuthflowWebsocketController);
Stimulus.register("authflow-polling", AuthflowPollingController);
Stimulus.register("authflow-passkey-request", AuthflowPasskeyRequestController);
Stimulus.register(
  "authflow-passkey-creation",
  AuthflowPasskeyCreationController
);
Stimulus.register("authflow-passkey-error", AuthflowV2PasskeyErrorController);
Stimulus.register("alert-message", AlertMessageController);
Stimulus.register(
  "dismiss-keyboard-on-scroll",
  DismissKeyboardOnScrollController
);
Stimulus.register("body-scroll-lock", BodyScrollLockController);
Stimulus.register("click-to-switch", ClickToSwitchController);
Stimulus.register("inline-preview", InlinePreviewController);
Stimulus.register("previewable-resource", PreviewableResourceController);
Stimulus.register(
  "bot-protection-token-input",
  BotProtectionTokenInputController
);
Stimulus.register(
  "bot-protection-standalone-page",
  BotProtectionStandalonePageController
);
Stimulus.register(
  "bot-protection-standalone-page-submit-btn",
  BotProtectionStandalonePageSubmitBtnController
);
Stimulus.register("cloudflare-turnstile", CloudflareTurnstileController);
Stimulus.register("recaptcha-v2", RecaptchaV2Controller);
Stimulus.register("bot-protection", BotProtectionController);
Stimulus.register("bot-protection-dialog", BotProtectionDialogController);
Stimulus.register("timezone-input", TimezoneInputController);
Stimulus.register("country-input", CountryInputController);

injectCSSAttrs(document.documentElement);
