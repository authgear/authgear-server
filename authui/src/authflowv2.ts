import { start } from "@hotwired/turbo";
import { Application } from "@hotwired/stimulus";
import axios from "axios";
import { init as SentryInit } from "@sentry/browser";
import { BrowserTracing } from "@sentry/tracing";
import {
  RestoreFormController,
  RetainFormFormController,
  RetainFormInputController,
  XHRSubmitFormController,
} from "./form";
import { PreventDoubleTapController } from "./preventDoubleTap";
import { LockoutController } from "./lockout";
import { FormatDateRelativeController } from "./date";
import { injectCSSAttrs } from "./cssattrs";

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
Stimulus.register("restore-form", RestoreFormController);
Stimulus.register("retain-form-form", RetainFormFormController);
Stimulus.register("retain-form-input", RetainFormInputController);

Stimulus.register("prevent-double-tap", PreventDoubleTapController);

Stimulus.register("lockout", LockoutController);

Stimulus.register("format-date-relative", FormatDateRelativeController);

injectCSSAttrs(document.documentElement);
