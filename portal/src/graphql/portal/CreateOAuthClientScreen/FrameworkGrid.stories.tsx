import type { Meta, StoryObj } from "@storybook/react-vite";
import { FrameworkGrid } from "./FrameworkGrid";

const meta = {
  title: "portal/CreateOAuthClient/FrameworkGrid",
  component: FrameworkGrid,
  args: {
    selectedId: null,
    onSelect: () => {},
    onSelectM2M: () => {},
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

export const FlaskSelected: Story = {
  args: {
    selectedId: "flask",
  },
};

export const MobileSelected: Story = {
  args: {
    selectedId: "ios",
  },
};
