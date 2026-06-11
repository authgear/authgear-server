import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { SettingsSectionCard } from "./SettingsSectionCard";
import { TextField } from "../TextField/TextField";

const meta = {
  component: SettingsSectionCard,
  tags: ["autodocs"],
  args: {
    title: "Deletion Schedule",
  },
} satisfies Meta<typeof SettingsSectionCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    contentClassName: "gap-4",
    children: (
      <>
        <TextField size="2" label="Grace period (days)" value="30" />
        <TextField size="2" label="Reason" optional={true} value="" />
      </>
    ),
  },
};

export const SingleField: Story = {
  args: {
    children: <TextField size="2" label="Grace period (days)" value="30" />,
  },
};
