import React, { useCallback, useContext, useMemo } from "react";
import produce from "immer";
import { Checkbox, Toggle, MessageBar } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import Widget from "../../Widget";
import WidgetWithOrdering from "../../WidgetWithOrdering";
import CheckboxWithContentLayout from "../../CheckboxWithContentLayout";
import CountryCallingCodeList from "./AuthenticationCountryCallingCodeList";
import { useTagPickerWithNewTags } from "../../hook/useInput";
import { clearEmptyObject } from "../../util/misc";
import { countryCallingCodes as supportedCountryCallingCodes } from "../../data/countryCallingCode.json";
import { useParams } from "react-router-dom";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import FormContainer from "../../FormContainer";
import {
  LocalValidationError,
  makeLocalValidationError,
} from "../../error/validation";
import {
  LoginIDEmailConfig,
  LoginIDKeyType,
  loginIDKeyTypes,
  LoginIDUsernameConfig,
  PortalAPIAppConfig,
  UICountryCallingCodeConfig,
} from "../../types";
import {
  renderPath,
  DEFAULT_TEMPLATE_LOCALE,
  RESOURCE_EMAIL_DOMAIN_BLOCKLIST,
  RESOURCE_EMAIL_DOMAIN_ALLOWLIST,
  RESOURCE_USERNAME_EXCLUDED_KEYWORDS_TXT,
} from "../../resources";

import styles from "./LoginIDConfigurationScreen.module.scss";
import CheckboxWithTooltip from "../../CheckboxWithTooltip";
import { Resource, ResourceSpecifier, specifierId } from "../../util/resource";
import { useResourceForm } from "../../hook/useResourceForm";
import CustomTagPicker from "../../CustomTagPicker";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

// email domain lists are not language specific
// so the locale in ResourceSpecifier is not important
const emailDomainBlocklistSpecifier: ResourceSpecifier = {
  def: RESOURCE_EMAIL_DOMAIN_BLOCKLIST,
  locale: DEFAULT_TEMPLATE_LOCALE,
};

const emailDomainAllowlistSpecifier: ResourceSpecifier = {
  def: RESOURCE_EMAIL_DOMAIN_ALLOWLIST,
  locale: DEFAULT_TEMPLATE_LOCALE,
};

const usernameExcludeKeywordsTXTSpecifier: ResourceSpecifier = {
  def: RESOURCE_USERNAME_EXCLUDED_KEYWORDS_TXT,
  locale: DEFAULT_TEMPLATE_LOCALE,
};

const specifiers: ResourceSpecifier[] = [
  emailDomainBlocklistSpecifier,
  emailDomainAllowlistSpecifier,
  usernameExcludeKeywordsTXTSpecifier,
];

interface LoginIDKeyTypeFormState {
  isEnabled: boolean;
  type: LoginIDKeyType;
}

interface EmailConfig extends LoginIDEmailConfig {
  modify_disabled?: boolean;
}

interface UsernameConfig extends LoginIDUsernameConfig {
  modify_disabled?: boolean;
}

interface PhoneConfig extends UICountryCallingCodeConfig {
  modify_disabled?: boolean;
}

interface ConfigFormState {
  types: LoginIDKeyTypeFormState[];
  email: Required<EmailConfig>;
  username: Required<UsernameConfig>;
  phone: Required<PhoneConfig>;
}

interface FeatureConfigFormState {
  loginIDPhoneDisabled: boolean;
}

function splitByNewline(text: string): string[] {
  return text
    .split(/\r?\n/)
    .map((x) => x.trim())
    .filter(Boolean);
}

function joinByNewline(list: string[]): string {
  return list.join("\n");
}

