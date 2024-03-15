import React from "react";
import { Text } from "@fluentui/react";

export const DebugScreen: React.VFC = function DebugScreen() {
  return (
    <main>
      <Text as="h1" variant="xxLarge">
        For ui components debugging
      </Text>
    </main>
  );
};

export default DebugScreen;
