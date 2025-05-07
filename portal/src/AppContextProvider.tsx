import React, { useMemo, type ReactElement } from "react";
import { Navigate, useParams } from "react-router-dom";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import { AppContext } from "./context/AppContext";
import ShowLoading from "./ShowLoading";
import ShowError from "./ShowError";

export interface AppContextProviderProps {
  children?: React.ReactNode;
}

export default function AppContextProvider(
  props: AppContextProviderProps
): ReactElement {
  const { children } = props;
  const { appID: appNodeID } = useParams() as { appID: string };

  // NOTE: check if appID actually exist in authorized app list
  const { effectiveAppConfig, viewer, error, loading } =
    useAppAndSecretConfigQuery(appNodeID);

  const appContextValue = useMemo(() => {
    return {
      appNodeID: appNodeID,
      appID: effectiveAppConfig?.id ?? "",
      currentCollaboratorRole: viewer?.role ?? "",
    };
  }, [appNodeID, effectiveAppConfig, viewer?.role]);

  if (loading) {
    return <ShowLoading />;
  }

  // if node is null after loading without error, treat this as invalid
  // request, frontend cannot distinguish between inaccessible and not found
  const isInvalidAppID = error == null && effectiveAppConfig == null;

  // redirect to app list if app id is invalid
  if (isInvalidAppID) {
    return <Navigate to="/projects" replace={true} />;
  }

  if (error != null) {
    return <ShowError error={error} />;
  }

  return (
    <AppContext.Provider value={appContextValue}>
      {children}
    </AppContext.Provider>
  );
}