function constructConfigFormState(config: PortalAPIAppConfig): ConfigFormState {
  const isLoginIDEnabled =
    config.authentication?.identities?.includes("login_id") ?? true;
  const types: LoginIDKeyTypeFormState[] = (
    config.identity?.login_id?.keys ?? []
  ).map((k) => ({
    isEnabled: isLoginIDEnabled,
    type: k.type,
  }));
  for (const type of loginIDKeyTypes) {
    if (!types.some((t) => t.type === type)) {
      types.push({ isEnabled: false, type });
    }
  }

  return {
    types,
    email: {
      block_plus_sign: false,
      case_sensitive: false,
      ignore_dot_sign: false,
      domain_blocklist_enabled: false,
      domain_allowlist_enabled: false,
      block_free_email_provider_domains: false,
      modify_disabled:
        config.identity?.login_id?.keys?.find((a) => a.type === "email")
          ?.modify_disabled ?? false,
      ...config.identity?.login_id?.types?.email,
    },
    username: {
      block_reserved_usernames: true,
      exclude_keywords_enabled: false,
      ascii_only: true,
      case_sensitive: false,
      modify_disabled:
        config.identity?.login_id?.keys?.find((a) => a.type === "username")
          ?.modify_disabled ?? false,
      ...config.identity?.login_id?.types?.username,
    },
    phone: {
      allowlist: [],
      pinned_list: [],
      modify_disabled:
        config.identity?.login_id?.keys?.find((a) => a.type === "phone")
          ?.modify_disabled ?? false,
      ...config.ui?.country_calling_code,
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: ConfigFormState,
  currentState: ConfigFormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.identity ??= {};
    config.identity.login_id ??= {};
    config.identity.login_id.keys ??= [];
    config.identity.login_id.types ??= {};
    config.identity.login_id.types.username ??= {};
    config.identity.login_id.types.email ??= {};
    config.ui ??= {};
    config.ui.country_calling_code ??= {};

    const keys = new Map(config.identity.login_id.keys.map((k) => [k.type, k]));
    config.identity.login_id.keys = currentState.types
      .filter((t) => t.isEnabled)
      .map((t) => keys.get(t.type) ?? { type: t.type, key: t.type });

    if (currentState.types.find((t) => t.type === "email")?.isEnabled) {
      const emailConfig = config.identity.login_id.types.email;
      if (
        initialState.email.modify_disabled !==
        currentState.email.modify_disabled
      ) {
        const keyConfig = config.identity.login_id.keys.find(
          (a) => a.type === "email"
        );
        if (keyConfig != null) {
          keyConfig.modify_disabled = currentState.email.modify_disabled;
        }
      }
      if (
        initialState.email.block_plus_sign !==
        currentState.email.block_plus_sign
      ) {
        emailConfig.block_plus_sign = currentState.email.block_plus_sign;
      }
      if (
        initialState.email.case_sensitive !== currentState.email.case_sensitive
      ) {
        emailConfig.case_sensitive = currentState.email.case_sensitive;
      }
      if (
        initialState.email.ignore_dot_sign !==
        currentState.email.ignore_dot_sign
      ) {
        emailConfig.ignore_dot_sign = currentState.email.ignore_dot_sign;
      }

      if (
        initialState.email.domain_blocklist_enabled !==
        currentState.email.domain_blocklist_enabled
      ) {
        emailConfig.domain_blocklist_enabled =
          currentState.email.domain_blocklist_enabled;
      }

      if (
        initialState.email.block_free_email_provider_domains !==
        currentState.email.block_free_email_provider_domains
      ) {
        emailConfig.block_free_email_provider_domains =
          currentState.email.block_free_email_provider_domains;
      }

      if (
        initialState.email.domain_allowlist_enabled !==
        currentState.email.domain_allowlist_enabled
      ) {
        emailConfig.domain_allowlist_enabled =
          currentState.email.domain_allowlist_enabled;
      }
    }

    if (currentState.types.find((t) => t.type === "username")?.isEnabled) {
      const usernameConfig = config.identity.login_id.types.username;
      if (
        initialState.username.modify_disabled !==
        currentState.username.modify_disabled
      ) {
        const keyConfig = config.identity.login_id.keys.find(
          (a) => a.type === "username"
        );
        if (keyConfig != null) {
          keyConfig.modify_disabled = currentState.username.modify_disabled;
        }
      }
      if (
        initialState.username.block_reserved_usernames !==
        currentState.username.block_reserved_usernames
      ) {
        usernameConfig.block_reserved_usernames =
          currentState.username.block_reserved_usernames;
      }
      if (
        initialState.username.exclude_keywords_enabled !==
        currentState.username.exclude_keywords_enabled
      ) {
        usernameConfig.exclude_keywords_enabled =
          currentState.username.exclude_keywords_enabled;
      }

      if (
        initialState.username.ascii_only !== currentState.username.ascii_only
      ) {
        usernameConfig.ascii_only = currentState.username.ascii_only;
      }
      if (
        initialState.username.case_sensitive !==
        currentState.username.case_sensitive
      ) {
        usernameConfig.case_sensitive = currentState.username.case_sensitive;
      }
    }

    if (currentState.types.find((t) => t.type === "phone")?.isEnabled) {
      if (
        initialState.phone.modify_disabled !==
        currentState.phone.modify_disabled
      ) {
        const keyConfig = config.identity.login_id.keys.find(
          (a) => a.type === "phone"
        );
        if (keyConfig != null) {
          keyConfig.modify_disabled = currentState.phone.modify_disabled;
        }
      }
      const phoneConfig = config.ui.country_calling_code;
      if (
        !deepEqual(initialState.phone.allowlist, currentState.phone.allowlist, {
          strict: true,
        })
      ) {
        phoneConfig.allowlist = currentState.phone.allowlist;
      }
      if (
        !deepEqual(
          initialState.phone.pinned_list,
          currentState.phone.pinned_list,
          { strict: true }
        )
      ) {
        phoneConfig.pinned_list = currentState.phone.pinned_list;
      }
    }

    function hasLoginIDTypes(s: ConfigFormState) {
      return s.types.some((t) => t.isEnabled);
    }
    if (hasLoginIDTypes(initialState) !== hasLoginIDTypes(currentState)) {
      const identities = (
        effectiveConfig.authentication?.identities ?? []
      ).slice();
      const index = identities.indexOf("login_id");
      const isEnabled = hasLoginIDTypes(currentState);

      if (isEnabled && index === -1) {
        identities.push("login_id");
      } else if (!isEnabled && index >= 0) {
        identities.splice(index, 1);
      }
      config.authentication ??= {};
      config.authentication.identities = identities;
    }

    clearEmptyObject(config);
  });
}

