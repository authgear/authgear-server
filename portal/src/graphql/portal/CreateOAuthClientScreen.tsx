import React, { useCallback, useEffect, useMemo, useState } from "react";
import {
  Callout,
  Dialog,
  DialogFooter,
  DirectionalHint,
  IconButton,
  Label,
  PrimaryButton,
  Text,
} from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import produce, { createDraft } from "immer";
import { FormattedMessage } from "@oursky/react-messageformat";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ModifyOAuthClientForm, {
  getReducedClientConfig,
} from "./ModifyOAuthClientForm";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { genRandomHexadecimalString } from "../../util/random";
import { copyToClipboard } from "../../util/clipboard";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";

import styles from "./CreateOAuthClientScreen.module.scss";
import deepEqual from "deep-equal";

interface FormState {
  clients: OAuthClientConfig[];
  newClient: OAuthClientConfig;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    clients: config.oauth?.clients ?? [],
    newClient: {
      name: undefined,
      client_id: genRandomHexadecimalString(),
      redirect_uris: [],
      grant_types: ["authorization_code", "refresh_token"],
      response_types: ["code", "none"],
      access_token_lifetime_seconds: undefined,
      refresh_token_lifetime_seconds: undefined,
      post_logout_redirect_uris: undefined,
      issue_jwt_access_token: undefined,
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.oauth ??= {};
    config.oauth.clients = currentState.clients.slice();
    const isDirty = !deepEqual(
      getReducedClientConfig(initialState.newClient),
      getReducedClientConfig(currentState.newClient),
      { strict: true }
    );
    if (isDirty) {
      config.oauth.clients.push(createDraft(currentState.newClient));
    }
    clearEmptyObject(config);
  });
}

interface CreateClientSuccessDialogProps {
  visible: boolean;
  clientId: string;
}

const CALLOUT_VISIBLE_DURATION = 3000;

const CreateClientSuccessDialog: React.FC<CreateClientSuccessDialogProps> = function CreateClientSuccessDialog(
  props: CreateClientSuccessDialogProps
) {
  const { visible, clientId } = props;
  const navigate = useNavigate();

  const [isCalloutVisible, setIsCalloutVisible] = useState(false);
  const [calloutActiveCount, setCalloutActiveCount] = useState(0);

  useEffect(() => {
    if (calloutActiveCount === 0) {
      // consistent return type in arrow function
      return () => {};
    }

    setIsCalloutVisible(true);
    const handle = setTimeout(
      () => setIsCalloutVisible(false),
      CALLOUT_VISIBLE_DURATION
    );
    return () => {
      // clear previous timeout when count is updated
      clearTimeout(handle);
    };
  }, [calloutActiveCount]);

  const onConfirmCreateClientSuccess = useCallback(() => {
    navigate("../");
  }, [navigate]);

  const onCopyClick = useCallback(() => {
    copyToClipboard(clientId);
    setCalloutActiveCount((c) => c + 1);
  }, [clientId]);

  return (
    <Dialog
      hidden={!visible}
      title={
        <FormattedMessage id="CreateOAuthClientScreen.success-dialog.title" />
      }
    >
      <Label>
        <FormattedMessage id="CreateOAuthClientScreen.success-dialog.client-id-label" />
      </Label>
      <div className={styles.dialogClientId}>
        <Text>{clientId}</Text>
        <IconButton
          onClick={onCopyClick}
          className={styles.dialogCopyIcon}
          iconProps={{ iconName: "Copy" }}
        />
      </div>
      {isCalloutVisible && (
        <Callout
          className={styles.copyButtonCallout}
          target={`.${styles.dialogCopyIcon}`}
          directionalHint={DirectionalHint.bottomLeftEdge}
          hideOverflow={true}
        >
          <Text>
            <FormattedMessage id="CreateOAuthClientScreen.success-dialog.copied" />
          </Text>
        </Callout>
      )}
      <DialogFooter>
        <PrimaryButton onClick={onConfirmCreateClientSuccess}>
          <FormattedMessage id="done" />
        </PrimaryButton>
      </DialogFooter>
    </Dialog>
  );
};

interface CreateOAuthClientContentProps {
  form: AppConfigFormModel<FormState>;
}

const CreateOAuthClientContent: React.FC<CreateOAuthClientContentProps> = function CreateOAuthClientContent(
  props
) {
  const { state, setState } = props.form;

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: "..",
        label: <FormattedMessage id="ApplicationsConfigurationScreen.title" />,
      },
      {
        to: ".",
        label: <FormattedMessage id="CreateOAuthClientScreen.title" />,
      },
    ];
  }, []);

  const [clientId] = useState(state.newClient.client_id);
  const client =
    state.clients.find((c) => c.client_id === clientId) ?? state.newClient;

  const onClientConfigChange = useCallback(
    (newClient: OAuthClientConfig) => {
      setState((state) => ({ ...state, newClient }));
    },
    [setState]
  );

  const isSuccessDialogVisible = state.clients.some(
    (c) => c.client_id === clientId
  );

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <ModifyOAuthClientForm
        clientConfig={client}
        onClientConfigChange={onClientConfigChange}
      />
      <CreateClientSuccessDialog
        visible={isSuccessDialogVisible}
        clientId={clientId}
      />
    </div>
  );
};

const CreateOAuthClientScreen: React.FC = function CreateOAuthClientScreen() {
  const { appID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <CreateOAuthClientContent form={form} />
    </FormContainer>
  );
};

export default CreateOAuthClientScreen;
