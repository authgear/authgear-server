import type { Meta, StoryObj } from "@storybook/react-vite";
import { M2MCard } from "./M2MCard";

const meta = {
  title: "portal/CreateOAuthClient/M2MCard",
  component: M2MCard,
  args: {
    onSelect: () => {},
  },
} satisfies Meta<typeof M2MCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
