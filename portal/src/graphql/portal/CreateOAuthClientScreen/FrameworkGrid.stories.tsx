import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { FrameworkGrid } from "./FrameworkGrid";

const meta = {
  title: "portal/CreateOAuthClient/FrameworkGrid",
  component: FrameworkGrid,
  args: {
    selectedId: null,
    onSelect: () => {},
  },
} satisfies Meta<typeof FrameworkGrid>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const ReactSelected: Story = {
  args: {
    selectedId: "react",
  },
};

export const DjangoSelected: Story = {
  args: {
    selectedId: "django",
  },
};

export const MobileSelected: Story = {
  args: {
    selectedId: "ios",
  },
};
