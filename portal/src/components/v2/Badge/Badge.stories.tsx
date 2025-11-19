import type { Meta, StoryObj } from "@storybook/react-vite";
import { Badge } from "./Badge";

const meta = {
  component: Badge,
  tags: ["autodocs"],
  args: {
    size: "1",
    text: "Badge",
  },
  argTypes: {
    text: {
      control: {
        type: "text",
      },
    },
  },
} satisfies Meta<typeof Badge>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Info: Story = {
  args: {
    variant: "info",
  },
};

export const Neutral: Story = {
  args: {
    variant: "neutral",
  },
};

export const Success: Story = {
  args: {
    variant: "success",
  },
};

export const Warning: Story = {
  args: {
    variant: "warning",
  },
};

export const Error: Story = {
  args: {
    variant: "error",
  },
};
