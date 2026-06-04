import React, { useCallback, useRef } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { useParams } from "react-router-dom";
import {
  useSAMLCertificateForm,
  FormState,
} from "../../hook/useSAMLCertificateForm";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormContainer from "../../FormContainer";
import { SAMLIdpSigningCertificate } from "../../types";
import { useUpdateAppAndSecretConfigMutation } from "./mutations/updateAppAndSecretMutation";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import ScreenContent from "../../ScreenContent";
import { FormattedMessage } from "../../intl";
import styles from "./SAMLCertificateScreen.module.css";
import { EditSAMLCertificateForm } from "../../components/saml/EditSAMLCertificateForm";
import { AutoGenerateFirstCertificate } from "../../components/saml/AutoGenerateFirstCertificate";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";

function EditSAMLCertificateContent({
  configAppID,
  form,
  certificates,
  generateNewCertificate,
}: {
  configAppID: string;
  form: AppSecretConfigFormModel<FormState>;
  certificates: SAMLIdpSigningCertificate[];
  generateNewCertificate: () => Promise<void>;
}) {
  const { isDirty } = useFormContainerBaseContext();
  const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

  return (
    <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
      <div
        ref={contentWidthAnchorRef}
        className={styles.contentWidthAnchor}
        aria-hidden
      />
      <div className={cn(styles.widget, styles.pageHeader)}>
        <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
          <FormattedMessage id="SAMLCertificateScreen.title" />
        </Text>
        <Text as="p" size="2" color="gray" className={styles.pageDescription}>
          <FormattedMessage id="SAMLCertificateScreen.desc" />
        </Text>
      </div>

      <div
        className={cn(
          styles.widget,
          "border border-[var(--gray-5)] rounded-lg p-6 flex gap-8 bg-white",
          isDirty && styles.settingsCardSaveBarClearance
        )}
      >
        <Text as="p" size="3" weight="medium" className={styles.sectionHeading}>
          <FormattedMessage id="EditSAMLCertificateForm.certificates.title" />
        </Text>
        <div className="flex-1 min-w-0">
          <EditSAMLCertificateForm
            configAppID={configAppID}
            form={form}
            certificates={certificates}
            onGenerateNewCertitificate={generateNewCertificate}
          />
        </div>
      </div>

      <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
    </ScreenContent>
  );
}

function EditSAMLCertificateFormContainer({
  appID,
  configAppID,
  certificates,
}: {
  appID: string;
  configAppID: string;
  certificates: SAMLIdpSigningCertificate[];
}) {
  const form = useSAMLCertificateForm(appID);
  const { updateAppAndSecretConfig } =
    useUpdateAppAndSecretConfigMutation(appID);

  const generateNewCertificate = useCallback(async () => {
    await updateAppAndSecretConfig({
      secretConfigUpdateInstructions: {
        samlIdpSigningSecrets: {
          action: "generate",
        },
      },
    });
    form.reload();
  }, [form, updateAppAndSecretConfig]);

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  if (form.isLoading) {
    return <ShowLoading />;
  }

  return (
    <FormContainer form={form} hideFooterComponent={true}>
      <EditSAMLCertificateContent
        configAppID={configAppID}
        certificates={certificates}
        form={form}
        generateNewCertificate={generateNewCertificate}
      />
    </FormContainer>
  );
}

export default function SAMLCertificateScreen(): React.ReactElement {
  const { appID } = useParams() as {
    appID: string;
  };
  const {
    rawAppConfig,
    secretConfig,
    isLoading: loading,
    loadError: error,
    refetch,
  } = useAppAndSecretConfigQuery(appID);

  if (error) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (loading) {
    return <ShowLoading />;
  }

  if (rawAppConfig == null) {
    throw new Error("unexpected null rawAppConfig");
  }

  const activeKeyID = rawAppConfig.saml?.signing?.key_id ?? null;
  const certificates = secretConfig?.samlIdpSigningSecrets?.certificates ?? [];

  if (activeKeyID == null || certificates.length === 0) {
    return (
      <AutoGenerateFirstCertificate
        appID={appID}
        onComplete={refetch}
        rawAppConfig={rawAppConfig}
        certificates={certificates}
      />
    );
  }

  return (
    <EditSAMLCertificateFormContainer
      appID={appID}
      configAppID={rawAppConfig.id}
      certificates={secretConfig!.samlIdpSigningSecrets!.certificates}
    />
  );
}
