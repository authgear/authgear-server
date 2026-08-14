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
import { Badge } from "../v2/Badge/Badge";
import { CardTable } from "../v2/CardTable/CardTable";

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
    <CardTable>
      <CardTable.Header>
        <CardTable.HeaderCell className={styles.colFingerprint}>
          <FormattedMessage id="EditSAMLCertificateForm.certificates.column.fingerprint" />
        </CardTable.HeaderCell>
        <CardTable.HeaderCell className={styles.colStatus}>
          <FormattedMessage id="EditSAMLCertificateForm.certificates.column.status" />
        </CardTable.HeaderCell>
        <CardTable.HeaderCell
          className={styles.colActions}
          aria-hidden={true}
        />
      </CardTable.Header>
      {certificates.map((cert) => {
        const isActive = activeKeyID === cert.keyID;
        const isActivating = activatingKeyID === cert.keyID;
        return (
          <CardTable.Row key={cert.keyID}>
            <CardTable.Cell className={styles.colFingerprint}>
              <Text size="2" className={styles.keysTableCellFingerprintText}>
                {cert.certificateFingerprint}
              </Text>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colStatus}>
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
            </CardTable.Cell>
            <CardTable.Cell className={styles.colActions}>
              <DropdownMenu.Root>
                <DropdownMenu.Trigger>
                  <RadixIconButton
                    className={styles.rowActionsButton}
                    variant="soft"
                    color="gray"
                    size="2"
                  >
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
            </CardTable.Cell>
          </CardTable.Row>
        );
      })}
    </CardTable>
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
  // Kept out of form state so that opening the confirmation dialog does not
  // dirty the form (which would show the unsaved-changes bar); the keyID is
  // merged into the form state only at save time, like onChangeActiveKey.
  const [removingCertificateKeyID, setRemovingCertificateKeyID] = useState<
    string | null
  >(null);

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

  const onRemoveCert = useCallback((cert: SAMLIdpSigningCertificate) => {
    setRemovingCertificateKeyID(cert.keyID);
  }, []);

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
    setRemovingCertificateKeyID(null);
  }, []);

  const onConfirmRemoveCertificate = useCallback(() => {
    form.saveWithState({ ...form.state, removingCertificateKeyID }).then(
      () => {
        dismissRemoveCertificateDialog();
        form.reload();
      },
      () => {
        dismissRemoveCertificateDialog();
      }
    );
  }, [form, removingCertificateKeyID, dismissRemoveCertificateDialog]);

  const isRemoveCertificateDialogOpen = removingCertificateKeyID != null;

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
          if (!open && !form.isUpdating) {
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
      <Badge
        size="1"
        variant="success"
        className={cn(isLoading ? "invisible" : undefined)}
        text={
          <FormattedMessage id="EditSAMLCertificateForm.certificates.column.status.active" />
        }
      />
      {isLoading ? (
        <div className={styles.activeStatusSpinner}>
          <Spinner size="1" />
        </div>
      ) : null}
    </div>
  );
}
