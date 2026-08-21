import React, { useCallback, useContext, useMemo } from "react";
import {
  Flex,
  IconButton as RadixIconButton,
  RadioGroup,
  Separator,
  Text,
} from "@radix-ui/themes";
import { CheckIcon, DownloadIcon, TrashIcon } from "@radix-ui/react-icons";
import { useParams } from "react-router-dom";
import { Toggle } from "../v2/Toggle/Toggle";
import {
  FormattedMessage,
  Context as MessageFormatContext,
  IntlContextValue as MessageFormatContextValue,
} from "../../intl";
import WidgetTitle from "../../WidgetTitle";
import {
  SAMLNameIDFormat,
  SAMLNameIDAttributePointer,
  SAMLBinding,
  PortalAPIAppConfig,
  SAMLIdpSigningCertificate,
} from "../../types";
import { TextField } from "../v2/TextField/TextField";
import { TextFieldList } from "../v2/TextFieldList/TextFieldList";
import { TextArea } from "../v2/TextArea/TextArea";
import { FormField } from "../v2/FormField/FormField";
import { Callout } from "../v2/Callout/Callout";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { useFormField } from "../../form";
import { joinParentChild } from "../../util/jsonpointer";
import ErrorRenderer from "../../ErrorRenderer";
import { downloadStringAsFile } from "../../util/download";
import Link from "../../Link";
import { AutoGenerateFirstCertificate } from "../saml/AutoGenerateFirstCertificate";
import {
  formatCertificateFilename,
  parseServiceProviderMetadata,
} from "../../model/saml";
import styles from "./OAuthClientSAMLForm.module.css";

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

interface RadioGroupOption {
  value: string;
  text: string;
}

const nameIDFormatOptions: RadioGroupOption[] = [
  { value: SAMLNameIDFormat.Unspecified, text: SAMLNameIDFormat.Unspecified },
  {
    value: SAMLNameIDFormat.EmailAddress,
    text: SAMLNameIDFormat.EmailAddress,
  },
];

function makeNameIDAttributePointerOptions(
  renderToString: MessageFormatContextValue["renderToString"]
): RadioGroupOption[] {
  return [
    {
      value: SAMLNameIDAttributePointer.Sub,
      text: renderToString(
        "OAuthClientSAMLForm.nameIDAttribute.options.userID"
      ),
    },
    {
      value: SAMLNameIDAttributePointer.Email,
      text: renderToString("OAuthClientSAMLForm.nameIDAttribute.options.email"),
    },
    {
      value: SAMLNameIDAttributePointer.PhoneNumber,
      text: renderToString("OAuthClientSAMLForm.nameIDAttribute.options.phone"),
    },
    {
      value: SAMLNameIDAttributePointer.PreferredUsername,
      text: renderToString(
        "OAuthClientSAMLForm.nameIDAttribute.options.username"
      ),
    },
  ];
}

function makeSLOCallbackBindingOptions(
  renderToString: MessageFormatContextValue["renderToString"]
): RadioGroupOption[] {
  return [
    {
      value: SAMLBinding.HTTPRedirect,
      text: renderToString(
        "OAuthClientSAMLForm.logout.callbackBinding.options.httpRedirect"
      ),
    },
    {
      value: SAMLBinding.HTTPPOST,
      text: renderToString(
        "OAuthClientSAMLForm.logout.callbackBinding.options.httpPost"
      ),
    },
  ];
}

function RadioGroupField({
  label,
  options,
  value,
  disabled,
  onValueChange,
}: {
  label: string;
  options: RadioGroupOption[];
  value: string | null;
  disabled?: boolean;
  onValueChange: (value: string) => void;
}): React.ReactElement {
  return (
    <FormField size="2" label={label}>
      <RadioGroup.Root value={value ?? ""} onValueChange={onValueChange}>
        <Flex direction="column" gap="2">
          {options.map((option) => (
            <Text
              as="label"
              size="2"
              key={option.value}
              className={
                disabled === true ? styles.radioOptionLabelDisabled : undefined
              }
            >
              <Flex align="center" gap="2">
                <RadioGroup.Item value={option.value} disabled={disabled} />
                {option.text}
              </Flex>
            </Text>
          ))}
        </Flex>
      </RadioGroup.Root>
    </FormField>
  );
}

