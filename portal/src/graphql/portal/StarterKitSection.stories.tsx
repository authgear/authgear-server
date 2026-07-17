import type { Meta, StoryObj } from "@storybook/react-vite";
import { StarterKitSection } from "./StarterKitSection";
import type { StarterKit } from "./CreateOAuthClientScreen/frameworks";

const KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-react",
  downloadUrl: "https://github.com/authgear/authgear-example-react/archive/HEAD.zip",
  redirectURI: "http://localhost:4000/auth-redirect",
  homepageUrl: "http://localhost:4000",
  config: {
    format: "dotenv",
    fileName: ".env",
    vars: [
      { key: "VITE_AUTHGEAR_CLIENT_ID", token: "clientID" },
      { key: "VITE_AUTHGEAR_ENDPOINT", token: "endpoint" },
      { key: "VITE_AUTHGEAR_REDIRECT_URL", token: "redirectURI" },
    ],
  },
  installCmd: "npm i",
  startCmd: "npm start",
  guideUrl: "https://docs.authgear.com/tutorials/spa/react",
};

const JS_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-spa-js",
  downloadUrl:
    "https://github.com/authgear/authgear-example-spa-js/archive/HEAD.zip",
  redirectURI: "http://localhost:3000/",
  homepageUrl: "http://localhost:3000",
  config: {
    format: "js",
    fileName: "public/app.js",
    vars: [
      { key: "AUTHGEAR_CLIENT_ID", token: "clientID" },
      { key: "AUTHGEAR_ENDPOINT", token: "endpoint" },
    ],
  },
  installCmd: "npm install",
  startCmd: "npm run dev",
  guideUrl: "https://docs.authgear.com/get-started/single-page-app/website",
};

const meta = {
  title: "portal/StarterKitSection",
  component: StarterKitSection,
  args: {
    starterKit: KIT,
    frameworkDisplayName: "React",
    clientID: "02b48e0ec3cf48e4bae3fcfada73239e",
    publicOrigin: "https://demo.authgear.cloud",
    usersPath: "/project/demo/users",
    redirectURIIsSet: false,
    saving: false,
    onSetRedirectURI: () => {},
    onGoToSettings: () => {},
  },
} satisfies Meta<typeof StarterKitSection>;

export default meta;
type Story = StoryObj<typeof meta>;

export const NotSet: Story = {};

export const RedirectURISet: Story = {
  args: {
    redirectURIIsSet: true,
  },
};

export const JsConfig: Story = {
  args: {
    starterKit: JS_KIT,
    frameworkDisplayName: "Other SPAs",
  },
};
