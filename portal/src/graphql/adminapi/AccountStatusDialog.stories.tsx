import React, { useCallback, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { MockedProvider } from "@apollo/client/testing";
import {
  AccountStatusDialog,
  type AccountStatus,
  type AccountStatusDialogProps,
} from "./UserDetailsAccountStatus";
import DefaultButton from "../../DefaultButton";
import { SystemConfigContext } from "../../context/SystemConfigContext";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "../../system-config";

const systemConfig = instantiateSystemConfig(defaultSystemConfig);

function AccountStatusDialogStoryChrome({
  Story,
  storyArgs,
}: {
  Story: React.ComponentType<{ args?: AccountStatusDialogProps }>;
  storyArgs: AccountStatusDialogProps;
}): React.ReactElement {
  const [open, setOpen] = useState(false);
  const onDismiss = useCallback(
    async (info: { deletedUser: boolean }) => {
      setOpen(false);
      await storyArgs.onDismiss?.(info);
    },
    [storyArgs]
  );

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: 12,
        width: "100%",
        boxSizing: "border-box",
      }}
    >
      <DefaultButton
        text="Open dialog"
        onClick={() => setOpen(true)}
        disabled={open}
      />
      <Story
        args={{
          ...storyArgs,
          isHidden: !open,
          onDismiss,
        }}
      />
    </div>
  );
}

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
  title: "components/v1/Dialog/AccountStatusDialog",
  component: AccountStatusDialog,
  tags: ["autodocs"],
  decorators: [
    (Story, context) => (
      <MockedProvider mocks={[]}>
        <SystemConfigContext.Provider value={systemConfig}>
          <AccountStatusDialogStoryChrome
            Story={Story}
            storyArgs={context.args as AccountStatusDialogProps}
          />
        </SystemConfigContext.Provider>
      </MockedProvider>
    ),
  ],
  argTypes: {
    isHidden: {
      control: false,
      description: "Visibility is controlled by the Open dialog button.",
    },
  },
  args: {
    isHidden: true,
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
