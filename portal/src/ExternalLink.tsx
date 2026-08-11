import React, { ReactElement } from "react";
import { Link as RadixLink } from "@radix-ui/themes";

export interface ExternalLinkProps
  extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, "color"> {}

export const DEFAULT_EXTERNAL_LINK_PROPS = {
  target: "_blank",
  rel: "noreferrer",
};

export default function ExternalLink(props: ExternalLinkProps): ReactElement {
  return <RadixLink {...DEFAULT_EXTERNAL_LINK_PROPS} {...props} />;
}
