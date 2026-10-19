import React, { useContext, useState } from "react";
import { Button, Dialog, Text, TextField } from "@radix-ui/themes";
import { Context } from "../../../../intl";
import { AuthenticationFlowsConfig } from "../../../../types";
import { SecondaryButton } from "../../../../components/v2/Button/SecondaryButton/SecondaryButton";
import { FlowType } from "../model";
import { NameError, nameError } from "./actions";
import styles from "./FlowNameDialog.module.css";

export const nameErrorMessageIDs: Record<NameError, string> = {
  invalid: "AuthFlowsScreen.list.name.invalid",
  taken: "AuthFlowsScreen.list.name.taken",
  reserved: "AuthFlowsScreen.list.name.reserved",
};

export interface FlowNameDialogProps {
  open: boolean;
  title: string;
  confirmText: string;
  // The type the name must be free in.
  type: FlowType;
  initialName: string;
  flows: AuthenticationFlowsConfig | undefined;
  onSubmit: (name: string) => void;
  onCancel: () => void;
}

// Asks for a flow name, for Rename and Duplicate. Mounted per use, so the
// field starts from initialName each time.
export function FlowNameDialog({
  open,
  title,
  confirmText,
  type,
  initialName,
  flows,
  onSubmit,
  onCancel,
}: FlowNameDialogProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const [name, setName] = useState(initialName);
  const error = nameError(flows, type, name);
  const submit = () => {
    if (error == null) {
      onSubmit(name);
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
          <TextField.Root
            value={name}
            onChange={(e) => setName(e.target.value)}
            aria-label={renderToString("AuthFlowsScreen.list.name")}
            autoFocus={true}
          />
          {error != null && name !== initialName ? (
            <Text as="p" size="1" color="red" className={styles.error}>
              {renderToString(nameErrorMessageIDs[error])}
            </Text>
          ) : null}
          <div className={styles.actions}>
            <SecondaryButton
              size="2"
              text={renderToString("AuthFlowsScreen.edit.cancel")}
              onClick={onCancel}
            />
            <Button type="submit" size="2" disabled={error != null}>
              {confirmText}
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
