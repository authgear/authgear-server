import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Flex, Text } from "@radix-ui/themes";
import { PlusIcon } from "@radix-ui/react-icons";
import { Context as IntlContext, FormattedMessage } from "../../intl";
import ExternalLink from "../../ExternalLink";
import {
  OAuthClientRow,
  OAuthClientRowHeader,
} from "./SingleSignOnConfigurationWidget";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
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
import cn from "classnames";
import {
  OAuthProviderFormModel,
  useOAuthProviderForm,
} from "../../hook/useOAuthProviderForm";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useLoadableView } from "../../hook/useLoadableView";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";

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

    const providersWithDemoCredentials = useMemo(() => {
      return new Set(
        effectiveSecretConfig?.oauthSSOProviderDemoSecrets?.map((it) => it.type)
      );
    }, [effectiveSecretConfig?.oauthSSOProviderDemoSecrets]);

    return (
      <ScreenContent layout="list">
        <div className={cn(styles.widget, styles.pageHeader)}>
          <div className={styles.pageTitleRow}>
            <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
              <FormattedMessage id="SingleSignOnConfigurationScreen.title" />
            </Text>
            {form.state.providers.length > 0 ? (
              <PrimaryButton
                size="2"
                disabled={limitReached}
                onClick={onAddConnection}
                text={
                  <Flex align="center" gap="2">
                    <PlusIcon width="1rem" height="1rem" />
                    <FormattedMessage id="SingleSignOnConfigurationScreen.add-connection" />
                  </Flex>
                }
              />
            ) : null}
          </div>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage
              id="SingleSignOnConfigurationScreen.description"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                docLink: (chunks: React.ReactNode) => (
                  <ExternalLink href="https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers">
                    {chunks}
                  </ExternalLink>
                ),
              }}
            />
          </Text>
          {oauthClientsMaximum < 99 ? (
            <FeatureDisabledMessageBar
              messageID="FeatureConfig.sso.maximum"
              messageValues={{
                maximum: oauthClientsMaximum,
              }}
            />
          ) : null}
        </div>
        <ShowOnlyIfSIWEIsDisabled>
          <div className={styles.content}>
            {form.state.providers.length > 0 ? (
              <div className={styles.list}>
                <div className={styles.tableWrapper}>
                  <div className={styles.table}>
                    <OAuthClientRowHeader />
                    {form.state.providers.map((provider) => (
                      <OAuthClientRow
                        key={`${provider.config.type}/${provider.config.alias}`}
                        providerConfig={provider.config}
                        providersWithDemoCredentials={
                          providersWithDemoCredentials
                        }
                        onEditClick={onEditConnection}
                        onDeleteClick={onDeleteConnection}
                      />
                    ))}
                  </div>
                </div>
              </div>
            ) : (
              <div className={styles.emptyMessage}>
                <Text as="p" size="4" weight="bold" className={styles.emptyTitle}>
                  <FormattedMessage id="SingleSignOnConfigurationScreen.empty-message" />
                </Text>
                <div className={styles.emptyButton}>
                  <PrimaryButton
                    size="2"
                    disabled={limitReached}
                    onClick={onAddConnection}
                    text={
                      <Flex align="center" gap="2">
                        <PlusIcon width="1rem" height="1rem" />
                        <FormattedMessage id="SingleSignOnConfigurationScreen.add-connection" />
                      </Flex>
                    }
                  />
                </div>
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
  const [pendingDelete, setPendingDelete] = useState<{
    k: OAuthSSOProviderItemKey;
    alias: string;
  } | null>(null);

  const onDisplayDeleteDialog = useCallback(
    (k: OAuthSSOProviderItemKey, alias: string) => {
      setPendingDelete({ k, alias });
      setIsDeleteDialogVisible(true);
    },
    []
  );
  const onDismissDeleteDialog = useCallback(() => {
    setIsDeleteDialogVisible(false);
    setPendingDelete(null);
  }, []);

  const onDeleteDialogOpenChange = useCallback(
    (open: boolean) => {
      if (!open) {
        onDismissDeleteDialog();
      }
    },
    [onDismissDeleteDialog]
  );

  const deleteConnection = useCallback(() => {
    if (pendingDelete == null) {
      return;
    }
    const { k, alias } = pendingDelete;
    const filteredState = {
      ...form.state,
      providers: form.state.providers.filter((p) => {
        if (
          createOAuthSSOProviderItemKey(p.config.type, p.config.app_type) ===
            k &&
          p.config.alias === alias
        ) {
          return false;
        }
        return true;
      }),
    };
    form.saveWithState(filteredState).then(
      () => {
        setIsDeleteDialogVisible(false);
        setPendingDelete(null);
      },
      () => {}
    );
  }, [form, pendingDelete]);

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
          <ConfirmationDialog
            open={isDeleteDialogVisible}
            onOpenChange={onDeleteDialogOpenChange}
            title={
              <FormattedMessage id="SingleSignOnConfigurationScreen.delete-confirm-dialog.title" />
            }
            description={renderToString(
              "SingleSignOnConfigurationScreen.delete-confirm-dialog.description"
            )}
            confirmText={
              <FormattedMessage id="SingleSignOnConfigurationScreen.delete-confirm-dialog.delete" />
            }
            cancelText={<FormattedMessage id="cancel" />}
            confirmColor="red"
            loading={form.isUpdating}
            onConfirm={deleteConnection}
            onCancel={onDismissDeleteDialog}
          />
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
