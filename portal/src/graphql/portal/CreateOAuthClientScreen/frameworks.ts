import type { Framework, ApplicationType } from "../../../types";

export type FrameworkSection = "website" | "mobile" | "integration";
export type Stage2Need = "none" | "token-or-cookie";
export type AuthMethodChoice = "token" | "cookie";

export type ConfigValueToken =
  | "clientID"
  | "endpoint"
  | "redirectURI"
  | "literal";

/**
 * How the starter kit's config values are rendered:
 * - "dotenv": `KEY=value` lines for a `.env` file.
 * - "js": `const KEY = "value";` lines to paste into a source file.
 * - "swift": `static let key = "value"` lines to paste into a Swift file.
 * - "kotlin": `const val KEY = "value"` lines to paste into a Kotlin file.
 */
export type StarterKitConfigFormat = "dotenv" | "js" | "swift" | "kotlin";

export interface StarterKitConfigVar {
  /** The variable/constant name, e.g. "VITE_AUTHGEAR_CLIENT_ID". */
  key: string;
  /** Which live value to substitute for this variable. */
  token: ConfigValueToken;
  /** Static value to render when `token` is "literal" (e.g. a placeholder the user replaces). */
  literalValue?: string;
}

export interface StarterKitConfig {
  /** Rendering format for the config block. */
  format: StarterKitConfigFormat;
  /** File the user edits, e.g. ".env" or "public/app.js". */
  fileName: string;
  /** Ordered config variables to render. */
  vars: StarterKitConfigVar[];
}

export interface StarterKitMobileRun {
  /** Command to build the web assets, e.g. "npm run build". */
  buildCmd: string;
  /** Command to sync assets to native platforms, e.g. "npx cap sync". */
  syncCmd: string;
  /** Command to open the iOS project, e.g. "npx cap open ios". */
  iosCmd: string;
  /** Command to open the Android project, e.g. "npx cap open android". */
  androidCmd: string;
}

export interface StarterKit {
  /** GitHub repo page. */
  repoUrl: string;
  /** Archive zip download URL. */
  downloadUrl: string;
  /**
   * Redirect URIs the starter kit expects; all are authorized in one click.
   * The first entry is used for the `redirectURI` config token.
   */
  redirectURIs: string[];
  /** Local dev homepage to visit after starting the app; omit for device-only apps. */
  homepageUrl?: string;
  /** How and where the app is configured. */
  config: StarterKitConfig;
  /** Install command, e.g. "npm install"; omit for kits with no install step (e.g. SPM). */
  installCmd?: string;
  /** Start command, e.g. "npm run dev"; omit for IDE-run kits. */
  startCmd?: string;
  /** IDE to open and run the project in (e.g. "Xcode"), for kits with no CLI run command. */
  ide?: string;
  /** Optional native build/run steps (for mobile/hybrid kits). */
  mobileRun?: StarterKitMobileRun;
  /** "Read <Framework> Guide" target. */
  guideUrl: string;
}

export interface CookieSnippet {
  /** Human-readable language label, e.g. "JavaScript", "Python". */
  language: string;
  /** Source code that reads x-authgear-* headers forwarded by nginx. */
  code: string;
}

/**
 * Application types a framework catalog entry can resolve to. M2M is
 * created exclusively through CreateM2MClientScreen, never through
 * a FrameworkEntry, so it is excluded here rather than handled as an
 * always-dead switch case at every call site.
 */
export type FrameworkApplicationType = Exclude<ApplicationType, "m2m">;

export interface FrameworkEntry {
  id: Framework;
  /** Message id for the framework's display name, e.g. "React". */
  displayNameMessageId: string;
  /** Message id for the framework's short helper text shown under its name. */
  helperTextMessageId: string;
  section: FrameworkSection;
  /** Tabler icon name without the "ti-" prefix, e.g. "brand-react". */
  iconName: string;
  /** Authgear docs URL for this framework's quick-start guide. */
  docLink: string;
  /**
   * Code snippet shown on the Quick Start tab of a Cookie SSO client to
   * demonstrate how this framework reads the x-authgear-* headers nginx
   * sets from the Authgear resolver response.
   */
  cookieSnippet?: CookieSnippet;
  stage2: Stage2Need;
  resolveType: (stage2?: AuthMethodChoice) => FrameworkApplicationType;
  compatibleTypes: FrameworkApplicationType[];
  /** Optional downloadable starter-kit walkthrough for this framework. */
  starterKit?: StarterKit;
}

