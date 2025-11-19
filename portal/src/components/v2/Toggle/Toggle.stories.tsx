import type { Meta, StoryObj } from "@storybook/react-vite";
import { Toggle } from "./Toggle";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: Toggle,
  tags: ["autodocs"],
  argTypes: {
    onCheckedChange: { table: { disable: true } },
  },
} satisfies Meta<typeof Toggle>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithText: Story = {
  args: {
    text: "On",
  },
};
