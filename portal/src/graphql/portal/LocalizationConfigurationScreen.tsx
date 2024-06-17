import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem } from "@fluentui/react";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import ManageLanguageWidget from "./ManageLanguageWidget";
import EditTemplatesWidget, {
  EditTemplatesWidgetSection,
} from "./EditTemplatesWidget";
import { AuthenticatorEmailOTPMode } from "../../types";
import {
  ALL_LANGUAGES_TEMPLATES,
  DEFAULT_TEMPLATE_LOCALE,
  RESOURCE_AUTHENTICATE_PRIMARY_LOGIN_LINK_HTML,
  RESOURCE_AUTHENTICATE_PRIMARY_LOGIN_LINK_TXT,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT,
  RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_HTML,
  RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_TXT,
  RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_HTML,
  RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_TXT,
  RESOURCE_FORGOT_PASSWORD_SMS_CODE_TXT,
  RESOURCE_FORGOT_PASSWORD_SMS_LINK_TXT,
  RESOURCE_SETUP_PRIMARY_LOGIN_LINK_HTML,
  RESOURCE_SETUP_PRIMARY_LOGIN_LINK_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT,
  RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT,
  RESOURCE_TRANSLATION_JSON,
  RESOURCE_VERIFICATION_EMAIL_HTML,
  RESOURCE_VERIFICATION_EMAIL_TXT,
  RESOURCE_VERIFICATION_SMS_TXT,
  TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_LOGIN_LINK_SUBJECT,
  TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_OOB_SUBJECT,
  TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_CODE_SUBJECT,
  TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_LINK_SUBJECT,
  TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_LOGIN_LINK_SUBJECT,
  TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_OOB_SUBJECT,
  TRANSLATION_JSON_KEY_EMAIL_VERIFICATION_SUBJECT,
} from "../../resources";
import {
  LanguageTag,
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  specifierId,
  expandSpecifier,
} from "../../util/resource";
import {
  ResourcesFormState,
  useResourceForm,
} from "../../hook/useResourceForm";
import FormContainer from "../../FormContainer";
import styles from "./LocalizationConfigurationScreen.module.css";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";

interface FormState extends ResourcesFormState {
  supportedLanguages: string[];
  fallbackLanguage: string;
  selectedLanguage: string;
}

interface FormModel {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: FormState;
  setState: (fn: (state: FormState) => FormState) => void;
  reload: () => void;
  reset: () => void;
  save: () => Promise<void>;
}

interface ResourcesConfigurationContentProps {
  form: FormModel;
  initialSupportedLanguages: string[];
  passwordlessViaEmailEnabled: boolean;
  passwordlessViaSMSEnabled: boolean;
  passwordlessViaEmailOTPMode: AuthenticatorEmailOTPMode;
  verificationEnabled: boolean;
}

const PIVOT_KEY_FORGOT_PASSWORD_LINK = "forgot_password_link";
const PIVOT_KEY_FORGOT_PASSWORD_CODE = "forgot_password_code";
const PIVOT_KEY_VERIFICATION = "verification";
const PIVOT_KEY_PASSWORDLESS_VIA_EMAIL = "passwordless_via_email";
const PIVOT_KEY_PASSWORDLESS_VIA_SMS = "passwordless_via_sms";

const PIVOT_KEY_DEFAULT = PIVOT_KEY_FORGOT_PASSWORD_LINK;

const ALL_PIVOT_KEYS = [
  PIVOT_KEY_FORGOT_PASSWORD_LINK,
  PIVOT_KEY_FORGOT_PASSWORD_CODE,
  PIVOT_KEY_VERIFICATION,
  PIVOT_KEY_PASSWORDLESS_VIA_EMAIL,
  PIVOT_KEY_PASSWORDLESS_VIA_SMS,
];

