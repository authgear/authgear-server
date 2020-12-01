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
