import React, { ComponentProps, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import CheckboxWithTooltip from "./CheckboxWithTooltip";

const meta = {
  title: "components/v1/Checkbox/CheckboxWithTooltip",
  component: CheckboxWithTooltip,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div style={{ width: 480 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    label: "Block plus sign (+) in local-part",
    tooltipMessageId:
      "LoginIDConfigurationScreen.email.blockPlusTooltipMessage",
  },
} satisfies Meta<typeof CheckboxWithTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

function CheckboxWithTooltipDefaultRender(
  args: ComponentProps<typeof CheckboxWithTooltip>
) {
  const [checked, setChecked] = useState(false);
  return (
    <CheckboxWithTooltip
      {...args}
      checked={checked}
      onChange={(_, v) => setChecked(v ?? false)}
    />
  );
}

function CheckboxWithTooltipCheckedRender(
  args: ComponentProps<typeof CheckboxWithTooltip>
) {
  const [checked, setChecked] = useState(true);
  return (
    <CheckboxWithTooltip
      {...args}
      checked={checked}
      onChange={(_, v) => setChecked(v ?? false)}
    />
  );
}

export const Default: Story = {
  render: (args) => <CheckboxWithTooltipDefaultRender {...args} />,
};

export const Checked: Story = {
  render: (args) => <CheckboxWithTooltipCheckedRender {...args} />,
};

export const Disabled: Story = {
  args: {
    label: "Option depends on another setting",
    disabled: true,
    checked: false,
    tooltipMessageId:
      "LoginIDConfigurationScreen.email.domainBlocklistTooltipMessage",
  },
};
