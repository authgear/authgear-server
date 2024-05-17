// /* global process */
import React from "react";
import { render } from "react-dom";
import { initializeIcons, registerIcons } from "@fluentui/react";
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

render(<ReactApp />, document.getElementById("react-app-root"));
