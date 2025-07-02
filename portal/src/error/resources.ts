import { Resource } from "../util/resource";
import { ErrorParseRule } from "./parse";

export interface APIResourceNotFoundError {
  errorName: "ResourceNotFound";
  reason: "ResourceNotFound";
}

export interface APIResourceTooLargeError {
  errorName: string;
  reason: "ResourceTooLarge";
  info: {
    size: number;
    max_size: number;
    path: string;
  };
}

export interface APIUnsupportedImageFileError {
  errorName: string;
  reason: "UnsupportedImageFile";
  info: {
    type: string;
  };
}

export function makeImageSizeTooLargeErrorRule(
  _resources: Resource[]
): ErrorParseRule {
  return (apiError) => {
    if (apiError.reason === "RequestEntityTooLarge") {
      // When the request is blocked by the load balancer due to RequestEntityTooLarge
      // We try to get the largest resource from the state
      // and construct the error message for display

      return {
        parsedAPIErrors: [
          {
            messageID: "errors.resource-too-large",
            arguments: {},
          },
        ],
        fullyHandled: true,
      };
    }
    return {
      parsedAPIErrors: [],
      fullyHandled: false,
    };
  };
}
