export type LanguageTag = string;

export interface ResourceUpdate {
  path: string;
  value?: string | null;
  specifier: ResourceSpecifier;
}

export interface Resource {
  path: string;
  value: string;
  specifier: ResourceSpecifier;
}

export interface ResourceSpecifier {
  def: ResourceDefinition;
  locale?: LanguageTag;
}

export interface ResourceDefinition {
  type: "text" | "binary";
  resourcePath: ResourcePath;
  extensions: string[];
  // If this is true, then the effectiveData is used as value when the raw data is unavailable.
  // This is useful for templates.
  usesEffectiveDataAsFallbackValue: boolean;
  // Indicates whether the resource is optional.
  // The default locale must have all non-optional resources configured.
  optional?: boolean;
}

export interface ResourcePath {
  parse(path: string): Record<string, string> | null;
  render(args: Record<string, string>): string;
}

export function resourcePath(
  parts: readonly string[],
  ...args: readonly string[]
): ResourcePath {
  const parse = (path: string): Record<string, string> | null => {
    const output: Partial<Record<string, string>> = {};
    const valueRegexpString = parts.join("(.*)");
    const valueRegexp = new RegExp(`^${valueRegexpString}$`);
    const valueMatches = valueRegexp.exec(path);
    // valueMatches[0] is string matched with regexp
    // valueMatches[1] is first matched group
    if (valueMatches == null) {
      return null;
    }
    let currentValueMatchesIndex = 1;
    for (const param of args) {
      if (
        output[param] != null &&
        output[param] !== valueMatches[currentValueMatchesIndex]
      ) {
        throw new Error(
          "[Parse string template]: Value of parameter is inconsistent"
        );
      }
      output[param] = valueMatches[currentValueMatchesIndex];
      currentValueMatchesIndex += 1;
    }
    return output as Record<string, string>;
  };

  const render = (values: Record<string, string>): string => {
    return parts.reduce((accu, part, index) => {
      accu += part;
      if (index < args.length) {
        accu += values[args[index]];
      }
      return accu;
    }, "");
  };

  return {
    parse,
    render,
  };
}

export function binary(a: string): string {
  return a;
}

export function decodeForText(a: string): string {
  return atob(a);
}

export function encodeForText(a: string): string {
  return btoa(a);
}

export function specifierId(specifier: ResourceSpecifier): string {
  return specifier.def.resourcePath.render({
    locale: specifier.locale ?? "{locale}",
    extension: "{extension}",
  });
}

export type LocaleInvalidReason =
  | "required-default-resource"
  | "locale-without-resources";

export interface LocaleValidationResult {
  invalidReason: LocaleInvalidReason | null;
  invalidLocales: LanguageTag[];
}

export function validateLocales(
  defaultLocale: LanguageTag,
  locales: LanguageTag[],
  resources: Resource[]
): LocaleValidationResult {
  // For convenience, report only one invalid reason.
  let invalidReason: LocaleInvalidReason | null = null;
  const invalidLocales = new Set<LanguageTag>();

  // Locale is valid iff there is at least 1 resource with that locale.
  const orphanedLocales = new Set<LanguageTag>(locales);
  for (const r of resources) {
    const locale = r.specifier.locale;
    if (!locale) {
      continue;
    }

    if (r.value !== "") {
      orphanedLocales.delete(locale);
    }
    if (
      r.value === "" &&
      locale === defaultLocale &&
      !r.specifier.def.optional
    ) {
      // An non-optional resource for default locale without value:
      invalidLocales.add(locale);
      invalidReason = "required-default-resource";
    }
  }

  if (orphanedLocales.size > 0) {
    for (const l of orphanedLocales) {
      invalidLocales.add(l);
    }
    if (!invalidReason) {
      invalidReason = "locale-without-resources";
    }
  }

  return {
    invalidReason,
    invalidLocales: Array.from(invalidLocales),
  };
}

export interface ResourcesDiffResult {
  needUpdate: boolean;
  newResources: ResourceUpdate[];
  editedResources: ResourceUpdate[];
  deletedResources: ResourceUpdate[];
}

export function diffResourceUpdates(
  initialResources: Resource[],
  currentResources: Resource[]
): ResourcesDiffResult {
  const initialResourceMap = new Map<string, Resource>(
    initialResources
      .filter((r) => r.value !== "")
      .map((r) => [specifierId(r.specifier), r])
  );
  const currentResourceMap = new Map<string, Resource>(
    currentResources
      .filter((r) => r.value !== "")
      .map((r) => [specifierId(r.specifier), r])
  );

  const result: ResourcesDiffResult = {
    needUpdate: false,
    newResources: [],
    editedResources: [],
    deletedResources: [],
  };

  for (const [id, r] of currentResourceMap.entries()) {
    const initialResource = initialResourceMap.get(id);
    if (!initialResource) {
      result.newResources.push({
        specifier: r.specifier,
        path: r.path,
        value: r.value,
      });
    } else if (initialResource.value !== r.value) {
      result.editedResources.push({
        specifier: r.specifier,
        path: r.path,
        value: r.value,
      });
    }
  }

  for (const [id, r] of initialResourceMap.entries()) {
    if (!currentResourceMap.has(id)) {
      result.deletedResources.push({
        specifier: r.specifier,
        path: r.path,
        value: null,
      });
    }
  }

  result.needUpdate =
    result.newResources.length > 0 ||
    result.editedResources.length > 0 ||
    result.deletedResources.length > 0;
  return result;
}
