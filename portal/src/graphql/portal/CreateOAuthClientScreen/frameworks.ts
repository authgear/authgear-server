import type { Framework, ApplicationType } from "../../../types";

export type FrameworkSection = "website" | "mobile";
export type Stage2Need = "none" | "token-or-cookie";
export type AuthMethodChoice = "token" | "cookie";

export interface FrameworkEntry {
  id: Framework;
  displayName: string;
  helperText: string;
  section: FrameworkSection;
  /** Tabler icon name without the "ti-" prefix, e.g. "brand-react". */
  iconName: string;
  /** Authgear docs URL for this framework's quick-start guide. */
  docLink: string;
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

const websiteServer = (id: Framework, displayName: string, helperText: string, iconName: string, docLink: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "website",
  iconName,
  docLink,
  stage2: "token-or-cookie",
  resolveType: (stage2) => requireStage2(id, stage2),
  compatibleTypes: ["confidential", "traditional_webapp"],
});

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
  websiteServer("express", "Express.js", "Server-side, Node backend", "brand-javascript", `${DOCS}/regular-web-app/express`),
  websiteSPA("other-spa", "Other SPAs", "Any JavaScript SPA framework", "world-www", `${DOCS}/single-page-app/website`),
  websiteServer("flask", "Python (Flask)", "Server-side, Python backend", "brand-python", `${DOCS}/regular-web-app/python-flask-app`),
  websiteServer("laravel", "PHP (Laravel)", "Server-side, PHP backend", "brand-laravel", `${DOCS}/regular-web-app/laravel`),
  websiteServer("java", "Java (Spring Boot)", "Server-side, JVM backend", "coffee", `${DOCS}/regular-web-app/java-spring-boot`),
  websiteServer("aspnet", "ASP.NET Core MVC", "Server-side, .NET backend", "brand-windows", `${DOCS}/regular-web-app/asp.net-core-mvc`),
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
