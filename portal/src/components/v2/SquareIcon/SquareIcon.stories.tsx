import type { Meta, StoryObj } from "@storybook/react-vite";
import { SquareIcon, SquareIconProps } from "./SquareIcon";
import React from "react";
import { ImageIcon, ButtonIcon, InputIcon } from "@radix-ui/react-icons";

enum DemoIcon {
  "Image" = "Image",
  "Button" = "Button",
  "Input" = "Input",
}
function toIcon(demoIcon: DemoIcon) {
  switch (demoIcon) {
    case DemoIcon.Image:
      return ImageIcon;
    case DemoIcon.Button:
      return ButtonIcon;
    case DemoIcon.Input:
      return InputIcon;
  }
}

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: SquareIcon,
  tags: ["autodocs"],
  args: {
    Icon: ImageIcon,
    icon: DemoIcon.Image,
    size: "7",
    radius: "3",
    className: "text-[var(--accent-9)]",
  },
  argTypes: {
    Icon: { table: { disable: true } },
    icon: {
      control: "radio",
      options: [DemoIcon.Image, DemoIcon.Button, DemoIcon.Input],
    },
  },
  render: (args) => {
    return <Demo {...args} />;
  },
} satisfies Meta<React.ExoticComponent<{ icon: DemoIcon } & SquareIconProps>>;

function Demo({
  Icon: _,
  icon,
  ...args
}: { icon: DemoIcon } & SquareIconProps) {
  return <SquareIcon Icon={toIcon(icon)} {...args} />;
}

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};

export const Gray: Story = {
  args: {
    className: "text-[var(--gray-11)]",
    backgroundColor: "var(--gray-3)",
  },
};
