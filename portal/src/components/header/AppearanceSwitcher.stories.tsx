import type { Meta, StoryObj } from "@storybook/react-vite";
import { AppearanceSwitcher } from "./AppearanceSwitcher";

const meta = {
  title: "components/header/AppearanceSwitcher",
  component: AppearanceSwitcher,
  tags: ["autodocs"],
} satisfies Meta<typeof AppearanceSwitcher>;

export default meta;
type Story = StoryObj<typeof meta>;

// Selecting an item changes the Storybook page's own appearance, the same
// way it changes the portal's.
export const Default: Story = {};
