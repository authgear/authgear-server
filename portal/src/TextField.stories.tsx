import React, { ComponentProps, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import TextField from "./TextField";

const meta = {
  title: "components/v1/TextField",
  component: TextField,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div style={{ width: 480 }}>
        <Story />
      </div>
    ),
  ],
  args: {
    label: "Display name",
    placeholder: "Jane Doe",
  },
} satisfies Meta<typeof TextField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const WithDescription: Story = {
  args: {
    label: "Webhook URL",
    description:
      "Must be HTTPS. Authgear sends verification events to this endpoint.",
    placeholder: "https://example.com/hooks/authgear",
  },
};

function TextFieldControlledRender(args: ComponentProps<typeof TextField>) {
  const [value, setValue] = useState(args.value ?? "");
  return (
    <TextField {...args} value={value} onChange={(_, v) => setValue(v ?? "")} />
  );
}

export const Controlled: Story = {
  render: (args) => <TextFieldControlledRender {...args} />,
  args: {
    label: "Project slug",
    value: "my-app",
  },
};

export const ReadOnly: Story = {
  args: {
    label: "App ID",
    value: "my-app.authgear.cloud",
    readOnly: true,
  },
};

export const Disabled: Story = {
  args: {
    label: "Disabled field",
    value: "Cannot edit",
    disabled: true,
  },
};

export const WithError: Story = {
  args: {
    label: "Email",
    value: "not-an-email",
    errorMessage: "Enter a valid email address.",
  },
};
