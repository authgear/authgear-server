import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { produce } from "immer";
import {
  ModifiedIndicatorWrapper,
  ModifiedIndicatorPortal,
} from "../../ModifiedIndicatorPortal";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ButtonWithLoading from "../../ButtonWithLoading";
import ManageLanguageWidget from "./ManageLanguageWidget";
import ImageFilePicker from "../../ImageFilePicker";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useAppTemplatesQuery } from "./query/appTemplatesQuery";
import { useTemplateLocaleQuery } from "./query/templateLocaleQuery";
import { useUpdateAppTemplatesMutation } from "./mutations/updateAppTemplatesMutation";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { PortalAPIAppConfig } from "../../types";
import {
  DEFAULT_TEMPLATE_LOCALE,
  ALL_RESOURCES,
  RESOURCE_TRANSLATION_JSON,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT,
  RESOURCE_FORGOT_PASSWORD_EMAIL_HTML,
  RESOURCE_FORGOT_PASSWORD_EMAIL_TXT,
  RESOURCE_FORGOT_PASSWORD_SMS_TXT,
  RESOURCE_APP_BANNER,
  RESOURCE_APP_LOGO,
  renderPath,
} from "../../resources";
import {
  LanguageTag,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
} from "../../util/resource";
import { generateUpdates } from "./templates";

import styles from "./ResourceConfigurationScreen.module.scss";

interface ResourceConfigurationSectionProps {
  rawAppConfig: PortalAPIAppConfig;
  initialTemplates: Resource[];
  initialTemplateLocales: LanguageTag[];
  initialDefaultTemplateLocale: LanguageTag;
  defaultTemplateLocale: LanguageTag;
  templateLocale: LanguageTag;
  setDefaultTemplateLocale: (locale: LanguageTag) => void;
  setTemplateLocale: (locale: LanguageTag) => void;
  onResetForm: () => void;
}

const PIVOT_KEY_APPEARANCE = "appearance";
const PIVOT_KEY_FORGOT_PASSWORD = "forgot_password";
const PIVOT_KEY_PASSWORDLESS = "passwordless";
const PIVOT_KEY_TRANSLATION_JSON = "translation.json";

const ALL_PIVOT_KEYS = [
  PIVOT_KEY_TRANSLATION_JSON,
  PIVOT_KEY_APPEARANCE,
  PIVOT_KEY_FORGOT_PASSWORD,
  PIVOT_KEY_PASSWORDLESS,
];