const requireStage2 = (
  id: Framework,
  stage2?: AuthMethodChoice
): FrameworkApplicationType => {
  if (stage2 === "token") return "confidential";
  if (stage2 === "cookie") return "traditional_webapp";
  throw new Error(`resolveType called without stage2 on ${id}`);
};

const websiteSPA = (
  id: Framework,
  displayNameMessageId: string,
  helperTextMessageId: string,
  iconName: string,
  docLink: string,
  starterKit?: StarterKit
): FrameworkEntry => ({
  id,
  displayNameMessageId,
  helperTextMessageId,
  section: "website",
  iconName,
  docLink,
  stage2: "none",
  resolveType: () => "spa",
  compatibleTypes: ["spa"],
  starterKit,
});

const websiteServer = (
  id: Framework,
  displayNameMessageId: string,
  helperTextMessageId: string,
  iconName: string,
  docLink: string,
  cookieSnippet?: CookieSnippet
): FrameworkEntry => ({
  id,
  displayNameMessageId,
  helperTextMessageId,
  section: "website",
  iconName,
  docLink,
  cookieSnippet,
  stage2: "token-or-cookie",
  resolveType: (stage2) => requireStage2(id, stage2),
  compatibleTypes: ["confidential", "traditional_webapp"],
});

const EXPRESS_SNIPPET: CookieSnippet = {
  language: "JavaScript",
  code: `app.get("/protected", (req, res) => {
  const sessionValid = req.headers["x-authgear-session-valid"];
  const userId = req.headers["x-authgear-user-id"];
  if (sessionValid !== "true") return res.sendStatus(401);
  res.send(\`Hello, \${userId}\`);
});`,
};

const FLASK_SNIPPET: CookieSnippet = {
  language: "Python",
  code: `from flask import request, abort

@app.route("/protected")
def protected():
    if request.headers.get("X-Authgear-Session-Valid") != "true":
        abort(401)
    return f"Hello, {request.headers.get('X-Authgear-User-Id')}"`,
};

const LARAVEL_SNIPPET: CookieSnippet = {
  language: "PHP",
  code: `Route::get('/protected', function (Request $request) {
    if ($request->header('X-Authgear-Session-Valid') !== 'true') {
        abort(401);
    }
    return 'Hello, ' . $request->header('X-Authgear-User-Id');
});`,
};

const JAVA_SNIPPET: CookieSnippet = {
  language: "Java",
  code: `@GetMapping("/protected")
public ResponseEntity<String> protectedRoute(
    @RequestHeader(value = "X-Authgear-Session-Valid", required = false) String sessionValid,
    @RequestHeader(value = "X-Authgear-User-Id", required = false) String userId
) {
    if (!"true".equals(sessionValid)) {
        return ResponseEntity.status(401).build();
    }
    return ResponseEntity.ok("Hello, " + userId);
}`,
};

const ASPNET_SNIPPET: CookieSnippet = {
  language: "C#",
  code: `[HttpGet("protected")]
public IActionResult Protected()
{
    var sessionValid = Request.Headers["X-Authgear-Session-Valid"].ToString();
    var userId = Request.Headers["X-Authgear-User-Id"].ToString();
    if (sessionValid != "true") return Unauthorized();
    return Ok($"Hello, {userId}");
}`,
};

const mobileNative = (
  id: Framework,
  displayNameMessageId: string,
  helperTextMessageId: string,
  iconName: string,
  docLink: string,
  starterKit?: StarterKit
): FrameworkEntry => ({
  id,
  displayNameMessageId,
  helperTextMessageId,
  section: "mobile",
  iconName,
  docLink,
  stage2: "none",
  resolveType: () => "native",
  compatibleTypes: ["native"],
  starterKit,
});

const DOCS = "https://docs.authgear.com/get-started";

const VITE_DOTENV_VARS: StarterKitConfig = {
  format: "dotenv",
  fileName: ".env",
  vars: [
    { key: "VITE_AUTHGEAR_CLIENT_ID", token: "clientID" },
    { key: "VITE_AUTHGEAR_ENDPOINT", token: "endpoint" },
    { key: "VITE_AUTHGEAR_REDIRECT_URL", token: "redirectURI" },
  ],
};

