import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useNavigate, useParams } from "react-router-dom";
import deepEqual from "deep-equal";
import produce from "immer";
import { Text, TextField } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ModifyOAuthClientForm from "./ModifyOAuthClientForm";
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
  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const navigate = useNavigate();

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
  const [submittedForm, setSubmittedForm] = useState(false);

  const onClientConfigChange = useCallback(
    (newClientConfig: OAuthClientConfig) => {
      setClientConfig(newClientConfig);
    },
    []
  );

  const onSaveClick = useCallback(() => {
    const newAppConfig = produce(rawAppConfig, (draftConfig) => {
      const clients = draftConfig.oauth!.clients!;
      const clientConfigIndex = clients.findIndex(
        (client) => client.client_id === clientConfig.client_id
      );
      clients[clientConfigIndex] = clientConfig;

      clearEmptyObject(draftConfig);
    });

    updateAppConfig(newAppConfig)
      .then((result) => {
        if (result != null) {
          setSubmittedForm(true);
        }
      })
      .catch(() => {});
  }, [clientConfig, updateAppConfig, rawAppConfig]);

  useEffect(() => {
    if (submittedForm) {
      navigate("../../");
    }
  }, [submittedForm, navigate]);

  const isFormModified = useMemo(() => {
    return !deepEqual(clientConfig, initialClientConfig);
  }, [clientConfig, initialClientConfig]);

  return (
    <form className={styles.form}>
      <NavigationBlockerDialog
        blockNavigation={!submittedForm && isFormModified}
      />
      <TextField
        className={styles.clientIdField}
        readOnly={true}
        value={clientConfig.client_id}
        label={renderToString("EditOAuthClientScreen.client-id")}
      />
      <ModifyOAuthClientForm
        className={styles.modifyClientForm}
        clientConfig={clientConfig}
        onClientConfigChange={onClientConfigChange}
        updateAppConfigError={updateAppConfigError}
      />
      <ButtonWithLoading
        onClick={onSaveClick}
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
  const { data, loading, error, refetch } = useAppConfigQuery(appID);

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

  const { rawAppConfig, effectiveAppConfig } = useMemo(() => {
    const appConfigNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      rawAppConfig: appConfigNode?.rawAppConfig ?? null,
      effectiveAppConfig: appConfigNode?.effectiveAppConfig ?? null,
    };
  }, [data]);

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
