import React, { useCallback, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { FormattedMessage } from "../../../intl";
import { SecondaryButton } from "../Button/SecondaryButton/SecondaryButton";
import { ConfirmationDialog } from "./ConfirmationDialog";

function ConfirmationDialogDemo({
  loading = false,
  confirmColor = "red",
}: {
  loading?: boolean;
  confirmColor?: "red" | "indigo";
}): React.ReactElement {
  const [open, setOpen] = useState(false);

  const onOpenChange = useCallback(
    (nextOpen: boolean) => {
      if (!loading) {
        setOpen(nextOpen);
      }
    },
    [loading]
  );

  const onCancel = useCallback(() => {
    setOpen(false);
  }, []);

  const onConfirm = useCallback(() => {
    setOpen(false);
  }, []);

  return (
    <>
      <SecondaryButton
        size="2"
        text="Open Dialog"
        onClick={() => setOpen(true)}
      />
      <ConfirmationDialog
        open={open}
        onOpenChange={onOpenChange}
        title={
          <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete-dialog.title" />
        }
        description={
          <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete-dialog.message" />
        }
        confirmText={
          <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete-dialog.confirm" />
        }
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirm}
        onCancel={onCancel}
        loading={loading}
        confirmColor={confirmColor}
      />
    </>
  );
}

const meta = {
  component: ConfirmationDialogDemo,
  tags: ["autodocs"],
  args: {
    loading: false,
    confirmColor: "red",
  },
} satisfies Meta<typeof ConfirmationDialogDemo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const DeleteAdminAPIKey: Story = {
  name: "Delete Admin API key",
};

export const Loading: Story = {
  args: {
    loading: true,
  },
};