interface ResourcesFormState {
  resources: Partial<Record<string, Resource>>;
}

function constructResourcesFormState(
  resources: Resource[]
): ResourcesFormState {
  const resourceMap: Partial<Record<string, Resource>> = {};
  for (const r of resources) {
    const id = specifierId(r.specifier);
    // Multiple resources may use same specifier ID (images),
    // use the first resource with non-empty values.
    if ((resourceMap[id]?.nullableValue ?? "") === "") {
      resourceMap[specifierId(r.specifier)] = r;
    }
  }

  return { resources: resourceMap };
}

function constructResources(state: ResourcesFormState): Resource[] {
  return Object.values(state.resources).filter(Boolean) as Resource[];
}

interface FormState
  extends ConfigFormState,
    ResourcesFormState,
    FeatureConfigFormState {}

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

function validateForm(
  state: ConfigFormState,
  renderToString: (id: string) => string
) {
  const errors: LocalValidationError[] = [];
  if (state.phone.allowlist.length === 0) {
    errors.push({
      message: renderToString(
        "LoginIDConfigurationScreen.error.calling-code-min-items"
      ),
    });
  }

  return makeLocalValidationError(errors);
}

const switchStyle = { root: { margin: "0" } };

interface LoginIDTypeEditProps {
  state: ConfigFormState;
  index: number;
  loginIDType: LoginIDKeyType;
  toggleLoginIDType: (type: LoginIDKeyType, isEnabled: boolean) => void;
  swapPosition: (index1: number, index2: number) => void;
  featureDisabled: boolean;
}

const LoginIDTypeEdit: React.FC<LoginIDTypeEditProps> =
  function LoginIDTypeEdit(props) {
    const {
      index,
      loginIDType,
      toggleLoginIDType,
      swapPosition,
      state,
      featureDisabled,
    } = props;
    const { renderToString } = useContext(Context);

    const isEnabled =
      state.types.find((t) => t.type === loginIDType)?.isEnabled ?? false;
    const onToggleIsEnabled = useCallback(
      (_, isEnabled?: boolean) =>
        toggleLoginIDType(loginIDType, isEnabled ?? false),
      [toggleLoginIDType, loginIDType]
    );

    const titleId = {
      email: "LoginIDConfigurationScreen.email.title",
      username: "LoginIDConfigurationScreen.username.title",
      phone: "LoginIDConfigurationScreen.phone.title",
    }[loginIDType];

    const renderAriaLabel = useCallback(
      () => renderToString(titleId),
      [renderToString, titleId]
    );

    const widgetHeader = useMemo(
      () => (
        <Toggle
          label={<FormattedMessage id={titleId} />}
          inlineLabel={true}
          styles={switchStyle}
          checked={isEnabled}
          onChange={onToggleIsEnabled}
          disabled={!isEnabled && featureDisabled}
        />
      ),
      [titleId, isEnabled, onToggleIsEnabled, featureDisabled]
    );

    const widgetMessageHeader = useMemo(
      () =>
        featureDisabled && (
          <MessageBar>
            <FormattedMessage
              id="FeatureConfig.disabled"
              values={{
                HREF: "../settings/subscription",
              }}
            />
          </MessageBar>
        ),
      [featureDisabled]
    );

    return (
      <WidgetWithOrdering
        className={styles.widget}
        index={index}
        itemCount={loginIDKeyTypes.length}
        onSwapClicked={swapPosition}
        readOnly={!isEnabled || featureDisabled}
        renderAriaLabel={renderAriaLabel}
        HeaderComponent={widgetHeader}
        HeaderMessageComponent={widgetMessageHeader}
      >
        {props.children}
      </WidgetWithOrdering>
    );
  };

