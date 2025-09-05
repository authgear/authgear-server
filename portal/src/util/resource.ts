import { toByteArray, fromByteArray } from "base64-js";

export type LanguageTag = string;

export const BUILTIN_LOCALE: LanguageTag = "en";

export interface ResourceUpdate {
  path: string;
  value?: string | null;
  specifier: ResourceSpecifier;
  checksum?: string | null;
}

export interface Resource {
  path: string;
  nullableValue?: string | null;
  specifier: ResourceSpecifier;
  effectiveData?: string | null;
  checksum?: string | null;
}

export interface ResourceSpecifier {
  def: ResourceDefinition;
  locale: LanguageTag | null;
  extension: string | null;
}

export interface FallbackStrategyEffectiveData {
  kind: "EffectiveData";
}

export const FALLBACK_EFFECTIVE_DATA: FallbackStrategyEffectiveData = {
  kind: "EffectiveData",
};

export interface FallbackStrategyConst {
  kind: "Const";
  fallbackValue: string;
}

export type FallbackStrategy =
  | FallbackStrategyEffectiveData
  | FallbackStrategyConst;

export interface ResourceDefinition {
  type: "text" | "binary" | "prettified-json";
  resourcePath: ResourcePath;
  extensions: string[];
  fallback?: FallbackStrategy;
  // Indicates whether the resource is optional.
  // The default locale must have all non-optional resources configured.
  optional?: boolean;
}

export interface ResourcePath {
  parse(path: string): Record<string, string> | null;
  render(args: Record<string, string>): string;
}

export function expandSpecifier(specifier: ResourceSpecifier): string {
  const { resourcePath } = specifier.def;
  const renderArgs: Record<string, string> = {};
  if (specifier.locale != null) {
    renderArgs["locale"] = specifier.locale;
  }
  if (specifier.extension != null) {
    renderArgs["extension"] = specifier.extension;
  }
  return resourcePath.render(renderArgs);
}

export function expandDef(
  def: ResourceDefinition,
  locale: LanguageTag
): ResourceSpecifier[] {
  if (def.extensions.length === 0) {
    return [
      {
        def,
        locale,
        extension: null,
      },
    ];
  }

  const specifiers = [];
  for (const extension of def.extensions) {
    specifiers.push({
      def,
      locale,
      extension,
    });
  }
  return specifiers;
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
  const byteArray = toByteArray(a);
  const decoder = new TextDecoder();
  return decoder.decode(byteArray);
}

export function encodeForText(a: string): string {
  const encoder = new TextEncoder();
  const byteArray = encoder.encode(a);
  return fromByteArray(byteArray);
}

export function decodeForPrettifiedJSON(a: string): string {
  const text = decodeForText(a);
  const j = JSON.parse(text);
  const prettified = JSON.stringify(j, null, 2);
  return prettified;
}

export function encodeForPrettifiedJSON(a: string): string {
  try {
    const j = JSON.parse(a);
    const compact = JSON.stringify(j);
    return encodeForText(compact);
  } catch (_e: unknown) {
    // In case you wonder why we are doing this,
    // if the input is not a valid JSON,
    // we just do not format it at all.
    // Passing the invalid JSON to the server,
    // and let the server to report an error.
    return encodeForText(a);
  }
}

export function specifierId(specifier: ResourceSpecifier): string {
  return specifier.def.resourcePath.render({
    locale: specifier.locale ?? "{locale}",
    extension: specifier.extension ?? "{extension}",
  });
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
      .filter((r) => r.nullableValue != null && r.nullableValue !== "")
      .map((r) => [specifierId(r.specifier), r])
  );
  const currentResourceMap = new Map<string, Resource>(
    currentResources
      .filter((r) => r.nullableValue != null && r.nullableValue !== "")
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
        value: r.nullableValue,
      });
    } else if (initialResource.nullableValue !== r.nullableValue) {
      result.editedResources.push({
        specifier: r.specifier,
        path: r.path,
        value: r.nullableValue,
        checksum: initialResource.checksum,
      });
    }
  }

  for (const [id, r] of initialResourceMap.entries()) {
    if (!currentResourceMap.has(id)) {
      result.deletedResources.push({
        specifier: r.specifier,
        path: r.path,
        value: null,
        checksum: r.checksum,
      });
    }
  }

  result.needUpdate =
    result.newResources.length > 0 ||
    result.editedResources.length > 0 ||
    result.deletedResources.length > 0;
  return result;
}

export function getDenoScriptPathFromURL(url: string): string {
  const path = url.slice("authgeardeno:///".length);
  return path;
}

export function makeDenoScriptSpecifier(url: string): ResourceSpecifier {
  const path = getDenoScriptPathFromURL(url);
  return {
    def: {
      resourcePath: resourcePath([path]),
      type: "text" as const,
      extensions: [],
    },
    locale: null,
    extension: null,
  };
}

export function resolveResource(
  resources: Partial<Record<string, Resource>>,
  specifiers: [ResourceSpecifier] | ResourceSpecifier[]
): Resource | null {
  for (const specifier of specifiers) {
    const resource = resources[specifierId(specifier)];
    if (resource?.nullableValue) {
      return resource;
    }
  }
  return resources[specifierId(specifiers[specifiers.length - 1])] ?? null;
}
