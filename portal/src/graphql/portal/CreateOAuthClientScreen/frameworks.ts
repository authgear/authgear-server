import type { Framework, ApplicationType } from "../../../types";

export type FrameworkSection = "website" | "mobile";
export type Stage2Need = "none" | "token-or-cookie";
export type AuthMethodChoice = "token" | "cookie";

export interface FrameworkEntry {
  id: Framework;
  displayName: string;
  helperText: string;
  section: FrameworkSection;
  logo: string;
  stage2: Stage2Need;
  resolveType: (stage2?: AuthMethodChoice) => ApplicationType;
  compatibleTypes: ApplicationType[];
}

const requireStage2 = (id: Framework, stage2?: AuthMethodChoice): ApplicationType => {
  if (stage2 === "token") return "confidential";
  if (stage2 === "cookie") return "traditional_webapp";
  throw new Error(`resolveType called without stage2 on ${id}`);
};

const websiteSPA = (id: Framework, displayName: string, helperText: string, logo: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "website",
  logo,
  stage2: "none",
  resolveType: () => "spa",
  compatibleTypes: ["spa"],
});

const websiteServer = (id: Framework, displayName: string, helperText: string, logo: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "website",
  logo,
  stage2: "token-or-cookie",
  resolveType: (stage2) => requireStage2(id, stage2),
  compatibleTypes: ["confidential", "traditional_webapp"],
});

const mobileNative = (id: Framework, displayName: string, helperText: string, logo: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "mobile",
  logo,
  stage2: "none",
  resolveType: () => "native",
  compatibleTypes: ["native"],
});

export const frameworks: FrameworkEntry[] = [
  websiteSPA("react", "React", "SPA, uses authgear-sdk-js", "/logos/frameworks/react.svg"),
  websiteSPA("vue", "Vue", "SPA, uses authgear-sdk-js", "/logos/frameworks/vue.svg"),
  websiteSPA("angular", "Angular", "SPA, uses authgear-sdk-js", "/logos/frameworks/angular.svg"),
  websiteSPA("nextjs", "Next.js", "SPA/SSR, uses authgear-sdk-nextjs", "/logos/frameworks/nextjs.svg"),
  websiteServer("express", "Express.js", "Server-side, Node backend", "/logos/frameworks/express.svg"),
  websiteSPA("other-spa", "Other SPAs", "Any JavaScript SPA framework", "/logos/frameworks/other-spa.svg"),
  websiteServer("django", "Python (Django)", "Server-side, Python backend", "/logos/frameworks/django.svg"),
  websiteServer("laravel", "PHP (Laravel)", "Server-side, PHP backend", "/logos/frameworks/laravel.svg"),
  websiteServer("java", "Java", "Server-side, JVM backend", "/logos/frameworks/java.svg"),
  websiteServer("aspnet", "ASP.NET", "Server-side, .NET backend", "/logos/frameworks/aspnet.svg"),
  {
    id: "other-oidc",
    displayName: "Other OIDC/SAML compatible",
    helperText: "Any OIDC/SAML compatible app",
    section: "website",
    logo: "/logos/frameworks/other-oidc.svg",
    stage2: "none",
    resolveType: () => "confidential",
    compatibleTypes: ["confidential"],
  },
  mobileNative("react-native", "React Native", "Cross-platform mobile SDK", "/logos/frameworks/react-native.svg"),
  mobileNative("ios", "iOS", "Native iOS (Swift)", "/logos/frameworks/ios.svg"),
  mobileNative("android", "Android", "Native Android (Kotlin)", "/logos/frameworks/android.svg"),
  mobileNative("flutter", "Flutter", "Cross-platform mobile SDK", "/logos/frameworks/flutter.svg"),
  mobileNative("ionic", "Ionic", "Cross-platform hybrid SDK", "/logos/frameworks/ionic.svg"),
];

export function findFramework(id: Framework | string | undefined): FrameworkEntry | undefined {
  return frameworks.find((f) => f.id === id);
}

export function frameworksForType(applicationType: ApplicationType): FrameworkEntry[] {
  return frameworks.filter((f) => f.compatibleTypes.includes(applicationType));
}
