import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { FormattedMessage } from "../../intl";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import SingleSignOnConfigurationWidget, {
  useSingleSignOnConfigurationWidget,
} from "./SingleSignOnConfigurationWidget";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import Link from "../../Link";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";
import FormContainer from "../../FormContainer";
import {
  createOAuthSSOProviderItemKey,
  isOAuthSSOProvider,
  OAuthSSOFeatureConfig,
  OAuthSSOProviderItemKey,
  oauthSSOProviderItemKeys,
  parseOAuthSSOProviderItemKey,
} from "../../types";
import styles from "./EditSingleSignOnConfigurationScreen.module.css";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { AppSecretKey, EffectiveSecretConfig } from "./globalTypes.generated";
import { startReauthentication } from "./Authenticated";
import {
  OAuthProviderFormModel,
  useOAuthProviderForm,
} from "../../hook/useOAuthProviderForm";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useLoadableView } from "../../hook/useLoadableView";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";

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

interface OAuthClientItemProps {
  initialAlias: string;
  providerItemKey: OAuthSSOProviderItemKey;
  form: OAuthProviderFormModel;
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig;
  effectiveSecretConfig: EffectiveSecretConfig | undefined;
  publicOrigin: string;
}

const OAuthClientItem: React.VFC<OAuthClientItemProps> =
  function OAuthClientItem(props) {
    const {
      initialAlias,
      providerItemKey,
      form,
      oauthSSOFeatureConfig,
      effectiveSecretConfig,
      publicOrigin,
    } = props;
    const widgetProps = useSingleSignOnConfigurationWidget(
      initialAlias,
      providerItemKey,
      form,
      effectiveSecretConfig,
      oauthSSOFeatureConfig
    );
    return (
      <SingleSignOnConfigurationWidget
        className={styles.widget}
        {...widgetProps}
        publicOrigin={publicOrigin}
      />
    );
  };

interface EditSingleSignOnConfigurationContentProps {
  alias: string;
  form: OAuthProviderFormModel;
  providerItemKey: OAuthSSOProviderItemKey;
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig;
  effectiveSecretConfig: EffectiveSecretConfig | undefined;
  publicOrigin: string;
}

