import React, { useCallback, useMemo } from "react";
import { useParams } from "react-router-dom";
import {
  useSAMLCertificateForm,
  FormState,
} from "../../hook/useSAMLCertificateForm";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { FormContainerBase } from "../../FormContainerBase";
import { SAMLIdpSigningCertificate } from "../../types";
import { useUpdateAppAndSecretConfigMutation } from "./mutations/updateAppAndSecretMutation";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ScreenDescription from "../../ScreenDescription";
import { FormattedMessage } from "@oursky/react-messageformat";
import styles from "./SAMLCertificateScreen.module.css";
import { EditSAMLCertificateForm } from "../../components/saml/EditSAMLCertificateForm";
import { AutoGenerateFirstCertificate } from "../../components/saml/AutoGenerateFirstCertificate";

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
  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="SAMLCertificateScreen.title" />,
      },
    ];
  }, []);

  return (
    <ScreenLayoutScrollView>
      <ScreenContent>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="SAMLCertificateScreen.desc" />
        </ScreenDescription>
        <div className={styles.widget}>
          <EditSAMLCertificateForm
            configAppID={configAppID}
            form={form}
            certificates={certificates}
            onGenerateNewCertitificate={generateNewCertificate}
          />
        </div>
      </ScreenContent>
    </ScreenLayoutScrollView>
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
    <FormContainerBase form={form} canSave={true}>
      <EditSAMLCertificateContent
        configAppID={configAppID}
        certificates={certificates}
        form={form}
        generateNewCertificate={generateNewCertificate}
      />
    </FormContainerBase>
  );
}

export default function SAMLCertificateScreen(): React.ReactElement {
  const { appID } = useParams() as {
    appID: string;
  };
  const { rawAppConfig, secretConfig, loading, error, refetch } =
    useAppAndSecretConfigQuery(appID);

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
