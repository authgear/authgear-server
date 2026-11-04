import React, { useMemo, type ReactElement } from "react";
import { Navigate, useParams } from "react-router-dom";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import { AppContext } from "./context/AppContext";
import ShowLoading from "./ShowLoading";
import ShowError from "./ShowError";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";
import { SupportLinkHrefProvider } from "./error/SupportLink";
import { makeHelpDeskURL } from "./util/helpDesk";

export interface AppContextProviderProps {
  children?: React.ReactNode;
}

// cache-only: the header already loads the viewer, so this only waits for it.
function ProjectSupportLinkProvider(props: {
  appID: string;
  children?: React.ReactNode;
}): ReactElement {
  const { appID, children } = props;
  const { viewer } = useViewerQuery({ fetchPolicy: "cache-only" });
  const href = makeHelpDeskURL({
    purpose: "other",
    projectID: appID,
    email: viewer?.email,
    fullName: viewer?.formattedName,
  });
  return (
    <SupportLinkHrefProvider value={href}>{children}</SupportLinkHrefProvider>
  );
}

export default function AppContextProvider(
  props: AppContextProviderProps
): ReactElement {
  const { children } = props;
  const { appID: appNodeID } = useParams() as { appID: string };

  // NOTE: check if appID actually exist in authorized app list
  const {
    effectiveAppConfig,
    viewer,
    loadError: error,
    isLoading: loading,
  } = useAppAndSecretConfigQuery(appNodeID);

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
    <AppContext.Provider key={appNodeID} value={appContextValue}>
      <ProjectSupportLinkProvider appID={appContextValue.appID}>
        {children}
      </ProjectSupportLinkProvider>
    </AppContext.Provider>
  );
}
