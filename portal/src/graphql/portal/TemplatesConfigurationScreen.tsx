import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  ModifiedIndicatorWrapper,
  ModifiedIndicatorPortal,
} from "../../ModifiedIndicatorPortal";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import TemplateLocaleManagement from "./TemplateLocaleManagement";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useAppTemplatesQuery, Template } from "./query/appTemplatesQuery";
import { useTemplateLocaleQuery } from "./query/templateLocaleQuery";
import { useUpdateAppTemplatesMutation } from "./mutations/updateAppTemplatesMutation";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { PortalAPIAppConfig } from "../../types";
import { usePivotNavigation } from "../../hook/usePivot";
import {
  DEFAULT_TEMPLATE_LOCALE,
  TemplateLocale,
  ALL_TEMPLATE_PATHS,
  forgotPasswordEmailHtmlPath,
  forgotPasswordEmailTextPath,
  forgotPasswordSmsTextPath,
  setupPrimaryOobEmailHtmlPath,
  setupPrimaryOobEmailTextPath,
  setupPrimaryOobSmsTextPath,
  authenticatePrimaryOobEmailHtmlPath,
  authenticatePrimaryOobEmailTextPath,
  authenticatePrimaryOobSmsTextPath,
  getLocalizedTemplatePath,
} from "../../templates";
import { ResourcePath } from "../../util/stringTemplate";
import { generateUpdates } from "./templates";

import styles from "./TemplatesConfigurationScreen.module.scss";

interface TemplatesConfigurationProps {
  rawAppConfig: PortalAPIAppConfig;
  initialTemplates: Record<string, Template | undefined>;
  initialTemplateLocales: TemplateLocale[];
  initialDefaultTemplateLocale: TemplateLocale;
  defaultTemplateLocale: TemplateLocale;
  templateLocale: TemplateLocale;
  setDefaultTemplateLocale: (locale: TemplateLocale) => void;
  setTemplateLocale: (locale: TemplateLocale) => void;
  onResetForm: () => void;
}

const FORGOT_PASSWORD_PIVOT_KEY = "forgot_password";
const PASSWORDLESS_AUTHENTICATOR_PIVOT_KEY = "passwordless_authenticator";