interface AuthenticationLoginIDSettingsContentProps {
  form: FormModel;
}

const AuthenticationLoginIDSettingsContent: React.FC<AuthenticationLoginIDSettingsContentProps> =
  function AuthenticationLoginIDSettingsContent(props) {
    const { state, setState } = props.form;

    const { renderToString } = useContext(Context);

    const swapPosition = useCallback(
      (index1: number, index2: number) => {
        setState((state) =>
          produce(state, (state) => {
            const tmp = state.types[index1];
            state.types[index1] = state.types[index2];
            state.types[index2] = tmp;
          })
        );
      },
      [setState]
    );

    const toggleLoginIDType = useCallback(
      (loginIDType: LoginIDKeyType, isEnabled: boolean) => {
        setState((state) =>
          produce(state, (state) => {
            const type = state.types.find((t) => t.type === loginIDType);
            if (type) {
              type.isEnabled = isEnabled;
            }
          })
        );
      },
      [setState]
    );

    const change = useCallback(
      (fn: (state: ConfigFormState) => void) =>
        setState((state) =>
          produce(state, (state) => {
            fn(state);
          })
        ),
      [setState]
    );

    const onEmailModifyDisabledChange = useCallback(
      (_, value?: boolean) => {
        change((state) => {
          state.email.modify_disabled = value ?? false;
        });
      },
      [change]
    );
    const onEmailCaseSensitiveChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.email.case_sensitive = value ?? false;
        }),
      [change]
    );
    const onEmailIgnoreDotLocalChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.email.ignore_dot_sign = value ?? false;
        }),
      [change]
    );
    const onEmailBlockPlusChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.email.block_plus_sign = value ?? false;
        }),
      [change]
    );

    const onEmailDomainBlocklistEnabledChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.email.domain_blocklist_enabled = value ?? false;
          if (state.email.domain_blocklist_enabled) {
            state.email.domain_allowlist_enabled = false;
          } else {
            state.email.block_free_email_provider_domains = false;
          }
        }),
      [change]
    );
    const onEmailBlockFreeEmailProviderDomainsChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.email.block_free_email_provider_domains = value ?? false;
          if (state.email.block_free_email_provider_domains) {
            state.email.domain_blocklist_enabled = true;
            state.email.domain_allowlist_enabled = false;
          }
        }),
      [change]
    );
    const onEmailDomainAllowlistEnabledChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.email.domain_allowlist_enabled = value ?? false;
          if (state.email.domain_allowlist_enabled) {
            state.email.domain_blocklist_enabled = false;
            state.email.block_free_email_provider_domains = false;
          }
        }),
      [change]
    );

    const valueForDomainBlocklist = useMemo(() => {
      const value =
        state.resources[specifierId(emailDomainBlocklistSpecifier)]
          ?.nullableValue;
      if (value == null) {
        return [];
      }
      return splitByNewline(value);
    }, [state.resources]);

    const valueForDomainAllowlist = useMemo(() => {
      const value =
        state.resources[specifierId(emailDomainAllowlistSpecifier)]
          ?.nullableValue;
      if (value == null) {
        return [];
      }
      return splitByNewline(value);
    }, [state.resources]);

    const updateEmailDomainBlocklist = useCallback(
      (value: string[]) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const specifier = emailDomainBlocklistSpecifier;
          const newResource: Resource = {
            specifier,
            path: renderPath(specifier.def.resourcePath, {}),
            nullableValue: joinByNewline(value),
          };
          updatedResources[specifierId(newResource.specifier)] = newResource;
          return {
            ...prev,
            resources: updatedResources,
          };
        });
      },
      [setState]
    );

    const addDomainToEmailDomainBlocklist = useCallback(
      (value: string) => {
        updateEmailDomainBlocklist([...valueForDomainBlocklist, value]);
      },
      [valueForDomainBlocklist, updateEmailDomainBlocklist]
    );

    const updateEmailDomainAllowlist = useCallback(
      (value: string[]) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const specifier = emailDomainAllowlistSpecifier;
          const newResource: Resource = {
            specifier,
            path: renderPath(specifier.def.resourcePath, {}),
            nullableValue: joinByNewline(value),
          };
          updatedResources[specifierId(newResource.specifier)] = newResource;
          return {
            ...prev,
            resources: updatedResources,
          };
        });
      },
      [setState]
    );

    const addDomainToEmailDomainAllowlist = useCallback(
      (value: string) => {
        updateEmailDomainAllowlist([...valueForDomainAllowlist, value]);
      },
      [valueForDomainAllowlist, updateEmailDomainAllowlist]
    );

    const {
      selectedItems: domainBlocklist,
      onChange: onDomainBlocklistChange,
      onResolveSuggestions: onDomainBlocklistSuggestions,
    } = useTagPickerWithNewTags(
      valueForDomainBlocklist,
      updateEmailDomainBlocklist
    );

    const {
      selectedItems: domainAllowlist,
      onChange: onDomainAllowlistChange,
      onResolveSuggestions: onDomainAllowlistSuggestions,
    } = useTagPickerWithNewTags(
      valueForDomainAllowlist,
      updateEmailDomainAllowlist
    );

    const emailSection = (
      <div className={styles.widgetContent}>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.email.caseSensitive"
          )}
          className={styles.control}
          checked={state.email.case_sensitive}
          onChange={onEmailCaseSensitiveChange}
        />
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.email.ignoreDotLocal"
          )}
          className={styles.control}
          checked={state.email.ignore_dot_sign}
          onChange={onEmailIgnoreDotLocalChange}
        />
        <CheckboxWithTooltip
          label={renderToString("LoginIDConfigurationScreen.email.blockPlus")}
          className={styles.control}
          checked={state.email.block_plus_sign}
          onChange={onEmailBlockPlusChange}
          tooltipMessageId="LoginIDConfigurationScreen.email.blockPlusTooltipMessage"
        />
        <CheckboxWithContentLayout className={styles.control}>
          <CheckboxWithTooltip
            label={renderToString(
              "LoginIDConfigurationScreen.email.domainBlocklist"
            )}
            checked={state.email.domain_blocklist_enabled}
            onChange={onEmailDomainBlocklistEnabledChange}
            disabled={state.email.domain_allowlist_enabled}
            tooltipMessageId="LoginIDConfigurationScreen.email.domainBlocklistTooltipMessage"
          />
          <CustomTagPicker
            inputProps={{
              "aria-label": renderToString(
                "LoginIDConfigurationScreen.email.domainBlocklist"
              ),
            }}
            className={styles.widgetInputField}
            disabled={!state.email.domain_blocklist_enabled}
            selectedItems={domainBlocklist}
            onChange={onDomainBlocklistChange}
            onResolveSuggestions={onDomainBlocklistSuggestions}
            onAdd={addDomainToEmailDomainBlocklist}
          />
        </CheckboxWithContentLayout>
        <CheckboxWithTooltip
          label={renderToString(
            "LoginIDConfigurationScreen.email.blockFreeEmailProviderDomains"
          )}
          className={styles.control}
          checked={state.email.block_free_email_provider_domains}
          onChange={onEmailBlockFreeEmailProviderDomainsChange}
          disabled={state.email.domain_allowlist_enabled}
          tooltipMessageId="LoginIDConfigurationScreen.email.blockFreeEmailProviderDomainsTooltipMessage"
        />
        <CheckboxWithContentLayout className={styles.control}>
          <CheckboxWithTooltip
            label={renderToString(
              "LoginIDConfigurationScreen.email.domainAllowlist"
            )}
            checked={state.email.domain_allowlist_enabled}
            onChange={onEmailDomainAllowlistEnabledChange}
            disabled={state.email.domain_blocklist_enabled}
            tooltipMessageId="LoginIDConfigurationScreen.email.domainAllowlistTooltipMessage"
          />
          <CustomTagPicker
            inputProps={{
              "aria-label": renderToString(
                "LoginIDConfigurationScreen.email.domainAllowlist"
              ),
            }}
            className={styles.widgetInputField}
            disabled={!state.email.domain_allowlist_enabled}
            selectedItems={domainAllowlist}
            onChange={onDomainAllowlistChange}
            onResolveSuggestions={onDomainAllowlistSuggestions}
            onAdd={addDomainToEmailDomainAllowlist}
          />
        </CheckboxWithContentLayout>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.email.modify-disabled"
          )}
          className={styles.control}
          checked={state.email.modify_disabled}
          onChange={onEmailModifyDisabledChange}
        />
      </div>
    );

    const onUsernameModifyDisabledChange = useCallback(
      (_, value?: boolean) => {
        change((state) => {
          state.username.modify_disabled = value ?? false;
        });
      },
      [change]
    );
    const onUsernameBlockReservedUsernameChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.username.block_reserved_usernames = value ?? false;
        }),
      [change]
    );
    const onUsernameCaseSensitiveChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.username.case_sensitive = value ?? false;
        }),
      [change]
    );
    const onUsernameASCIIOnlyChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.username.ascii_only = value ?? false;
        }),
      [change]
    );
    const onUsernameIsExcludedKeywordsEnabledChange = useCallback(
      (_, value?: boolean) =>
        change((state) => {
          state.username.exclude_keywords_enabled = value ?? false;
        }),
      [change]
    );

    const valueForUsernameExcludedKeywords = useMemo(() => {
      const value =
        state.resources[specifierId(usernameExcludeKeywordsTXTSpecifier)]
          ?.nullableValue;
      if (value == null) {
        return [];
      }
      return splitByNewline(value);
    }, [state.resources]);

    const updateUsernameExcludeKeywords = useCallback(
      (value: string[]) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const specifier = usernameExcludeKeywordsTXTSpecifier;
          const newResource: Resource = {
            specifier,
            path: renderPath(specifier.def.resourcePath, {}),
            nullableValue: joinByNewline(value),
          };
          updatedResources[specifierId(newResource.specifier)] = newResource;
          return {
            ...prev,
            resources: updatedResources,
          };
        });
      },
      [setState]
    );

    const addKeywordToUsernameExcludeKeywords = useCallback(
      (value: string) => {
        updateUsernameExcludeKeywords([
          ...valueForUsernameExcludedKeywords,
          value,
        ]);
      },
      [valueForUsernameExcludedKeywords, updateUsernameExcludeKeywords]
    );

    const {
      selectedItems: excludedKeywordItems,
      onChange: onExcludedKeywordsChange,
      onResolveSuggestions: onResolveExcludedKeywordSuggestions,
    } = useTagPickerWithNewTags(
      valueForUsernameExcludedKeywords,
      updateUsernameExcludeKeywords
    );
    const usernameSection = (
      <div className={styles.widgetContent}>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.username.blockReservedUsername"
          )}
          checked={state.username.block_reserved_usernames}
          onChange={onUsernameBlockReservedUsernameChange}
          className={styles.control}
        />
        <CheckboxWithContentLayout className={styles.control}>
          <CheckboxWithTooltip
            label={renderToString(
              "LoginIDConfigurationScreen.username.excludeKeywords"
            )}
            checked={state.username.exclude_keywords_enabled}
            onChange={onUsernameIsExcludedKeywordsEnabledChange}
            tooltipMessageId="LoginIDConfigurationScreen.username.excludeKeywordsTooltipMessage"
          />
          <CustomTagPicker
            inputProps={{
              "aria-label": renderToString(
                "LoginIDConfigurationScreen.username.excludeKeywords"
              ),
            }}
            className={styles.widgetInputField}
            disabled={!state.username.exclude_keywords_enabled}
            selectedItems={excludedKeywordItems}
            onChange={onExcludedKeywordsChange}
            onResolveSuggestions={onResolveExcludedKeywordSuggestions}
            onAdd={addKeywordToUsernameExcludeKeywords}
          />
        </CheckboxWithContentLayout>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.username.caseSensitive"
          )}
          className={styles.control}
          checked={state.username.case_sensitive}
          onChange={onUsernameCaseSensitiveChange}
        />
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.username.asciiOnly"
          )}
          className={styles.control}
          checked={state.username.ascii_only}
          onChange={onUsernameASCIIOnlyChange}
        />
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.username.modify-disabled"
          )}
          className={styles.control}
          checked={state.username.modify_disabled}
          onChange={onUsernameModifyDisabledChange}
        />
      </div>
    );

    const onPhoneModifyDisabledChange = useCallback(
      (_, value?: boolean) => {
        change((state) => {
          state.phone.modify_disabled = value ?? false;
        });
      },
      [change]
    );
    const onPhoneListChange = useCallback(
      (allowlist: string[], pinnedList: string[]) =>
        change((state) => {
          state.phone.allowlist = allowlist;
          state.phone.pinned_list = pinnedList;
        }),
      [change]
    );
    const phoneSection = (
      <div className={styles.widgetContent}>
        <Widget className={styles.control}>
          <CountryCallingCodeList
            allCountryCallingCodes={supportedCountryCallingCodes}
            selectedCountryCallingCodes={state.phone.allowlist}
            pinnedCountryCallingCodes={state.phone.pinned_list}
            onChange={onPhoneListChange}
          />
        </Widget>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.phone.modify-disabled"
          )}
          className={styles.control}
          checked={state.phone.modify_disabled}
          onChange={onPhoneModifyDisabledChange}
        />
      </div>
    );

    const sections = {
      email: emailSection,
      username: usernameSection,
      phone: phoneSection,
    };

    return (
      <ScreenContent className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="LoginIDConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="LoginIDConfigurationScreen.columns.orderTooltipMessage" />
        </ScreenDescription>
        {state.types.map(({ type }, index) => (
          <LoginIDTypeEdit
            key={type}
            state={state}
            index={index}
            loginIDType={type}
            toggleLoginIDType={toggleLoginIDType}
            swapPosition={swapPosition}
            featureDisabled={type === "phone" && state.loginIDPhoneDisabled}
          >
            {sections[type]}
          </LoginIDTypeEdit>
        ))}
      </ScreenContent>
    );
  };