interface TextAreaListItemProps {
  index: number;
  itemsJSONPointer: string | RegExp;
  value: string;
  canDelete: boolean;
  onItemChange: (index: number, value: string) => void;
  onItemDelete: (index: number) => void;
}

function TextAreaListItem({
  index,
  itemsJSONPointer,
  value,
  canDelete,
  onItemChange,
  onItemDelete,
}: TextAreaListItemProps): React.ReactElement {
  const onChange = useCallback(
    (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      onItemChange(index, e.currentTarget.value);
    },
    [index, onItemChange]
  );
  const onDeleteClick = useCallback(() => {
    onItemDelete(index);
  }, [index, onItemDelete]);

  return (
    <div className={styles.textAreaListRow}>
      <div className={styles.textAreaListRowField}>
        <TextArea
          size="2"
          className={styles.textAreaListTextArea}
          parentJSONPointer={itemsJSONPointer}
          fieldName={index.toString(10)}
          value={value}
          onChange={onChange}
        />
      </div>
      {canDelete ? (
        <RadixIconButton
          className={styles.textAreaListDeleteButton}
          variant="ghost"
          color="gray"
          size="2"
          type="button"
          onClick={onDeleteClick}
        >
          <TrashIcon width="1rem" height="1rem" />
        </RadixIconButton>
      ) : null}
    </div>
  );
}

