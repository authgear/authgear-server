import React, { ReactElement } from "react";
import { Link as RadixLink } from "@radix-ui/themes";

export interface LinkButtonProps
  extends Omit<
    React.ButtonHTMLAttributes<HTMLButtonElement>,
    "type" | "color"
  > {}

// A button that looks like a link. See Link / ExternalLink for the
// anchor counterparts.
export default function LinkButton(props: LinkButtonProps): ReactElement {
  return (
    // color="indigo" pins the link accent so it stays link-colored even
    // inside a gray Text, which would otherwise remap the accent scale.
    <RadixLink asChild={true} color="indigo">
      <button type="button" {...props} />
    </RadixLink>
  );
}