const ResourcesConfigurationContent: React.VFC<ResourcesConfigurationContentProps> =
  function ResourcesConfigurationContent(props) {
    const { state, setState } = props.form;
    const {
      initialSupportedLanguages,
      passwordlessViaEmailEnabled,
      passwordlessViaSMSEnabled,
      passwordlessViaEmailOTPMode,
      verificationEnabled,
    } = props;
    const { supportedLanguages } = state;
    const { renderToString } = useContext(Context);

    const setSelectedLanguage = useCallback(
      (selectedLanguage: LanguageTag) => {
        setState((s) => ({ ...s, selectedLanguage }));
      },
      [setState]
    );

    const [selectedKey, setSelectedKey] = useState<string>(PIVOT_KEY_DEFAULT);
    const onLinkClick = useCallback((item?: PivotItem) => {
      const itemKey = item?.props.itemKey;
      if (itemKey != null) {
        const idx = ALL_PIVOT_KEYS.indexOf(itemKey);
        if (idx >= 0) {
          setSelectedKey(itemKey);
        }
      }
    }, []);

    const getValueFromState = useCallback(
      (
        resources: Partial<Record<string, Resource>>,
        selectedLanguage: string,
        fallbackLanguage: string,
        def: ResourceDefinition,
        getValueFn: (
          resource: Resource | undefined
        ) => string | undefined | null
      ): string | undefined | null => {
        const specifier: ResourceSpecifier = {
          def,
          locale: selectedLanguage,
          extension: null,
        };
        const value = getValueFn(resources[specifierId(specifier)]);

        if (value == null) {
          const specifier: ResourceSpecifier = {
            def,
            locale: fallbackLanguage,
            extension: null,
          };
          return getValueFn(resources[specifierId(specifier)]);
        }

        return value;
      },
      []
    );

    const getValue = useCallback(
      (def: ResourceDefinition) => {
        const selectedValue = getValueFromState(
          state.resources,
          state.selectedLanguage,
          state.fallbackLanguage,
          def,
          (res) => res?.nullableValue ?? res?.effectiveData
        );
        if (selectedValue != null) {
          return selectedValue;
        }

        return (
          getValueFromState(
            state.resources,
            DEFAULT_TEMPLATE_LOCALE,
            state.fallbackLanguage,
            def,
            (res) => res?.effectiveData
          ) ?? ""
        );
      },
      [
        state.resources,
        state.selectedLanguage,
        state.fallbackLanguage,
        getValueFromState,
      ]
    );

    const getOnChange = useCallback(
      (def: ResourceDefinition) => {
        const specifier: ResourceSpecifier = {
          def,
          locale: state.selectedLanguage,
          extension: null,
        };
        return (value: string | undefined, _e: unknown) => {
          setState((prev) => {
            const updatedResources = { ...prev.resources };
            const resource: Resource = {
              specifier,
              path: expandSpecifier(specifier),
              nullableValue: value ?? "",
              effectiveData:
                prev.resources[specifierId(specifier)]?.effectiveData,
            };
            updatedResources[specifierId(resource.specifier)] = resource;
            return { ...prev, resources: updatedResources };
          });
        };
      },
      [state.selectedLanguage, setState]
    );

    const getTranslationValue = useCallback(
      (key: string) => {
        // get from the translation json first
        const translationJSONStr = getValueFromState(
          state.resources,
          state.selectedLanguage,
          state.fallbackLanguage,
          RESOURCE_TRANSLATION_JSON,
          (res) => res?.nullableValue
        );
        try {
          if (translationJSONStr != null) {
            const translationJSON = JSON.parse(translationJSONStr);
            if (translationJSON[key] != null) {
              return translationJSON[key];
            }
          }
        } catch (_e: unknown) {
          // if failed to decode the translation.json, use the effective data
        }

        // fallback to the effective data
        const effTranslationJSONStr = getValueFromState(
          state.resources,
          state.selectedLanguage,
          state.fallbackLanguage,
          RESOURCE_TRANSLATION_JSON,
          (res) => res?.effectiveData
        );
        try {
          if (effTranslationJSONStr != null) {
            const translationJSON = JSON.parse(effTranslationJSONStr);
            return translationJSON[key] ?? "";
          }
        } catch (_e: unknown) {
          // if failed to decode the translation.json, use English.
        }

        // fallback to en
        const enTranslationJSONStr = getValueFromState(
          state.resources,
          DEFAULT_TEMPLATE_LOCALE,
          state.fallbackLanguage,
          RESOURCE_TRANSLATION_JSON,
          (res) => res?.effectiveData
        );
        try {
          if (enTranslationJSONStr != null) {
            const translationJSON = JSON.parse(enTranslationJSONStr);
            return translationJSON[key] ?? "";
          }
        } catch (_e: unknown) {
          // if failed to decode the translation.json, return empty string
        }

        return "";
      },
      [
        state.resources,
        state.selectedLanguage,
        state.fallbackLanguage,
        getValueFromState,
      ]
    );

    const getTranslationOnChange = useCallback(
      (key: string) => {
        const specifier: ResourceSpecifier = {
          def: RESOURCE_TRANSLATION_JSON,
          locale: state.selectedLanguage,
          extension: null,
        };
        return (value: string | undefined, _e: unknown) => {
          setState((prev) => {
            // get the translation JSON, decode and alter
            const translationJSONStr = getValueFromState(
              prev.resources,
              prev.selectedLanguage,
              prev.fallbackLanguage,
              RESOURCE_TRANSLATION_JSON,
              (res) => res?.nullableValue
            );

            // By default, create a new translation.json
            let translationJSON: Record<string, string> = {};
            // If translation.json exists, use it.
            if (translationJSONStr != null) {
              translationJSON = JSON.parse(translationJSONStr);
            }

            // Update the value.
            if (value) {
              translationJSON[key] = value;
            } else {
              delete translationJSON[key];
            }

            const resultTranslationJSON = JSON.stringify(
              translationJSON,
              null,
              2
            );

            // get the translation JSON effective data, decode and alter
            const effTranslationJSONStr = getValueFromState(
              prev.resources,
              prev.selectedLanguage,
              prev.fallbackLanguage,
              RESOURCE_TRANSLATION_JSON,
              (res) => res?.effectiveData
            );

            // By default, create a new translation.json
            let effTranslationJSON: Record<string, string> = {};
            if (effTranslationJSONStr != null) {
              effTranslationJSON = JSON.parse(effTranslationJSONStr);
            }

            // Update the value.
            if (value) {
              effTranslationJSON[key] = value;
            } else {
              delete effTranslationJSON[key];
            }

            const resultEffTranslationJSON = JSON.stringify(
              effTranslationJSON,
              null,
              2
            );

            // When the value is updated, both value and effective data of
            // translation.json need to be updated
            // Otherwise there will be inconsistent in the ui
            const updatedResources = { ...prev.resources };
            const resource: Resource = {
              specifier,
              path: expandSpecifier(specifier),
              nullableValue: resultTranslationJSON,
              effectiveData: resultEffTranslationJSON,
            };
            updatedResources[specifierId(resource.specifier)] = resource;
            return { ...prev, resources: updatedResources };
          });
        };
      },
      [setState, getValueFromState, state.selectedLanguage]
    );

    const sectionsForgotPasswordLink: EditTemplatesWidgetSection[] = [
      {
        key: "email",
        title: <FormattedMessage id="EditTemplatesWidget.email" />,
        items: [
          {
            key: "email-subject",
            title: <FormattedMessage id="EditTemplatesWidget.email-subject" />,
            language: "plaintext",
            value: getTranslationValue(
              TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_LINK_SUBJECT
            ),
            onChange: getTranslationOnChange(
              TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_LINK_SUBJECT
            ),
            editor: "textfield",
          },
          {
            key: "html-email",
            title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
            language: "html",
            value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_HTML),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_HTML),
            editor: "code",
          },
          {
            key: "plaintext-email",
            title: (
              <FormattedMessage id="EditTemplatesWidget.plaintext-email" />
            ),
            language: "plaintext",
            value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_TXT),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_LINK_TXT),
            editor: "code",
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
            value: getValue(RESOURCE_FORGOT_PASSWORD_SMS_LINK_TXT),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_SMS_LINK_TXT),
            editor: "code",
          },
        ],
      },
    ];

    const sectionsForgotPasswordCode: EditTemplatesWidgetSection[] = [
      {
        key: "email",
        title: <FormattedMessage id="EditTemplatesWidget.email" />,
        items: [
          {
            key: "email-subject",
            title: <FormattedMessage id="EditTemplatesWidget.email-subject" />,
            language: "plaintext",
            value: getTranslationValue(
              TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_CODE_SUBJECT
            ),
            onChange: getTranslationOnChange(
              TRANSLATION_JSON_KEY_EMAIL_FORGOT_PASSWORD_CODE_SUBJECT
            ),
            editor: "textfield",
          },
          {
            key: "html-email",
            title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
            language: "html",
            value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_HTML),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_HTML),
            editor: "code",
          },
          {
            key: "plaintext-email",
            title: (
              <FormattedMessage id="EditTemplatesWidget.plaintext-email" />
            ),
            language: "plaintext",
            value: getValue(RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_TXT),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_EMAIL_CODE_TXT),
            editor: "code",
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
            value: getValue(RESOURCE_FORGOT_PASSWORD_SMS_CODE_TXT),
            onChange: getOnChange(RESOURCE_FORGOT_PASSWORD_SMS_CODE_TXT),
            editor: "code",
          },
        ],
      },
    ];

    const sectionsVerification: EditTemplatesWidgetSection[] = [
      {
        key: "email",
        title: <FormattedMessage id="EditTemplatesWidget.email" />,
        items: [
          {
            key: "email-subject",
            title: <FormattedMessage id="EditTemplatesWidget.email-subject" />,
            language: "plaintext",
            value: getTranslationValue(
              TRANSLATION_JSON_KEY_EMAIL_VERIFICATION_SUBJECT
            ),
            onChange: getTranslationOnChange(
              TRANSLATION_JSON_KEY_EMAIL_VERIFICATION_SUBJECT
            ),
            editor: "textfield",
          },
          {
            key: "html-email",
            title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
            language: "html",
            value: getValue(RESOURCE_VERIFICATION_EMAIL_HTML),
            onChange: getOnChange(RESOURCE_VERIFICATION_EMAIL_HTML),
            editor: "code",
          },
          {
            key: "plaintext-email",
            title: (
              <FormattedMessage id="EditTemplatesWidget.plaintext-email" />
            ),
            language: "plaintext",
            value: getValue(RESOURCE_VERIFICATION_EMAIL_TXT),
            onChange: getOnChange(RESOURCE_VERIFICATION_EMAIL_TXT),
            editor: "code",
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
            value: getValue(RESOURCE_VERIFICATION_SMS_TXT),
            onChange: getOnChange(RESOURCE_VERIFICATION_SMS_TXT),
            editor: "code",
          },
        ],
      },
    ];

    const passwordlessViaEmailTemplates = {
      code: {
        setupHtml: RESOURCE_SETUP_PRIMARY_OOB_EMAIL_HTML,
        setupPlainText: RESOURCE_SETUP_PRIMARY_OOB_EMAIL_TXT,
        authenticateHtml: RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
        authenticatePlainText: RESOURCE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TXT,
      },
      login_link: {
        setupHtml: RESOURCE_SETUP_PRIMARY_LOGIN_LINK_HTML,
        setupPlainText: RESOURCE_SETUP_PRIMARY_LOGIN_LINK_TXT,
        authenticateHtml: RESOURCE_AUTHENTICATE_PRIMARY_LOGIN_LINK_HTML,
        authenticatePlainText: RESOURCE_AUTHENTICATE_PRIMARY_LOGIN_LINK_TXT,
      },
    }[passwordlessViaEmailOTPMode];

    const passwordlessViaEmailSubject = {
      code: {
        setup: TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_OOB_SUBJECT,
        authenticate:
          TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_OOB_SUBJECT,
      },
      login_link: {
        setup: TRANSLATION_JSON_KEY_EMAIL_SETUP_PRIMARY_LOGIN_LINK_SUBJECT,
        authenticate:
          TRANSLATION_JSON_KEY_EMAIL_AUTHENTICATE_PRIMARY_LOGIN_LINK_SUBJECT,
      },
    }[passwordlessViaEmailOTPMode];

    const sectionsPasswordlessViaEmail: EditTemplatesWidgetSection[] = [
      {
        key: "setup",
        title: (
          <FormattedMessage id="EditTemplatesWidget.passwordless.setup.title" />
        ),
        items: [
          {
            key: "email-subject",
            title: <FormattedMessage id="EditTemplatesWidget.email-subject" />,
            language: "plaintext",
            value: getTranslationValue(passwordlessViaEmailSubject.setup),
            onChange: getTranslationOnChange(passwordlessViaEmailSubject.setup),
            editor: "textfield",
          },
          {
            key: "html-email",
            title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
            language: "html",
            value: getValue(passwordlessViaEmailTemplates.setupHtml),
            onChange: getOnChange(passwordlessViaEmailTemplates.setupHtml),
            editor: "code",
          },
          {
            key: "plaintext-email",
            title: (
              <FormattedMessage id="EditTemplatesWidget.plaintext-email" />
            ),
            language: "plaintext",
            value: getValue(passwordlessViaEmailTemplates.setupPlainText),
            onChange: getOnChange(passwordlessViaEmailTemplates.setupPlainText),
            editor: "code",
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
            key: "email-subject",
            title: <FormattedMessage id="EditTemplatesWidget.email-subject" />,
            language: "plaintext",
            value: getTranslationValue(
              passwordlessViaEmailSubject.authenticate
            ),
            onChange: getTranslationOnChange(
              passwordlessViaEmailSubject.authenticate
            ),
            editor: "textfield",
          },
          {
            key: "html-email",
            title: <FormattedMessage id="EditTemplatesWidget.html-email" />,
            language: "html",
            value: getValue(passwordlessViaEmailTemplates.authenticateHtml),
            onChange: getOnChange(
              passwordlessViaEmailTemplates.authenticateHtml
            ),
            editor: "code",
          },
          {
            key: "plaintext-email",
            title: (
              <FormattedMessage id="EditTemplatesWidget.plaintext-email" />
            ),
            language: "plaintext",
            value: getValue(
              passwordlessViaEmailTemplates.authenticatePlainText
            ),
            onChange: getOnChange(
              passwordlessViaEmailTemplates.authenticatePlainText
            ),
            editor: "code",
          },
        ],
      },
    ];

    const sectionsPasswordlessViaSMS: EditTemplatesWidgetSection[] = [
      {
        key: "setup",
        title: (
          <FormattedMessage id="EditTemplatesWidget.passwordless.setup.title" />
        ),
        items: [
          {
            key: "sms",
            title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
            language: "plaintext",
            value: getValue(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
            onChange: getOnChange(RESOURCE_SETUP_PRIMARY_OOB_SMS_TXT),
            editor: "code",
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
            key: "sms",
            title: <FormattedMessage id="EditTemplatesWidget.sms-body" />,
            language: "plaintext",
            value: getValue(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
            onChange: getOnChange(RESOURCE_AUTHENTICATE_PRIMARY_OOB_SMS_TXT),
            editor: "code",
          },
        ],
      },
    ];

    return (
      <ScreenContent>
        <ScreenTitle className={cn("col-span-8", "tablet:col-span-full")}>
          <FormattedMessage id="LocalizationConfigurationScreen.title" />
        </ScreenTitle>
        <div className={styles.descriptionContainer}>
          <ScreenDescription className={styles.widget}>
            <FormattedMessage id="LocalizationConfigurationScreen.description" />
          </ScreenDescription>
          <ManageLanguageWidget
            showLabel={false}
            existingLanguages={initialSupportedLanguages}
            supportedLanguages={supportedLanguages}
            selectedLanguage={state.selectedLanguage}
            fallbackLanguage={state.fallbackLanguage}
            onChangeSelectedLanguage={setSelectedLanguage}
          />
        </div>
        {/* Code editors might incorrectly fire change events when changing language
            Set key to selectedLanguage to ensure code editors always remount */}
        <Widget className={styles.widget} key={state.selectedLanguage}>
          <WidgetTitle>
            <FormattedMessage id="LocalizationConfigurationScreen.template-content-title" />
          </WidgetTitle>
          <Pivot
            overflowBehavior="menu"
            onLinkClick={onLinkClick}
            selectedKey={selectedKey}
          >
            <PivotItem
              headerText={renderToString(
                "LocalizationConfigurationScreen.forgot-password-link.title"
              )}
              itemKey={PIVOT_KEY_FORGOT_PASSWORD_LINK}
            >
              <EditTemplatesWidget sections={sectionsForgotPasswordLink} />
            </PivotItem>
            <PivotItem
              headerText={renderToString(
                "LocalizationConfigurationScreen.forgot-password-code.title"
              )}
              itemKey={PIVOT_KEY_FORGOT_PASSWORD_CODE}
            >
              <EditTemplatesWidget sections={sectionsForgotPasswordCode} />
            </PivotItem>
            {verificationEnabled ? (
              <PivotItem
                headerText={renderToString(
                  "LocalizationConfigurationScreen.verification.title"
                )}
                itemKey={PIVOT_KEY_VERIFICATION}
              >
                <EditTemplatesWidget sections={sectionsVerification} />
              </PivotItem>
            ) : null}
            {passwordlessViaEmailEnabled ? (
              <PivotItem
                headerText={renderToString(
                  "LocalizationConfigurationScreen.passwordless-via-email.title"
                )}
                itemKey={PIVOT_KEY_PASSWORDLESS_VIA_EMAIL}
              >
                <EditTemplatesWidget sections={sectionsPasswordlessViaEmail} />
              </PivotItem>
            ) : null}
            {passwordlessViaSMSEnabled ? (
              <PivotItem
                headerText={renderToString(
                  "LocalizationConfigurationScreen.passwordless-via-sms.title"
                )}
                itemKey={PIVOT_KEY_PASSWORDLESS_VIA_SMS}
              >
                <EditTemplatesWidget sections={sectionsPasswordlessViaSMS} />
              </PivotItem>
            ) : null}
          </Pivot>
        </Widget>
      </ScreenContent>
    );
  };

const LocalizationConfigurationScreen: React.VFC =
  function LocalizationConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const [selectedLanguage, setSelectedLanguage] =
      useState<LanguageTag | null>(null);

    const config = useAppAndSecretConfigQuery(appID);

    const initialSupportedLanguages = useMemo(() => {
      return (
        config.effectiveAppConfig?.localization?.supported_languages ?? [
          config.effectiveAppConfig?.localization?.fallback_language ?? "en",
        ]
      );
    }, [config.effectiveAppConfig?.localization]);

    const verificationEnabled = useMemo(() => {
      const verificationConfig = config.effectiveAppConfig?.verification;
      return Boolean(
        (verificationConfig?.claims?.email?.enabled ?? true) ||
          (verificationConfig?.claims?.phone_number?.enabled ?? true)
      );
    }, [config]);

    const passwordlessViaEmailEnabled = useMemo(() => {
      return (
        config.effectiveAppConfig?.authentication?.primary_authenticators?.indexOf(
          "oob_otp_email"
        ) !== -1
      );
    }, [config.effectiveAppConfig]);

    const passwordlessViaSMSEnabled = useMemo(() => {
      return (
        config.effectiveAppConfig?.authentication?.primary_authenticators?.indexOf(
          "oob_otp_sms"
        ) !== -1
      );
    }, [config.effectiveAppConfig]);

    const passwordlessViaEmailOTPMode =
      config.effectiveAppConfig?.authenticator?.oob_otp?.email
        ?.email_otp_mode ?? "code";

    const specifiers = useMemo<ResourceSpecifier[]>(() => {
      const specifiers = [];

      const supportedLanguages = [...initialSupportedLanguages];
      if (!supportedLanguages.includes(DEFAULT_TEMPLATE_LOCALE)) {
        supportedLanguages.push(DEFAULT_TEMPLATE_LOCALE);
      }

      for (const locale of supportedLanguages) {
        for (const def of ALL_LANGUAGES_TEMPLATES) {
          specifiers.push({
            def,
            locale,
            extension: null,
          });
        }
      }
      return specifiers;
    }, [initialSupportedLanguages]);

    const resources = useResourceForm(appID, specifiers);

    const state = useMemo<FormState>(() => {
      const fallbackLanguage =
        config.effectiveAppConfig?.localization?.fallback_language ?? "en";
      return {
        supportedLanguages: config.effectiveAppConfig?.localization
          ?.supported_languages ?? [fallbackLanguage],
        fallbackLanguage,
        resources: resources.state.resources,
        selectedLanguage: selectedLanguage ?? fallbackLanguage,
      };
    }, [
      config.effectiveAppConfig?.localization,
      resources.state.resources,
      selectedLanguage,
    ]);

    const form: FormModel = useMemo(
      () => ({
        isLoading: config.loading || resources.isLoading,
        isUpdating: resources.isUpdating,
        isDirty: resources.isDirty,
        loadError: config.error ?? resources.loadError,
        updateError: resources.updateError,
        state,
        setState: (fn) => {
          const newState = fn(state);
          resources.setState(() => ({ resources: newState.resources }));
          setSelectedLanguage(newState.selectedLanguage);
        },
        reload: () => {
          // Previously is also a floating promise, so just log the error out
          // to make linter happy
          config.refetch().catch((err) => {
            console.error("Reload config error", err);
            throw err;
          });
          resources.reload();
        },
        reset: () => {
          resources.reset();
          setSelectedLanguage(state.fallbackLanguage);
        },
        save: async (ignoreConflict: boolean = false) => {
          await resources.save(ignoreConflict);
        },
      }),
      [config, resources, state]
    );

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form} canSave={true}>
        <ResourcesConfigurationContent
          form={form}
          initialSupportedLanguages={initialSupportedLanguages}
          passwordlessViaEmailEnabled={passwordlessViaEmailEnabled}
          passwordlessViaSMSEnabled={passwordlessViaSMSEnabled}
          passwordlessViaEmailOTPMode={passwordlessViaEmailOTPMode}
          verificationEnabled={verificationEnabled}
        />
      </FormContainer>
    );
  };

export default LocalizationConfigurationScreen;
