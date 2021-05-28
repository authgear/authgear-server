import React, { useCallback, useMemo } from "react";
import { useParams } from "react-router-dom";
import deepEqual from "deep-equal";
import produce, { createDraft } from "immer";
import { Label, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ModifyOAuthClientForm, {
  getReducedClientConfig,
} from "./ModifyOAuthClientForm";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";

import styles from "./EditOAuthClientScreen.module.scss";

interface FormState {
  clients: OAuthClientConfig[];
  editedClient: OAuthClientConfig | null;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    clients: config.oauth?.clients ?? [],
    editedClient: null,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.oauth ??= {};
    config.oauth.clients = currentState.clients.slice();

    const client = currentState.editedClient;
    if (client) {
      const index = config.oauth.clients.findIndex(
        (c) => c.client_id === client.client_id
      );
      if (
        index !== -1 &&
        !deepEqual(
          getReducedClientConfig(client),
          getReducedClientConfig(config.oauth.clients[index]),
          { strict: true }
        )
      ) {
        config.oauth.clients[index] = createDraft(client);
      }
    }
    clearEmptyObject(config);
  });
}

interface EditOAuthClientContentProps {
  form: AppConfigFormModel<FormState>;
  clientID: string;
}

const EditOAuthClientContent: React.FC<EditOAuthClientContentProps> =
  function EditOAuthClientContent(props) {
    const {
      clientID,
      form: { state, setState },
    } = props;

    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: "../..",
          label: (
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          ),
        },
        {
          to: ".",
          label: <FormattedMessage id="EditOAuthClientScreen.title" />,
        },
      ];
    }, []);

    const client =
      state.editedClient ?? state.clients.find((c) => c.client_id === clientID);

    const onClientConfigChange = useCallback(
      (editedClient: OAuthClientConfig) => {
        setState((state) => ({ ...state, editedClient }));
      },
      [setState]
    );

    if (client == null) {
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
      <div className={styles.root}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <Label>
          <FormattedMessage id="EditOAuthClientScreen.client-id" />
        </Label>
        <Text className={styles.clientIdField}>{client.client_id}</Text>
        <ModifyOAuthClientForm
          isCreation={false}
          clientConfig={client}
          onClientConfigChange={onClientConfigChange}
        />
      </div>
    );
  };

const EditOAuthClientScreen: React.FC = function EditOAuthClientScreen() {
  const { appID, clientID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <EditOAuthClientContent form={form} clientID={clientID} />
    </FormContainer>
  );
};

export default EditOAuthClientScreen;
