import { LocalValidationError } from "../../../error/validation";
import { BorderRadiusStyle } from "../../../model/themeAuthFlowV2";

export function validateBorderRadius(
  location: string,
  borderRadius: BorderRadiusStyle
): LocalValidationError[] {
  switch (borderRadius.type) {
    case "none":
      return [];
    case "rounded-full":
      return [];
    case "rounded": {
      const parsed = /^([\d.]+)(\w{0,})$/.exec(borderRadius.radius);
      if (parsed == null || parsed.length < 3 || isNaN(Number(parsed[1]))) {
        return [
          {
            location: location,
            messageID: "errors.validation.borderRadius.format",
          },
        ];
      }
      const num = Number(parsed[1]);
      if (
        !["", "px", "em", "rem"].includes(parsed[2]) ||
        (num !== 0 && parsed[2] === "")
      ) {
        return [
          {
            location: location,
            messageID: "errors.validation.borderRadius.unit",
          },
        ];
      }
      return [];
    }
  }
}
