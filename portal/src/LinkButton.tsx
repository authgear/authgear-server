import React, { ReactElement } from "react";
// eslint-disable-next-line no-restricted-imports
import { Link as FluentLink, ILinkProps } from "@fluentui/react";

export interface LinkButtonProps extends Omit<ILinkProps, "href" | "as"> {}

export default function LinkButton(props: LinkButtonProps): ReactElement {
  return <FluentLink {...props} as="button" />;
}
