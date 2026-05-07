import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
// eslint-disable-next-line no-restricted-imports
import { DefaultButton } from "@fluentui/react";
import Tooltip, { TooltipIcon } from "./Tooltip";

const meta = {
  title: "components/v1/Tooltip/Tooltip",
  component: Tooltip,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div
        style={{
          width: "100%",
          maxWidth: 560,
          minHeight: 160,
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
    tooltipMessageId:
      "LoginIDConfigurationScreen.email.blockPlusTooltipMessage",
  },
} satisfies Meta<typeof Tooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  /** Renders the default info icon trigger. */
};

export const WithCustomTrigger: Story = {
  args: {
    tooltipMessageId:
      "LoginIDConfigurationScreen.email.domainBlocklistTooltipMessage",
    children: <DefaultButton text="Why is this disabled?" />,
  },
};

export const Hidden: Story = {
  args: {
    tooltipMessageId:
      "LoginIDConfigurationScreen.email.blockPlusTooltipMessage",
    isHidden: true,
  },
};

export const TooltipIconOnly: Story = {
  name: "TooltipIcon",
  render: () => (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        gap: 12,
        flexWrap: "wrap",
        textAlign: "center",
      }}
    >
      <TooltipIcon />
      <span style={{ fontSize: 12, opacity: 0.8, maxWidth: 280 }}>
        Info icon used as the default trigger when `children` is omitted.
      </span>
    </div>
  ),
};
