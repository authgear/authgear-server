import React, { ReactElement } from "react";
// eslint-disable-next-line no-restricted-imports
import { Link as FluentLink } from "@fluentui/react";
import ReactRouterLink, { ReactRouterLinkProps } from "./ReactRouterLink";

export interface LinkProps extends Omit<ReactRouterLinkProps, "component"> {}

// We finally generalize 3 use cases of Link.
// They are Link, ExternalLink and LinkButton.
// Use Link when you want to render an internal link.
// Use ExternalLink when you want to render an external link.
// Use LinkButton when you want to show a button that looks like a link.
export default function Link(props: LinkProps): ReactElement {
  return <ReactRouterLink {...props} component={FluentLink} />;
}