const ResourceConfigurationSection: React.FC<ResourceConfigurationSectionProps> = function ResourceConfigurationSection(
  props: ResourceConfigurationSectionProps
) {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const {
    rawAppConfig,
    initialTemplates,
    initialTemplateLocales,
    initialDefaultTemplateLocale,
    defaultTemplateLocale,
    setDefaultTemplateLocale,
    templateLocale,
    setTemplateLocale,
    onResetForm,
  } = props;

  const [templateLocales, setTemplateLocales] = useState<LanguageTag[]>(
    initialTemplateLocales
  );

  const [templates, setTemplates] = useState<Resource[]>(initialTemplates);

  const onChangeTemplateLocales = useCallback(
    (locales: LanguageTag[]) => {
      // Reset templateLocale to default if the selected one was removed.
      const idx = locales.findIndex((item) => item === templateLocale);
      if (idx < 0) {
        setTemplateLocale(defaultTemplateLocale);
      }

      // Find out new locales.
      const newLocales: LanguageTag[] = [];
      for (const newLocale of locales) {
        const idx = templateLocales.findIndex((item) => item === newLocale);
        if (idx < 0) {
          newLocales.push(newLocale);
        }
      }

      // Populate initial values for new locales from default locale.
      const newResources: Resource[] = [];
      for (const locale of newLocales) {
        for (const resource of ALL_RESOURCES) {
          const path = renderPath(resource.resourcePath, { locale });
          const defaultPath = renderPath(resource.resourcePath, {
            locale: defaultTemplateLocale,
          });
          const defaultResource = templates.find(
            (resource) => resource.path === defaultPath
          );
          const value = defaultResource?.value ?? "";
          const template: Resource = {
            specifier: {
              def: resource,
              locale,
            },
            path,
            value,
          };
          newResources.push(template);
        }
      }
      setTemplates((prev) => {
        // Discard any resources that are new locales.
        const withoutNewLocales = prev.filter((resource) => {
          const isNewLocale = newLocales.includes(resource.specifier.locale);
          return !isNewLocale;
        });
        return [...withoutNewLocales, ...newResources];
      });

      // Finally update the list of locales.
      setTemplateLocales(locales);
    },
    [
      templates,
      templateLocales,
      templateLocale,
      defaultTemplateLocale,
      setTemplateLocale,
    ]
  );

  const updates = useMemo(() => {
    return generateUpdates(
      initialTemplateLocales,
      initialTemplates,
      templateLocales,
      templates
    );
  }, [initialTemplateLocales, initialTemplates, templateLocales, templates]);

  const {
    invalidAdditionLocales,
    invalidEditionLocales,
    additions,
    editions,
    deletions,
  } = updates;

  const invalidTemplateLocales = useMemo(() => {
    return invalidAdditionLocales.concat(invalidEditionLocales);
  }, [invalidAdditionLocales, invalidEditionLocales]);

  const {
    updateAppTemplates,
    loading: updatingTemplates,
    error: updateTemplatesError,
  } = useUpdateAppTemplatesMutation(appID);

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const isModified =
    initialDefaultTemplateLocale !== defaultTemplateLocale ||
    updates.isModified;

  const onSubmit = useCallback(
    (e: React.FormEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();

      // Save default language
      if (initialDefaultTemplateLocale !== defaultTemplateLocale) {
        const newAppConfig = produce(rawAppConfig, (draftConfig) => {
          draftConfig.localization = draftConfig.localization ?? {};
          draftConfig.localization.fallback_language = defaultTemplateLocale;
        });
        updateAppConfig(newAppConfig).catch(() => {});
      }

      // Save templates
      const updates = [...additions, ...editions, ...deletions];
      if (updates.length > 0) {
        const specifiers = [];
        for (const resource of ALL_RESOURCES) {
          for (const locale of templateLocales) {
            specifiers.push({
              def: resource,
              locale,
            });
          }
        }
        updateAppTemplates(specifiers, updates).catch(() => {});
      }
    },
    [
      initialDefaultTemplateLocale,
      defaultTemplateLocale,
      templateLocales,
      rawAppConfig,
      updateAppConfig,
      updateAppTemplates,
      additions,
      editions,
      deletions,
    ]
  );

  // We used to use fragment to control the pivot key.
  // Now that the save button applies all changes, not just the changes in the curren pivot item.
  // Therefore, we do not need to use fragment to control the pivot key anymore.
  const [selectedKey, setSelectedKey] = useState<string>(
    PIVOT_KEY_TRANSLATION_JSON
  );
  const onLinkClick = useCallback((item?: PivotItem) => {
    const itemKey = item?.props.itemKey;
    if (itemKey != null) {
      const idx = ALL_PIVOT_KEYS.indexOf(itemKey);
      if (idx >= 0) {
        setSelectedKey(itemKey);
      }
    }
  }, []);

  const getValueIgnoreEmptyString = useCallback(
    (resourceDef: ResourceDefinition) => {
      const resource = templates.find(
        (resource) =>
          resource.specifier.def === resourceDef &&
          resource.specifier.locale === templateLocale
      );
      if (resource == null || resource.value === "") {
        return undefined;
      }
      return resource.value;
    },
    [templates, templateLocale]
  );

  const getValue = useCallback(
    (resourceDef: ResourceDefinition) => {
      const resource = templates.find(
        (resource) =>
          resource.specifier.def === resourceDef &&
          resource.specifier.locale === templateLocale
      );
      return resource?.value ?? "";
    },
    [templates, templateLocale]
  );

  const getOnChange = useCallback(
    (resourceDef: ResourceDefinition) => {
      return (_e: unknown, value?: string) => {
        if (value != null) {
          const path = renderPath(resourceDef.resourcePath, {
            locale: templateLocale,
          });
          setTemplates((prev) => {
            const idx = prev.findIndex(
              (resource) =>
                resource.specifier.def === resourceDef &&
                resource.specifier.locale === templateLocale
            );

            let template: Resource;
            if (idx < 0) {
              template = {
                specifier: {
                  def: resourceDef,
                  locale: templateLocale,
                },
                path: path,
                value,
              };
            } else {
              template = {
                ...prev[idx],
                value,
              };
            }

            const newTemplates = [...prev];
            if (idx < 0) {
              newTemplates.push(template);
            } else {
              newTemplates[idx] = template;
            }

            return newTemplates;
          });
        }
      };
    },
    [templateLocale]
  );

  const getOnChangeImage = useCallback(
    (resourceDef: ResourceDefinition) => {
      return (base64EncodedData?: string, extension?: string) => {
        setTemplates((prev) => {
          // First we have to remove the current one.
          const next = prev.filter((resource) => {
            const ok =
              resource.specifier.def === resourceDef &&
              resource.specifier.locale === templateLocale;
            return !ok;
          });
          // Add if it is not a deletion.
          if (base64EncodedData != null && extension != null) {
            const path = renderPath(resourceDef.resourcePath, {
              locale: templateLocale,
              extension,
            });
            next.push({
              specifier: {
                def: resourceDef,
                locale: templateLocale,
              },
              path,
              value: base64EncodedData,
            });
          }
          return next;
        });
      };
    },
    [templateLocale]
  );

  const sectionsTranslationJSON: EditTemplatesWidgetSection[] = [
    {
      key: "translation.json",
      title: (
        <FormattedMessage id="EditTemplatesWidget.translationjson.title" />
      ),
      items: [
        {
          key: "translation.json",
          title: (
            <FormattedMessage id="EditTemplatesWidget.translationjson.subtitle" />
          ),
          language: "json",
          value: getValue(RESOURCE_TRANSLATION_JSON),
          onChange: getOnChange(RESOURCE_TRANSLATION_JSON),
        },
      ],
    },
  ];

  const sectionsForgotPassword: EditTemplatesWidgetSection[] = [
    {
      key: "email",
      title: <FormattedMessage id="EditTemplatesWidget.email" />,
      items: [
        {
          key: "html-email",
          title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
          language: "html",
          value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_TXT),
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
          value: getValue(RESOURCE_FORGOT_PASSWORD_SMS_TXT),
          onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_SMS_TXT),
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
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT),
        },
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
          onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
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
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML),
        },
        {
          key: "plaintext-email",
          title: <FormattedMessage id="EditTemplatesWidget.plaintext-email" />,
          language: "plaintext",
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT),
        },
        {
          key: "sms",
          title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
          language: "plaintext",
          value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
          onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
        },
      ],
    },
  ];

  return (
    <form
      role="main"
      className={cn(styles.root, {
        [styles.loading]: updatingTemplates,
      })}
      onSubmit={onSubmit}
    >
      {updateTemplatesError && <ShowError error={updateTemplatesError} />}
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <ModifiedIndicatorWrapper className={styles.screen}>
        <ModifiedIndicatorPortal
          resetForm={onResetForm}
          isModified={isModified}
        />
        <NavigationBlockerDialog blockNavigation={isModified} />
        <Text className={styles.screenHeaderText} as="h1">
          <FormattedMessage id="ResourceConfigurationScreen.title" />
        </Text>
        <ManageLanguageWidget
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
              "ResourceConfigurationScreen.translationjson.title"
            )}
            itemKey={PIVOT_KEY_TRANSLATION_JSON}
          >
            <EditTemplatesWidget sections={sectionsTranslationJSON} />
          </PivotItem>
          <PivotItem
            headerText={renderToString(
              "ResourceConfigurationScreen.appearance.title"
            )}
            itemKey={PIVOT_KEY_APPEARANCE}
          >
            <div className={styles.pivotItemAppearance}>
              <ImageFilePicker
                title={renderToString("ResourceConfigurationScreen.app-banner")}
                base64EncodedData={getValueIgnoreEmptyString(
                  RESOURCE_APP_BANNER
                )}
                onChange={getOnChangeImage(RESOURCE_APP_BANNER)}
              />
              <ImageFilePicker
                title={renderToString("ResourceConfigurationScreen.app-logo")}
                base64EncodedData={getValueIgnoreEmptyString(RESOURCE_APP_LOGO)}
                onChange={getOnChangeImage(RESOURCE_APP_LOGO)}
              />
            </div>
          </PivotItem>
          <PivotItem
            headerText={renderToString(
              "ResourceConfigurationScreen.forgot-password.title"
            )}
            itemKey={PIVOT_KEY_FORGOT_PASSWORD}
          >
            <EditTemplatesWidget sections={sectionsForgotPassword} />
          </PivotItem>
          <PivotItem
            headerText={renderToString(
              "ResourceConfigurationScreen.passwordless-authenticator.title"
            )}
            itemKey={PIVOT_KEY_PASSWORDLESS}
          >
            <EditTemplatesWidget sections={sectionsPasswordless} />
          </PivotItem>
        </Pivot>
        <ButtonWithLoading
          className={styles.saveButton}
          type="submit"
          disabled={
            !isModified ||
            invalidAdditionLocales.length > 0 ||
            invalidEditionLocales.length > 0
          }
          loading={updatingAppConfig || updatingTemplates}
          labelId="save"
          loadingLabelId="saving"
        />
      </ModifiedIndicatorWrapper>
    </form>
  );
};

