import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import TextFieldWithCopyButton from "./TextFieldWithCopyButton";
import { SystemConfigContext } from "./context/SystemConfigContext";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "./system-config";

const systemConfig = instantiateSystemConfig(defaultSystemConfig);

const meta = {
  title: "components/v1/TextFieldWithCopyButton",
  component: TextFieldWithCopyButton,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <SystemConfigContext.Provider value={systemConfig}>
        <div style={{ width: 480 }}>
          <Story />
        </div>
      </SystemConfigContext.Provider>
    ),
  ],
  args: {
    label: "API Endpoint",
    value: "https://example.authgear.cloud/_api/admin/graphql",
    readOnly: true,
  },
} satisfies Meta<typeof TextFieldWithCopyButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Empty: Story = {
  args: {
    value: "",
  },
};

export const Disabled: Story = {
  args: {
    disabled: true,
  },
};

export const HideCopyButton: Story = {
  args: {
    hideCopyButton: true,
  },
};

export const WithAdditionalIconButtons: Story = {
  args: {
    additionalIconButtons: [
      {
        iconProps: { iconName: "Edit" },
        ariaLabel: "Edit",
        title: "Edit",
      },
      {
        iconProps: { iconName: "Delete" },
        ariaLabel: "Delete",
        title: "Delete",
      },
    ],
  },
};

export const ProjectID: Story = {
  args: {
    label: "Project ID",
    value: "my-project",
  },
};
