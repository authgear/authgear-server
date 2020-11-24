export type LanguageTag = string;

export interface Resource {
  locale: LanguageTag;
  def: ResourceDefinition;
  path: string;
  value: string;
}

export interface ResourceSpecifier {
  locale: LanguageTag;
  def: ResourceDefinition;
  path: string;
}

export interface ResourceDefinition {
  type: "text" | "binary";
  resourcePath: ResourcePath<"locale">;
}

export interface ResourcePath<Arg extends string> {
  parse(path: string): Record<Arg, string> | null;
  render(args: Record<Arg, string>): string;
}

export function resourcePath<Arg extends string>(
  parts: readonly string[],
  ...args: readonly Arg[]
): ResourcePath<Arg> {
  const parse = (path: string): Record<Arg, string> | null => {
    const output: Partial<Record<Arg, string>> = {};
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
    return output as Record<Arg, string>;
  };

  const render = (values: Record<Arg, string>): string => {
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
