import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { DirectionalHint } from "@fluentui/react";
import LabelWithTooltip from "./LabelWithTooltip";

const meta = {
  title: "components/v1/Tooltip/LabelWithTooltip",
  component: LabelWithTooltip,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div
        style={{
          width: "100%",
          maxWidth: 560,
          minHeight: 120,
          margin: "0 auto",
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          padding: "32px 16px",
          boxSizing: "border-box",
        }}
      >
        <Story />
      </div>
    ),
  ],
  args: {
    labelId:
      "SingleSignOnConfigurationScreen.widget.wechat-redirect-uris-label",
    tooltipMessageId:
      "SingleSignOnConfigurationScreen.widget.wechat-redirect-uris-tooltip-message",
  },
} satisfies Meta<typeof LabelWithTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithTooltipHeader: Story = {
  args: {
    tooltipHeaderId:
      "SingleSignOnConfigurationScreen.widget.wechat-redirect-uris-label",
  },
};

export const Required: Story = {
  args: {
    required: true,
  },
};

export const WithLabelIcon: Story = {
  args: {
    labelIIconProps: { iconName: "Mail" },
  },
};

export const BottomLeftEdge: Story = {
  args: {
    directionalHint: DirectionalHint.bottomLeftEdge,
  },
};
