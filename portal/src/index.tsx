// /* global process */
import React from "react";
import { render } from "react-dom";
import { initializeIcons } from "@fluentui/react";
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

import ReactApp from "./ReactApp";

initializeIcons();
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
