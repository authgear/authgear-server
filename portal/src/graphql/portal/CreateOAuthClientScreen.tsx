import React, { useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import produce from "immer";
import { FormattedMessage } from "@oursky/react-messageformat";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { PortalAPIAppConfig } from "../../types";

import styles from "./CreateOAuthClientScreen.module.scss";

interface CreateOAuthClientFormProps {
  effectiveAppConfig: PortalAPIAppConfig;
  rawAppConfig: PortalAPIAppConfig;
}

const CreateOAuthClientForm: React.FC<CreateOAuthClientFormProps> = function CreateOAuthClientForm(
  props: CreateOAuthClientFormProps
) {
  return (
    <form className={styles.form}>
      <span>TODO: to be implemented</span>
    </form>
  );
};

const CreateOAuthClientScreen: React.FC = function CreateOAuthClientScreen() {
  const { appID } = useParams();
  const { data, loading, error, refetch } = useAppConfigQuery(appID);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: "..",
        label: <FormattedMessage id="OAuthClientConfiguration.title" />,
      },
      {
        to: ".",
        label: <FormattedMessage id="CreateOAuthClientScreen.title" />,
      },
    ];
  }, []);

  const { rawAppConfig, effectiveAppConfig } = useMemo(() => {
    const appConfigNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      rawAppConfig: appConfigNode?.rawAppConfig ?? null,
      effectiveAppConfig: appConfigNode?.effectiveAppConfig ?? null,
    };
  }, [data]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (rawAppConfig == null || effectiveAppConfig == null) {
    return null;
  }

  return (
    <main className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <CreateOAuthClientForm
        effectiveAppConfig={effectiveAppConfig}
        rawAppConfig={rawAppConfig}
      />
    </main>
  );
};

export default CreateOAuthClientScreen;
