import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { MockedProvider } from "@apollo/client/testing";
import {
  AccountStatusDialog,
  type AccountStatus,
} from "./UserDetailsAccountStatus";
import { SystemConfigContext } from "../../context/SystemConfigContext";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "../../system-config";

const systemConfig = instantiateSystemConfig(defaultSystemConfig);

const baseAccountStatus: AccountStatus = {
  id: "user_1234567890",
  endUserAccountID: "alice",
  isDisabled: false,
  isAnonymized: false,
  disableReason: null,
  accountValidFrom: null,
  accountValidUntil: null,
  temporarilyDisabledFrom: null,
  temporarilyDisabledUntil: null,
  anonymizeAt: null,
  deleteAt: null,
};

const meta = {
  title: "components/v1/AccountStatusDialog",
  component: AccountStatusDialog,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <MockedProvider mocks={[]}>
        <SystemConfigContext.Provider value={systemConfig}>
          <Story />
        </SystemConfigContext.Provider>
      </MockedProvider>
    ),
  ],
  parameters: {
    layout: "padded",
  },
  args: {
    isHidden: false,
    mode: "delete-or-schedule",
    accountStatus: baseAccountStatus,
    onDismiss: async () => {},
  },
} satisfies Meta<typeof AccountStatusDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const DeleteOrSchedule: Story = {};

export const DeleteImmediately: Story = {
  args: {
    mode: "delete-immediately",
  },
};

export const CancelScheduledDeletion: Story = {
  args: {
    mode: "cancel-deletion",
    accountStatus: {
      ...baseAccountStatus,
      deleteAt: "2030-01-01T00:00:00.000Z",
    },
  },
};
