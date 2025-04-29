import React, { useMemo, type ReactElement } from "react";
import { useParams } from "react-router-dom";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import { AppContext } from "./context/AppContext";

export interface AppContextProviderProps {
  children?: React.ReactNode;
}

export default function AppContextProvider(
  props: AppContextProviderProps
): ReactElement {
  const { children } = props;
  const { appID: appNodeID } = useParams() as { appID: string };

  // NOTE: check if appID actually exist in authorized app list
  const { effectiveAppConfig, viewer } = useAppAndSecretConfigQuery(appNodeID);

  const appContextValue = useMemo(() => {
    return {
      appNodeID: appNodeID,
      appID: effectiveAppConfig?.id ?? "",
      currentCollaboratorRole: viewer?.role ?? "",
    };
  }, [appNodeID, effectiveAppConfig, viewer?.role]);

  return (
    <AppContext.Provider value={appContextValue}>
      {children}
    </AppContext.Provider>
  );
}
