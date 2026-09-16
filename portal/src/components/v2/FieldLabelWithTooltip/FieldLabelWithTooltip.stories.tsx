import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { Text } from "@radix-ui/themes";
import { FieldLabelWithTooltip } from "./FieldLabelWithTooltip";

const meta = {
  component: FieldLabelWithTooltip,
  tags: ["autodocs"],
  args: {
    children: "Description",
    tooltip: "Shown to users on the consent screen.",
    tooltipLabel: "Shown to users on the consent screen.",
  },
  render: (args) => {
    return (
      <Text as="p" size="2" weight="medium">
        <FieldLabelWithTooltip {...args} />
      </Text>
    );
  },
} satisfies Meta<typeof FieldLabelWithTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};

export const WithRequiredMark: Story = {
  args: {
    children: (
      <>
        Identifier
        <span style={{ color: "var(--red-11)" }} aria-hidden="true">
          *
        </span>
      </>
    ),
  },
};
