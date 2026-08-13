import React, { ReactElement } from "react";
import { Link as RadixLink } from "@radix-ui/themes";
import {
  ReactRouterLinkComponent,
  ReactRouterLinkPropsBase,
} from "./ReactRouterLink";

export interface LinkProps
  extends Omit<ReactRouterLinkPropsBase, "component">,
    Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, "href" | "color"> {}

// We finally generalize 3 use cases of Link.
// They are Link, ExternalLink and LinkButton.
// Use Link when you want to render an internal link.
// Use ExternalLink when you want to render an external link.
// Use LinkButton when you want to show a button that looks like a link.
export default function Link(props: LinkProps): ReactElement {
  // color="indigo" pins the link accent so it stays link-colored even inside a
  // gray <Text>, which would otherwise remap the accent scale to gray. Matches
  // ExternalLink and LinkButton.
  return (
    <ReactRouterLinkComponent color="indigo" {...props} component={RadixLink} />
  );
}
