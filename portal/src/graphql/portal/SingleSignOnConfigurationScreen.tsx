import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import {
  Dialog,
  DialogFooter,
  Label,
  Spinner,
  SpinnerSize,
  Text,
} from "@fluentui/react";
import { Context as IntlContext, FormattedMessage } from "../../intl";
import {
  OAuthClientRow,
  OAuthClientRowHeader,
} from "./SingleSignOnConfigurationWidget";
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
  OAuthSSOProviderConfig,
  OAuthSSOProviderItemKey,
} from "../../types";
import styles from "./SingleSignOnConfigurationScreen.module.css";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { AppSecretKey, EffectiveSecretConfig } from "./globalTypes.generated";
import PrimaryButton from "../../PrimaryButton";
import cn from "classnames";
import ScreenContentHeader from "../../ScreenContentHeader";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  OAuthProviderFormModel,
  useOAuthProviderForm,
} from "../../hook/useOAuthProviderForm";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useLoadableView } from "../../hook/useLoadableView";

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
  onDeleteProvider: (k: OAuthSSOProviderItemKey, alias: string) => void;
  effectiveSecretConfig: EffectiveSecretConfig | undefined;
}

const SingleSignOnConfigurationContent: React.VFC<SingleSignOnConfigurationContentProps> =
  function SingleSignOnConfigurationContent(props) {
    const {
      oauthClientsMaximum,
      onDeleteProvider,
      form,
      effectiveSecretConfig,
    } = props;
    const { renderToString } = useContext(IntlContext);

    const limitReached = form.state.providers.length >= oauthClientsMaximum;

    const navigate = useNavigate();

    const onAddConnection = useCallback(() => {
      navigate("./add");
    }, [navigate]);

    const onEditConnection = useCallback(
      (provider: OAuthSSOProviderConfig) => {
        navigate(
          `./edit/${createOAuthSSOProviderItemKey(
            provider.type,
            provider.app_type
          )}/${provider.alias}`
        );
      },
      [navigate]
    );

    const onDeleteConnection = useCallback(
      (provider: OAuthSSOProviderConfig) => {
        onDeleteProvider(
          createOAuthSSOProviderItemKey(provider.type, provider.app_type),
          provider.alias
        );
      },
      [onDeleteProvider]
    );

    const providerKeysWithDuplications = useMemo(() => {
      const set = new Set<OAuthSSOProviderItemKey>();
      const keysWithDuplication = new Set<OAuthSSOProviderItemKey>();
      for (const p of form.state.providers) {
        const key = createOAuthSSOProviderItemKey(
          p.config.type,
          p.config.app_type
        );
        if (set.has(key)) {
          keysWithDuplication.add(key);
        }
        set.add(key);
      }
      return keysWithDuplication;
    }, [form.state.providers]);

    const providersWithDemoCredentials = useMemo(() => {
      return new Set(
        effectiveSecretConfig?.oauthSSOProviderDemoSecrets?.map((it) => it.type)
      );
    }, [effectiveSecretConfig?.oauthSSOProviderDemoSecrets]);

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
            {form.state.providers.length > 0 ? (
              form.state.providers.map((provider, idx) => (
                <React.Fragment
                  key={`${provider.config.type}/${provider.config.alias}`}
                >
                  {idx === 0 ? (
                    <OAuthClientRowHeader className={styles.contentHeader} />
                  ) : null}
                  <OAuthClientRow
                    className={styles.contentItem}
                    showAlias={providerKeysWithDuplications.has(
                      createOAuthSSOProviderItemKey(
                        provider.config.type,
                        provider.config.app_type
                      )
                    )}
                    providerConfig={provider.config}
                    providersWithDemoCredentials={providersWithDemoCredentials}
                    onEditClick={onEditConnection}
                    onDeleteClick={onDeleteConnection}
                  />
                </React.Fragment>
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
  const featureConfigQuery = useAppFeatureConfigQuery(appID);

  const effectiveSecretConfigQuery = useAppAndSecretConfigQuery(
    appID,
    secretVisitToken
  );

  const oauthClientsMaximum = useMemo(
    () =>
      featureConfigQuery.effectiveFeatureConfig?.identity?.oauth
        ?.maximum_providers ?? 99,
    [
      featureConfigQuery.effectiveFeatureConfig?.identity?.oauth
        ?.maximum_providers,
    ]
  );

  const [isDeleteDialogVisible, setIsDeleteDialogVisible] = useState(false);
  const onDisplayDeleteDialog = useCallback(
    (k: OAuthSSOProviderItemKey, alias: string) => {
      form.setState((state) => ({
        ...state,
        providers: state.providers.filter((p) => {
          if (
            createOAuthSSOProviderItemKey(p.config.type, p.config.app_type) ===
              k &&
            p.config.alias === alias
          ) {
            return false;
          }
          return true;
        }),
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

  return useLoadableView({
    loadables: [form, featureConfigQuery, effectiveSecretConfigQuery] as const,
    render: ([form, _, effectiveSecretConfigQuery]) => {
      return (
        <FormContainer form={form} hideFooterComponent={true}>
          <SingleSignOnConfigurationContent
            form={form}
            oauthClientsMaximum={oauthClientsMaximum}
            onDeleteProvider={onDisplayDeleteDialog}
            effectiveSecretConfig={
              effectiveSecretConfigQuery.effectiveSecretConfig
            }
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
    },
  });
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