const LoginIDConfigurationScreen: React.FC =
  function LoginIDConfigurationScreen() {
    const { appID } = useParams();
    const { renderToString } = useContext(Context);

    const config = useAppConfigForm(
      appID,
      constructConfigFormState,
      constructConfig
    );
    const localValidationError = validateForm(config.state, renderToString);

    const resources = useResourceForm(
      appID,
      specifiers,
      constructResourcesFormState,
      constructResources
    );

    const featureConfig = useAppFeatureConfigQuery(appID);

    const state = useMemo<FormState>(
      () => ({
        resources: resources.state.resources,
        types: config.state.types,
        email: config.state.email,
        username: config.state.username,
        phone: config.state.phone,
        loginIDPhoneDisabled:
          featureConfig.effectiveFeatureConfig?.identity?.login_id?.types?.phone
            ?.disabled ?? false,
      }),
      [
        resources.state.resources,
        config.state.types,
        config.state.email,
        config.state.username,
        config.state.phone,
        featureConfig.effectiveFeatureConfig?.identity?.login_id?.types?.phone
          ?.disabled,
      ]
    );

    const form: FormModel = {
      isLoading:
        config.isLoading || resources.isLoading || featureConfig.loading,
      isUpdating: config.isUpdating || resources.isUpdating,
      isDirty: config.isDirty || resources.isDirty,
      loadError: config.loadError ?? resources.loadError ?? featureConfig.error,
      updateError: config.updateError ?? resources.updateError,
      state,
      setState: (fn) => {
        const newState = fn(state);
        config.setState(() => ({
          types: newState.types,
          email: newState.email,
          username: newState.username,
          phone: newState.phone,
        }));
        resources.setState(() => ({ resources: newState.resources }));
      },
      reload: () => {
        config.reload();
        resources.reload();
        featureConfig.refetch().finally(() => {});
      },
      reset: () => {
        config.reset();
        resources.reset();
      },
      save: async () => {
        await config.save();
        await resources.save();
      },
    };

    if (form.isLoading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <FormContainer form={form} localError={localValidationError}>
        <AuthenticationLoginIDSettingsContent form={form} />
      </FormContainer>
    );
  };

export default LoginIDConfigurationScreen;
