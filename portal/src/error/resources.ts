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

const DefaultImageMaxSizeInKB = 100;
const imageTypeMaxSizeInKB: Partial<Record<string, number>> = {
  app_background_image: 500,
};

export function makeImageSizeTooLargeErrorRule(
  resources: Resource[]
): ErrorParseRule {
  return (apiError) => {
    if (apiError.reason === "RequestEntityTooLarge") {
      // When the request is blocked by the load balancer due to RequestEntityTooLarge
      // We try to get the largest resource from the state
      // and construct the error message for display

      let path = "";
      let longestLength = 0;
      // get the largest resources from the state
      for (const resource of resources) {
        const l = resource.nullableValue?.length ?? 0;
        if (l > longestLength) {
          longestLength = l;
          path = resource.path;
        }
      }

      // parse resource type from resource path
      let resourceType = "other";
      if (path !== "") {
        const dir = path.split("/");
        const fileName = dir[dir.length - 1];
        if (fileName.lastIndexOf(".") !== -1) {
          resourceType = fileName.slice(0, fileName.lastIndexOf("."));
        } else {
          resourceType = fileName;
        }
      }

      return {
        parsedAPIErrors: [
          {
            messageID: "errors.resource-too-large",
            arguments: {
              maxSize:
                imageTypeMaxSizeInKB[resourceType] ?? DefaultImageMaxSizeInKB,
              resourceType,
            },
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
