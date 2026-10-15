import React, { useContext } from "react";
import {
  AuthenticationFlowsConfig,
  CustomAttributesAttributeConfig,
  OAuthSSOProviderConfig,
} from "../../../types";
import { FlowType } from "./model";

export interface FlowSketchContextValue {
  flowType: FlowType;
  oauthProviders: OAuthSSOProviderConfig[];
  customAttributes: CustomAttributesAttributeConfig[];
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
