import React, { useCallback, useContext, useMemo, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
// eslint-disable-next-line no-restricted-imports
import { Dialog, DialogFooter } from "@fluentui/react";
import { Context, FormattedMessage } from "./intl";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import { SystemConfigContext } from "./context/SystemConfigContext";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "./system-config";

const systemConfig = instantiateSystemConfig(defaultSystemConfig);

/**
 * Fluent **Dialog** patterns used across the portal (e.g. `FormContainer` discard /
 * “Discard unsaved changes?”).
 */
const meta = {
  title: "components/v1/Dialog/Patterns",
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <SystemConfigContext.Provider value={systemConfig}>
        <div
          style={{
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
            minHeight: 96,
            padding: 12,
            boxSizing: "border-box",
          }}
        >
          <Story />
        </div>
      </SystemConfigContext.Provider>
    ),
  ],
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

/** Same copy and actions as `FormContainer` reset / discard dialog. */
function DiscardUnsavedChangesDemo(): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { themes } = systemConfig;
  const [open, setOpen] = useState(false);

  const dialogContentProps = useMemo(
    () => ({
      title: <FormattedMessage id="FormContainer.reset-dialog.title" />,
      subText: renderToString("FormContainer.reset-dialog.message"),
    }),
    [renderToString]
  );

  const close = useCallback(() => setOpen(false), []);
  const onDiscard = useCallback(() => {
    close();
  }, [close]);

  return (
    <>
      <DefaultButton text="Open Dialog" onClick={() => setOpen(true)} />
      <Dialog
        hidden={!open}
        dialogContentProps={dialogContentProps}
        onDismiss={close}
        styles={{ main: { minHeight: 0 } }}
      >
        <DialogFooter>
          <PrimaryButton
            onClick={onDiscard}
            theme={themes.destructive}
            text={<FormattedMessage id="FormContainer.reset-dialog.confirm" />}
          />
          <DefaultButton
            onClick={close}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    </>
  );
}

export const DiscardUnsavedChanges: Story = {
  name: "Discard unsaved changes (FormContainer)",
  render: () => <DiscardUnsavedChangesDemo />,
};

/** Neutral confirm + cancel (no destructive primary). */
function SimpleConfirmDemo(): React.ReactElement {
  const [open, setOpen] = useState(false);
  const dialogContentProps = useMemo(
    () => ({
      title: <FormattedMessage id="confirm" />,
      subText: "This is sample body copy for a non-destructive confirmation.",
    }),
    []
  );

  return (
    <>
      <DefaultButton text="Open confirm dialog" onClick={() => setOpen(true)} />
      <Dialog
        hidden={!open}
        dialogContentProps={dialogContentProps}
        onDismiss={() => setOpen(false)}
        styles={{ main: { minHeight: 0 } }}
      >
        <DialogFooter>
          <PrimaryButton
            onClick={() => setOpen(false)}
            text={<FormattedMessage id="confirm" />}
          />
          <DefaultButton
            onClick={() => setOpen(false)}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    </>
  );
}

export const SimpleConfirm: Story = {
  name: "Simple confirm",
  render: () => <SimpleConfirmDemo />,
};
