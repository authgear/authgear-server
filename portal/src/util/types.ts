import { PortalAPIApp } from "../types";
import { UserDetailsScreenQuery_node_User } from "../graphql/adminapi/__generated__/UserDetailsScreenQuery";

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

export function isUserDetails(
  value: any
): value is UserDetailsScreenQuery_node_User {
  if (!(value instanceof Object)) {
    return false;
  }
  return value.__typename === "User";
}
