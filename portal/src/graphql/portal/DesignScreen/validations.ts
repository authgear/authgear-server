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
      const parsed = /^(?<num>[\d.]+)(?<unit>\w{0,})$/.exec(
        borderRadius.radius
      );
      if (
        // eslint-disable-next-line @typescript-eslint/prefer-optional-chain
        parsed == null ||
        parsed.groups == null ||
        isNaN(Number(parsed.groups["num"]))
      ) {
        return [
          {
            location: location,
            messageID: "errors.validation.borderRadius.format",
          },
        ];
      }
      const num = Number(parsed.groups["num"]);
      if (
        !["", "px", "em", "rem"].includes(parsed.groups["unit"]) ||
        (num !== 0 && parsed.groups["unit"] === "")
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
