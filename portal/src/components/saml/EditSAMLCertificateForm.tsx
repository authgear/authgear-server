import React, { useCallback, useState } from "react";
import cn from "classnames";
import {
  DotsVerticalIcon,
  DownloadIcon,
  PlusIcon,
  TrashIcon,
} from "@radix-ui/react-icons";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Spinner,
  Text,
} from "@radix-ui/themes";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import { SAMLIdpSigningCertificate } from "../../types";
import { FormState } from "../../hook/useSAMLCertificateForm";
import { FormattedMessage } from "../../intl";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { downloadStringAsFile } from "../../util/download";
import { formatCertificateFilename } from "../../model/saml";
import styles from "./EditSAMLCertificateForm.module.css";
import { ConfirmationDialog } from "../v2/ConfirmationDialog/ConfirmationDialog";

interface EditSAMLCertificateFormProps {
  configAppID: string;
  form: AppSecretConfigFormModel<FormState>;
  certificates: SAMLIdpSigningCertificate[];
  onGenerateNewCertitificate: () => Promise<void>;
}

interface SAMLCertificatesTableProps {
  certificates: SAMLIdpSigningCertificate[];
  activeKeyID: string | undefined;
  activatingKeyID: string | null;
  formDisabled: boolean;
  onDownload: (cert: SAMLIdpSigningCertificate) => void;
  onRemove: (cert: SAMLIdpSigningCertificate) => void;
  onActivate: (cert: SAMLIdpSigningCertificate) => void;
}

