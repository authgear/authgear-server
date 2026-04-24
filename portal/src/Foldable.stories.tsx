import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { Text } from "@fluentui/react";
import { FormattedMessage } from "./intl";
import { SystemConfigContext } from "./context/SystemConfigContext";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "./system-config";
import FoldableDiv from "./FoldableDiv";

const systemConfig = instantiateSystemConfig(defaultSystemConfig);

/**
 * `FoldableDiv`: primary label + `ChevronDown` / `ChevronUp`. **Add user** → **Advanced** uses
 * the same with `id="AdduserScreen.advanced"`; body is placeholder only here.
 */
// Same as `MessageBar` stories: global `layout: "centered"` in `.storybook/preview` centers
// the wrapper on both axes; do not set `docs.story.height` so the Docs block matches MessageBar.
// `minHeight` >= max(collapsed, expanded) so the wrapper height is stable on toggle; keep `flex-start` so the trigger row does not re-center in the column.
const FOLDABLE_SHEET_MIN_PX = 128;

const meta = {
  title: "components/v1/Foldable",
  tags: ["autodocs"],
  // Same as `MessageBar` — `preview` is `centered`; this keeps Default / Expanded horizontally & vertically centered in the story frame, and the same default Docs block height.
  parameters: {
    layout: "centered",
  },
  decorators: [
    (Story) => (
      <SystemConfigContext.Provider value={systemConfig}>
        <div
          style={{
            width: 480,
            minHeight: FOLDABLE_SHEET_MIN_PX,
            display: "flex",
            flexDirection: "column",
            justifyContent: "flex-start",
            textAlign: "left" as const,
          }}
        >
          <Story />
        </div>
      </SystemConfigContext.Provider>
    ),
  ],
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

const sampleBody = (
  <Text variant="small" block>
    Placeholder body.
  </Text>
);

function AdvancedFoldableSample(props: { initialFolded: boolean }) {
  const { initialFolded } = props;
  const [folded, setFolded] = useState(initialFolded);
  return (
    <FoldableDiv
      label={<FormattedMessage id="AdduserScreen.advanced" />}
      folded={folded}
      setFolded={setFolded}
    >
      {sampleBody}
    </FoldableDiv>
  );
}

export const Default: Story = {
  name: "Default",
  render: function FoldableDefault() {
    return <AdvancedFoldableSample initialFolded={true} />;
  },
};

export const Expanded: Story = {
  name: "Expanded",
  render: function FoldableExpanded() {
    return <AdvancedFoldableSample initialFolded={false} />;
  },
};
