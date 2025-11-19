import type { Meta, StoryObj } from "@storybook/react-vite";
import { WhiteButton } from "./WhiteButton";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: WhiteButton,
  tags: ["autodocs"],
  argTypes: {
    text: {
      control: {
        type: "text",
      },
    },
  },
  args: {
    text: "Start",
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
} satisfies Meta<typeof WhiteButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    size: "2",
    disabled: false,
  },
};
