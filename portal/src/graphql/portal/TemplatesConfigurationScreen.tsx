import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
// import produce from "immer";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { ModifiedIndicatorWrapper } from "../../ModifiedIndicatorPortal";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import TemplateLocaleManagement from "./TemplateLocaleManagement";
import ForgotPasswordTemplatesSettings from "./ForgotPasswordTemplatesSettings";
import PasswordlessAuthenticatorTemplatesSettings from "./PasswordlessAuthenticatorTemplatesSettings";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useAppTemplatesQuery } from "./query/appTemplatesQuery";
import { useTemplateLocaleQuery } from "./query/templateLocaleQuery";
import { useUpdateAppTemplatesMutation } from "./mutations/updateAppTemplatesMutation";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { PortalAPIAppConfig } from "../../types";
import { usePivotNavigation } from "../../hook/usePivot";
import {
  DEFAULT_TEMPLATE_LOCALE,
  TemplateLocale,
  ALL_TEMPLATE_PATHS,
} from "../../templates";

import styles from "./TemplatesConfigurationScreen.module.scss";

interface TemplatesConfigurationProps {
  effectiveAppConfig: PortalAPIAppConfig;
  rawAppConfig: PortalAPIAppConfig;
  initialTemplateLocales: TemplateLocale[];
  onResetForm: () => void;
}

const FORGOT_PASSWORD_PIVOT_KEY = "forgot_password";
const PASSWORDLESS_AUTHENTICATOR_PIVOT_KEY = "passwordless_authenticator";

