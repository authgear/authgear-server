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
  stage2: Stage2Need;
  resolveType: (stage2?: AuthMethodChoice) => ApplicationType;
  compatibleTypes: ApplicationType[];
}

const requireStage2 = (id: Framework, stage2?: AuthMethodChoice): ApplicationType => {
  if (stage2 === "token") return "confidential";
  if (stage2 === "cookie") return "traditional_webapp";
  throw new Error(`resolveType called without stage2 on ${id}`);
};

const websiteSPA = (id: Framework, displayName: string, helperText: string, iconName: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "website",
  iconName,
  stage2: "none",
  resolveType: () => "spa",
  compatibleTypes: ["spa"],
});

const websiteServer = (id: Framework, displayName: string, helperText: string, iconName: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "website",
  iconName,
  stage2: "token-or-cookie",
  resolveType: (stage2) => requireStage2(id, stage2),
  compatibleTypes: ["confidential", "traditional_webapp"],
});

const mobileNative = (id: Framework, displayName: string, helperText: string, iconName: string): FrameworkEntry => ({
  id,
  displayName,
  helperText,
  section: "mobile",
  iconName,
  stage2: "none",
  resolveType: () => "native",
  compatibleTypes: ["native"],
});

export const frameworks: FrameworkEntry[] = [
  websiteSPA("react", "React", "SPA, uses authgear-sdk-js", "brand-react"),
  websiteSPA("vue", "Vue", "SPA, uses authgear-sdk-js", "brand-vue"),
  websiteSPA("angular", "Angular", "SPA, uses authgear-sdk-js", "brand-angular"),
  websiteSPA("nextjs", "Next.js", "SPA/SSR, uses authgear-sdk-nextjs", "brand-nextjs"),
  websiteServer("express", "Express.js", "Server-side, Node backend", "brand-javascript"),
  websiteSPA("other-spa", "Other SPAs", "Any JavaScript SPA framework", "world-www"),
  websiteServer("django", "Python (Django)", "Server-side, Python backend", "brand-django"),
  websiteServer("laravel", "PHP (Laravel)", "Server-side, PHP backend", "brand-laravel"),
  websiteServer("java", "Java", "Server-side, JVM backend", "coffee"),
  websiteServer("aspnet", "ASP.NET", "Server-side, .NET backend", "brand-windows"),
  {
    id: "other-oidc",
    displayName: "Other OIDC/SAML compatible",
    helperText: "Any OIDC/SAML compatible app",
    section: "website",
    iconName: "shield-check",
    stage2: "none",
    resolveType: () => "confidential",
    compatibleTypes: ["confidential"],
  },
  mobileNative("react-native", "React Native", "Cross-platform mobile SDK", "brand-react-native"),
  mobileNative("ios", "iOS", "Native iOS (Swift)", "brand-apple"),
  mobileNative("android", "Android", "Native Android (Kotlin)", "brand-android"),
  mobileNative("flutter", "Flutter", "Cross-platform mobile SDK", "brand-flutter"),
  mobileNative("ionic", "Ionic", "Cross-platform hybrid SDK", "device-mobile"),
];

export function findFramework(id: Framework | string | undefined): FrameworkEntry | undefined {
  return frameworks.find((f) => f.id === id);
}

export function frameworksForType(applicationType: ApplicationType): FrameworkEntry[] {
  return frameworks.filter((f) => f.compatibleTypes.includes(applicationType));
}
