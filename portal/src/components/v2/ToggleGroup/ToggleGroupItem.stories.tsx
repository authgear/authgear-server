import React from "react";
import cn from "classnames";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { ToggleGroupItem } from "./ToggleGroup";

enum DemoIcon {
  "email" = "email",
  "phone" = "phone",
}

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: ToggleGroupItem,
  tags: ["autodocs"],
  args: {
    value: "example",
    text: "Example",
  },
  argTypes: {
    icon: { control: "radio", options: [DemoIcon.email, DemoIcon.phone] },
    onCheckedChange: { table: { disable: true } },
  },
  render: (args) => {
    return (
      <div style={{ width: "386px" }}>
        <ToggleGroupItem
          {...args}
          icon={
            args.icon != null ? (
              <DemoIconComponent icon={args.icon as DemoIcon} />
            ) : null
          }
        />
      </div>
    );
  },
} satisfies Meta<typeof ToggleGroupItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};

export const WithSupportingText: Story = {
  args: {
    supportingText: "Supporting text",
  },
};

export const WithIcon: Story = {
  args: {
    icon: DemoIcon.email,
  },
};

function DemoIconComponent({ icon }: { icon: DemoIcon }) {
  let iconClassName = "";
  switch (icon) {
    case DemoIcon.email:
      iconClassName = "fa-envelope";

      break;
    case DemoIcon.phone:
      iconClassName = "fa-mobile";
      break;
  }
  return (
    <div className="h-5 w-5 grid items-center justify-center">
      <i className={cn("fa", iconClassName, "text-xl")} />
    </div>
  );
}
