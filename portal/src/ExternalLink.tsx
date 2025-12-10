import React, { ReactElement } from "react";
// eslint-disable-next-line no-restricted-imports
import { ILinkProps, Link as FluentLink } from "@fluentui/react";

export interface ExternalLinkProps extends Omit<ILinkProps, "rel"> {}

export const DEFAULT_EXTERNAL_LINK_PROPS = {
  target: "_blank",
  rel: "noreferrer",
};

export default function ExternalLink(props: ExternalLinkProps): ReactElement {
  return <FluentLink {...DEFAULT_EXTERNAL_LINK_PROPS} {...props} />;
}
