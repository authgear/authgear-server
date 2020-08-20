// This following 2 lines are extremely important.
// Since we do not provide our babel.config.json,
// Parcel provides a default one for us.
// The default config uses preset-env with useBuiltins: entry.
// Therefore, we have to include the following imports to
// let Babel rewrite them into polyfill imports according to .browserslistrc.
import "core-js/stable";
import "regenerator-runtime/runtime";

import "normalize.css";

import "./index.scss";

import React from "react";
import { render } from "react-dom";

import App from "./App";

render(<App />, document.getElementById("react-app-root"));
