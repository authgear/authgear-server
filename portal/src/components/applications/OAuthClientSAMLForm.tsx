import React, { useCallback, useContext, useMemo } from "react";
import Toggle from "../../Toggle";
import {
  FormattedMessage,
  Context as MessageFormatContext,
  ContextValue as MessageFormatContextValue,
} from "@oursky/react-messageformat";
import HorizontalDivider from "../../HorizontalDivider";
import WidgetTitle from "../../WidgetTitle";
import ScreenTitle from "../../ScreenTitle";
import { IChoiceGroupOption, ChoiceGroup, Label } from "@fluentui/react";
import {
  SAMLNameIDFormat,
  SAMLNameIDAttributePointer,
  SAMLBinding,
} from "../../types";
import FormTextFieldList from "../../FormTextFieldList";
import FormTextField from "../../FormTextField";

export interface OAuthClientSAMLFormState {
  isSAMLEnabled: boolean;
  // Basic
  nameIDFormat: SAMLNameIDFormat;
  nameIDAttributePointer: SAMLNameIDAttributePointer | undefined;
  // SSO
  acsURLs: string[] | undefined;
  destination: string | undefined;
  recipient: string | undefined;
  audience: string | undefined;
  assertionValidDurationSeconds: number | undefined;
  // Logout
  isSLOEnabled: boolean | undefined;
  sloCallbackURL: string | undefined;
  sloCallbackBinding: SAMLBinding | undefined;
  // Signature
  signatureVerificationEnabled: boolean | undefined;
  signingCertificates: string[] | undefined;
}

export interface OAuthClientSAMLFormProps {
  formState: OAuthClientSAMLFormState;
  onFormStateChange: (newState: OAuthClientSAMLFormState) => void;
}

const nameIDFormatOptions: IChoiceGroupOption[] = [
  { key: SAMLNameIDFormat.Unspecified, text: SAMLNameIDFormat.Unspecified },
  { key: SAMLNameIDFormat.EmailAddress, text: SAMLNameIDFormat.EmailAddress },
];

function makeNameIDAttributePointerOptions(
  renderToString: MessageFormatContextValue["renderToString"]
): IChoiceGroupOption[] {
  return [
    {
      key: SAMLNameIDAttributePointer.Sub,
      text: renderToString(
        "OAuthClientSAMLForm.nameIDAttribute.options.userID"
      ),
    },
    {
      key: SAMLNameIDAttributePointer.Email,
      text: renderToString("OAuthClientSAMLForm.nameIDAttribute.options.email"),
    },
    {
      key: SAMLNameIDAttributePointer.PhoneNumber,
      text: renderToString("OAuthClientSAMLForm.nameIDAttribute.options.phone"),
    },
    {
      key: SAMLNameIDAttributePointer.PreferredUsername,
      text: renderToString(
        "OAuthClientSAMLForm.nameIDAttribute.options.username"
      ),
    },
  ];
}

function makeSLOCallbackBindingOptions(
  renderToString: MessageFormatContextValue["renderToString"]
): IChoiceGroupOption[] {
  return [
    {
      key: SAMLBinding.HTTPRedirect,
      text: renderToString(
        "OAuthClientSAMLForm.logout.callbackBinding.options.httpRedirect"
      ),
    },
    {
      key: SAMLBinding.HTTPPOST,
      text: renderToString(
        "OAuthClientSAMLForm.logout.callbackBinding.options.httpPost"
      ),
    },
  ];
}

