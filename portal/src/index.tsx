// /* global process */
import "normalize.css";
import "./index.css";
import "intl-tel-input/build/css/intlTelInput.css";
import "intl-tel-input/build/js/utils.js";
import "cropperjs/dist/cropper.min.css";

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

import ReactApp from "./ReactApp";

initializeIcons();

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
