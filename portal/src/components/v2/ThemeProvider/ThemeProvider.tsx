import React from "react";
import { Theme } from "@radix-ui/themes";

export function ThemeProvider({
  children,
}: {
  children?: React.ReactNode;
}): React.ReactElement {
  return (
    <Theme
      // We only want Theme as a variable and context provider, and don't want it to affect the layout
      className="contents"
      accentColor="indigo"
    >
      {children}
    </Theme>
  );
}
