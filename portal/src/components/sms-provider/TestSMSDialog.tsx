import React, { useCallback, useEffect, useState } from "react";
import { FormattedMessage } from "../../intl";
import { SmsProviderConfigurationInput } from "../../graphql/portal/globalTypes.generated";
import { Button, Dialog } from "@radix-ui/themes";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import FormPhoneTextField from "../../FormPhoneTextField";
import { PortalAPIAppConfig } from "../../types";
import { useSendTestSMSMutation } from "../../graphql/portal/mutations/sendTestSMS";
import { useCalloutToast } from "../v2/Callout/Callout";
import { FormProvider, useFormTopErrors } from "../../form";
import {
  ErrorParseRule,
  ErrorParseRuleResult,
  makeReasonErrorParseRule,
} from "../../error/parse";
import { APIError, APISMSGatewayError } from "../../error/error";
import cn from "classnames";
import styles from "../users/Add2FAPhoneDialog.module.css";
import phoneDialogStyles from "../../PhoneDialog.module.css";

const topErrorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "SMSGatewayAuthenticationFailed",
    "TestSMSDialog.errors.gateway-authentication-failed-error",
    (err) => ({
      code: (err as APISMSGatewayError).info.ProviderErrorCode || "__empty__",
      description: (err as APISMSGatewayError).info.Description || "__empty__",
    })
  ),
  makeReasonErrorParseRule(
    "SMSGatewayDeliveryRejected",
    "TestSMSDialog.errors.gateway-delivery-rejected-error",
    (err) => ({
      code: (err as APISMSGatewayError).info.ProviderErrorCode || "__empty__",
      description: (err as APISMSGatewayError).info.Description || "__empty__",
    })
  ),
  makeReasonErrorParseRule(
    "SMSGatewayRateLimited",
    "TestSMSDialog.errors.gateway-rate-limited-error",
    (err) => ({
      code: (err as APISMSGatewayError).info.ProviderErrorCode || "__empty__",
      description: (err as APISMSGatewayError).info.Description || "__empty__",
    })
  ),
  (apiError: APIError): ErrorParseRuleResult => {
    const info = (apiError as Partial<APISMSGatewayError> | null)?.info;
    return {
      parsedAPIErrors: [
        {
          messageID: "TestSMSDialog.errors.unknown-error",
          arguments: {
            code: info?.ProviderErrorCode || "__empty__",
            description: info?.Description || "__empty__",
          },
        },
      ],
      fullyHandled: true,
    };
  },
];

const phoneFieldErrorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "SMSGatewayInvalidPhoneNumber",
    "TestSMSDialog.errors.gateway-invalid-phone-number-error",
    (err) => ({
      code: (err as APISMSGatewayError).info.ProviderErrorCode || "__empty__",
      description: (err as APISMSGatewayError).info.Description || "__empty__",
    })
  ),
];

export interface TestSMSDialogProps {
  appID: string;
  isHidden: boolean;
  input: SmsProviderConfigurationInput;
  effectiveAppConfig: PortalAPIAppConfig | undefined;
  onDismiss: () => void;
}

export function TestSMSDialog({
  appID,
  isHidden,
  input,
  effectiveAppConfig,
  onDismiss,
}: TestSMSDialogProps): React.ReactElement {
  const onOpenChange = useCallback(
    (open: boolean) => {
      if (!open) {
        onDismiss();
      }
    },
    [onDismiss]
  );
  const [toInputValue, setToInputValue] = useState("");
  const [to, setTo] = useState("");
  const onChangeValues = useCallback(
    (values: { e164?: string; rawInputValue: string }) => {
      const { e164, rawInputValue } = values;
      setTo(e164 ?? "");
      setToInputValue(rawInputValue);
    },
    []
  );

  const { showToast } = useCalloutToast();

  const {
    sendTestSMS,
    loading: sendTestSMSLoading,
    error: sendTestSMSError,
  } = useSendTestSMSMutation(appID);

  const onSend = useCallback(() => {
    sendTestSMS({
      to,
      config: input,
    })
      .then(() => {
        showToast({
          type: "success",
          text: <FormattedMessage id="TestSMSDialog.toast.success" />,
        });
        onDismiss();
      })
      // The error is handled by toast
      .catch(console.warn);
  }, [input, onDismiss, sendTestSMS, showToast, to]);

  const onSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      if (!to || sendTestSMSLoading) {
        return;
      }
      onSend();
    },
    [onSend, sendTestSMSLoading, to]
  );

  return (
    <FormProvider
      loading={sendTestSMSLoading}
      error={sendTestSMSError}
      rules={topErrorRules}
    >
      <Dialog.Root open={!isHidden} onOpenChange={onOpenChange}>
        <Dialog.Content
          maxWidth="480px"
          size="3"
          className={phoneDialogStyles.phoneDialogContent}
          data-phone-dialog="true"
        >
          <Dialog.Title>
            <FormattedMessage id="TestSMSDialog.title" />
          </Dialog.Title>
          <Dialog.Description size="2">
            <FormattedMessage id="TestSMSDialog.description" />
          </Dialog.Description>
          <form
            className={cn(styles.form, phoneDialogStyles.phoneDialogForm)}
            onSubmit={onSubmit}
          >
            <FormPhoneTextField
              parentJSONPointer=""
              fieldName="to"
              allowlist={effectiveAppConfig?.ui?.phone_input?.allowlist}
              pinnedList={effectiveAppConfig?.ui?.phone_input?.pinned_list}
              initialInputValue={toInputValue}
              onChange={onChangeValues}
              errorRules={phoneFieldErrorRules}
            />
            <div
              className={cn(
                styles.actions,
                phoneDialogStyles.phoneDialogActions
              )}
            >
              <SecondaryButton
                size="2"
                disabled={sendTestSMSLoading}
                text={<FormattedMessage id="cancel" />}
                onClick={onDismiss}
              />
              <Button
                type="submit"
                size="2"
                loading={sendTestSMSLoading}
                disabled={!to || sendTestSMSLoading}
              >
                <FormattedMessage id="TestSMSDialog.send" />
              </Button>
            </div>
          </form>
        </Dialog.Content>
      </Dialog.Root>
      <ErrorToast onDismiss={onDismiss} />
    </FormProvider>
  );
}

function ErrorToast({ onDismiss }: { onDismiss: () => void }) {
  const errors = useFormTopErrors();

  const { showToast } = useCalloutToast();

  useEffect(() => {
    for (const err of errors) {
      showToast({
        type: "error",
        text: (
          <FormattedMessage id={err.messageID ?? ""} values={err.arguments} />
        ),
      });
      // Close the dialog to let error toast shows outside
      onDismiss();
    }
  }, [errors, onDismiss, showToast]);

  return <></>;
}
