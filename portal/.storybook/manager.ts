import { addons } from "storybook/manager-api";
import { create } from "storybook/theming/create";

const AUTHGEAR_LOGO_SVG =
  "https://cdn.prod.website-files.com/60658b46b03f0cf83ac1485d/619e6607eb647619cecee2cf_authgear-logo.svg";

addons.setConfig({
  theme: create({
    base: "light",
    brandTitle: "Authgear",
    brandUrl: "https://www.authgear.com",
    brandImage: AUTHGEAR_LOGO_SVG,
  }),
});
