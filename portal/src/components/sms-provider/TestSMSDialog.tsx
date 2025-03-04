import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { SmsProviderConfigurationInput } from "../../graphql/portal/globalTypes.generated";
import { Dialog, DialogFooter, IDialogProps, Text } from "@fluentui/react";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import FormPhoneTextField from "../../FormPhoneTextField";
import { PortalAPIAppConfig } from "../../types";
import { useSendTestSMSMutation } from "../../graphql/portal/mutations/sendTestSMS";
import { CalloutColor, useCalloutToast } from "../v2/common/Callout";

export interface TestSMSDialogProps {
  appID: string;
  isHidden: boolean;
  input: SmsProviderConfigurationInput;
  effectiveAppConfig: PortalAPIAppConfig | undefined;
  onCancel: () => void;
}

export function TestSMSDialog({
  appID,
  isHidden,
  input,
  effectiveAppConfig,
  onCancel,
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

  // eslint-disable-next-line no-useless-assignment
  const { Component: ToastComponent, showToast } = useCalloutToast();

  const { sendTestSMS, loading: sendTestSMSLoading } =
    useSendTestSMSMutation(appID);

  const onSend = useCallback(() => {
    sendTestSMS({
      to,
      config: input,
    })
      .then(() => {
        showToast({ color: CalloutColor.success, text: "success" });
      })
      .catch(() => {
        showToast({ color: CalloutColor.error, text: "error" });
      });
  }, [input, sendTestSMS, showToast, to]);

  return (
    <Dialog
      hidden={isHidden}
      dialogContentProps={useMemo<IDialogProps["dialogContentProps"]>(() => {
        return {
          title: <FormattedMessage id="TestSMSDialog.title" />,
        };
      }, [])}
      onDismiss={onCancel}
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
          inputValue={toInputValue}
          onChange={onChangeValues}
        />
      </div>
      <DialogFooter>
        <PrimaryButton
          onClick={onSend}
          disabled={!to || sendTestSMSLoading}
          text={<FormattedMessage id="TestSMSDialog.send" />}
        />
        <DefaultButton
          onClick={onCancel}
          text={<FormattedMessage id="cancel" />}
        />
      </DialogFooter>
      <ToastComponent />
    </Dialog>
  );
}
