import React from "react";
import { Route, RouteProps } from "react-router";
import Authenticated from "./graphql/portal/Authenticated";

interface AppRouteProps extends RouteProps {
  requireAuth?: boolean;
}

// requireAuth on AppRoute only works on element, don't wrap children inside for now
export const AppRoute: React.FC<AppRouteProps> = function AppRoute({
  requireAuth,
  ...routeProps
}) {
  if (requireAuth) {
    if (routeProps.element) {
      routeProps.element = <Authenticated>{routeProps.element}</Authenticated>;
    }
  }
  return <Route {...routeProps} />;
};
