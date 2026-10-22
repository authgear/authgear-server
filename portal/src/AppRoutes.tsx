import React from "react";
import { Outlet, Route, Routes } from "react-router-dom";
import { AppFeature } from "./appFeatures";
import AppFeatureRouteGuard from "./AppFeatureRouteGuard";

export interface AppRouteProps {
  // The feature gating this route and every route under it, or null if it adds
  // no gate. Required so that every route decides whether it is gated.
  feature: AppFeature | null;
  path?: string;
  index?: boolean;
  element?: React.ReactNode;
  children?: React.ReactNode;
}

// AppRoute is Route with a required `feature`. It is Route at runtime, because
// Routes accepts only Route children. It must be rendered inside AppRoutes,
// which applies the gate.
export const AppRoute = Route as unknown as React.VFC<AppRouteProps>;

function withFeatureGates(children: React.ReactNode): React.ReactNode {
  return React.Children.map(children, (child) => {
    if (!React.isValidElement<AppRouteProps>(child)) {
      return child;
    }
    const grandchildren = withFeatureGates(child.props.children);
    if (child.type === React.Fragment) {
      return React.cloneElement(child, undefined, grandchildren);
    }
    const { feature, element } = child.props;
    const props: Partial<AppRouteProps> = {};
    if (feature != null) {
      props.element = (
        <AppFeatureRouteGuard feature={feature}>
          {element ?? <Outlet />}
        </AppFeatureRouteGuard>
      );
    }
    return React.cloneElement(child, props, grandchildren);
  });
}

interface AppRoutesProps {
  children: React.ReactNode;
}

// AppRoutes is Routes for AppRoute children. A route whose `feature` is
// unavailable shows a "not available" message in place of its element.
export const AppRoutes: React.VFC<AppRoutesProps> = function AppRoutes({
  children,
}) {
  return <Routes>{withFeatureGates(children)}</Routes>;
};
