import React, { useCallback, useEffect, useMemo, useState } from "react";
import { FormattedMessage } from "../../intl";
import { SmsProviderConfigurationInput } from "../../graphql/portal/globalTypes.generated";
import { Dialog, DialogFooter, IDialogProps, Text } from "@fluentui/react";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
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

const topErrorRules: ErrorParseRule[] = [
  makeReasonErrorParseRule(
    "SMSGatewayAuthenticationFailed",
    "TestSMSDialog.errors.gateway-authentication-failed-error",
    (err) => ({
      code: (err as APISMSGatewayError).info.ProviderErrorCode,
    })
  ),
  makeReasonErrorParseRule(
    "SMSGatewayDeliveryRejected",
    "TestSMSDialog.errors.gateway-delivery-rejected-error",
    (err) => ({
      code: (err as APISMSGatewayError).info.ProviderErrorCode,
    })
  ),
  makeReasonErrorParseRule(
    "SMSGatewayRateLimited",
    "TestSMSDialog.errors.gateway-rate-limited-error",
    (err) => ({
      code: (err as APISMSGatewayError).info.ProviderErrorCode,
    })
  ),
  (apiError: APIError): ErrorParseRuleResult => {
    return {
      parsedAPIErrors: [
        {
          messageID: "TestSMSDialog.errors.unknown-error",
          arguments: {
            code:
              (apiError as Partial<APISMSGatewayError> | null)?.info
                ?.ProviderErrorCode ?? "-",
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
      code: (err as APISMSGatewayError).info.ProviderErrorCode,
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

  return (
    <FormProvider
      loading={sendTestSMSLoading}
      error={sendTestSMSError}
      rules={topErrorRules}
    >
      <Dialog
        hidden={isHidden}
        dialogContentProps={useMemo<IDialogProps["dialogContentProps"]>(() => {
          return {
            title: <FormattedMessage id="TestSMSDialog.title" />,
          };
        }, [])}
        onDismiss={onDismiss}
      >
        <div>
          <Text className="mb-3" block={true}>
            <FormattedMessage id="TestSMSDialog.description" />
          </Text>
          <FormPhoneTextField
            parentJSONPointer=""
            fieldName="to"
            allowlist={effectiveAppConfig?.ui?.phone_input?.allowlist}
            pinnedList={effectiveAppConfig?.ui?.phone_input?.pinned_list}
            initialInputValue={toInputValue}
            onChange={onChangeValues}
            errorRules={phoneFieldErrorRules}
          />
        </div>
        <DialogFooter>
          <PrimaryButton
            onClick={onSend}
            disabled={!to || sendTestSMSLoading}
            text={<FormattedMessage id="TestSMSDialog.send" />}
          />
          <DefaultButton
            onClick={onDismiss}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
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
