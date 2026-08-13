import React, { ReactElement } from "react";
import { Link as RadixLink } from "@radix-ui/themes";

export interface ExternalLinkProps
  extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, "color"> {}

export const DEFAULT_EXTERNAL_LINK_PROPS = {
  target: "_blank",
  rel: "noreferrer",
};

export default function ExternalLink(props: ExternalLinkProps): ReactElement {
  // Force the accent color so the link stays a link even when nested inside a
  // gray <Text> (which otherwise reassigns the accent scale to gray for its
  // descendants, rendering the link gray/black instead of blue).
  return (
    <RadixLink color="indigo" {...DEFAULT_EXTERNAL_LINK_PROPS} {...props} />
  );
}
