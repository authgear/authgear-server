import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Dialog,
  DialogFooter,
  Label,
  Spinner,
  SpinnerSize,
  Text,
} from "@fluentui/react";
import {
  Context as IntlContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import { OAuthClientRow } from "./SingleSignOnConfigurationWidget";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";
import FormContainer from "../../FormContainer";
import {
  createOAuthSSOProviderItemKey,
  OAuthSSOFeatureConfig,
  OAuthSSOProviderItemKey,
} from "../../types";
import styles from "./SingleSignOnConfigurationScreen.module.css";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { AppSecretKey } from "./globalTypes.generated";
import PrimaryButton from "../../PrimaryButton";
import cn from "classnames";
import ScreenContentHeader from "../../ScreenContentHeader";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  OAuthProviderFormModel,
  useOAuthProviderForm,
} from "../../hook/useOAuthProviderForm";

interface LocationState {
  isRevealSecrets: boolean;
}

function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isRevealSecrets != null
  );
}

interface SingleSignOnConfigurationContentProps {
  form: OAuthProviderFormModel;
  oauthClientsMaximum: number;
  onDeleteProvider: (k: OAuthSSOProviderItemKey) => void;
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig;
}

const SingleSignOnConfigurationContent: React.VFC<SingleSignOnConfigurationContentProps> =
  function SingleSignOnConfigurationContent(props) {
    const { oauthClientsMaximum, onDeleteProvider, form } = props;
    const { renderToString } = useContext(IntlContext);
    const [providers, setProviders] = useState(form.state.initialProvidersKey);

    const limitReached =
      form.state.initialProvidersKey.length >= oauthClientsMaximum;

    useEffect(() => {
      if (!form.isDirty) {
        setProviders(
          form.state.providers.map((p) =>
            createOAuthSSOProviderItemKey(p.config.type, p.config.app_type)
          )
        );
      }
    }, [form]);

    const navigate = useNavigate();

    const onAddConnection = useCallback(() => {
      navigate("./add");
    }, [navigate]);

    const onEditConnection = useCallback(
      (providerItemKey: OAuthSSOProviderItemKey) => {
        navigate(`./edit/${providerItemKey}`);
      },
      [navigate]
    );

    const onDeleteConnection = useCallback(
      (providerItemKey: OAuthSSOProviderItemKey) => {
        onDeleteProvider(providerItemKey);
      },
      [onDeleteProvider]
    );

    return (
      <ScreenContent
        layout="list"
        header={
          <ScreenContentHeader
            title={
              <ScreenTitle className={cn(styles.widget, styles.screenTitle)}>
                <span>
                  <FormattedMessage id="SingleSignOnConfigurationScreen.title" />
                </span>
                <PrimaryButton
                  text={renderToString(
                    "SingleSignOnConfigurationScreen.add-connection"
                  )}
                  iconProps={{ iconName: "Add" }}
                  onClick={onAddConnection}
                  disabled={limitReached}
                />
              </ScreenTitle>
            }
            description={
              <ScreenDescription className={styles.widget}>
                <Text>
                  <FormattedMessage id="SingleSignOnConfigurationScreen.description" />
                </Text>
                {oauthClientsMaximum < 99 ? (
                  <FeatureDisabledMessageBar
                    messageID="FeatureConfig.sso.maximum"
                    messageValues={{
                      maximum: oauthClientsMaximum,
                    }}
                  />
                ) : null}
              </ScreenDescription>
            }
          />
        }
      >
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          <div className={styles.content}>
            {providers.length > 0 ? (
              providers.map((providerItemKey) => (
                <OAuthClientRow
                  key={providerItemKey}
                  className={styles.contentItem}
                  providerItemKey={providerItemKey}
                  onEditClick={onEditConnection}
                  onDeleteClick={onDeleteConnection}
                />
              ))
            ) : (
              <div className={styles.emptyMessage}>
                <Label>
                  <FormattedMessage id="SingleSignOnConfigurationScreen.empty-message" />
                </Label>
                <PrimaryButton
                  text={renderToString(
                    "SingleSignOnConfigurationScreen.add-connection"
                  )}
                  onClick={onAddConnection}
                  disabled={limitReached}
                />
              </div>
            )}
          </div>
        </ShowOnlyIfSIWEIsDisabled>
      </ScreenContent>
    );
  };

