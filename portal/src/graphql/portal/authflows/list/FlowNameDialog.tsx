import React, { useContext, useState } from "react";
import { Button, Dialog, Text, TextField } from "@radix-ui/themes";
import { Context } from "../../../../intl";
import { SecondaryButton } from "../../../../components/v2/Button/SecondaryButton/SecondaryButton";
import { NameError } from "./actions";
import styles from "./FlowNameDialog.module.css";

// The schema's limit on a flow's display_name.
export const DISPLAY_NAME_MAX = 100;

export const nameErrorMessageIDs: Record<NameError, string> = {
  invalid: "AuthFlowsScreen.list.name.invalid",
  taken: "AuthFlowsScreen.list.name.taken",
  reserved: "AuthFlowsScreen.list.name.reserved",
};

export interface FlowNameDialogProps {
  open: boolean;
  title: string;
  confirmText: string;
  initialName: string;
  // A flow's display name; when given, the dialog asks for it above the
  // name.
  initialDisplayName?: string;
  // Why a name can't be used; null when it can.
  validate: (name: string) => NameError | null;
  // How each error reads; flow names by default.
  errorMessageIDs?: Record<NameError, string>;
  // What else changes with the name, under the field.
  notice?: string | null;
  onSubmit: (name: string, displayName: string) => void;
  onCancel: () => void;
}

// Asks for a flow or flow group name, for Rename, Duplicate and New group.
// Mounted per use, so the field starts from initialName each time.
export function FlowNameDialog({
  open,
  title,
  confirmText,
  initialName,
  initialDisplayName,
  validate,
  errorMessageIDs = nameErrorMessageIDs,
  notice,
  onSubmit,
  onCancel,
}: FlowNameDialogProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const [name, setName] = useState(initialName);
  const [displayName, setDisplayName] = useState(initialDisplayName ?? "");
  const hasDisplayName = initialDisplayName != null;
  const error = validate(name);
  const displayNameTooLong = displayName.trim().length > DISPLAY_NAME_MAX;
  const submit = () => {
    if (error == null && !displayNameTooLong) {
      onSubmit(name, displayName.trim());
    }
  };
  return (
    <Dialog.Root
      open={open}
      onOpenChange={(o) => {
        if (!o) {
          onCancel();
        }
      }}
    >
      <Dialog.Content maxWidth="400px" size="3">
        <Dialog.Title>{title}</Dialog.Title>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            submit();
          }}
        >
          {hasDisplayName ? (
            <label className={styles.field}>
              <Text size="2" weight="medium">
                {renderToString("AuthFlowsScreen.list.display-name")}
              </Text>
              <TextField.Root
                value={displayName}
                onChange={(e) => setDisplayName(e.target.value)}
                autoFocus={true}
              />
              {displayNameTooLong ? (
                <Text as="p" size="1" color="red">
                  {renderToString(
                    "AuthFlowsScreen.list.display-name.too-long",
                    { max: DISPLAY_NAME_MAX }
                  )}
                </Text>
              ) : null}
            </label>
          ) : null}
          <label className={styles.field}>
            {hasDisplayName ? (
              <Text size="2" weight="medium">
                {renderToString("AuthFlowsScreen.list.flow-name")}
              </Text>
            ) : null}
            <TextField.Root
              value={name}
              onChange={(e) => setName(e.target.value)}
              aria-label={renderToString(
                hasDisplayName
                  ? "AuthFlowsScreen.list.flow-name"
                  : "AuthFlowsScreen.list.name"
              )}
              autoFocus={!hasDisplayName}
            />
          </label>
          {error != null && name !== initialName ? (
            <Text as="p" size="1" color="red" className={styles.error}>
              {renderToString(errorMessageIDs[error])}
            </Text>
          ) : null}
          {notice != null ? (
            <Text as="p" size="1" color="gray" className={styles.notice}>
              {notice}
            </Text>
          ) : null}
          <div className={styles.actions}>
            <SecondaryButton
              size="2"
              text={renderToString("AuthFlowsScreen.edit.cancel")}
              onClick={onCancel}
            />
            <Button
              type="submit"
              size="2"
              disabled={error != null || displayNameTooLong}
            >
              {confirmText}
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
