import React, { useCallback, useMemo, useState } from "react";
import { Button, Dialog, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import PasswordField from "../../PasswordField";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { useCreateAuthenticatorMutation } from "../../graphql/adminapi/mutations/createAuthenticatorMutation";
import {
  AuthenticatorKind,
  AuthenticatorType,
} from "../../graphql/adminapi/globalTypes.generated";
import { PasswordPolicyConfig } from "../../types";
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
    "Add2FAScreen.error.duplicated-password"
  ),
];

export interface Add2FAPasswordDialogProps {
  open: boolean;
  userID: string;
  passwordPolicy: PasswordPolicyConfig;
  onOpenChange: (open: boolean) => void;
  onCreated?: () => unknown;
}

export function Add2FAPasswordDialog({
  open,
  userID,
  passwordPolicy,
  onOpenChange,
  onCreated,
}: Add2FAPasswordDialogProps): React.ReactElement {
  const [password, setPassword] = useState("");
  const { createAuthenticator, loading, error } =
    useCreateAuthenticatorMutation(userID);

  const [prevOpen, setPrevOpen] = useState(open);
  if (prevOpen !== open) {
    setPrevOpen(open);
    if (!open) {
      setPassword("");
    }
  }

  const formError = useMemo(() => {
    if (error == null) {
      return null;
    }
    const apiErrors = parseRawError(error);
    const { topErrors } = parseAPIErrors(apiErrors, [], errorRules);
    return topErrors.length > 0 ? <ErrorRenderer errors={topErrors} /> : null;
  }, [error]);

  const onSubmit = useCallback(
    (event: React.FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      if (password === "" || loading) {
        return;
      }

      const submit = async () => {
        try {
          const authenticator = await createAuthenticator({
            type: AuthenticatorType.Password,
            password,
            kind: AuthenticatorKind.Secondary,
          });
          if (authenticator != null) {
            await onCreated?.();
            onOpenChange(false);
          }
        } catch {
          // Error is rendered in the dialog.
        }
      };
      void submit();
    },
    [createAuthenticator, password, loading, onCreated, onOpenChange]
  );

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content maxWidth="480px" size="3">
        <Dialog.Title>
          <FormattedMessage id="Add2FAScreen.title.password" />
        </Dialog.Title>
        <Dialog.Description size="2">
          <FormattedMessage id="Add2FAScreen.password.description" />
        </Dialog.Description>
        <form className={styles.form} onSubmit={onSubmit}>
          <div>
            <PasswordField
              label={<FormattedMessage id="Add2FAScreen.password.label" />}
              passwordPolicy={passwordPolicy}
              canGeneratePassword={true}
              canRevealPassword={true}
              value={password}
              onChange={setPassword}
              parentJSONPointer=""
              fieldName="password"
            />
            {formError != null ? (
              <Text as="p" size="1" color="red" mt="1">
                {formError}
              </Text>
            ) : null}
          </div>
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
              disabled={password === ""}
            >
              <FormattedMessage id="add" />
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