export function OAuthClientSAMLForm({
  formState,
  onFormStateChange,
}: OAuthClientSAMLFormProps): React.ReactElement {
  const { renderToString } = useContext(MessageFormatContext);

  const onIsSAMLEnabledChange = useCallback(
    (_, checked?: boolean) => {
      onFormStateChange({ ...formState, isSAMLEnabled: Boolean(checked) });
    },
    [formState, onFormStateChange]
  );

  const onNameIDFormatChange = useCallback(
    (_, option?: IChoiceGroupOption) => {
      if (option == null) {
        return;
      }
      onFormStateChange({
        ...formState,
        nameIDFormat: option.key as SAMLNameIDFormat,
      });
    },
    [formState, onFormStateChange]
  );

  const onNameIDAttributePointerChange = useCallback(
    (_, option?: IChoiceGroupOption) => {
      if (option == null) {
        return;
      }
      onFormStateChange({
        ...formState,
        nameIDAttributePointer: option.key as SAMLNameIDAttributePointer,
      });
    },
    [formState, onFormStateChange]
  );

  const onAcsUrlsChange = useCallback(
    (newList: string[]) => {
      onFormStateChange({
        ...formState,
        acsURLs: newList,
      });
    },
    [formState, onFormStateChange]
  );

  const onTextfieldChange = useMemo(() => {
    const makeOnChangeCallback = (key: keyof OAuthClientSAMLFormState) => {
      return (_: unknown, newValue?: string) => {
        onFormStateChange({
          ...formState,
          [key]: newValue,
        });
      };
    };
    return {
      destination: makeOnChangeCallback("destination"),
      recipient: makeOnChangeCallback("recipient"),
      audience: makeOnChangeCallback("audience"),
      sloCallbackURL: makeOnChangeCallback("sloCallbackURL"),
    };
  }, [formState, onFormStateChange]);

  const onAssertionValidDurationSecondsChange = useCallback(
    (_: unknown, newValue?: string) => {
      if (newValue == null) {
        return;
      }
      if (newValue.trim() === "") {
        onFormStateChange({
          ...formState,
          assertionValidDurationSeconds: undefined,
        });
        return;
      }
      const newValueInt = parseInt(newValue, 10);
      if (isNaN(newValueInt)) {
        return;
      }
      onFormStateChange({
        ...formState,
        assertionValidDurationSeconds: newValueInt,
      });
    },
    [formState, onFormStateChange]
  );

  const onIsSLOEnabledChange = useCallback(
    (_, checked?: boolean) => {
      onFormStateChange({ ...formState, isSLOEnabled: Boolean(checked) });
    },
    [formState, onFormStateChange]
  );

  const onSLOCallbackBindingChange = useCallback(
    (_, option?: IChoiceGroupOption) => {
      if (option == null) {
        return;
      }
      onFormStateChange({
        ...formState,
        sloCallbackBinding: option.key as SAMLBinding,
      });
    },
    [formState, onFormStateChange]
  );

  const onSignatureVerificationEnabledChange = useCallback(
    (_, checked?: boolean) => {
      onFormStateChange({
        ...formState,
        signatureVerificationEnabled: Boolean(checked),
      });
    },
    [formState, onFormStateChange]
  );

  const onSigningCertificatesChange = useCallback(
    (newList: string[]) => {
      onFormStateChange({
        ...formState,
        signingCertificates: newList,
      });
    },
    [formState, onFormStateChange]
  );

  const nameIDAttributePointerOptions = useMemo(
    () => makeNameIDAttributePointerOptions(renderToString),
    [renderToString]
  );

  const sloBindingOptions = useMemo(
    () => makeSLOCallbackBindingOptions(renderToString),
    [renderToString]
  );

  return (
    <div>
      <Toggle
        label={renderToString("OAuthClientSAMLForm.enable.label")}
        description={renderToString("OAuthClientSAMLForm.enable.description")}
        checked={formState.isSAMLEnabled}
        onChange={onIsSAMLEnabledChange}
      />
      {formState.isSAMLEnabled ? (
        <>
          <HorizontalDivider className="my-12" />
          <div className="grid gap-y-12 grid-cols-1">
            <ScreenTitle>
              <FormattedMessage id="OAuthClientSAMLForm.screen.title" />
            </ScreenTitle>
            <div>
              <WidgetTitle className="mb-3" id="basic">
                <FormattedMessage id="OAuthClientSAMLForm.basic.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <ChoiceGroup
                  label={renderToString(
                    "OAuthClientSAMLForm.nameIDFormat.label"
                  )}
                  options={nameIDFormatOptions}
                  selectedKey={formState.nameIDFormat}
                  onChange={onNameIDFormatChange}
                />
                <ChoiceGroup
                  label={renderToString(
                    "OAuthClientSAMLForm.nameIDAttribute.label"
                  )}
                  disabled={
                    formState.nameIDFormat !== SAMLNameIDFormat.Unspecified
                  }
                  options={nameIDAttributePointerOptions}
                  selectedKey={
                    formState.nameIDFormat !== SAMLNameIDFormat.Unspecified
                      ? null
                      : formState.nameIDAttributePointer
                  }
                  onChange={onNameIDAttributePointerChange}
                />
              </div>
            </div>

            <div>
              <WidgetTitle className="mb-3" id="basic">
                <FormattedMessage id="OAuthClientSAMLForm.sso.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <FormTextFieldList
                  parentJSONPointer=""
                  fieldName="acs_urls"
                  list={formState.acsURLs ?? []}
                  onListItemAdd={onAcsUrlsChange}
                  onListItemChange={onAcsUrlsChange}
                  onListItemDelete={onAcsUrlsChange}
                  addButtonLabelMessageID="OAuthClientSAMLForm.sso.acsUrls.add"
                  label={
                    <Label>
                      <FormattedMessage id="OAuthClientSAMLForm.sso.acsUrls.title" />
                    </Label>
                  }
                  minItem={1}
                />
                <FormTextField
                  parentJSONPointer=""
                  fieldName="destination"
                  label={renderToString(
                    "OAuthClientSAMLForm.sso.destination.label"
                  )}
                  description={renderToString(
                    "OAuthClientSAMLForm.sso.destination.description"
                  )}
                  value={formState.destination}
                  onChange={onTextfieldChange.destination}
                />
                <FormTextField
                  parentJSONPointer=""
                  fieldName="recipient"
                  label={renderToString(
                    "OAuthClientSAMLForm.sso.recipient.label"
                  )}
                  description={renderToString(
                    "OAuthClientSAMLForm.sso.recipient.description"
                  )}
                  value={formState.recipient}
                  onChange={onTextfieldChange.recipient}
                />
                <FormTextField
                  parentJSONPointer=""
                  fieldName="audience"
                  label={renderToString(
                    "OAuthClientSAMLForm.sso.audience.label"
                  )}
                  description={renderToString(
                    "OAuthClientSAMLForm.sso.audience.description"
                  )}
                  value={formState.audience}
                  onChange={onTextfieldChange.audience}
                />
                <FormTextField
                  parentJSONPointer=""
                  fieldName="assertion_valid_duration"
                  label={renderToString(
                    "OAuthClientSAMLForm.sso.assertionValidDuration.label"
                  )}
                  value={
                    formState.assertionValidDurationSeconds?.toFixed(0) ?? ""
                  }
                  onChange={onAssertionValidDurationSecondsChange}
                />
              </div>
            </div>

            <div>
              <WidgetTitle className="mb-3" id="basic">
                <FormattedMessage id="OAuthClientSAMLForm.logout.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <Toggle
                  label={renderToString(
                    "OAuthClientSAMLForm.logout.enable.label"
                  )}
                  checked={formState.isSLOEnabled}
                  onChange={onIsSLOEnabledChange}
                />
                <FormTextField
                  parentJSONPointer=""
                  fieldName="slo_callback_url"
                  label={renderToString(
                    "OAuthClientSAMLForm.logout.callbackURL.label"
                  )}
                  description={renderToString(
                    "OAuthClientSAMLForm.logout.callbackURL.description"
                  )}
                  value={formState.isSLOEnabled ? formState.sloCallbackURL : ""}
                  onChange={onTextfieldChange.sloCallbackURL}
                  disabled={!formState.isSLOEnabled}
                />
                <ChoiceGroup
                  label={renderToString(
                    "OAuthClientSAMLForm.logout.callbackBinding.label"
                  )}
                  disabled={!formState.isSLOEnabled}
                  options={sloBindingOptions}
                  selectedKey={
                    formState.isSLOEnabled ? formState.sloCallbackBinding : null
                  }
                  onChange={onSLOCallbackBindingChange}
                />
              </div>
            </div>

            <div>
              <WidgetTitle className="mb-3" id="basic">
                <FormattedMessage id="OAuthClientSAMLForm.signature.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <Toggle
                  label={renderToString(
                    "OAuthClientSAMLForm.signature.checkSignature.label"
                  )}
                  description={renderToString(
                    "OAuthClientSAMLForm.signature.checkSignature.description"
                  )}
                  checked={formState.signatureVerificationEnabled}
                  onChange={onSignatureVerificationEnabledChange}
                />
                <FormTextFieldList
                  parentJSONPointer=""
                  fieldName="signingCertificates"
                  list={formState.signingCertificates ?? []}
                  onListItemAdd={onSigningCertificatesChange}
                  onListItemChange={onSigningCertificatesChange}
                  onListItemDelete={onSigningCertificatesChange}
                  addButtonLabelMessageID="OAuthClientSAMLForm.signature.certificates.add"
                  label={
                    <Label>
                      <FormattedMessage id="OAuthClientSAMLForm.signature.certificates.label" />
                    </Label>
                  }
                  multiline={true}
                />
              </div>
            </div>
          </div>
        </>
      ) : null}
    </div>
  );
}