const ResourceConfigurationScreen: React.FC = function ResourceConfigurationScreen() {
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

  const initialDefaultTemplateLocale = useMemo<LanguageTag>(() => {
    return (
      effectiveAppConfig?.localization?.fallback_language ??
      DEFAULT_TEMPLATE_LOCALE
    );
  }, [effectiveAppConfig]);

  const [remountIdentifier, setRemountIdentifier] = useState(0);

  const [defaultTemplateLocale, setDefaultTemplateLocale] = useState<
    LanguageTag
  >(initialDefaultTemplateLocale);

  const [templateLocale, setTemplateLocale] = useState<LanguageTag>(
    defaultTemplateLocale
  );

  const onResetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
    setDefaultTemplateLocale(initialDefaultTemplateLocale);
    setTemplateLocale(initialDefaultTemplateLocale);
  }, [initialDefaultTemplateLocale]);

  const specifiers = useMemo<ResourceSpecifier[]>(() => {
    const specifiers = [];
    for (const locale of initialTemplateLocales) {
      for (const def of ALL_RESOURCES) {
        specifiers.push({
          def,
          locale,
        });
      }
    }
    return specifiers;
  }, [initialTemplateLocales]);

  const {
    resources: initialTemplates,
    loading: loadingTemplates,
    error: loadTemplatesError,
    refetch: refetchTemplates,
  } = useAppTemplatesQuery(appID, specifiers);

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
    <ResourceConfigurationSection
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

export default ResourceConfigurationScreen;