const REACT_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-react",
  downloadUrl:
    "https://github.com/authgear/authgear-example-react/archive/HEAD.zip",
  redirectURIs: ["http://localhost:4000/auth-redirect"],
  homepageUrl: "http://localhost:4000",
  config: VITE_DOTENV_VARS,
  installCmd: "npm i",
  startCmd: "npm start",
  guideUrl: "https://docs.authgear.com/tutorials/spa/react",
};

const VUE_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-vue",
  downloadUrl:
    "https://github.com/authgear/authgear-example-vue/archive/HEAD.zip",
  redirectURIs: ["http://localhost:4000/auth-redirect"],
  homepageUrl: "http://localhost:4000",
  config: VITE_DOTENV_VARS,
  installCmd: "npm install",
  startCmd: "npm run dev",
  guideUrl: "https://docs.authgear.com/tutorials/spa/vue",
};

const NEXTJS_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-nextjs",
  downloadUrl:
    "https://github.com/authgear/authgear-example-nextjs/archive/HEAD.zip",
  redirectURIs: ["http://localhost:3000/api/auth/callback"],
  homepageUrl: "http://localhost:3000",
  config: {
    format: "dotenv",
    fileName: ".env.local",
    vars: [
      { key: "AUTHGEAR_CLIENT_ID", token: "clientID" },
      { key: "AUTHGEAR_ENDPOINT", token: "endpoint" },
      { key: "AUTHGEAR_REDIRECT_URI", token: "redirectURI" },
      {
        key: "SESSION_SECRET",
        token: "literal",
        literalValue: "a-random-string-of-at-least-32-characters",
      },
    ],
  },
  installCmd: "npm install",
  startCmd: "npm run dev",
  guideUrl: "https://docs.authgear.com/get-started/regular-web-app/nextjs",
};

const ANGULAR_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-angular",
  downloadUrl:
    "https://github.com/authgear/authgear-example-angular/archive/HEAD.zip",
  redirectURIs: ["http://localhost:4000/auth-redirect"],
  homepageUrl: "http://localhost:4000",
  config: {
    format: "dotenv",
    fileName: ".env",
    vars: [
      { key: "NG_APP_AUTHGEAR_CLIENT_ID", token: "clientID" },
      { key: "NG_APP_AUTHGEAR_ENDPOINT", token: "endpoint" },
    ],
  },
  installCmd: "npm install",
  startCmd: "npm start",
  guideUrl: "https://docs.authgear.com/get-started/single-page-app/angular",
};

const OTHER_SPA_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-spa-js",
  downloadUrl:
    "https://github.com/authgear/authgear-example-spa-js/archive/HEAD.zip",
  redirectURIs: ["http://localhost:3000/"],
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

const ANDROID_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-android",
  downloadUrl:
    "https://github.com/authgear/authgear-example-android/archive/HEAD.zip",
  redirectURIs: ["com.example.authgeardemo://host/path"],
  config: {
    format: "kotlin",
    fileName: "app/src/main/java/com/example/authgeardemo/Constants.kt",
    vars: [
      { key: "AUTHGEAR_CLIENT_ID", token: "clientID" },
      { key: "AUTHGEAR_ENDPOINT", token: "endpoint" },
      { key: "AUTHGEAR_REDIRECT_URI", token: "redirectURI" },
    ],
  },
  ide: "Android Studio",
  guideUrl: "https://docs.authgear.com/get-started/native-mobile-app/android",
};

const IOS_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-ios",
  downloadUrl:
    "https://github.com/authgear/authgear-example-ios/archive/HEAD.zip",
  redirectURIs: ["com.example.authgeardemo://host/path"],
  config: {
    format: "swift",
    fileName: "Constants.swift",
    vars: [
      { key: "authgearClientId", token: "clientID" },
      { key: "authgearEndpoint", token: "endpoint" },
      { key: "authgearRedirectUri", token: "redirectURI" },
    ],
  },
  ide: "Xcode",
  guideUrl: "https://docs.authgear.com/get-started/native-mobile-app/ios",
};

