import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import { SAMLIdpSigningCertificate } from "../../types";
import { FormState } from "../../hook/useSAMLCertificateForm";
import WidgetTitle from "../../WidgetTitle";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import {
  DetailsList,
  IColumn,
  SelectionMode,
  Text,
  ChoiceGroup,
  IChoiceGroupOption,
  ILinkStyles,
} from "@fluentui/react";
import LinkButton from "../../LinkButton";
import { downloadStringAsFile } from "../../util/download";
import { useSystemConfig } from "../../context/SystemConfigContext";
import styles from "./EditSAMLCertificateForm.module.css";
import ActionButton from "../../ActionButton";
import PrimaryButton from "../../PrimaryButton";

interface EditSAMLCertificateFormProps {
  form: AppSecretConfigFormModel<FormState>;
  certificates: SAMLIdpSigningCertificate[];
  onGenerateNewCertitificate: () => Promise<void>;
}

const actionLinkButtonStyle: ILinkStyles = { root: { fontSize: 14 } };

export function EditSAMLCertificateForm({
  form,
  certificates,
  onGenerateNewCertitificate,
}: EditSAMLCertificateFormProps): React.ReactElement {
  const { canSave, onSubmit } = useFormContainerBaseContext();
  const { renderToString } = useContext(MessageContext);
  const { themes } = useSystemConfig();

  const [isLoading, setIsLoading] = useState(false);

  const generateNewCert = useCallback(async () => {
    setIsLoading(true);
    try {
      await onGenerateNewCertitificate();
    } finally {
      setIsLoading(false);
    }
  }, [onGenerateNewCertitificate]);

  const onClickDownloadCert = useMemo(() => {
    const callbacks: Record<string, () => void> = {};
    for (const cert of certificates) {
      callbacks[cert.keyID] = () => {
        downloadStringAsFile({
          content: cert.certificatePEM,
          mimeType: "application/x-pem-file",
          filename: `${cert.certificateFingerprint}.pem`,
        });
      };
    }
    return callbacks;
  }, [certificates]);

  const onToggleRemoveCert = useMemo(() => {
    const callbacks: Record<string, () => void> = {};
    for (const cert of certificates) {
      callbacks[cert.keyID] = () => {
        form.setState((prevState) => {
          const newKeyIDs = new Set(prevState.removingCertificateKeyIDs);
          if (newKeyIDs.has(cert.keyID)) {
            newKeyIDs.delete(cert.keyID);
          } else {
            newKeyIDs.add(cert.keyID);
          }
          return {
            ...prevState,
            removingCertificateKeyIDs: Array.from(newKeyIDs),
          };
        });
      };
    }
    return callbacks;
  }, [certificates, form]);

  const choiceGroupOptionsByKeyID = useMemo(() => {
    const optionsByKeyID: Record<string, IChoiceGroupOption[]> = {};
    for (const cert of certificates) {
      optionsByKeyID[cert.keyID] = [
        {
          key: cert.keyID,
          text: renderToString(
            "EditSAMLCertificateForm.certificates.column.status.active"
          ),
        },
      ];
    }
    return optionsByKeyID;
  }, [certificates, renderToString]);

  const removingCertificateKeyIDsSet = useMemo(() => {
    return new Set(form.state.removingCertificateKeyIDs);
  }, [form.state.removingCertificateKeyIDs]);

  const onChangeActiveKey = useCallback(
    (_: unknown, option?: IChoiceGroupOption) => {
      if (!option) {
        return;
      }
      form.setState((prevState) => ({ ...prevState, activeKeyID: option.key }));
    },
    [form]
  );

  const columns: IColumn[] = useMemo(() => {
    const renderFingerprint = (
      item?: SAMLIdpSigningCertificate,
      _index?: number,
      _column?: IColumn
    ) => {
      if (!item) {
        return null;
      }
      return (
        <div className="grid grid-cols-1 gap-y-2">
          <Text
            className={cn(
              "text-neutral-secondary",
              removingCertificateKeyIDsSet.has(item.keyID)
                ? "line-through"
                : null
            )}
            block={true}
          >
            {item.certificateFingerprint}
          </Text>
          <div className="grid grid-rows-1 grid-flow-col gap-x-4 justify-start">
            <LinkButton
              styles={actionLinkButtonStyle}
              onClick={onClickDownloadCert[item.keyID]}
              disabled={removingCertificateKeyIDsSet.has(item.keyID)}
            >
              <FormattedMessage id="EditSAMLCertificateForm.certificates.download" />
            </LinkButton>
            {form.state.activeKeyID !== item.keyID ? (
              <LinkButton
                styles={actionLinkButtonStyle}
                onClick={onToggleRemoveCert[item.keyID]}
                theme={
                  removingCertificateKeyIDsSet.has(item.keyID)
                    ? themes.actionButton
                    : themes.destructive
                }
              >
                {removingCertificateKeyIDsSet.has(item.keyID) ? (
                  <FormattedMessage id="EditSAMLCertificateForm.certificates.restore" />
                ) : (
                  <FormattedMessage id="EditSAMLCertificateForm.certificates.remove" />
                )}
              </LinkButton>
            ) : null}
          </div>
        </div>
      );
    };
    const renderStatus = (
      item?: SAMLIdpSigningCertificate,
      _index?: number,
      _column?: IColumn
    ) => {
      if (!item) {
        return null;
      }
      return (
        <ChoiceGroup
          options={choiceGroupOptionsByKeyID[item.keyID]}
          selectedKey={form.state.activeKeyID}
          onChange={onChangeActiveKey}
          disabled={removingCertificateKeyIDsSet.has(item.keyID)}
        />
      );
    };
    return [
      {
        key: "certificateFingerprint",
        fieldName: "certificateFingerprint",
        name: renderToString(
          "EditSAMLCertificateForm.certificates.column.fingerprint"
        ),
        minWidth: 150,
        onRender: renderFingerprint,
      },
      {
        key: "status",
        name: renderToString(
          "EditSAMLCertificateForm.certificates.column.status"
        ),
        minWidth: 150,
        onRender: renderStatus,
      },
    ];
  }, [
    renderToString,
    removingCertificateKeyIDsSet,
    onClickDownloadCert,
    form.state.activeKeyID,
    onToggleRemoveCert,
    themes.actionButton,
    themes.destructive,
    choiceGroupOptionsByKeyID,
    onChangeActiveKey,
  ]);

  return (
    <form onSubmit={onSubmit}>
      <WidgetTitle>
        <FormattedMessage id="EditSAMLCertificateForm.certificates.title" />
      </WidgetTitle>

      <div className="grid grid-cols-1 gap-y-12">
        <div>
          <DetailsList
            items={certificates}
            columns={columns}
            selectionMode={SelectionMode.none}
          />
          <ActionButton
            className="mt-4"
            theme={themes.actionButton}
            iconProps={{
              iconName: "CirclePlus",
              className: styles.addButtonIcon,
            }}
            onClick={generateNewCert}
            text={
              <FormattedMessage
                id={"EditSAMLCertificateForm.certificates.generate"}
              />
            }
            disabled={certificates.length >= 2 || isLoading}
          />
        </div>
        <PrimaryButton
          className="justify-self-start"
          type="submit"
          disabled={!canSave}
          text={<FormattedMessage id="save" />}
        />
      </div>
    </form>
  );
}
