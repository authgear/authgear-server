import type { Framework, ApplicationType } from "../../../types";
import reactLogo from "./logos/react.svg";
import vueLogo from "./logos/vue.svg";
import angularLogo from "./logos/angular.svg";
import nextjsLogo from "./logos/nextjs.svg";
import expressLogo from "./logos/express.svg";
import otherSpaLogo from "./logos/other-spa.svg";
import djangoLogo from "./logos/django.svg";
import laravelLogo from "./logos/laravel.svg";
import javaLogo from "./logos/java.svg";
import aspnetLogo from "./logos/aspnet.svg";
import otherOidcLogo from "./logos/other-oidc.svg";
import reactNativeLogo from "./logos/react-native.svg";
import iosLogo from "./logos/ios.svg";
import androidLogo from "./logos/android.svg";
import flutterLogo from "./logos/flutter.svg";
import ionicLogo from "./logos/ionic.svg";

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
  websiteSPA("react", "React", "SPA, uses authgear-sdk-js", reactLogo),
  websiteSPA("vue", "Vue", "SPA, uses authgear-sdk-js", vueLogo),
  websiteSPA("angular", "Angular", "SPA, uses authgear-sdk-js", angularLogo),
  websiteSPA("nextjs", "Next.js", "SPA/SSR, uses authgear-sdk-nextjs", nextjsLogo),
  websiteServer("express", "Express.js", "Server-side, Node backend", expressLogo),
  websiteSPA("other-spa", "Other SPAs", "Any JavaScript SPA framework", otherSpaLogo),
  websiteServer("django", "Python (Django)", "Server-side, Python backend", djangoLogo),
  websiteServer("laravel", "PHP (Laravel)", "Server-side, PHP backend", laravelLogo),
  websiteServer("java", "Java", "Server-side, JVM backend", javaLogo),
  websiteServer("aspnet", "ASP.NET", "Server-side, .NET backend", aspnetLogo),
  {
    id: "other-oidc",
    displayName: "Other OIDC/SAML compatible",
    helperText: "Any OIDC/SAML compatible app",
    section: "website",
    logo: otherOidcLogo,
    stage2: "none",
    resolveType: () => "confidential",
    compatibleTypes: ["confidential"],
  },
  mobileNative("react-native", "React Native", "Cross-platform mobile SDK", reactNativeLogo),
  mobileNative("ios", "iOS", "Native iOS (Swift)", iosLogo),
  mobileNative("android", "Android", "Native Android (Kotlin)", androidLogo),
  mobileNative("flutter", "Flutter", "Cross-platform mobile SDK", flutterLogo),
  mobileNative("ionic", "Ionic", "Cross-platform hybrid SDK", ionicLogo),
];

export function findFramework(id: Framework | string | undefined): FrameworkEntry | undefined {
  return frameworks.find((f) => f.id === id);
}

export function frameworksForType(applicationType: ApplicationType): FrameworkEntry[] {
  return frameworks.filter((f) => f.compatibleTypes.includes(applicationType));
}
