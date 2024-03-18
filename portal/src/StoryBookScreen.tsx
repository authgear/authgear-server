import React from "react";
import { Text } from "@fluentui/react";
import { SearchableDropdownStory } from "./components/stories/SearchableDropdownStory";

export const StoryBookScreen: React.VFC = function StoryBookScreen() {
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
            <SearchableDropdownStory />
          </div>
        </div>
      </section>
    </main>
  );
};

export default StoryBookScreen;