const SingleSignOnConfigurationScreen1: React.VFC<{
  appID: string;
  secretVisitToken: string | null;
}> = function SingleSignOnConfigurationScreen1({ appID, secretVisitToken }) {
  const { renderToString } = useContext(IntlContext);
  const { themes } = useSystemConfig();
  const form = useOAuthProviderForm(appID, secretVisitToken);
  const featureConfig = useAppFeatureConfigQuery(appID);

  const oauthClientsMaximum = useMemo(
    () =>
      featureConfig.effectiveFeatureConfig?.identity?.oauth
        ?.maximum_providers ?? 99,
    [featureConfig.effectiveFeatureConfig?.identity?.oauth?.maximum_providers]
  );

  const [isDeleteDialogVisible, setIsDeleteDialogVisible] = useState(false);
  const onDisplayDeleteDialog = useCallback(
    (k: OAuthSSOProviderItemKey) => {
      form.setState((state) => ({
        ...state,
        providers: state.providers.filter(
          (p) =>
            createOAuthSSOProviderItemKey(p.config.type, p.config.app_type) !==
            k
        ),
      }));
      setIsDeleteDialogVisible(true);
    },
    [form]
  );
  const onDismissDeleteDialog = useCallback(() => {
    setIsDeleteDialogVisible(false);
    form.reset();
  }, [form]);

  const deleteConnection = useCallback(() => {
    form.save().then(
      () => {
        setIsDeleteDialogVisible(false);
      },
      () => {}
    );
  }, [form]);

  const deleteDialogContentProps = useMemo(() => {
    return {
      title: (
        <FormattedMessage id="SingleSignOnConfigurationScreen.delete-confirm-dialog.title" />
      ),
      subText: renderToString(
        "SingleSignOnConfigurationScreen.delete-confirm-dialog.description"
      ),
    };
  }, [renderToString]);

  if (form.isLoading || featureConfig.loading) {
    return <ShowLoading />;
  }

  if (form.loadError ?? featureConfig.error) {
    return (
      <ShowError
        error={form.loadError ?? featureConfig.error}
        onRetry={() => {
          form.reload();
          featureConfig.refetch().finally(() => {});
        }}
      />
    );
  }

  return (
    <FormContainer form={form} hideFooterComponent={true}>
      <SingleSignOnConfigurationContent
        form={form}
        oauthClientsMaximum={oauthClientsMaximum}
        onDeleteProvider={onDisplayDeleteDialog}
      />
      <Dialog
        hidden={!isDeleteDialogVisible}
        dialogContentProps={deleteDialogContentProps}
        modalProps={{ isBlocking: !form.isUpdating }}
        onDismiss={onDismissDeleteDialog}
      >
        <DialogFooter>
          <PrimaryButton
            text={
              <div className={styles.deleteButton}>
                {form.isUpdating ? (
                  <Spinner size={SpinnerSize.xSmall} ariaLive="assertive" />
                ) : null}
                <span>
                  <FormattedMessage id="SingleSignOnConfigurationScreen.delete-confirm-dialog.delete" />
                </span>
              </div>
            }
            theme={themes.destructive}
            disabled={form.isUpdating}
            onClick={deleteConnection}
          />
          <DefaultButton
            onClick={onDismissDeleteDialog}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    </FormContainer>
  );
};

const SECRETS = [AppSecretKey.OauthSsoProviderClientSecrets];

const SingleSignOnConfigurationScreen: React.VFC = () => {
  const { appID } = useParams() as { appID: string };
  const state = useLocationEffect(() => {
    // Pop the state
  });
  const [shouldRefreshToken] = useState<boolean>(() => {
    if (isLocationState(state) && state.isRevealSecrets) {
      return true;
    }
    return false;
  });

  const { token, error, loading, retry } = useAppSecretVisitToken(
    appID,
    SECRETS,
    shouldRefreshToken
  );

  if (error) {
    return <ShowError error={error} onRetry={retry} />;
  }

  if (token === undefined || loading) {
    return <ShowLoading />;
  }

  return (
    <SingleSignOnConfigurationScreen1 appID={appID} secretVisitToken={token} />
  );
};

export default SingleSignOnConfigurationScreen;
