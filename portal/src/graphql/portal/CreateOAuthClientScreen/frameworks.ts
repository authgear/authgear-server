import type { Framework, ApplicationType } from "../../../types";

export type FrameworkSection = "website" | "mobile";
export type Stage2Need = "none" | "token-or-cookie";
export type AuthMethodChoice = "token" | "cookie";

export interface CookieSnippet {
  /** Human-readable language label, e.g. "JavaScript", "Python". */
  language: string;
  /** Source code that reads x-authgear-* headers forwarded by nginx. */
  code: string;
}

export interface FrameworkEntry {
  id: Framework;
  displayName: string;
  helperText: string;
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
  resolveType: (stage2?: AuthMethodChoice) => ApplicationType;
  compatibleTypes: ApplicationType[];
}

const requireStage2 = (id: Framework, stage2?: AuthMethodChoice): ApplicationType => {
  if (stage2 === "token") return "confidential";
  if (stage2 === "cookie") return "traditional_webapp";
  throw new Error(`resolveType called without stage2 on ${id}`);
};

const websiteSPA = (id: Framework, displayName: string, helperText: string, iconName: string, docLink: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "website",
  iconName,
  docLink,
  stage2: "none",
  resolveType: () => "spa",
  compatibleTypes: ["spa"],
});

const websiteServer = (
  id: Framework,
  displayName: string,
  helperText: string,
  iconName: string,
  docLink: string,
  cookieSnippet?: CookieSnippet
): FrameworkEntry => ({
  id,
  displayName,
  helperText,
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

const mobileNative = (id: Framework, displayName: string, helperText: string, iconName: string, docLink: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "mobile",
  iconName,
  docLink,
  stage2: "none",
  resolveType: () => "native",
  compatibleTypes: ["native"],
});

const DOCS = "https://docs.authgear.com/get-started";

export const frameworks: FrameworkEntry[] = [
  websiteSPA("react", "React", "SPA, uses authgear-sdk-js", "brand-react", `${DOCS}/single-page-app/react`),
  websiteSPA("vue", "Vue", "SPA, uses authgear-sdk-js", "brand-vue", `${DOCS}/single-page-app/vue`),
  websiteSPA("angular", "Angular", "SPA, uses authgear-sdk-js", "brand-angular", `${DOCS}/single-page-app/angular`),
  websiteSPA("nextjs", "Next.js", "SPA/SSR, uses authgear-sdk-nextjs", "brand-nextjs", `${DOCS}/regular-web-app/nextjs`),
  websiteSPA("other-spa", "Other SPAs", "Any JavaScript SPA framework", "world-www", `${DOCS}/single-page-app/website`),
  websiteServer("express", "Express.js", "Server-side, Node backend", "brand-javascript", `${DOCS}/regular-web-app/express`, EXPRESS_SNIPPET),
  websiteServer("flask", "Python (Flask)", "Server-side, Python backend", "brand-python", `${DOCS}/regular-web-app/python-flask-app`, FLASK_SNIPPET),
  websiteServer("laravel", "PHP (Laravel)", "Server-side, PHP backend", "brand-laravel", `${DOCS}/regular-web-app/laravel`, LARAVEL_SNIPPET),
  websiteServer("java", "Java (Spring Boot)", "Server-side, JVM backend", "coffee", `${DOCS}/regular-web-app/java-spring-boot`, JAVA_SNIPPET),
  websiteServer("aspnet", "ASP.NET Core MVC", "Server-side, .NET backend", "brand-windows", `${DOCS}/regular-web-app/asp.net-core-mvc`, ASPNET_SNIPPET),
  {
    id: "other-oidc",
    displayName: "Other OIDC/SAML compatible",
    helperText: "Any OIDC/SAML compatible app",
    section: "website",
    iconName: "shield-check",
    docLink: `${DOCS}/oidc-provider`,
    stage2: "none",
    resolveType: () => "confidential",
    compatibleTypes: ["confidential"],
  },
  mobileNative("react-native", "React Native", "Cross-platform mobile SDK", "brand-react-native", `${DOCS}/native-mobile-app/react-native`),
  mobileNative("ios", "iOS", "Native iOS (Swift)", "brand-apple", `${DOCS}/native-mobile-app/ios`),
  mobileNative("android", "Android", "Native Android (Kotlin)", "brand-android", `${DOCS}/native-mobile-app/android`),
  mobileNative("flutter", "Flutter", "Cross-platform mobile SDK", "brand-flutter", `${DOCS}/native-mobile-app/flutter`),
  mobileNative("ionic", "Ionic", "Cross-platform hybrid SDK", "device-mobile", `${DOCS}/native-mobile-app/ionic`),
];

export function findFramework(id: Framework | string | undefined): FrameworkEntry | undefined {
  return frameworks.find((f) => f.id === id);
}

export function frameworksForType(applicationType: ApplicationType): FrameworkEntry[] {
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
      docLink: framework?.docLink ?? "https://docs.authgear.com/get-started/oidc-provider",
      bodyMessageId:
        "EditOAuthClientFormFrameworkQuickStart.tutorial.body.other-oidc",
    };
  }
  if (client.x_application_type === "confidential") {
    return {
      docLink: framework?.docLink ?? "https://docs.authgear.com/get-started/regular-web-app",
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