const FLUTTER_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-flutter",
  downloadUrl:
    "https://github.com/authgear/authgear-example-flutter/archive/HEAD.zip",
  redirectURIs: ["com.example.authgeardemo.flutter://host/path"],
  config: {
    format: "js",
    fileName: "lib/constants.dart",
    vars: [
      { key: "authgearClientId", token: "clientID" },
      { key: "authgearEndpoint", token: "endpoint" },
      { key: "authgearRedirectUri", token: "redirectURI" },
    ],
  },
  installCmd: "flutter pub get",
  startCmd: "flutter run",
  guideUrl: "https://docs.authgear.com/get-started/native-mobile-app/flutter",
};

const IONIC_STARTER_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-ionic",
  downloadUrl:
    "https://github.com/authgear/authgear-example-ionic/archive/HEAD.zip",
  redirectURIs: [
    "com.authgear.example.capacitor://host/path",
    "capacitor://localhost",
    "http://localhost:8100/oauth-redirect",
    "https://localhost",
  ],
  homepageUrl: "http://localhost:8100",
  config: {
    format: "dotenv",
    fileName: ".env",
    vars: [
      { key: "VITE_AUTHGEAR_CLIENT_ID", token: "clientID" },
      { key: "VITE_AUTHGEAR_ENDPOINT", token: "endpoint" },
    ],
  },
  installCmd: "npm install",
  startCmd: "ionic serve",
  mobileRun: {
    buildCmd: "npm run build",
    syncCmd: "npx cap sync",
    iosCmd: "npx cap open ios",
    androidCmd: "npx cap open android",
  },
  guideUrl: "https://docs.authgear.com/get-started/native-mobile-app/ionic",
};

export const frameworks: FrameworkEntry[] = [
  websiteSPA(
    "react",
    "CreateOAuthClientScreen.framework.react.title",
    "CreateOAuthClientScreen.framework.react.description",
    "brand-react",
    `${DOCS}/single-page-app/react`,
    REACT_STARTER_KIT
  ),
  websiteSPA(
    "vue",
    "CreateOAuthClientScreen.framework.vue.title",
    "CreateOAuthClientScreen.framework.vue.description",
    "brand-vue",
    `${DOCS}/single-page-app/vue`,
    VUE_STARTER_KIT
  ),
  websiteSPA(
    "angular",
    "CreateOAuthClientScreen.framework.angular.title",
    "CreateOAuthClientScreen.framework.angular.description",
    "brand-angular",
    `${DOCS}/single-page-app/angular`,
    ANGULAR_STARTER_KIT
  ),
  websiteSPA(
    "nextjs",
    "CreateOAuthClientScreen.framework.nextjs.title",
    "CreateOAuthClientScreen.framework.nextjs.description",
    "brand-nextjs",
    `${DOCS}/regular-web-app/nextjs`,
    NEXTJS_STARTER_KIT
  ),
  websiteSPA(
    "other-spa",
    "CreateOAuthClientScreen.framework.other-spa.title",
    "CreateOAuthClientScreen.framework.other-spa.description",
    "world-www",
    `${DOCS}/single-page-app/website`,
    OTHER_SPA_STARTER_KIT
  ),
  websiteServer(
    "express",
    "CreateOAuthClientScreen.framework.express.title",
    "CreateOAuthClientScreen.framework.express.description",
    "brand-javascript",
    `${DOCS}/regular-web-app/express`,
    EXPRESS_SNIPPET
  ),
  websiteServer(
    "flask",
    "CreateOAuthClientScreen.framework.flask.title",
    "CreateOAuthClientScreen.framework.flask.description",
    "brand-python",
    `${DOCS}/regular-web-app/python-flask-app`,
    FLASK_SNIPPET
  ),
  websiteServer(
    "laravel",
    "CreateOAuthClientScreen.framework.laravel.title",
    "CreateOAuthClientScreen.framework.laravel.description",
    "brand-laravel",
    `${DOCS}/regular-web-app/laravel`,
    LARAVEL_SNIPPET
  ),
  websiteServer(
    "java",
    "CreateOAuthClientScreen.framework.java.title",
    "CreateOAuthClientScreen.framework.java.description",
    "coffee",
    `${DOCS}/regular-web-app/java-spring-boot`,
    JAVA_SNIPPET
  ),
  websiteServer(
    "aspnet",
    "CreateOAuthClientScreen.framework.aspnet.title",
    "CreateOAuthClientScreen.framework.aspnet.description",
    "brand-windows",
    `${DOCS}/regular-web-app/asp.net-core-mvc`,
    ASPNET_SNIPPET
  ),
  {
    id: "other-oidc",
    displayNameMessageId: "CreateOAuthClientScreen.framework.other-oidc.title",
    helperTextMessageId:
      "CreateOAuthClientScreen.framework.other-oidc.description",
    section: "integration",
    iconName: "shield-check",
    docLink: `${DOCS}/oidc-provider`,
    stage2: "none",
    resolveType: () => "confidential",
    compatibleTypes: ["confidential"],
  },
  mobileNative(
    "react-native",
    "CreateOAuthClientScreen.framework.react-native.title",
    "CreateOAuthClientScreen.framework.react-native.description",
    "brand-react-native",
    `${DOCS}/native-mobile-app/react-native`
  ),
  mobileNative(
    "ios",
    "CreateOAuthClientScreen.framework.ios.title",
    "CreateOAuthClientScreen.framework.ios.description",
    "brand-apple",
    `${DOCS}/native-mobile-app/ios`,
    IOS_STARTER_KIT
  ),
  mobileNative(
    "android",
    "CreateOAuthClientScreen.framework.android.title",
    "CreateOAuthClientScreen.framework.android.description",
    "brand-android",
    `${DOCS}/native-mobile-app/android`,
    ANDROID_STARTER_KIT
  ),
  mobileNative(
    "flutter",
    "CreateOAuthClientScreen.framework.flutter.title",
    "CreateOAuthClientScreen.framework.flutter.description",
    "brand-flutter",
    `${DOCS}/native-mobile-app/flutter`,
    FLUTTER_STARTER_KIT
  ),
  mobileNative(
    "ionic",
    "CreateOAuthClientScreen.framework.ionic.title",
    "CreateOAuthClientScreen.framework.ionic.description",
    "device-mobile",
    `${DOCS}/native-mobile-app/ionic`,
    IONIC_STARTER_KIT
  ),
];