// A multiline counterpart of the v2 TextFieldList: a growable list of
// textareas with per-item form error binding (items bind to
// <parent>/<fieldName>/<index>), a list-level error message, and add/delete
// controls.
function TextAreaList({
  label,
  description,
  parentJSONPointer,
  fieldName,
  list: propList,
  onListChange,
  addButtonLabelMessageID,
  minItem,
  maxItem,
}: {
  label: React.ReactNode;
  description?: React.ReactNode;
  parentJSONPointer: string | RegExp;
  fieldName: string;
  list: string[];
  onListChange: (newList: string[]) => void;
  addButtonLabelMessageID: string;
  minItem: number;
  maxItem: number;
}): React.ReactElement {
  const field = useMemo(
    () => ({ parentJSONPointer, fieldName }),
    [parentJSONPointer, fieldName]
  );
  const { errors } = useFormField(field);

  const itemsJSONPointer = useMemo(
    () => joinParentChild(parentJSONPointer, fieldName),
    [parentJSONPointer, fieldName]
  );

  const list = useMemo(() => {
    // If the number of items is less than minItem, fill with empty items.
    if (minItem === 0) {
      return propList;
    }
    if (propList.length === 0) {
      return new Array(minItem).fill("") as string[];
    }
    return propList;
  }, [minItem, propList]);

  const canDelete = list.length > minItem;
  const canAdd = list.length < maxItem;

  const onItemChange = useCallback(
    (index: number, value: string) => {
      const newList = list.slice();
      newList[index] = value;
      onListChange(newList);
    },
    [list, onListChange]
  );

  const onItemDelete = useCallback(
    (index: number) => {
      const newList = list.slice();
      newList.splice(index, 1);
      onListChange(newList);
    },
    [list, onListChange]
  );

  const onAddClick = useCallback(() => {
    onListChange([...list, ""]);
  }, [list, onListChange]);

  return (
    <div className={styles.textAreaList}>
      <Text
        as="p"
        size="2"
        weight="medium"
        className={styles.textAreaListLabel}
      >
        {label}
      </Text>
      <div className={styles.textAreaListItems}>
        {list.map((value, index) => (
          <TextAreaListItem
            key={index}
            index={index}
            itemsJSONPointer={itemsJSONPointer}
            value={value}
            canDelete={canDelete}
            onItemChange={onItemChange}
            onItemDelete={onItemDelete}
          />
        ))}
      </div>
      {errors.length > 0 ? (
        <Text as="p" size="1" color="red" className={styles.textAreaListErrors}>
          <ErrorRenderer errors={errors} />
        </Text>
      ) : null}
      {canAdd ? (
        <span className={styles.textAreaListAddButton}>
          <SecondaryButton
            size="2"
            onClick={onAddClick}
            text={<FormattedMessage id={addButtonLabelMessageID} />}
          />
        </span>
      ) : null}
      {description != null ? (
        <Text as="p" size="1" className={styles.textAreaListDescription}>
          {description}
        </Text>
      ) : null}
    </div>
  );
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
          <span className="inline-block">
            <SecondaryButton
              size="2"
              onClick={onDownloadIdpCertificate}
              text={
                <>
                  <DownloadIcon />
                  <FormattedMessage id="OAuthClientSAMLForm.idpCertificate.download" />
                </>
              }
            />
          </span>
          <Text as="p" size="2" className="mt-1">
            <FormattedMessage
              id="OAuthClientSAMLForm.idpCertificate.fingerprint"
              values={{
                fingerprint: samlIdpSigningCertificate.certificateFingerprint,
              }}
            />
          </Text>
        </div>

        <Callout
          type="info"
          showCloseButton={false}
          text={
            <FormattedMessage
              id="OAuthClientSAMLForm.idpCertificate.rotateHint"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                reactRouterLink: (chunks: React.ReactNode) => (
                  <Link to={`/project/${appID}/advanced/saml-certificate`}>
                    {chunks}
                  </Link>
                ),
              }}
            />
          }
        />
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
  const { getIsDirty } = useFormContainerBaseContext();
  const isFormDirty = useMemo(() => getIsDirty(), [getIsDirty]);
  const { appID } = useParams() as { appID: string };

  const onIsSAMLEnabledChange = useCallback(
    (checked: boolean) => {
      onFormStateChange({ ...formState, isSAMLEnabled: checked });
    },
    [formState, onFormStateChange]
  );

  const onNameIDFormatChange = useCallback(
    (value: string) => {
      onFormStateChange({
        ...formState,
        nameIDFormat: value as SAMLNameIDFormat,
      });
    },
    [formState, onFormStateChange]
  );

  const onNameIDAttributePointerChange = useCallback(
    (value: string) => {
      onFormStateChange({
        ...formState,
        nameIDAttributePointer: value as SAMLNameIDAttributePointer,
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

  const onAcsUrlItemAdd = useCallback(
    (list: string[], item: string) => {
      onAcsUrlsChange([...list, item]);
    },
    [onAcsUrlsChange]
  );

  const onAcsUrlItemChange = useCallback(
    (list: string[], index: number, item: string) => {
      const newList = list.slice();
      newList[index] = item;
      onAcsUrlsChange(newList);
    },
    [onAcsUrlsChange]
  );

  const onAcsUrlItemDelete = useCallback(
    (list: string[], index: number, _item: string) => {
      const newList = list.slice();
      newList.splice(index, 1);
      onAcsUrlsChange(newList);
    },
    [onAcsUrlsChange]
  );

  const onTextfieldChange = useMemo(() => {
    const makeOnChangeCallback = (
      key: keyof OAuthClientSAMLFormState
    ): React.ChangeEventHandler<HTMLInputElement> => {
      return (e) => {
        onFormStateChange({
          ...formState,
          [key]: e.currentTarget.value,
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
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const newValue = e.currentTarget.value;
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
    (checked: boolean) => {
      onFormStateChange({ ...formState, isSLOEnabled: checked });
    },
    [formState, onFormStateChange]
  );

  const onSLOCallbackBindingChange = useCallback(
    (value: string) => {
      onFormStateChange({
        ...formState,
        sloCallbackBinding: value as SAMLBinding,
      });
    },
    [formState, onFormStateChange]
  );

  const onSignatureVerificationEnabledChange = useCallback(
    (checked: boolean) => {
      onFormStateChange({
        ...formState,
        signatureVerificationEnabled: checked,
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
      <div className="grid gap-y-1 grid-cols-1">
        <Toggle
          text={renderToString("OAuthClientSAMLForm.enable.label")}
          textWeight="medium"
          checked={formState.isSAMLEnabled}
          onCheckedChange={onIsSAMLEnabledChange}
        />
        <Text as="p" size="2" className={styles.toggleDescription}>
          {renderToString("OAuthClientSAMLForm.enable.description")}
        </Text>
      </div>
      {formState.isSAMLEnabled ? (
        <>
          <Separator size="4" className="my-12" />
          <div className="grid gap-y-12 grid-cols-1">
            <div>
              <Text as="p" size="5" weight="bold">
                <FormattedMessage id="OAuthClientSAMLForm.title" />
              </Text>
              <div className="mt-3 grid gap-y-2 grid-cols-1 items-start justify-items-start">
                <Text as="p" size="2">
                  <FormattedMessage id="OAuthClientSAMLForm.metadataUpload.description" />
                </Text>
                <div className="grid grid-flow-col items-center gap-x-2">
                  <SecondaryButton
                    size="2"
                    text={renderToString(
                      "OAuthClientSAMLForm.metadataUpload.label"
                    )}
                    onClick={onUploadMetadata}
                  />
                  {formState.isMetadataUploaded ? (
                    <div className={styles.metadataUploadedIndicator}>
                      <CheckIcon width="1rem" height="1rem" />
                      <Text size="2">
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
                <RadioGroupField
                  label={renderToString(
                    "OAuthClientSAMLForm.nameIDFormat.label"
                  )}
                  options={nameIDFormatOptions}
                  value={formState.nameIDFormat}
                  onValueChange={onNameIDFormatChange}
                />
                <RadioGroupField
                  label={renderToString(
                    "OAuthClientSAMLForm.nameIDAttribute.label"
                  )}
                  disabled={
                    formState.nameIDFormat !== SAMLNameIDFormat.Unspecified
                  }
                  options={nameIDAttributePointerOptions}
                  value={
                    formState.nameIDFormat !== SAMLNameIDFormat.Unspecified
                      ? null
                      : formState.nameIDAttributePointer
                  }
                  onValueChange={onNameIDAttributePointerChange}
                />
              </div>
            </div>

            <div>
              <WidgetTitle className="mb-3" id="sso">
                <FormattedMessage id="OAuthClientSAMLForm.sso.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <TextFieldList
                  parentJSONPointer={parentJSONPointer}
                  fieldName="acs_urls"
                  list={formState.acsURLs}
                  onListItemAdd={onAcsUrlItemAdd}
                  onListItemChange={onAcsUrlItemChange}
                  onListItemDelete={onAcsUrlItemDelete}
                  addButtonLabelMessageID="OAuthClientSAMLForm.sso.acsUrls.add"
                  label={
                    <FormattedMessage id="OAuthClientSAMLForm.sso.acsUrls.title" />
                  }
                  minItem={1}
                />
                <TextField
                  size="2"
                  parentJSONPointer={parentJSONPointer}
                  fieldName="destination"
                  label={renderToString(
                    "OAuthClientSAMLForm.sso.destination.label"
                  )}
                  hint={renderToString(
                    "OAuthClientSAMLForm.sso.destination.description"
                  )}
                  value={formState.destination}
                  onChange={onTextfieldChange.destination}
                />
                <TextField
                  size="2"
                  parentJSONPointer={parentJSONPointer}
                  fieldName="recipient"
                  label={renderToString(
                    "OAuthClientSAMLForm.sso.recipient.label"
                  )}
                  hint={renderToString(
                    "OAuthClientSAMLForm.sso.recipient.description"
                  )}
                  value={formState.recipient}
                  onChange={onTextfieldChange.recipient}
                />
                <TextField
                  size="2"
                  parentJSONPointer={parentJSONPointer}
                  fieldName="audience"
                  label={renderToString(
                    "OAuthClientSAMLForm.sso.audience.label"
                  )}
                  hint={renderToString(
                    "OAuthClientSAMLForm.sso.audience.description"
                  )}
                  value={formState.audience}
                  onChange={onTextfieldChange.audience}
                />
                <TextField
                  size="2"
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
                  text={renderToString(
                    "OAuthClientSAMLForm.logout.enable.label"
                  )}
                  textWeight="medium"
                  checked={formState.isSLOEnabled}
                  onCheckedChange={onIsSLOEnabledChange}
                />
                <TextField
                  size="2"
                  parentJSONPointer={parentJSONPointer}
                  fieldName="slo_callback_url"
                  label={renderToString(
                    "OAuthClientSAMLForm.logout.callbackURL.label"
                  )}
                  hint={renderToString(
                    "OAuthClientSAMLForm.logout.callbackURL.description"
                  )}
                  value={formState.isSLOEnabled ? formState.sloCallbackURL : ""}
                  onChange={onTextfieldChange.sloCallbackURL}
                  disabled={!formState.isSLOEnabled}
                />
                <RadioGroupField
                  label={renderToString(
                    "OAuthClientSAMLForm.logout.callbackBinding.label"
                  )}
                  disabled={!formState.isSLOEnabled}
                  options={sloBindingOptions}
                  value={
                    formState.isSLOEnabled ? formState.sloCallbackBinding : null
                  }
                  onValueChange={onSLOCallbackBindingChange}
                />
              </div>
            </div>

            <div>
              <WidgetTitle className="mb-3" id="signature">
                <FormattedMessage id="OAuthClientSAMLForm.signature.title" />
              </WidgetTitle>
              <div className="grid gap-y-4 grid-cols-1">
                <TextAreaList
                  parentJSONPointer={
                    /\/secrets\/(\d*)\/data\/(\d*)\/certificates\/(\d*)/
                  }
                  fieldName="pem"
                  list={formState.signingCertificates}
                  onListChange={onSigningCertificatesChange}
                  addButtonLabelMessageID="OAuthClientSAMLForm.signature.certificates.add"
                  label={
                    <FormattedMessage id="OAuthClientSAMLForm.signature.certificates.label" />
                  }
                  description={renderToString(
                    "OAuthClientSAMLForm.signature.certificates.description"
                  )}
                  maxItem={2}
                  minItem={formState.signatureVerificationEnabled ? 1 : 0}
                />
                <div className="grid gap-y-2 grid-cols-1">
                  <Toggle
                    text={renderToString(
                      "OAuthClientSAMLForm.signature.checkSignature.label"
                    )}
                    textWeight="medium"
                    disabled={formState.signingCertificates.length < 1}
                    checked={formState.signatureVerificationEnabled}
                    onCheckedChange={onSignatureVerificationEnabledChange}
                  />
                  <Text as="p" size="2" className={styles.toggleDescription}>
                    {renderToString(
                      "OAuthClientSAMLForm.signature.checkSignature.description"
                    )}
                  </Text>
                  {formState.signingCertificates.length < 1 ? (
                    <Callout
                      type="warning"
                      showCloseButton={false}
                      text={
                        <FormattedMessage id="OAuthClientSAMLForm.signature.checkSignature.hint" />
                      }
                    />
                  ) : null}
                </div>
              </div>
            </div>

            <Separator size="4" />

            <div>
              <WidgetTitle className="mb-6" id="configuration-parameters">
                <FormattedMessage id="OAuthClientSAMLForm.configurationParameters.title" />
              </WidgetTitle>

              <div className="grid gap-y-4 grid-cols-1">
                <div className="grid gap-y-2 grid-cols-1">
                  <TextField
                    size="2"
                    label={renderToString(
                      "OAuthClientSAMLForm.configurationParameters.metadata.label"
                    )}
                    value={endpoints.metadata}
                    readOnly={true}
                    suffixPlain={true}
                    suffix={<CopyIconButton textToCopy={endpoints.metadata} />}
                  />
                  <span className="justify-self-start">
                    <SecondaryButton
                      size="2"
                      text={
                        <>
                          <DownloadIcon />
                          <FormattedMessage id="OAuthClientSAMLForm.configurationParameters.metadata.download" />
                        </>
                      }
                      onClick={onClickDownloadMetadata}
                      disabled={isFormDirty}
                    />
                  </span>
                  {isFormDirty ? (
                    <Callout
                      type="warning"
                      showCloseButton={false}
                      text={
                        <FormattedMessage id="OAuthClientSAMLForm.configurationParameters.metadata.saveBeforeDownload.hint" />
                      }
                    />
                  ) : null}
                </div>
                <TextField
                  size="2"
                  label={renderToString(
                    "OAuthClientSAMLForm.configurationParameters.issuer.label"
                  )}
                  value={samlIdpEntityID}
                  readOnly={true}
                  suffixPlain={true}
                  suffix={<CopyIconButton textToCopy={samlIdpEntityID} />}
                />
                <TextField
                  size="2"
                  label={renderToString(
                    "OAuthClientSAMLForm.configurationParameters.loginURL.label"
                  )}
                  value={endpoints.login}
                  readOnly={true}
                  suffixPlain={true}
                  suffix={<CopyIconButton textToCopy={endpoints.login} />}
                />
                <TextField
                  size="2"
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
                  suffixPlain={true}
                  suffix={
                    formState.isSLOEnabled ? (
                      <CopyIconButton textToCopy={endpoints.logout} />
                    ) : undefined
                  }
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
