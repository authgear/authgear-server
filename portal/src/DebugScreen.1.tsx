import React from "react";
import { Text } from "@fluentui/react";
import { DebugSearchableDropdown } from "./components/debug/DebugSearchableDropdown";

export const DebugScreen: React.VFC = function DebugScreen() {
  return (
    <main className="p-4">
      <Text as="h1" className="block py-2" variant="xxLarge">
        For ui components debugging
      </Text>
      <section>
        <div className="py-4">
          <Text as="h2" className="block py-2" variant="large">
            SearchableDropdown
          </Text>
          <div className="w-60">
            <DebugSearchableDropdown />
          </div>
        </div>
      </section>
    </main>
  );
};
