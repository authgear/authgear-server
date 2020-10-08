import React, { useCallback, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import deepEqual from "deep-equal";
import produce from "immer";
import { Label, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ModifyOAuthClientForm, {
  getReducedClientConfig,
} from "./ModifyOAuthClientForm";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";

import styles from "./EditOAuthClientScreen.module.scss";

interface EditOAuthClientFormProps {
  clientConfig: OAuthClientConfig;
  rawAppConfig: PortalAPIAppConfig;
}

const EditOAuthClientForm: React.FC<EditOAuthClientFormProps> = function EditOAuthClientForm(
  props: EditOAuthClientFormProps
) {
  const { clientConfig: clientConfigProps, rawAppConfig } = props;
  const { appID } = useParams();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const initialClientConfig = useMemo(() => {
    return {
      ...clientConfigProps,
      post_logout_redirect_uris:
        (clientConfigProps.post_logout_redirect_uris ?? []).length > 0
          ? clientConfigProps.post_logout_redirect_uris
          : undefined,
    };
  }, [clientConfigProps]);

  const [clientConfig, setClientConfig] = useState<OAuthClientConfig>(
    initialClientConfig
  );

  const onClientConfigChange = useCallback(
    (newClientConfig: OAuthClientConfig) => {
      setClientConfig(newClientConfig);
    },
    []
  );

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      const newAppConfig = produce(rawAppConfig, (draftConfig) => {
        const clients = draftConfig.oauth!.clients!;
        const clientConfigIndex = clients.findIndex(
          (client) => client.client_id === clientConfig.client_id
        );
        clients[clientConfigIndex] = clientConfig;

        clearEmptyObject(draftConfig);
      });

      updateAppConfig(newAppConfig).catch(() => {});
    },
    [clientConfig, updateAppConfig, rawAppConfig]
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(
      getReducedClientConfig(clientConfig),
      getReducedClientConfig(initialClientConfig)
    );
  }, [clientConfig, initialClientConfig]);

  return (
    <form className={styles.form} onSubmit={onFormSubmit}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <Label>
        <FormattedMessage id="EditOAuthClientScreen.client-id" />
      </Label>
      <Text className={styles.clientIdField}>{clientConfig.client_id}</Text>

      <ModifyOAuthClientForm
        className={styles.modifyClientForm}
        clientConfig={clientConfig}
        onClientConfigChange={onClientConfigChange}
        updateAppConfigError={updateAppConfigError}
      />
      <ButtonWithLoading
        type="submit"
        disabled={!isFormModified}
        labelId="save"
        loading={updatingAppConfig}
        loadingLabelId="saving"
      />
    </form>
  );
};

const EditOAuthClientScreen: React.FC = function EditOAuthClientScreen() {
  const { appID, clientID } = useParams();
  const {
    rawAppConfig,
    effectiveAppConfig,
    loading,
    error,
    refetch,
  } = useAppConfigQuery(appID);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: "../../",
        label: <FormattedMessage id="OAuthClientConfiguration.title" />,
      },
      {
        to: ".",
        label: <FormattedMessage id="EditOAuthClientScreen.title" />,
      },
    ];
  }, []);

  const clientConfig = useMemo(() => {
    const clients = effectiveAppConfig?.oauth?.clients ?? [];
    return clients.find((client) => client.client_id === clientID);
  }, [effectiveAppConfig, clientID]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  if (rawAppConfig == null || effectiveAppConfig == null) {
    return null;
  }

  if (clientConfig == null) {
    return (
      <Text>
        <FormattedMessage
          id="EditOAuthClientScreen.client-not-found"
          values={{ clientID }}
        />
      </Text>
    );
  }

  return (
    <main className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <EditOAuthClientForm
        clientConfig={clientConfig}
        rawAppConfig={rawAppConfig}
      />
    </main>
  );
};

export default EditOAuthClientScreen;
