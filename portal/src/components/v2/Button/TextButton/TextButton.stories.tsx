import type { Meta, StoryObj } from "@storybook/react-vite";
import { TextButton, TextButtonIcon } from "./TextButton";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: TextButton,
  tags: ["autodocs"],
  argTypes: {
    text: {
      control: {
        type: "text",
      },
    },
    iconStart: {
      options: ["none", ...Object.keys(TextButtonIcon)],
      mapping: {
        none: undefined,
        ...Object.keys(TextButtonIcon).reduce<Record<string, any>>(
          (mapping, it) => {
            mapping[it] = TextButtonIcon[
              it as keyof typeof TextButtonIcon
            ] as any;
            return mapping;
          },
          {}
        ),
      },
    },
  },
  args: {
    text: "Button",
    size: "4",
  },
} satisfies Meta<typeof TextButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    variant: "default",
    darkMode: false,
    disabled: false,
  },
};

export const DarkSecondary: Story = {
  args: {
    variant: "secondary",
    darkMode: true,
    disabled: false,
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
};

export const DarkSecondaryBack: Story = {
  args: {
    variant: "secondary",
    darkMode: true,
    disabled: false,
    iconStart: TextButtonIcon.Back,
    text: "Back",
  },
  parameters: {
    backgrounds: {
      default: "dark",
    },
  },
};
