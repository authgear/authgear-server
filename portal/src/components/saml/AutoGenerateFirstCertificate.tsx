import React, { useCallback, useEffect } from "react";
import { PortalAPIAppConfig, SAMLIdpSigningCertificate } from "../../types";
import { useUpdateAppAndSecretConfigMutation } from "../../graphql/portal/mutations/updateAppAndSecretMutation";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";

export function AutoGenerateFirstCertificate({
  appID,
  rawAppConfig,
  certificates,
  onComplete,
}: {
  appID: string;
  rawAppConfig: PortalAPIAppConfig;
  certificates: SAMLIdpSigningCertificate[];
  onComplete: () => void;
}): React.ReactElement {
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