function SAMLCertificatesTable({
  certificates,
  activeKeyID,
  activatingKeyID,
  formDisabled,
  onDownload,
  onRemove,
  onActivate,
}: SAMLCertificatesTableProps): React.ReactElement {
  return (
    <div className={styles.keysTableWrapper}>
      <div className={styles.keysTable}>
        <div className={styles.keysTableHeader}>
          <div className={styles.keysTableHeaderCellFingerprint}>
            <FormattedMessage id="EditSAMLCertificateForm.certificates.column.fingerprint" />
          </div>
          <div className={styles.keysTableHeaderCellStatus}>
            <FormattedMessage id="EditSAMLCertificateForm.certificates.column.status" />
          </div>
          <div
            className={styles.keysTableHeaderCellActions}
            aria-hidden={true}
          />
        </div>
        {certificates.map((cert) => {
          const isActive = activeKeyID === cert.keyID;
          const isActivating = activatingKeyID === cert.keyID;
          return (
            <div key={cert.keyID} className={styles.keysTableRow}>
              <div className={styles.keysTableCellFingerprint}>
                <Text size="2" className={styles.keysTableCellFingerprintText}>
                  {cert.certificateFingerprint}
                </Text>
              </div>
              <div className={styles.keysTableCellStatus}>
                {isActive || isActivating ? (
                  <CertificateActiveStatus isLoading={isActivating} />
                ) : (
                  <button
                    type="button"
                    className={styles.activateButton}
                    disabled={formDisabled}
                    onClick={() => onActivate(cert)}
                  >
                    <FormattedMessage id="EditSAMLCertificateForm.certificates.column.status.activate" />
                  </button>
                )}
              </div>
              <div className={styles.keysTableCellActions}>
                <DropdownMenu.Root>
                  <DropdownMenu.Trigger>
                    <RadixIconButton variant="soft" color="gray" size="2">
                      <DotsVerticalIcon width="1rem" height="1rem" />
                    </RadixIconButton>
                  </DropdownMenu.Trigger>
                  <DropdownMenu.Content align="end">
                    <DropdownMenu.Item
                      onSelect={() => {
                        onDownload(cert);
                      }}
                    >
                      <DownloadIcon />
                      <FormattedMessage id="download" />
                    </DropdownMenu.Item>
                    <DropdownMenu.Item
                      color="red"
                      disabled={isActive || formDisabled}
                      onSelect={() => {
                        if (!isActive) {
                          onRemove(cert);
                        }
                      }}
                    >
                      <TrashIcon />
                      <FormattedMessage id="EditSAMLCertificateForm.certificates.remove" />
                    </DropdownMenu.Item>
                  </DropdownMenu.Content>
                </DropdownMenu.Root>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export function EditSAMLCertificateForm({
  configAppID,
  form,
  certificates,
  onGenerateNewCertitificate,
}: EditSAMLCertificateFormProps): React.ReactElement {
  const { onSubmit } = useFormContainerBaseContext();

  const [isGenerating, setIsGenerating] = useState(false);
  const [activatingKeyID, setActivatingKeyID] = useState<string | null>(null);

  const generateNewCert = useCallback(async () => {
    setIsGenerating(true);
    try {
      await onGenerateNewCertitificate();
    } finally {
      setIsGenerating(false);
    }
  }, [onGenerateNewCertitificate]);

  const onClickDownloadCert = useCallback(
    (cert: SAMLIdpSigningCertificate) => {
      downloadStringAsFile({
        content: cert.certificatePEM,
        mimeType: "application/x-pem-file",
        filename: formatCertificateFilename(
          configAppID,
          cert.certificateFingerprint
        ),
      });
    },
    [configAppID]
  );

  const onRemoveCert = useCallback(
    (cert: SAMLIdpSigningCertificate) => {
      form.setState((prevState) => ({
        ...prevState,
        removingCertificateKeyID: cert.keyID,
      }));
    },
    [form]
  );

  const onChangeActiveKey = useCallback(
    async (cert: SAMLIdpSigningCertificate) => {
      if (form.isUpdating || activatingKeyID != null) {
        return;
      }
      setActivatingKeyID(cert.keyID);
      try {
        await form.saveWithState({
          ...form.state,
          isUpdatingActiveKeyID: true,
          activeKeyID: cert.keyID,
        });
      } finally {
        setActivatingKeyID(null);
      }
    },
    [activatingKeyID, form]
  );

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

  const isRemoveCertificateDialogOpen =
    form.state.removingCertificateKeyID != null;

  const formDisabled =
    form.isLoading || form.isUpdating || activatingKeyID != null;

  return (
    <form onSubmit={onSubmit}>
      <SAMLCertificatesTable
        certificates={certificates}
        activeKeyID={form.state.activeKeyID}
        activatingKeyID={activatingKeyID}
        formDisabled={formDisabled}
        onDownload={onClickDownloadCert}
        onRemove={onRemoveCert}
        // eslint-disable-next-line @typescript-eslint/strict-void-return
        onActivate={onChangeActiveKey}
      />

      <button
        type="button"
        className={cn(styles.generateKeyButton, "mt-4")}
        // eslint-disable-next-line @typescript-eslint/strict-void-return
        onClick={generateNewCert}
        disabled={certificates.length >= 2 || isGenerating || formDisabled}
      >
        <PlusIcon width="1rem" height="1rem" />
        <FormattedMessage id="EditSAMLCertificateForm.certificates.generate" />
        {isGenerating ? <Spinner size="1" className="ml-1" /> : null}
      </button>

      <ConfirmationDialog
        open={isRemoveCertificateDialogOpen}
        onOpenChange={(open) => {
          if (!open) {
            dismissRemoveCertificateDialog();
          }
        }}
        title={
          <FormattedMessage id="EditSAMLCertificateForm.removeCertificateDialog.title" />
        }
        description={
          <FormattedMessage id="EditSAMLCertificateForm.removeCertificateDialog.description" />
        }
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        loading={form.isUpdating}
        confirmColor="red"
        onConfirm={onConfirmRemoveCertificate}
        onCancel={dismissRemoveCertificateDialog}
      />
    </form>
  );
}

function CertificateActiveStatus({ isLoading }: { isLoading: boolean }) {
  return (
    <div className={styles.activeStatus}>
      <Text
        as="p"
        size="2"
        weight="medium"
        className={cn(
          styles.activeStatusText,
          isLoading ? "invisible" : undefined
        )}
      >
        <FormattedMessage id="EditSAMLCertificateForm.certificates.column.status.active" />
      </Text>
      {isLoading ? (
        <div className={styles.activeStatusSpinner}>
          <Spinner size="1" />
        </div>
      ) : null}
    </div>
  );
}