const EditSingleSignOnConfigurationContent: React.VFC<EditSingleSignOnConfigurationContentProps> =
  function EditSingleSignOnConfigurationContent(props) {
    const {
      alias,
      form,
      providerItemKey,
      oauthSSOFeatureConfig,
      effectiveSecretConfig,
      publicOrigin,
    } = props;
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);
    const { appID } = useParams() as { appID: string };

    return (
      <ScreenContent
        className={cn(isDirty ? styles.contentWithSaveBar : null)}
      >
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Link
            to={`/project/${appID}/configuration/authentication/external-oauth`}
            className={styles.backLink}
          >
            <ChevronLeftIcon className={styles.backLinkIcon} />
            <span>
              <FormattedMessage id="SingleSignOnConfigurationScreen.title" />
            </span>
          </Link>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="EditSingleSignOnConfigurationScreen.title" />
          </Text>
        </div>
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          <OAuthClientItem
            initialAlias={alias}
            providerItemKey={providerItemKey}
            form={form}
            oauthSSOFeatureConfig={oauthSSOFeatureConfig}
            effectiveSecretConfig={effectiveSecretConfig}
            publicOrigin={publicOrigin}
          />
        </ShowOnlyIfSIWEIsDisabled>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

const EditSingleSignOnConfigurationScreen1: React.VFC<{
  appID: string;
  alias: string;
  providerItemKey: OAuthSSOProviderItemKey;
  secretVisitToken: string | null;
}> = function EditSingleSignOnConfigurationScreen1({
  appID,
  alias,
  providerItemKey,
  secretVisitToken,
}) {
  const form = useOAuthProviderForm(appID, secretVisitToken);
  const featureConfigQuery = useAppFeatureConfigQuery(appID);
  const effectiveSecretConfigQuery = useAppAndSecretConfigQuery(
    appID,
    secretVisitToken
  );

  const isReadyToEdit = useMemo(() => {
    const isSecretPresent =
      form.state.providers.filter(
        (p) =>
          createOAuthSSOProviderItemKey(p.config.type, p.config.app_type) ===
            providerItemKey &&
          p.secret.originalAlias != null &&
          p.secret.newClientSecret != null
      ).length !== 0;
    const isNoExistingSecret =
      form.state.providers.filter((p) => p.secret.originalAlias != null)
        .length === 0;
    return isSecretPresent || isNoExistingSecret;
  }, [form.state.providers, providerItemKey]);

  const navigate = useNavigate();

  // After switching projects, the alias in the URL may not exist in the new
  // project. Redirect to the provider list instead of showing an edit form
  // that would silently add a new provider.
  const providerMissing = useMemo(() => {
    if (form.isLoading || form.loadError != null) {
      return false;
    }
    const [providerType, appType] =
      parseOAuthSSOProviderItemKey(providerItemKey);
    return !form.state.providers.some((p) =>
      isOAuthSSOProvider(p.config, providerType, alias, appType)
    );
  }, [
    form.isLoading,
    form.loadError,
    form.state.providers,
    providerItemKey,
    alias,
  ]);

  useEffect(() => {
    if (providerMissing) {
      navigate("../", { replace: true });
    }
  }, [providerMissing, navigate]);

  const onSaveSuccess = useCallback(() => {
    navigate("../");
  }, [navigate]);

  const onRevealSecrets = useCallback(() => {
    const locationState: LocationState = {
      isRevealSecrets: true,
    };

    startReauthentication(navigate, locationState).catch((e) => {
      // Normally there should not be any error.
      console.error(e);
    });
  }, [navigate]);

  useEffect(() => {
    if (providerMissing) {
      return;
    }
    if (!isReadyToEdit) {
      onRevealSecrets();
    }
  }, [providerMissing, isReadyToEdit, onRevealSecrets]);

  return useLoadableView({
    loadables: [form, featureConfigQuery, effectiveSecretConfigQuery] as const,
    isLoading: !isReadyToEdit,
    render: ([form, featureConfigQuery, effectiveSecretConfigQuery]) => {
      const publicOrigin =
        effectiveSecretConfigQuery.effectiveAppConfig?.http?.public_origin ??
        "";
      return (
        <FormContainer form={form} afterSave={onSaveSuccess} hideFooterComponent={true}>
          <EditSingleSignOnConfigurationContent
            form={form}
            alias={alias}
            providerItemKey={providerItemKey}
            oauthSSOFeatureConfig={
              featureConfigQuery.effectiveFeatureConfig?.identity?.oauth
            }
            effectiveSecretConfig={
              effectiveSecretConfigQuery.effectiveSecretConfig
            }
            publicOrigin={publicOrigin}
          />
        </FormContainer>
      );
    },
  });
};

const SECRETS = [AppSecretKey.OauthSsoProviderClientSecrets];

const EditSingleSignOnConfigurationScreen: React.VFC = () => {
  const navigate = useNavigate();
  const {
    appID,
    provider: rawProviderItemKey,
    alias,
  } = useParams() as {
    appID: string;
    provider: string;
    alias: string;
  };

  const providerItemKey = useMemo(() => {
    return oauthSSOProviderItemKeys.includes(
      rawProviderItemKey as OAuthSSOProviderItemKey
    )
      ? (rawProviderItemKey as OAuthSSOProviderItemKey)
      : undefined;
  }, [rawProviderItemKey]);

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

  useEffect(() => {
    if (providerItemKey == null) {
      navigate("../", { replace: true });
    }
  }, [providerItemKey, navigate]);

  if (error) {
    return <ShowError error={error} onRetry={retry} />;
  }

  if (providerItemKey == null || token === undefined || loading) {
    return <ShowLoading />;
  }

  return (
    <EditSingleSignOnConfigurationScreen1
      appID={appID}
      alias={alias}
      providerItemKey={providerItemKey}
      secretVisitToken={token}
    />
  );
};

export default EditSingleSignOnConfigurationScreen;
