import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
// eslint-disable-next-line no-restricted-imports
import { Checkbox, Text } from "@fluentui/react";
import CheckboxWithContentLayout from "./CheckboxWithContentLayout";
import CheckboxWithTooltip from "./CheckboxWithTooltip";

const meta = {
  title: "components/v1/Checkbox/CheckboxWithContentLayout",
  tags: ["autodocs"],
  /** Omit `component` — required `children` would force every story to declare dummy `args`. */
  decorators: [
    (Story) => (
      <div style={{ width: 560 }}>
        <Story />
      </div>
    ),
  ],
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithPlainCheckbox: Story = {
  render: () => {
    const [checked, setChecked] = useState(true);
    return (
      <CheckboxWithContentLayout>
        <Checkbox
          label="Enable domain blocklist"
          checked={checked}
          onChange={(_, v) => setChecked(v ?? false)}
        />
        <Text variant="small">
          When enabled, sign-ins from domains in the list below are blocked.
          One domain per line.
        </Text>
      </CheckboxWithContentLayout>
    );
  },
};

export const WithCheckboxAndTooltip: Story = {
  render: () => {
    const [checked, setChecked] = useState(false);
    return (
      <CheckboxWithContentLayout>
        <CheckboxWithTooltip
          label="Enable domain blocklist"
          checked={checked}
          onChange={(_, v) => setChecked(v ?? false)}
          tooltipMessageId="LoginIDConfigurationScreen.email.domainBlocklistTooltipMessage"
        />
        <Text variant="small" styles={{ root: { marginTop: 8 } }}>
          Additional controls appear here when the checkbox is on (pattern
          from login ID email settings).
        </Text>
      </CheckboxWithContentLayout>
    );
  },
};
