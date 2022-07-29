import("@authgear/web").finally(() => {});
import("@monaco-editor/react").finally(() => {});
import("axios").finally(() => {});
import("base64-js").finally(() => {});
import("chart.js").finally(() => {});
import("chartjs-plugin-datalabels").finally(() => {});
import("classnames").finally(() => {});
import("cropperjs").finally(() => {});
import("deep-equal").finally(() => {});
import("history").finally(() => {});
import("i18n-iso-countries").finally(() => {});
import("immer").finally(() => {});
import("intl-tel-input").finally(() => {});
import("js-yaml").finally(() => {});
import("luxon").finally(() => {});
import("monaco-editor").finally(() => {});
import("postcss").finally(() => {});
import("react").finally(() => {});
import("react-chartjs-2").finally(() => {});
import("react-code-blocks").finally(() => {});
import("tzdata").finally(() => {});
import("uuid").finally(() => {});
import("zxcvbn").finally(() => {});

if (process.env.NODE_ENV === "production") {
  import("@apollo/client").finally(() => {});
  import("@fluentui/react").finally(() => {});
  import("@fluentui/react-hooks").finally(() => {});
  import("@oursky/react-messageformat").finally(() => {});
  import("react-dom").finally(() => {});
  import("react-helmet-async").finally(() => {});
  import("react-router-dom").finally(() => {});
}
