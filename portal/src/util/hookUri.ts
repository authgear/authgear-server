function getRawPathFromURL(url: string): string | null {
  const schemeEnd = url.indexOf("://");
  if (schemeEnd === -1) {
    return null;
  }

  const rest = url.slice(schemeEnd + 3);
  const pathStart = rest.indexOf("/");
  if (pathStart === -1) {
    return "";
  }

  const pathAndAfter = rest.slice(pathStart);
  const queryStart = pathAndAfter.indexOf("?");
  const hashStart = pathAndAfter.indexOf("#");
  let end = pathAndAfter.length;
  if (queryStart !== -1) {
    end = Math.min(end, queryStart);
  }
  if (hashStart !== -1) {
    end = Math.min(end, hashStart);
  }

  return pathAndAfter.slice(0, end);
}

function isValidAbsoluteURLPath(pathname: string): boolean {
  if (pathname === "") {
    return true;
  }
  if (!pathname.startsWith("/")) {
    return false;
  }

  const hasTrailingSlash = pathname.endsWith("/");
  const parts = pathname.split("/").filter((part) => part !== "" && part !== ".");
  const resolved: string[] = [];
  for (const part of parts) {
    if (part === "..") {
      if (resolved.length === 0) {
        return false;
      }
      resolved.pop();
    } else {
      resolved.push(part);
    }
  }

  let normalized = "/" + resolved.join("/");
  if (hasTrailingSlash && !normalized.endsWith("/")) {
    normalized += "/";
  }

  return normalized === pathname;
}

export function isValidWebhookHookURI(url: string): boolean {
  if (url === "") {
    return true;
  }

  let parsed: URL;
  try {
    parsed = new URL(url);
  } catch {
    return false;
  }

  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    return false;
  }

  if (!parsed.hostname) {
    return false;
  }

  const rawPath = getRawPathFromURL(url);
  if (rawPath == null) {
    return false;
  }

  return isValidAbsoluteURLPath(rawPath);
}
