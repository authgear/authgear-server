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
import deepEqual from "deep-equal";
import { FormattedMessage } from "@oursky/react-messageformat";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ModifyOAuthClientForm, {
  getReducedClientConfig,
} from "./ModifyOAuthClientForm";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { genRandomHexadecimalString } from "../../util/random";
import { copyToClipboard } from "../../util/clipboard";
import { FormContext } from "../../error/FormContext";
import { useValidationError } from "../../error/useValidationError";

import styles from "./CreateOAuthClientScreen.module.scss";

interface CreateOAuthClientFormProps {
  rawAppConfig: PortalAPIAppConfig;
  resetForm: () => void;
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

const CreateOAuthClientForm: React.FC<CreateOAuthClientFormProps> = function CreateOAuthClientForm(
  props: CreateOAuthClientFormProps
) {
  const { rawAppConfig, resetForm } = props;
  const { appID } = useParams();
  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const initialClientConfig = useMemo(() => {
    return {
      name: undefined,
      client_id: genRandomHexadecimalString(),
      redirect_uris: [],
      grant_types: [
        "authorization_code",
        "refresh_token",
        "urn:authgear:params:oauth:grant-type:anonymous-request",
      ],
      response_types: ["code", "none"],
      access_token_lifetime_seconds: undefined,
      refresh_token_lifetime_seconds: undefined,
      post_logout_redirect_uris: undefined,
    };
  }, []);

  const [clientConfig, setClientConfig] = useState<OAuthClientConfig>(
    initialClientConfig
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(
      getReducedClientConfig(initialClientConfig),
      getReducedClientConfig(clientConfig)
    );
  }, [clientConfig, initialClientConfig]);

  const [submittedForm, setSubmittedForm] = useState(false);

  const [
    createClientSuccessDialogVisible,
    setCreateClientSuccessDialogVisible,
  ] = useState(false);

  const onClientConfigChange = useCallback(
    (newClientConfig: OAuthClientConfig) => {
      setClientConfig(newClientConfig);
    },
    []
  );

  const onCreateClientSuccess = useCallback(() => {
    setSubmittedForm(true);
    setCreateClientSuccessDialogVisible(true);
  }, []);

  const onFormSubmit = useCallback(
    (e: React.SyntheticEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();

      const newAppConfig = produce(rawAppConfig, (draftConfig) => {
        draftConfig.oauth = draftConfig.oauth ?? {};
        draftConfig.oauth.clients = draftConfig.oauth.clients ?? [];
        draftConfig.oauth.clients.push(createDraft(clientConfig));

        clearEmptyObject(draftConfig);
      });

      updateAppConfig(newAppConfig)
        .then((result) => {
          if (result != null) {
            onCreateClientSuccess();
          }
        })
        .catch(() => {});
    },
    [rawAppConfig, clientConfig, onCreateClientSuccess, updateAppConfig]
  );

  const {
    otherError,
    unhandledCauses,
    value: formContextValue,
  } = useValidationError(updateAppConfigError);

  return (
    <FormContext.Provider value={formContextValue}>
      <form className={styles.form} onSubmit={onFormSubmit} noValidate={true}>
        <NavigationBlockerDialog
          blockNavigation={!submittedForm && isFormModified}
        />
        <CreateClientSuccessDialog
          visible={createClientSuccessDialogVisible}
          clientId={clientConfig.client_id}
        />
        <ModifiedIndicatorPortal
          resetForm={resetForm}
          isModified={isFormModified}
        />
        {(unhandledCauses ?? []).length === 0 && otherError && (
          <ShowError error={otherError} />
        )}
        <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
        <ModifyOAuthClientForm
          className={styles.modifyClientForm}
          clientConfig={clientConfig}
          onClientConfigChange={onClientConfigChange}
        />
        <ButtonWithLoading
          type="submit"
          disabled={!isFormModified || submittedForm}
          labelId="create"
          loading={updatingAppConfig}
        />
      </form>
    </FormContext.Provider>
  );
};

const CreateOAuthClientScreen: React.FC = function CreateOAuthClientScreen() {
  const { appID } = useParams();
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
        to: "..",
        label: <FormattedMessage id="OAuthClientConfiguration.title" />,
      },
      {
        to: ".",
        label: <FormattedMessage id="CreateOAuthClientScreen.title" />,
      },
    ];
  }, []);

  const [remountIdentifier, setRemountIdentifier] = useState(0);
  const resetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
  }, []);

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
      <ModifiedIndicatorWrapper className={styles.wrapper}>
        <NavBreadcrumb items={navBreadcrumbItems} />
        <CreateOAuthClientForm
          key={remountIdentifier}
          rawAppConfig={rawAppConfig}
          resetForm={resetForm}
        />
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default CreateOAuthClientScreen;
