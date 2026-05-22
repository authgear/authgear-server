import { describe, it, expect } from "@jest/globals";
import { frameworks, findFramework, frameworksForType } from "./frameworks";
import { applicationFrameworks } from "../../../types";

describe("frameworks catalog", () => {
  it("covers every framework id in applicationFrameworks", () => {
    const ids = frameworks.map((f) => f.id).sort();
    expect(ids).toEqual([...applicationFrameworks].sort());
  });

  it("resolveType for SPA frameworks returns 'spa' without stage2", () => {
    expect(findFramework("react")!.resolveType()).toBe("spa");
    expect(findFramework("nextjs")!.resolveType()).toBe("spa");
  });

  it("resolveType for native frameworks returns 'native' without stage2", () => {
    expect(findFramework("ios")!.resolveType()).toBe("native");
    expect(findFramework("flutter")!.resolveType()).toBe("native");
  });

  it("resolveType for other-oidc returns 'confidential' without stage2", () => {
    expect(findFramework("other-oidc")!.resolveType()).toBe("confidential");
  });

  it("server framework with stage2='token' returns 'confidential'", () => {
    expect(findFramework("flask")!.resolveType("token")).toBe("confidential");
    expect(findFramework("aspnet")!.resolveType("token")).toBe("confidential");
  });

  it("server framework with stage2='cookie' returns 'traditional_webapp'", () => {
    expect(findFramework("flask")!.resolveType("cookie")).toBe(
      "traditional_webapp"
    );
    expect(findFramework("aspnet")!.resolveType("cookie")).toBe(
      "traditional_webapp"
    );
  });

  it("server framework with no stage2 throws", () => {
    expect(() => findFramework("flask")!.resolveType()).toThrow();
  });

  it("findFramework returns undefined for unknown id", () => {
    expect(findFramework("nonsense")).toBeUndefined();
  });

  it("frameworksForType('spa') returns SPA frameworks only", () => {
    const ids = frameworksForType("spa").map((f) => f.id);
    expect(ids).toEqual(["react", "vue", "angular", "nextjs", "other-spa"]);
  });

  it("frameworksForType('traditional_webapp') excludes other-oidc", () => {
    const ids = frameworksForType("traditional_webapp").map((f) => f.id);
    expect(ids).toEqual(["express", "flask", "laravel", "java", "aspnet"]);
  });

  it("frameworksForType('confidential') includes other-oidc and server frameworks", () => {
    const ids = frameworksForType("confidential").map((f) => f.id);
    expect(ids).toEqual([
      "express",
      "flask",
      "laravel",
      "java",
      "aspnet",
      "other-oidc",
    ]);
  });

  it("frameworksForType('native') returns native frameworks", () => {
    const ids = frameworksForType("native").map((f) => f.id);
    expect(ids).toEqual(["react-native", "ios", "android", "flutter", "ionic"]);
  });
});
