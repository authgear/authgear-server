import type { Meta, StoryObj } from "@storybook/react-vite";
import { PrimaryButton } from "./PrimaryButton";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: PrimaryButton,
  tags: ["autodocs"],
  argTypes: {
    text: {
      control: {
        type: "text",
      },
    },
    darkMode: {
      description:
        "When highContrast=false, PrimaryButton in darkMode is visually the same as its non-darkMode counterpart. So you won't see any difference by toggling this. See https://www.radix-ui.com/themes/playground#button",
    },
  },
  args: {
    text: "Start",
  },
} satisfies Meta<typeof PrimaryButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    size: "2",
    highContrast: false,
    disabled: false,
  },
};

export const DarkHighContrast: Story = {
  args: {
    darkMode: true,
    size: "2",
    highContrast: true,
    disabled: false,
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
};
