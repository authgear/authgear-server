import { produce } from "immer";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../types";
import { useAppSecretConfigForm } from "./useAppSecretConfigForm";

export interface FormState {
  activeKeyID?: string;
  isUpdatingActiveKeyID: boolean;
  removingCertificateKeyID: string | null;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    activeKeyID: config.saml?.signing?.key_id,
    isUpdatingActiveKeyID: false,
    removingCertificateKeyID: null,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const [newConfig, _] = produce(
    [config, currentState],
    ([config, currentState]) => {
      if (currentState.activeKeyID) {
        config.saml ??= {};
        config.saml.signing ??= {};
        config.saml.signing.key_id = currentState.activeKeyID;
      }
    }
  );
  return [newConfig, secrets];
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  _secrets: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  let instruction: PortalAPISecretConfigUpdateInstruction | undefined;
  if (currentState.removingCertificateKeyID != null) {
    instruction ??= {};
    instruction.samlIdpSigningSecrets = {
      action: "delete",
      deleteData: { keyIDs: [currentState.removingCertificateKeyID] },
    };
  }
  return instruction;
}

export function useSAMLCertificateForm(
  appID: string
): ReturnType<typeof useAppSecretConfigForm<FormState>> {
  return useAppSecretConfigForm({
    appID,
    secretVisitToken: null,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
}
