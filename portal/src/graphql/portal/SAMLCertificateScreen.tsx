import React, { useCallback, useEffect } from "react";
import { useParams } from "react-router-dom";
import {
  useSAMLCertificateForm,
  FormState,
} from "../../hook/useSAMLCertificateForm";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { FormContainerBase } from "../../FormContainerBase";
import { PortalAPIAppConfig, SAMLIdpSigningCertificate } from "../../types";
import { useUpdateAppAndSecretConfigMutation } from "./mutations/updateAppAndSecretMutation";
import { AppSecretConfigFormModel } from "../../hook/useAppSecretConfigForm";

function AutoGenerateFirstCertificate({
  appID,
  rawAppConfig,
  certificates,
  onComplete,
}: {
  appID: string;
  rawAppConfig: PortalAPIAppConfig;
  certificates: SAMLIdpSigningCertificate[];
  onComplete: () => void;
}) {
  const { error, updateAppAndSecretConfig } =
    useUpdateAppAndSecretConfigMutation(appID);

  const generate = useCallback(async () => {
    let keyID: string;
    if (certificates.length === 0) {
      const result = await updateAppAndSecretConfig({
        secretConfigUpdateInstructions: {
          samlIdpSigningSecrets: {
            action: "generate",
          },
        },
      });
      if (result?.secretConfig == null) {
        throw new Error("unexpected null secretConfig");
      }
      if (
        (result.secretConfig.samlIdpSigningSecrets?.certificates.length ??
          0) === 0
      ) {
        throw new Error("unexpected 0 length idp signing certificates");
      }
      keyID = result.secretConfig.samlIdpSigningSecrets!.certificates[0].keyID;
    } else {
      keyID = certificates[0].keyID;
    }
    const newConfig = { ...rawAppConfig };
    newConfig.saml ??= {};
    newConfig.saml.signing ??= {};
    newConfig.saml.signing.key_id = keyID;
    await updateAppAndSecretConfig({
      appConfig: newConfig,
    });

    onComplete();
  }, [certificates, rawAppConfig, updateAppAndSecretConfig, onComplete]);

  useEffect(() => {
    generate();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (error) {
    return <ShowError error={error} onRetry={generate} />;
  }

  return <ShowLoading />;
}

function EditSAMLCertificateForm({}: {
  form: AppSecretConfigFormModel<FormState>;
  certificates: SAMLIdpSigningCertificate[];
}) {
  return <></>;
}

function EditSAMLCertificateFormContainer({
  appID,
  certificates,
}: {
  appID: string;
  certificates: SAMLIdpSigningCertificate[];
}) {
  const form = useSAMLCertificateForm(appID);

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  if (form.isLoading) {
    return <ShowLoading />;
  }

  return (
    <FormContainerBase form={form} canSave={true}>
      <EditSAMLCertificateForm certificates={certificates} form={form} />
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
      certificates={secretConfig!.samlIdpSigningSecrets!.certificates}
    />
  );
}
