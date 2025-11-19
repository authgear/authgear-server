import type { Meta, StoryObj } from "@storybook/react-vite";
import { IconButton, IconButtonIcon } from "./IconButton";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: IconButton,
  tags: ["autodocs"],
  argTypes: {},
  args: {
    size: "2",
  },
} satisfies Meta<typeof IconButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    variant: "default",
    icon: IconButtonIcon.MagnifyingGlass,
  },
};

export const Destroy: Story = {
  args: {
    variant: "destroy",
    icon: IconButtonIcon.Trash,
  },
};