const TemplatesConfiguration: React.FC<TemplatesConfigurationProps> = function TemplatesConfiguration(
  props: TemplatesConfigurationProps
) {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const { effectiveAppConfig, initialTemplateLocales, onResetForm } = props;

  const initialDefaultTemplateLocale = useMemo(() => {
    return (
      effectiveAppConfig.localization?.fallback_language ??
      DEFAULT_TEMPLATE_LOCALE
    );
  }, [effectiveAppConfig]);

  const [defaultTemplateLocale, setDefaultTemplateLocale] = useState<
    TemplateLocale
  >(initialDefaultTemplateLocale);

  const [templateLocale, setTemplateLocale] = useState<TemplateLocale>(
    defaultTemplateLocale
  );

  const [templateLocales, setTemplateLocales] = useState<TemplateLocale[]>(
    initialTemplateLocales
  );

  const {
    templates,
    loading: loadingTemplates,
    error: loadTemplatesError,
    refetch: refetchTemplates,
  } = useAppTemplatesQuery(appID, templateLocale, ...ALL_TEMPLATE_PATHS);

  const onChangeTemplateLocales = useCallback(
    (locales: TemplateLocale[]) => {
      // Reset templateLocale to default if the selected one was removed.
      const idx = locales.findIndex((item) => item === templateLocale);
      if (idx < 0) {
        setTemplateLocale(defaultTemplateLocale);
      }

      setTemplateLocales(locales);
    },
    [templateLocale, defaultTemplateLocale]
  );

  const {
    updateAppTemplates,
    loading: updatingTemplates,
    error: updateTemplatesError,
    resetError: resetUpdateTemplatesError,
  } = useUpdateAppTemplatesMutation(appID);

  const {
    // updateAppConfig,
    // loading: updatingAppConfig,
    error: updateAppConfigError,
    resetError: resetUpdateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  // FIXME: Unify save.
  // const saveDefaultTemplateLocale = useCallback(
  //   (defaultTemplateLocale: TemplateLocale) => {
  //     if (rawAppConfig == null) {
  //       return;
  //     }
  //     const newAppConfig = produce(rawAppConfig, (draftConfig) => {
  //       draftConfig.localization = draftConfig.localization ?? {};
  //       draftConfig.localization.fallback_language = defaultTemplateLocale;
  //     });

  //     updateAppConfig(newAppConfig).catch(() => {});
  //   },
  //   [rawAppConfig, updateAppConfig]
  // );

  const resetError = useCallback(() => {
    resetUpdateTemplatesError();
    resetUpdateAppConfigError();
  }, [resetUpdateTemplatesError, resetUpdateAppConfigError]);

  const resetForm = useCallback(() => {
    onResetForm();
  }, [onResetForm]);

  const { selectedKey, onLinkClick } = usePivotNavigation(
    [FORGOT_PASSWORD_PIVOT_KEY, PASSWORDLESS_AUTHENTICATOR_PIVOT_KEY],
    resetError
  );

  return (
    <main
      className={cn(styles.root, {
        [styles.loading]: updatingTemplates,
      })}
    >
      {loadTemplatesError && (
        <ShowError error={loadTemplatesError} onRetry={refetchTemplates} />
      )}
      {updateTemplatesError && <ShowError error={updateTemplatesError} />}
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <ModifiedIndicatorWrapper className={styles.screen}>
        <Text className={styles.screenHeaderText} as="h1">
          <FormattedMessage id="TemplatesConfigurationScreen.title" />
        </Text>
        <TemplateLocaleManagement
          templateLocales={templateLocales}
          onChangeTemplateLocales={onChangeTemplateLocales}
          templateLocale={templateLocale}
          defaultTemplateLocale={defaultTemplateLocale}
          onSelectTemplateLocale={setTemplateLocale}
          onSelectDefaultTemplateLocale={setDefaultTemplateLocale}
        />
        {loadingTemplates && <ShowLoading />}
        <Pivot onLinkClick={onLinkClick} selectedKey={selectedKey}>
          <PivotItem
            headerText={renderToString(
              "TemplatesConfigurationScreen.forgot-password.title"
            )}
            itemKey={FORGOT_PASSWORD_PIVOT_KEY}
          >
            <ForgotPasswordTemplatesSettings
              templates={templates}
              templateLocale={templateLocale}
              updateTemplates={updateAppTemplates}
              updatingTemplates={updatingTemplates}
              resetForm={resetForm}
            />
          </PivotItem>
          <PivotItem
            headerText={renderToString(
              "TemplatesConfigurationScreen.passwordless-authenticator.title"
            )}
            itemKey={PASSWORDLESS_AUTHENTICATOR_PIVOT_KEY}
          >
            <PasswordlessAuthenticatorTemplatesSettings
              templates={templates}
              templateLocale={templateLocale}
              updateTemplates={updateAppTemplates}
              updatingTemplates={updatingTemplates}
              resetForm={resetForm}
            />
          </PivotItem>
        </Pivot>
      </ModifiedIndicatorWrapper>
    </main>
  );
};

const TemplatesConfigurationScreen: React.FC = function TemplatesConfigurationScreen() {
  const { appID } = useParams();
  const {
    effectiveAppConfig,
    rawAppConfig,
    loading: loadingAppConfig,
    error: loadAppConfigError,
    refetch: refetchAppConfig,
  } = useAppConfigQuery(appID);

  const {
    templateLocales: initialTemplateLocales,
    loading: loadingTemplates,
    error: loadTemplateLocalesError,
    refetch: refetchTemplateLocales,
  } = useTemplateLocaleQuery(appID);

  const [remountIdentifier, setRemountIdentifier] = useState(0);

  const onResetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
  }, []);

  if (loadingAppConfig || loadingTemplates) {
    return <ShowLoading />;
  }

  if (loadAppConfigError) {
    return <ShowError error={loadAppConfigError} onRetry={refetchAppConfig} />;
  }

  if (loadTemplateLocalesError) {
    return (
      <ShowError
        error={loadTemplateLocalesError}
        onRetry={refetchTemplateLocales}
      />
    );
  }

  return (
    <TemplatesConfiguration
      key={remountIdentifier}
      effectiveAppConfig={effectiveAppConfig!}
      rawAppConfig={rawAppConfig!}
      initialTemplateLocales={initialTemplateLocales}
      onResetForm={onResetForm}
    />
  );
};

export default TemplatesConfigurationScreen;
