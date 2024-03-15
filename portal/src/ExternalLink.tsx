import React, { ReactElement } from "react";
// eslint-disable-next-line no-restricted-imports
import { ILinkProps, Link as FluentLink } from "@fluentui/react";

export interface ExternalLinkProps extends Omit<ILinkProps, "rel"> {}

export default function ExternalLink(props: ExternalLinkProps): ReactElement {
  return <FluentLink target="_blank" rel="noreferrer" {...props} />;
}
