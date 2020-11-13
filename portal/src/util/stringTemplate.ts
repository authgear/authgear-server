interface StringTemplateParamData {
  paramName: string;
  startIndex: number;
  endIndex: number;
}

function applyArgumentOnTemplateString(
  argument: string,
  template: string,
  value: string
): string {
  return template.replace(new RegExp(`{{[ ]*${argument}[ ]*}}`, "g"), value);
}

export function renderTemplateString(
  values: Record<string, string>,
  template: string
): string {
  return Object.entries(values).reduce((result, [key, value]) => {
    return applyArgumentOnTemplateString(key, result, value);
  }, template);
}

export function parseTemplateString(
  input: string,
  template: string
): Partial<Record<string, string>> {
  const output: Partial<Record<string, string>> = {};
  const paramRegexp = /{{[ ]*([^{} ]*)[ ]*}}/g;
  const stringSegments: string[] = [];
  const templateParams: StringTemplateParamData[] = [];

  let paramMatch: RegExpExecArray | null;
  let segmentStartIndex = 0;
  while ((paramMatch = paramRegexp.exec(template)) != null) {
    const matchedString = paramMatch[0];
    const matchedParamName = paramMatch[1];
    const paramExprStartIndex = paramMatch.index;
    const paramExprEndIndex = paramExprStartIndex + matchedString.length - 1;
    templateParams.push({
      paramName: matchedParamName,
      startIndex: paramExprStartIndex,
      endIndex: paramExprEndIndex,
    });
    stringSegments.push(template.slice(segmentStartIndex, paramExprStartIndex));
    // for next match
    segmentStartIndex = paramExprStartIndex + matchedString.length;
  }
  // push last segment
  stringSegments.push(template.slice(segmentStartIndex));

  const valueRegexpString = stringSegments.join("(.*)");
  const valueRegexp = new RegExp(`^${valueRegexpString}$`);
  const valueMatches = valueRegexp.exec(input);
  // valueMatches[0] is string matched with regexp
  // valueMatches[1] is first matched group
  if (valueMatches == null) {
    return output;
  }
  let currentValueMatchesIndex = 1;
  for (const param of templateParams) {
    if (
      output[param.paramName] != null &&
      output[param.paramName] !== valueMatches[currentValueMatchesIndex]
    ) {
      throw new Error(
        "[Parse string template]: Value of parameter is inconsistent"
      );
    }
    output[param.paramName] = valueMatches[currentValueMatchesIndex];
    currentValueMatchesIndex += 1;
  }

  return output;
}
