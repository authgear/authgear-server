import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import Toggle from "../../Toggle";
import {
  FormattedMessage,
  Context as MessageFormatContext,
  IntlContextValue as MessageFormatContextValue,
} from "../../intl";
import HorizontalDivider from "../../HorizontalDivider";
import WidgetTitle from "../../WidgetTitle";
import ScreenTitle from "../../ScreenTitle";
import {
  IChoiceGroupOption,
  ChoiceGroup,
  Label,
  MessageBar,
  MessageBarType,
  Text,
  FontIcon,
} from "@fluentui/react";
import {
  SAMLNameIDFormat,
  SAMLNameIDAttributePointer,
  SAMLBinding,
  PortalAPIAppConfig,
  SAMLIdpSigningCertificate,
} from "../../types";
import FormTextFieldList from "../../FormTextFieldList";
import FormTextField from "../../FormTextField";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import DefaultButton from "../../DefaultButton";
import { downloadStringAsFile } from "../../util/download";
import { useParams } from "react-router-dom";
import { AutoGenerateFirstCertificate } from "../saml/AutoGenerateFirstCertificate";
import {
  formatCertificateFilename,
  parseServiceProviderMetadata,
} from "../../model/saml";
import OutlinedActionButton from "../common/OutlinedActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";

export interface OAuthClientSAMLFormState {
  isSAMLEnabled: boolean;
  // Basic
  nameIDFormat: SAMLNameIDFormat;
  nameIDAttributePointer: SAMLNameIDAttributePointer;
  // SSO
  acsURLs: string[];
  destination: string;
  recipient: string;
  audience: string;
  assertionValidDurationSeconds: number;
  // Logout
  isSLOEnabled: boolean;
  sloCallbackURL: string;
  sloCallbackBinding: SAMLBinding;
  // Signature
  signatureVerificationEnabled: boolean;
  signingCertificates: string[];

  isMetadataUploaded: boolean;
}

