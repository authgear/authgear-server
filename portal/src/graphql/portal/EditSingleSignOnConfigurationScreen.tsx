import React, { useCallback, useEffect, useMemo, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import SingleSignOnConfigurationWidget, {
  useSingleSignOnConfigurationWidget,
} from "./SingleSignOnConfigurationWidget";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";
import FormContainer from "../../FormContainer";
import {
  createOAuthSSOProviderItemKey,
  OAuthSSOFeatureConfig,
  OAuthSSOProviderItemKey,
  oauthSSOProviderItemKeys,
} from "../../types";
import styles from "./SingleSignOnConfigurationScreen.module.css";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { AppSecretKey } from "./globalTypes.generated";
import { startReauthentication } from "./Authenticated";
import cn from "classnames";
import NavBreadcrumb from "../../NavBreadcrumb";
import ScreenContentHeader from "../../ScreenContentHeader";
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

interface OAuthClientItemProps {
  initialAlias: string;
  providerItemKey: OAuthSSOProviderItemKey;
  form: OAuthProviderFormModel;
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig;
}

const OAuthClientItem: React.VFC<OAuthClientItemProps> =
  function OAuthClientItem(props) {
    const { initialAlias, providerItemKey, form, oauthSSOFeatureConfig } =
      props;
    const widgetProps = useSingleSignOnConfigurationWidget(
      initialAlias,
      providerItemKey,
      form,
      oauthSSOFeatureConfig
    );
    return (
      <SingleSignOnConfigurationWidget
        className={styles.widget}
        {...widgetProps}
      />
    );
  };

interface EditSingleSignOnConfigurationContentProps {
  alias: string;
  form: OAuthProviderFormModel;
  providerItemKey: OAuthSSOProviderItemKey;
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig;
}

const EditSingleSignOnConfigurationContent: React.VFC<EditSingleSignOnConfigurationContentProps> =
  function EditSingleSignOnConfigurationContent(props) {
    const { alias, form, providerItemKey, oauthSSOFeatureConfig } = props;

    const navBreadcrumbItems = useMemo(() => {
      return [
        {
          to: "..",
          label: (
            <FormattedMessage id="SingleSignOnConfigurationScreen.title" />
          ),
        },
        {
          to: ".",
          label: (
            <FormattedMessage id="EditSingleSignOnConfigurationScreen.title" />
          ),
        },
      ];
    }, []);

    return (
      <ScreenContent
        header={
          <ScreenContentHeader
            title={
              <NavBreadcrumb
                className={cn(styles.widget, styles.breadcrumb)}
                items={navBreadcrumbItems}
              />
            }
          />
        }
      >
        <ShowOnlyIfSIWEIsDisabled>
          <OAuthClientItem
            initialAlias={alias}
            providerItemKey={providerItemKey}
            form={form}
            oauthSSOFeatureConfig={oauthSSOFeatureConfig}
          />
        </ShowOnlyIfSIWEIsDisabled>
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
  const config = useOAuthProviderForm(appID, secretVisitToken);
  const featureConfig = useAppFeatureConfigQuery(appID);
  const form = config;

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
    if (!isReadyToEdit) {
      onRevealSecrets();
    }
  }, [isReadyToEdit, onRevealSecrets]);

  if (!isReadyToEdit || form.isLoading || featureConfig.loading) {
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
    <FormContainer form={form} afterSave={onSaveSuccess}>
      <EditSingleSignOnConfigurationContent
        form={form}
        alias={alias}
        providerItemKey={providerItemKey}
        oauthSSOFeatureConfig={
          featureConfig.effectiveFeatureConfig?.identity?.oauth
        }
      />
    </FormContainer>
  );
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
      navigate("../");
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
