import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Button, Dialog } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { useCreateAuthenticatorMutation } from "../../graphql/adminapi/mutations/createAuthenticatorMutation";
import {
  AuthenticatorKind,
  AuthenticatorType,
} from "../../graphql/adminapi/globalTypes.generated";
import {
  ErrorParseRule,
  makeInvariantViolatedErrorParseRule,
  parseAPIErrors,
  parseRawError,
} from "../../error/parse";
import ErrorRenderer from "../../ErrorRenderer";
import styles from "./Add2FAPhoneDialog.module.css";

const errorRules: ErrorParseRule[] = [
  makeInvariantViolatedErrorParseRule(
    "DuplicatedAuthenticator",
    "Add2FAScreen.error.duplicated-email"
  ),
];

export interface Add2FAEmailDialogProps {
  open: boolean;
  userID: string;
  onOpenChange: (open: boolean) => void;
  onCreated?: () => unknown;
}

export function Add2FAEmailDialog({
  open,
  userID,
  onOpenChange,
  onCreated,
}: Add2FAEmailDialogProps): React.ReactElement {
  const [email, setEmail] = useState("");
  const { createAuthenticator, loading, error } =
    useCreateAuthenticatorMutation(userID);

  useEffect(() => {
    if (!open) {
      setEmail("");
    }
  }, [open]);

  const formError = useMemo(() => {
    if (error == null) {
      return null;
    }
    const apiErrors = parseRawError(error);
    const { topErrors } = parseAPIErrors(apiErrors, [], errorRules);
    return topErrors.length > 0 ? <ErrorRenderer errors={topErrors} /> : null;
  }, [error]);

  const onSubmit = useCallback(
    async (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      const value = email.trim();
      if (value === "" || loading) {
        return;
      }

      try {
        const authenticator = await createAuthenticator({
          type: AuthenticatorType.OobOtpEmail,
          email: value,
          kind: AuthenticatorKind.Secondary,
        });
        if (authenticator != null) {
          await onCreated?.();
          onOpenChange(false);
        }
      } catch {
        // Error is rendered in the dialog.
      }
    },
    [createAuthenticator, email, loading, onCreated, onOpenChange]
  );

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="480px" size="3">
        <Dialog.Title>
          <FormattedMessage id="Add2FAScreen.title.email" />
        </Dialog.Title>
        <Dialog.Description size="2">
          <FormattedMessage id="Add2FAScreen.email.description" />
        </Dialog.Description>
        <form className={styles.form} onSubmit={onSubmit}>
          <TextField
            size="2"
            type="email"
            label={<FormattedMessage id="Add2FAScreen.email.label" />}
            value={email}
            error={formError}
            onChange={(event) => {
              setEmail(event.currentTarget.value);
            }}
          />
          <div className={styles.actions}>
            <SecondaryButton
              size="2"
              disabled={loading}
              text={<FormattedMessage id="cancel" />}
              onClick={() => {
                onOpenChange(false);
              }}
            />
            <Button
              type="submit"
              size="2"
              loading={loading}
              disabled={email.trim() === ""}
            >
              <FormattedMessage id="add" />
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
