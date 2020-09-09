import { PortalAPIApp } from "../types";

export function nonNullable<T>(value: T): value is NonNullable<T> {
  return value != null;
}

export function isPortalApiApp(value: any): value is PortalAPIApp {
  if (!(value instanceof Object)) {
    return false;
  }
  if (!value.id) {
    return false;
  }
  return value.__typename === "App";
}