export function getDefaultOAuthClientSAMLFormState(): OAuthClientSAMLFormState {
  return {
    isSAMLEnabled: false,
    nameIDFormat: SAMLNameIDFormat.Unspecified,
    nameIDAttributePointer: SAMLNameIDAttributePointer.Sub,
    acsURLs: [],
    destination: "",
    recipient: "",
    audience: "",
    assertionValidDurationSeconds: 1200,
    isSLOEnabled: false,
    sloCallbackURL: "",
    sloCallbackBinding: SAMLBinding.HTTPRedirect,
    signatureVerificationEnabled: false,
    signingCertificates: [],
    isMetadataUploaded: false,
  };
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

function IdpCertificateSection({
  appID,
  configAppID,
  samlIdpSigningCertificate,
}: {
  appID: string;
  configAppID: string;
  samlIdpSigningCertificate: SAMLIdpSigningCertificate;
}) {
  const { themes } = useSystemConfig();

  const onDownloadIdpCertificate = useCallback(() => {
    downloadStringAsFile({
      content: samlIdpSigningCertificate.certificatePEM,
      filename: formatCertificateFilename(
        configAppID,
        samlIdpSigningCertificate.certificateFingerprint
      ),
      mimeType: "application/x-pem-file",
    });
  }, [samlIdpSigningCertificate, configAppID]);

  return (
    <div>
      <WidgetTitle className="mb-3" id="identity-provider-certificates">
        <FormattedMessage id="OAuthClientSAMLForm.idpCertificate.title" />
      </WidgetTitle>
      <div className="grid gap-y-4 grid-cols-1">
        <div>
          <OutlinedActionButton
            theme={themes.actionButton}
            className="justify-self-start"
            iconProps={{ iconName: "Download" }}
            onClick={onDownloadIdpCertificate}
            text={
              <FormattedMessage id="OAuthClientSAMLForm.idpCertificate.download" />
            }
          />
          <Text block={true} className={"mt-1"}>
            <FormattedMessage
              id="OAuthClientSAMLForm.idpCertificate.fingerprint"
              values={{
                fingerprint: samlIdpSigningCertificate.certificateFingerprint,
              }}
            />
          </Text>
        </div>

        <MessageBar messageBarType={MessageBarType.info}>
          <FormattedMessage
            id="OAuthClientSAMLForm.idpCertificate.rotateHint"
            values={{
              href: `/project/${appID}/advanced/saml-certificate`,
            }}
          />
        </MessageBar>
      </div>
    </div>
  );
}

export interface OAuthClientSAMLFormProps {
  parentJSONPointer: string | RegExp;
  clientID: string;
  rawAppConfig: PortalAPIAppConfig;
  publicOrigin: string;
  samlIdpEntityID: string;
  samlIdpSigningCertificates: SAMLIdpSigningCertificate[];
  formState: OAuthClientSAMLFormState;
  onFormStateChange: (newState: OAuthClientSAMLFormState) => void;
  onGeneratedNewIdpSigningCertificate: () => void;
}

export function OAuthClientSAMLForm({
  parentJSONPointer,
  clientID,
  rawAppConfig,
  publicOrigin,
  samlIdpEntityID,
  samlIdpSigningCertificates,
  formState,
  onFormStateChange,
  onGeneratedNewIdpSigningCertificate,
}: OAuthClientSAMLFormProps): React.ReactElement {
  const { renderToString } = useContext(MessageFormatContext);
  const { isDirty: isFormDirty } = useFormContainerBaseContext();
  const { appID } = useParams() as { appID: string };
  const { themes } = useSystemConfig();

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
          assertionValidDurationSeconds:
            getDefaultOAuthClientSAMLFormState().assertionValidDurationSeconds,
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

  const endpoints = useMemo(() => {
    return {
      metadata: `${publicOrigin}/saml2/metadata/${clientID}`,
      login: `${publicOrigin}/saml2/login/${clientID}`,
      logout: `${publicOrigin}/saml2/logout/${clientID}`,
    };
  }, [clientID, publicOrigin]);

  const onClickDownloadMetadata = useCallback(() => {
    const link = document.createElement("a");
    link.href = endpoints.metadata;
    link.target = "_blank";
    link.click();
  }, [endpoints.metadata]);

  const updateFormStateByMetadata = useCallback(
    (xmlData: string) => {
      const parseResult = parseServiceProviderMetadata(xmlData);
      const newState = {
        ...formState,
        isMetadataUploaded: true,
      };
      if (parseResult.acsURL != null) {
        newState.acsURLs = [parseResult.acsURL];
      }
      if (parseResult.sloEnabled != null) {
        newState.isSLOEnabled = parseResult.sloEnabled;
      }
      if (parseResult.sloCallbackURL != null) {
        newState.sloCallbackURL = parseResult.sloCallbackURL;
      }
      if (parseResult.sloCallbackBinding != null) {
        newState.sloCallbackBinding = parseResult.sloCallbackBinding;
      }
      if (parseResult.authnRequestsSigned != null) {
        newState.signatureVerificationEnabled = parseResult.authnRequestsSigned;
      }
      if (parseResult.certificate != null) {
        newState.signingCertificates = [parseResult.certificate];
      }
      onFormStateChange(newState);
    },
    [formState, onFormStateChange]
  );

  const onUploadMetadata = useCallback(() => {
    const fileInput = document.createElement("input");
    fileInput.type = "file";
    fileInput.accept = "application/xml,text/xml,.xml";
    const onChange = () => {
      fileInput.removeEventListener("change", onChange);
      if (fileInput.files && fileInput.files.length > 0) {
        const reader = new FileReader();
        reader.addEventListener("load", () => {
          if (typeof reader.result === "string") {
            updateFormStateByMetadata(reader.result);
          }
        });
        reader.readAsText(fileInput.files[0]);
      }
      fileInput.remove();
    };
    fileInput.addEventListener("change", onChange);
    fileInput.click();
  }, [updateFormStateByMetadata]);

  const nameIDAttributePointerOptions = useMemo(
    () => makeNameIDAttributePointerOptions(renderToString),
    [renderToString]
  );

  const sloBindingOptions = useMemo(
    () => makeSLOCallbackBindingOptions(renderToString),
    [renderToString]
  );

  const activeIdpCertificate = useMemo(() => {
    if (rawAppConfig.saml?.signing?.key_id == null) {
      return null;
    }
    return (
      samlIdpSigningCertificates.find(
        (cert) => cert.keyID === rawAppConfig.saml?.signing?.key_id
      ) ?? null
    );
  }, [rawAppConfig, samlIdpSigningCertificates]);

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
            <div>
              <ScreenTitle>
                <FormattedMessage id="OAuthClientSAMLForm.title" />
              </ScreenTitle>
              <div className="mt-3 grid gap-y-2 grid-cols-1 items-start justify-items-start">
                <Text block={true}>
                  <FormattedMessage id="OAuthClientSAMLForm.metadataUpload.description" />
                </Text>
                <div className="grid grid-flow-col items-center gap-x-2">
                  <DefaultButton
                    className="w-fit"
                    text={renderToString(
                      "OAuthClientSAMLForm.metadataUpload.label"
                    )}
                    onClick={onUploadMetadata}
                  />
                  {formState.isMetadataUploaded ? (
                    <div className="flex flex-row items-center">
                      <FontIcon
                        iconName="Accept"
                        className="text-text-disabled mr-1"
                      />
                      <Text className="text-text-disabled">
                        <FormattedMessage id="OAuthClientSAMLForm.metadataUpload.success" />
                      </Text>
                    </div>
                  ) : null}
                </div>
              </div>
            </div>
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
              <WidgetTitle className="mb-3" id="sso">
                <FormattedMessage id="OAuthClientSAMLForm.sso.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <FormTextFieldList
                  parentJSONPointer={parentJSONPointer}
                  fieldName="acs_urls"
                  list={formState.acsURLs}
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
                  parentJSONPointer={parentJSONPointer}
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
                  parentJSONPointer={parentJSONPointer}
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
                  parentJSONPointer={parentJSONPointer}
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
                  parentJSONPointer={parentJSONPointer}
                  fieldName="assertion_valid_duration"
                  label={renderToString(
                    "OAuthClientSAMLForm.sso.assertionValidDuration.label"
                  )}
                  value={formState.assertionValidDurationSeconds.toFixed(0)}
                  onChange={onAssertionValidDurationSecondsChange}
                />
              </div>
            </div>

            <div>
              <WidgetTitle className="mb-3" id="logout">
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
                  parentJSONPointer={parentJSONPointer}
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
              <WidgetTitle className="mb-3" id="signature">
                <FormattedMessage id="OAuthClientSAMLForm.signature.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <FormTextFieldList
                  parentJSONPointer={
                    /\/secrets\/(\d*)\/data\/(\d*)\/certificates\/(\d*)/
                  }
                  fieldName="pem"
                  list={formState.signingCertificates}
                  onListItemAdd={onSigningCertificatesChange}
                  onListItemChange={onSigningCertificatesChange}
                  onListItemDelete={onSigningCertificatesChange}
                  addButtonLabelMessageID="OAuthClientSAMLForm.signature.certificates.add"
                  label={
                    <Label>
                      <FormattedMessage id="OAuthClientSAMLForm.signature.certificates.label" />
                    </Label>
                  }
                  description={renderToString(
                    "OAuthClientSAMLForm.signature.certificates.description"
                  )}
                  multiline={true}
                  maxItem={2}
                  minItem={formState.signatureVerificationEnabled ? 1 : 0}
                />
                <div className="grid gap-y-2 grid-cols-1">
                  <Toggle
                    label={renderToString(
                      "OAuthClientSAMLForm.signature.checkSignature.label"
                    )}
                    description={renderToString(
                      "OAuthClientSAMLForm.signature.checkSignature.description"
                    )}
                    disabled={formState.signingCertificates.length < 1}
                    checked={formState.signatureVerificationEnabled}
                    onChange={onSignatureVerificationEnabledChange}
                  />
                  {formState.signingCertificates.length < 1 ? (
                    <MessageBar messageBarType={MessageBarType.warning}>
                      <FormattedMessage id="OAuthClientSAMLForm.signature.checkSignature.hint" />
                    </MessageBar>
                  ) : null}
                </div>
              </div>
            </div>

            <HorizontalDivider />

            <div>
              <WidgetTitle className="mb-6" id="configuration-parameters">
                <FormattedMessage id="OAuthClientSAMLForm.configurationParameters.title" />
              </WidgetTitle>

              <div className="grid gap-y-4 grid-cols-1">
                <div className="grid gap-y-2 grid-cols-1">
                  <TextFieldWithCopyButton
                    label={renderToString(
                      "OAuthClientSAMLForm.configurationParameters.metadata.label"
                    )}
                    value={endpoints.metadata}
                    readOnly={true}
                  />
                  <OutlinedActionButton
                    theme={themes.actionButton}
                    className="justify-self-start"
                    iconProps={{ iconName: "Download" }}
                    text={
                      <FormattedMessage id="OAuthClientSAMLForm.configurationParameters.metadata.download" />
                    }
                    onClick={onClickDownloadMetadata}
                    disabled={isFormDirty}
                  />
                  <MessageBar
                    className={cn(isFormDirty ? null : "hidden")}
                    messageBarType={MessageBarType.warning}
                  >
                    <FormattedMessage id="OAuthClientSAMLForm.configurationParameters.metadata.saveBeforeDownload.hint" />
                  </MessageBar>
                </div>
                <TextFieldWithCopyButton
                  label={renderToString(
                    "OAuthClientSAMLForm.configurationParameters.issuer.label"
                  )}
                  value={samlIdpEntityID}
                  readOnly={true}
                />
                <TextFieldWithCopyButton
                  label={renderToString(
                    "OAuthClientSAMLForm.configurationParameters.loginURL.label"
                  )}
                  value={endpoints.login}
                  readOnly={true}
                />
                <TextFieldWithCopyButton
                  label={renderToString(
                    "OAuthClientSAMLForm.configurationParameters.logoutURL.label"
                  )}
                  value={
                    formState.isSLOEnabled
                      ? endpoints.logout
                      : renderToString(
                          "OAuthClientSAMLForm.configurationParameters.logoutURL.not-available"
                        )
                  }
                  disabled={!formState.isSLOEnabled}
                  readOnly={true}
                />
              </div>
            </div>

            {activeIdpCertificate != null ? (
              <IdpCertificateSection
                appID={appID}
                configAppID={rawAppConfig.id}
                samlIdpSigningCertificate={activeIdpCertificate}
              />
            ) : (
              <AutoGenerateFirstCertificate
                appID={appID}
                rawAppConfig={rawAppConfig}
                certificates={samlIdpSigningCertificates}
                onComplete={onGeneratedNewIdpSigningCertificate}
              />
            )}
          </div>
        </>
      ) : null}
    </div>
  );
}
