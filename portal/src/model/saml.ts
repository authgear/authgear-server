import { XMLParser } from "fast-xml-parser";
import { SAMLBinding } from "../types";

interface ServiceProviderMetadataParseResult {
  acsURL?: string;
  sloEnabled?: boolean;
  sloCallbackURL?: string;
  sloCallbackBinding?: SAMLBinding;
  authnRequestsSigned?: boolean;
  certificate?: string;
}

export function parseServiceProviderMetadata(
  xmlData: string
): ServiceProviderMetadataParseResult {
  const parser = new XMLParser({
    alwaysCreateTextNode: true,
    removeNSPrefix: true,
    allowBooleanAttributes: true,
    ignoreAttributes: false,
    attributeNamePrefix: "@_",
  });
  const xmlObj = parser.parse(xmlData);

  const acsURL = findAttributeByPath(
    xmlObj,
    ["EntityDescriptor", "SPSSODescriptor", "AssertionConsumerService"],
    "@_Location"
  );
  const singleLogoutURL = findAttributeByPath(
    xmlObj,
    ["EntityDescriptor", "SPSSODescriptor", "SingleLogoutService"],
    "@_Location"
  );
  const sloCallbackBinding = findAttributeByPath(
    xmlObj,
    ["EntityDescriptor", "SPSSODescriptor", "SingleLogoutService"],
    "@_Binding"
  );
  const isSLOEnabled =
    singleLogoutURL != null && isSupportedLogoutBinding(sloCallbackBinding);

  const authnRequestsSignedStr = findAttributeByPath(
    xmlObj,
    ["EntityDescriptor", "SPSSODescriptor"],
    "@_AuthnRequestsSigned"
  );
  let authnRequestsSigned: boolean | undefined;
  if (authnRequestsSignedStr != null) {
    authnRequestsSigned = authnRequestsSignedStr === "true";
  }

  const certificateData = findAttributeByPath(
    xmlObj,
    [
      "EntityDescriptor",
      "SPSSODescriptor",
      "KeyDescriptor",
      "KeyInfo",
      "X509Data",
      "X509Certificate",
    ],
    "#text", // #text is used to identify the text node,
    (tag: string, node: any): boolean => {
      if (typeof node !== "object" || tag !== "KeyDescriptor") {
        // We only want to filter KeyDescriptor
        return true;
      }
      const keyUse = node["@_use"];
      if (keyUse != null && keyUse !== "signing") {
        return false;
      }
      return true;
    }
  );

  let certificate: string | undefined;
  if (certificateData != null) {
    certificate = toPem(certificateData);
  }

  return {
    acsURL: acsURL,
    sloEnabled: isSLOEnabled,
    sloCallbackURL: isSLOEnabled ? singleLogoutURL : undefined,
    sloCallbackBinding: isSLOEnabled ? sloCallbackBinding : undefined,
    authnRequestsSigned: authnRequestsSigned,
    certificate,
  };
}

function getNode(
  node: any,
  path: string,
  filter?: (tag: string, node: any) => boolean
): any {
  switch (typeof node) {
    case "object":
      if (Array.isArray(node)) {
        for (const n of node) {
          // Try all nodes to see if we can find the next node
          const nextNode = getNode(n, path);
          if (nextNode != null) {
            if (filter != null && !filter(path, nextNode)) {
              continue;
            }
            return nextNode;
          }
        }
      } else {
        return node[path];
      }
      break;
    case "boolean":
      return undefined;
    case "number":
      return undefined;
    case "string":
      return undefined;
    case "undefined":
      return undefined;
    default:
      return undefined;
  }
}

function findAttributeByPath(
  xmlObj: any,
  path: string[],
  attribute: string,
  nodeFilter?: (tag: string, node: any) => boolean
): string | undefined {
  let node = xmlObj;
  for (const nextPath of path) {
    node = getNode(node, nextPath, nodeFilter);
  }

  if (typeof node === "object") {
    if (Array.isArray(node)) {
      for (const n of node) {
        if (n[attribute] != null) {
          return n[attribute];
        }
      }
    } else {
      return node[attribute];
    }
  }

  return undefined;
}

function isSupportedLogoutBinding(raw: unknown): raw is SAMLBinding {
  if (typeof raw !== "string") {
    return false;
  }
  switch (raw) {
    case SAMLBinding.HTTPPOST:
      return true;
    case SAMLBinding.HTTPRedirect:
      return true;
    default:
      return false;
  }
}

function toPem(data: string): string {
  let pem = "-----BEGIN CERTIFICATE-----\n";
  pem = pem + data.trim() + "\n";
  pem = pem + "-----END CERTIFICATE-----";
  return pem;
}

export function formatCertificateFilename(
  configAppID: string,
  fingerprint: string
): string {
  return `${configAppID}-${fingerprint.replace(/:/g, "")}.pem`;
}
