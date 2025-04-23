import React from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { ToggleGroupItem } from "./ToggleGroup";
import loginEmailIcon from "../../../images/login_email.svg";
import loginPhoneIcon from "../../../images/login_phone.svg";

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
          icon={args.icon != null ? toIcon(args.icon as DemoIcon) : null}
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

function toIcon(demoIcon: DemoIcon): React.ReactElement {
  switch (demoIcon) {
    case DemoIcon.email:
      return <img src={loginEmailIcon} width={20} height={20} />;
    case DemoIcon.phone:
      return <img src={loginPhoneIcon} width={20} height={20} />;
  }
}
