import React, { useCallback, useContext, useEffect, useMemo, useState } from "react";
import cn from "classnames";
import { Button, Dialog } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import PhoneTextField from "../../PhoneTextField";
import phoneDialogStyles from "../../PhoneDialog.module.css";
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
    "Add2FAScreen.error.duplicated-phone-number"
  ),
];

export interface Add2FAPhoneDialogProps {
  open: boolean;
  userID: string;
  phoneInputAllowlist?: string[];
  phoneInputPinnedList?: string[];
  onOpenChange: (open: boolean) => void;
  onCreated?: () => unknown;
}

export function Add2FAPhoneDialog({
  open,
  userID,
  phoneInputAllowlist,
  phoneInputPinnedList,
  onOpenChange,
  onCreated,
}: Add2FAPhoneDialogProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const [e164, setE164] = useState("");
  const [rawInputValue, setRawInputValue] = useState("");
  const [fieldKey, setFieldKey] = useState(0);
  const { createAuthenticator, loading, error } =
    useCreateAuthenticatorMutation(userID);

  useEffect(() => {
    if (!open) {
      setE164("");
      setRawInputValue("");
    } else {
      setFieldKey((key) => key + 1);
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
      if (e164 === "" || loading) {
        return;
      }

      try {
        const authenticator = await createAuthenticator({
          type: AuthenticatorType.OobOtpSms,
          phone: e164,
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
    [createAuthenticator, e164, loading, onCreated, onOpenChange]
  );

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Content
        maxWidth="480px"
        size="3"
        className={phoneDialogStyles.phoneDialogContent}
        data-phone-dialog="true"
      >
        <Dialog.Title>
          <FormattedMessage id="Add2FAScreen.title.phone" />
        </Dialog.Title>
        <Dialog.Description size="2">
          <FormattedMessage id="Add2FAScreen.phone.description" />
        </Dialog.Description>
        <form
          className={cn(styles.form, phoneDialogStyles.phoneDialogForm)}
          onSubmit={onSubmit}
        >
          <PhoneTextField
            key={fieldKey}
            label={renderToString("Add2FAScreen.phone.label")}
            allowlist={phoneInputAllowlist}
            pinnedList={phoneInputPinnedList}
            initialInputValue={rawInputValue}
            errorMessage={formError}
            onChange={(values) => {
              setE164(values.e164 ?? "");
              setRawInputValue(values.rawInputValue);
            }}
          />
          <div className={cn(styles.actions, phoneDialogStyles.phoneDialogActions)}>
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
              disabled={e164 === ""}
            >
              <FormattedMessage id="add" />
            </Button>
          </div>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
