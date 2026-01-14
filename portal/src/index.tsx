// /* global process */
import "intl-tel-input/build/js/utils.js";
import "intl-tel-input/build/css/intlTelInput.css";
import "cropperjs/dist/cropper.min.css";
import "@tabler/icons/iconfont/tabler-icons.min.css";
import "@fortawesome/fontawesome-free/css/all.min.css";

import React from "react";
import { createRoot } from "react-dom/client";
import { initializeIcons, registerIcons } from "@fluentui/react";

// See below for details.
// Monaco editor initialization imports - Start
import { loader } from "@monaco-editor/react";
import * as monaco from "monaco-editor";
// @ts-expect-error
import editorWorker from "monaco-editor/esm/vs/editor/editor.worker?worker";
// @ts-expect-error
import jsonWorker from "monaco-editor/esm/vs/language/json/json.worker?worker";
// @ts-expect-error
import cssWorker from "monaco-editor/esm/vs/language/css/css.worker?worker";
// @ts-expect-error
import htmlWorker from "monaco-editor/esm/vs/language/html/html.worker?worker";
// @ts-expect-error
import tsWorker from "monaco-editor/esm/vs/language/typescript/ts.worker?worker";
// Monaco editor initialization imports - End

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Tooltip,
  PointElement,
  LineElement,
  ArcElement,
} from "chart.js";
import { setAutoFreeze } from "immer";
import { Settings } from "luxon";
import ReactApp from "./ReactApp";
import { Cookies16Regular } from "@fluentui/react-icons";

Settings.throwOnInvalid = true;
// Tell typescript that we expect luxon to always return something or throw, instead of returning null | something.
// https://github.com/DefinitelyTyped/DefinitelyTyped/pull/64995/files#diff-dc498e8eceb5d6d1bb58b4a9933293385301cca36a3810e765af2fc7861fe67aR66
declare module "luxon" {
  export interface TSSettings {
    throwOnInvalid: true;
  }
}

initializeIcons();
registerIcons({
  icons: {
    Cookies: <Cookies16Regular />,
  },
});

// We sometimes use immer in forms.
// It seems that frozen object is problematic if we use
// produce more than once.
// Interestingly, this bug only happens in production build :(
// https://github.com/authgear/authgear-server/issues/2561
setAutoFreeze(false);

// ChartJS registration for Bar chart in the AnalyticsActivityWidget
ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip);

// ChartJS registration for Line chart in the AnalyticsActivityWidget
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip
);

// ChartJS registration for Pie chart in the AnalyticsSignupConversionWidget
// and AnalyticsSignupMethodsWidget
ChartJS.register(ArcElement, Tooltip);

// See https://github.com/microsoft/monaco-editor/blob/main/docs/integrate-esm.md#using-vite
// See https://github.com/suren-atoyan/monaco-react?tab=readme-ov-file#use-monaco-editor-as-an-npm-package
//
// By using this approach, it is now our own responsibility to keep monaco-editor and @monaco-editor/react compatible.
// @monaco-editor/react uses @monaco-editor/loader, and @monaco-editor/loader uses a specific version of monaco-editor.
// So when you need to update them, you do
// 1. Pick a version of @monaco-editor/react.
// 2. Inspect @monaco-editor/loader in package-lock.json to see the actual version of @monaco-editor/loader
// 3. Inspect https://github.com/suren-atoyan/monaco-loader/blob/master/src/config/index.js to see what version of monaco-editor it is using.
// 4. Install the same version of monaco-editor.
//
// Note that monaco-editor has some breaking changes in 0.45.0
// See https://github.com/microsoft/monaco-editor/blob/main/CHANGELOG.md#0450
// So the highest version we can use is 0.44.0, until @monaco-editor/react supports >= 0.45.0
window.MonacoEnvironment = {
  getWorker(_, label) {
    switch (label) {
      case "json":
        return new jsonWorker();
      case "css":
      case "scss":
      case "less":
        return new cssWorker();
      case "html":
      case "handlebars":
      case "razor":
        return new htmlWorker();
      case "javascript":
      case "typescript":
        return new tsWorker();
    }
    return new editorWorker();
  },
};
loader.config({ monaco });
loader.init().then(() => {
  const container = document.getElementById("react-app-root");
  if (container) {
    const root = createRoot(container);
    root.render(<ReactApp />);
  }
});