export function findFramework(
  id: string | undefined
): FrameworkEntry | undefined {
  return frameworks.find((f) => f.id === id);
}

export function frameworksForType(
  applicationType: ApplicationType
): FrameworkEntry[] {
  if (applicationType === "m2m") {
    // No FrameworkEntry is ever compatible with m2m; it is created
    // exclusively through CreateM2MClientScreen.
    return [];
  }
  return frameworks.filter((f) => f.compatibleTypes.includes(applicationType));
}

/**
 * Tabler icon name (without "ti-" prefix) to display for an OAuth client.
 * Falls back to a generic glyph when the framework is unknown.
 */
export function getDisplayIconName(client: {
  x_framework?: string | null;
  x_application_type?: ApplicationType;
}): string {
  const fw = findFramework(client.x_framework ?? undefined);
  if (fw != null) return fw.iconName;
  return client.x_application_type === "m2m" ? "server" : "app-window";
}

/**
 * Quick Start tutorial guide for the framework-based Quick Start tab.
 * For Cookie SSO clients (traditional_webapp) the guide is the nginx
 * reverse-proxy setup, which is the same regardless of backend framework;
 * for everything else we fall through to the framework's own doc page.
 */
export function getQuickStartGuide(client: {
  x_application_type?: ApplicationType;
  x_framework?: string | null;
}): { docLink: string; bodyMessageId: string } {
  if (client.x_application_type === "traditional_webapp") {
    return {
      docLink: "https://docs.authgear.com/get-started/backend-api/nginx",
      bodyMessageId:
        "EditOAuthClientFormFrameworkQuickStart.tutorial.body.cookie-sso",
    };
  }
  const framework = findFramework(client.x_framework ?? undefined);
  if (client.x_framework === "other-oidc") {
    return {
      docLink:
        framework?.docLink ??
        "https://docs.authgear.com/get-started/oidc-provider",
      bodyMessageId:
        "EditOAuthClientFormFrameworkQuickStart.tutorial.body.other-oidc",
    };
  }
  if (client.x_application_type === "confidential") {
    return {
      docLink:
        framework?.docLink ??
        "https://docs.authgear.com/get-started/regular-web-app",
      bodyMessageId:
        "EditOAuthClientFormFrameworkQuickStart.tutorial.body.confidential",
    };
  }
  return {
    docLink: framework?.docLink ?? "https://docs.authgear.com/",
    bodyMessageId:
      "EditOAuthClientFormFrameworkQuickStart.tutorial.body.default",
  };
}
