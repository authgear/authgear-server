import React, { useContext } from "react";
import {
  AuthenticationFlowsConfig,
  AuthenticatorPhoneOTPMode,
  CustomAttributesAttributeConfig,
  OAuthSSOProviderConfig,
} from "../../../types";
import { AccountRecoveryDestination, FlowType } from "./model";

export interface FlowSketchContextValue {
  flowType: FlowType;
  oauthProviders: OAuthSSOProviderConfig[];
  customAttributes: CustomAttributesAttributeConfig[];
  // How phone codes are sent: SMS, WhatsApp or either.
  phoneOTPMode: AuthenticatorPhoneOTPMode | undefined;
  // What the flow's account recovery messages send, and how (link, code).
  recoveryDestinations: AccountRecoveryDestination[];
  custom: AuthenticationFlowsConfig | undefined;
  generated: AuthenticationFlowsConfig;
}

export const FlowSketchContext =
  React.createContext<FlowSketchContextValue | null>(null);

export function useFlowSketchContext(): FlowSketchContextValue {
  const value = useContext(FlowSketchContext);
  if (value == null) {
    throw new Error("useFlowSketchContext must be used in FlowSketchContext");
  }
  return value;
}