const TemplatesConfiguration: React.FC<TemplatesConfigurationProps> = function TemplatesConfiguration(
  props: TemplatesConfigurationProps
) {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const {
    initialTemplates,
    initialTemplateLocales,
    initialDefaultTemplateLocale,
    defaultTemplateLocale,
    setDefaultTemplateLocale,
    templateLocale,
    setTemplateLocale,
    onResetForm,
  } = props;

  const [templateLocales, setTemplateLocales] = useState<TemplateLocale[]>(
    initialTemplateLocales
  );

  const onChangeTemplateLocales = useCallback(
    (locales: TemplateLocale[]) => {
      // Reset templateLocale to default if the selected one was removed.
      const idx = locales.findIndex((item) => item === templateLocale);
      if (idx < 0) {
        setTemplateLocale(defaultTemplateLocale);
      }

      setTemplateLocales(locales);
    },
    [templateLocale, defaultTemplateLocale, setTemplateLocale]
  );

  const [templates, setTemplates] = useState(initialTemplates);

  const updates = useMemo(() => {
    return generateUpdates(
      initialTemplateLocales,
      initialTemplates,
      templateLocales,
      templates
    );
  }, [initialTemplateLocales, initialTemplates, templateLocales, templates]);

  const { invalidAdditionLocales, invalidEditionLocales } = updates;

  const invalidTemplateLocales = useMemo(() => {
    return invalidAdditionLocales.concat(invalidEditionLocales);
  }, [invalidAdditionLocales, invalidEditionLocales]);

  const {
    // updateAppTemplates,
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

  const isModified =
    initialDefaultTemplateLocale !== defaultTemplateLocale ||
    updates.isModified;

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

  const getValue = useCallback(
    (resourcePath: ResourcePath<"locale">) => {
      const path = getLocalizedTemplatePath(templateLocale, resourcePath);
      const template = templates[path];
      return template?.value ?? "";
    },
    [templates, templateLocale]
  );

  const getOnChange = useCallback(
    (resourcePath: ResourcePath<"locale">) => {
      return (_e: unknown, value?: string) => {
        if (value != null) {
          const path = getLocalizedTemplatePath(templateLocale, resourcePath);
          setTemplates((prev) => {
            let template = prev[path];

            if (template == null) {
              template = {
                resourcePath,
                path: path,
                locale: templateLocale,
                value,
              };
            } else {
              template = {
                ...template,
                value,
              };
            }

            return {
              ...prev,
              [path]: template,
            };
          });
        }
      };
    },
    [templateLocale]
  );

  const sectionsForgotPassword: EditTemplatesWidgetSection[] = [
    {
      key: "email",
      title: <FormattedMessage id="EditTemplatesWidget.email" />,
      items: [
        {
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(forgotPasswordEmailHtmlPath),
          onChange: getOnChange(forgotPasswordEmailHtmlPath),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(forgotPasswordEmailTextPath),
          onChange: getOnChange(forgotPasswordEmailTextPath),
        },
      ],
    },
    {
      key: "sms",
      title: <FormattedMessage id="EditTemplatesWidget.sms" />,
      items: [
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(forgotPasswordSmsTextPath),
          onChange: getOnChange(forgotPasswordSmsTextPath),
        },
      ],
    },
  ];

  const sectionsPasswordless: EditTemplatesWidgetSection[] = [
    {
      key: "setup",
      title: (
        <FormattedMessage id="EditTemplatesWidget.passwordless.setup.title" />
      ),
      items: [
        {
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(setupPrimaryOobEmailHtmlPath),
          onChange: getOnChange(setupPrimaryOobEmailHtmlPath),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(setupPrimaryOobEmailTextPath),
          onChange: getOnChange(setupPrimaryOobEmailTextPath),
        },
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(setupPrimaryOobSmsTextPath),
          onChange: getOnChange(setupPrimaryOobSmsTextPath),
        },
      ],
    },
    {
      key: "login",
      title: (
        <FormattedMessage id="EditTemplatesWidget.passwordless.login.title" />
      ),
      items: [
        {
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(authenticatePrimaryOobEmailHtmlPath),
          onChange: getOnChange(authenticatePrimaryOobEmailHtmlPath),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(authenticatePrimaryOobEmailTextPath),
          onChange: getOnChange(authenticatePrimaryOobEmailTextPath),
        },
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(authenticatePrimaryOobSmsTextPath),
          onChange: getOnChange(authenticatePrimaryOobSmsTextPath),
        },
      ],
    },
  ];

  return (
    <main
      className={cn(styles.root, {
        [styles.loading]: updatingTemplates,
      })}
    >
      {updateTemplatesError && <ShowError error={updateTemplatesError} />}
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <ModifiedIndicatorWrapper className={styles.screen}>
        <ModifiedIndicatorPortal
          resetForm={resetForm}
          isModified={isModified}
        />
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
          invalidTemplateLocales={invalidTemplateLocales}
        />
        <Pivot
          key={
            /* If we do not remount the pivot, we will have stale onChange callback being fired */
            templateLocale
          }
          onLinkClick={onLinkClick}
          selectedKey={selectedKey}
        >
          <PivotItem
            headerText={renderToString(
              "TemplatesConfigurationScreen.forgot-password.title"
            )}
            itemKey={FORGOT_PASSWORD_PIVOT_KEY}
          >
            <EditTemplatesWidget sections={sectionsForgotPassword} />
          </PivotItem>
          <PivotItem
            headerText={renderToString(
              "TemplatesConfigurationScreen.passwordless-authenticator.title"
            )}
            itemKey={PASSWORDLESS_AUTHENTICATOR_PIVOT_KEY}
          >
            <EditTemplatesWidget sections={sectionsPasswordless} />
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
    loading: loadingTemplateLocales,
    error: loadTemplateLocalesError,
    refetch: refetchTemplateLocales,
  } = useTemplateLocaleQuery(appID);

  const initialDefaultTemplateLocale = useMemo<TemplateLocale>(() => {
    return (
      effectiveAppConfig?.localization?.fallback_language ??
      DEFAULT_TEMPLATE_LOCALE
    );
  }, [effectiveAppConfig]);

  const [remountIdentifier, setRemountIdentifier] = useState(0);

  const onResetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
  }, []);

  const [defaultTemplateLocale, setDefaultTemplateLocale] = useState<
    TemplateLocale
  >(initialDefaultTemplateLocale);

  const [templateLocale, setTemplateLocale] = useState<TemplateLocale>(
    defaultTemplateLocale
  );

  const {
    templates: initialTemplates,
    loading: loadingTemplates,
    error: loadTemplatesError,
    refetch: refetchTemplates,
  } = useAppTemplatesQuery(
    appID,
    initialTemplateLocales,
    ...ALL_TEMPLATE_PATHS
  );

  if (loadingAppConfig || loadingTemplateLocales || loadingTemplates) {
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

  if (loadTemplatesError) {
    return <ShowError error={loadTemplatesError} onRetry={refetchTemplates} />;
  }

  return (
    <TemplatesConfiguration
      key={remountIdentifier}
      rawAppConfig={rawAppConfig!}
      initialTemplates={initialTemplates}
      initialTemplateLocales={initialTemplateLocales}
      initialDefaultTemplateLocale={initialDefaultTemplateLocale}
      defaultTemplateLocale={defaultTemplateLocale}
      templateLocale={templateLocale}
      setDefaultTemplateLocale={setDefaultTemplateLocale}
      setTemplateLocale={setTemplateLocale}
      onResetForm={onResetForm}
    />
  );
};

export default TemplatesConfigurationScreen;
