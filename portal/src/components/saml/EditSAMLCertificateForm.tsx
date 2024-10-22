import React, { useCallback, useContext, useMemo, useState } from "react";
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
  ILinkStyles,
  Dialog,
  IDialogContentProps,
  DialogFooter,
} from "@fluentui/react";
import LinkButton from "../../LinkButton";
import { downloadStringAsFile } from "../../util/download";
import { useSystemConfig } from "../../context/SystemConfigContext";
import styles from "./EditSAMLCertificateForm.module.css";
import ActionButton from "../../ActionButton";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";

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
  const { onSubmit } = useFormContainerBaseContext();
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

  const onRemoveCert = useMemo(() => {
    const callbacks: Record<string, () => void> = {};
    for (const cert of certificates) {
      callbacks[cert.keyID] = () => {
        form.setState((prevState) => {
          return {
            ...prevState,
            removingCertificateKeyID: cert.keyID,
          };
        });
      };
    }
    return callbacks;
  }, [certificates, form]);

  const onChangeActiveKey = useMemo(() => {
    const callbacks: Record<string, () => void> = {};
    for (const cert of certificates) {
      callbacks[cert.keyID] = () => {
        form.setState((prevState) => ({
          ...prevState,
          activeKeyID: cert.keyID,
        }));
      };
    }
    return callbacks;
  }, [certificates, form]);

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
          <Text className={"text-neutral-secondary"} block={true}>
            {item.certificateFingerprint}
          </Text>
          <div className="grid grid-rows-1 grid-flow-col gap-x-4 justify-start">
            <LinkButton
              styles={actionLinkButtonStyle}
              onClick={onClickDownloadCert[item.keyID]}
            >
              <FormattedMessage id="EditSAMLCertificateForm.certificates.download" />
            </LinkButton>
            {form.state.activeKeyID !== item.keyID ? (
              <LinkButton
                styles={actionLinkButtonStyle}
                onClick={onRemoveCert[item.keyID]}
                theme={themes.destructive}
              >
                <FormattedMessage id="EditSAMLCertificateForm.certificates.remove" />
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
      if (form.state.activeKeyID === item.keyID) {
        return (
          <Text className="text-status-green">
            <FormattedMessage id="EditSAMLCertificateForm.certificates.column.status.active" />
          </Text>
        );
      }
      return (
        <LinkButton
          styles={actionLinkButtonStyle}
          onClick={onChangeActiveKey[item.keyID]}
          disabled={form.isLoading || form.isUpdating}
        >
          <FormattedMessage id="EditSAMLCertificateForm.certificates.column.status.activate" />
        </LinkButton>
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
    onClickDownloadCert,
    form.state.activeKeyID,
    form.isLoading,
    form.isUpdating,
    onRemoveCert,
    themes.destructive,
    onChangeActiveKey,
  ]);

  const dismissRemoveCertificateDialog = useCallback(() => {
    form.setState((state) => ({
      ...state,
      removingCertificateKeyID: null,
    }));
  }, [form]);

  const onConfirmRemoveCertificate = useCallback(() => {
    form.save().then(
      () => {
        dismissRemoveCertificateDialog();
        form.reload();
      },
      () => {
        dismissRemoveCertificateDialog();
      }
    );
  }, [form, dismissRemoveCertificateDialog]);

  const removeCertDialogContentProps: IDialogContentProps = useMemo(() => {
    return {
      title: renderToString(
        "EditSAMLCertificateForm.removeCertificateDialog.title"
      ),
      subText: renderToString(
        "EditSAMLCertificateForm.removeCertificateDialog.description"
      ),
    };
  }, [renderToString]);

  const isRemoveCertificateDialogVisible =
    form.state.removingCertificateKeyID != null;

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
      </div>

      <Dialog
        hidden={!isRemoveCertificateDialogVisible}
        dialogContentProps={removeCertDialogContentProps}
        modalProps={{ isBlocking: form.isUpdating }}
        onDismiss={dismissRemoveCertificateDialog}
      >
        <DialogFooter>
          <ButtonWithLoading
            theme={themes.actionButton}
            loading={form.isUpdating}
            onClick={onConfirmRemoveCertificate}
            disabled={!isRemoveCertificateDialogVisible}
            labelId="confirm"
          />
          <DefaultButton
            onClick={dismissRemoveCertificateDialog}
            disabled={form.isUpdating || !isRemoveCertificateDialogVisible}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    </form>
  );
}
