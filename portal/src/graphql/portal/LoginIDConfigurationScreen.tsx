import React, { useCallback, useContext, useMemo } from "react";
import produce from "immer";
import { Checkbox, MessageBar, MessageBarType, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import Widget from "../../Widget";
import WidgetWithOrdering from "../../WidgetWithOrdering";
import CheckboxWithContentLayout from "../../CheckboxWithContentLayout";
import PhoneInputListWidget from "./PhoneInputListWidget";
import { useTagPickerWithNewTags } from "../../hook/useInput";
import { clearEmptyObject } from "../../util/misc";
import { useParams } from "react-router-dom";
import { useAppConfigForm } from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import FormContainer from "../../FormContainer";
import {
  LoginIDEmailConfig,
  LoginIDKeyType,
  loginIDKeyTypes,
  LoginIDUsernameConfig,
  PortalAPIAppConfig,
  PhoneInputConfig,
} from "../../types";
import {
  DEFAULT_TEMPLATE_LOCALE,
  RESOURCE_EMAIL_DOMAIN_BLOCKLIST,
  RESOURCE_EMAIL_DOMAIN_ALLOWLIST,
  RESOURCE_USERNAME_EXCLUDED_KEYWORDS_TXT,
} from "../../resources";
import { fixTagPickerStyles } from "../../bugs";
import CheckboxWithTooltip from "../../CheckboxWithTooltip";
import {
  Resource,
  ResourceSpecifier,
  specifierId,
  expandSpecifier,
} from "../../util/resource";
import { useResourceForm } from "../../hook/useResourceForm";
import CustomTagPicker from "../../CustomTagPicker";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import Toggle from "../../Toggle";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { makeValidationErrorMatchUnknownKindParseRule } from "../../error/parse";
import ALL_COUNTRIES from "../../data/country.json";

import styles from "./LoginIDConfigurationScreen.module.css";

const errorRules = [
  makeValidationErrorMatchUnknownKindParseRule(
    "minItems",
    /^\/ui\/phone_input\/allowlist$/,
    "LoginIDConfigurationScreen.error.calling-code-min-items"
  ),
  makeValidationErrorMatchUnknownKindParseRule(
    "const",
    /\/authentication\/identities/,
    "errors.validation.passkey"
  ),
];

// email domain lists are not language specific
// so the locale in ResourceSpecifier is not important
const emailDomainBlocklistSpecifier: ResourceSpecifier = {
  def: RESOURCE_EMAIL_DOMAIN_BLOCKLIST,
  locale: DEFAULT_TEMPLATE_LOCALE,
  extension: null,
};

const emailDomainAllowlistSpecifier: ResourceSpecifier = {
  def: RESOURCE_EMAIL_DOMAIN_ALLOWLIST,
  locale: DEFAULT_TEMPLATE_LOCALE,
  extension: null,
};

const usernameExcludeKeywordsTXTSpecifier: ResourceSpecifier = {
  def: RESOURCE_USERNAME_EXCLUDED_KEYWORDS_TXT,
  locale: DEFAULT_TEMPLATE_LOCALE,
  extension: null,
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

interface PhoneConfig extends PhoneInputConfig {
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

function constructFormState(config: PortalAPIAppConfig): ConfigFormState {
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
      preselect_by_ip_disabled: false,
      modify_disabled:
        config.identity?.login_id?.keys?.find((a) => a.type === "phone")
          ?.modify_disabled ?? false,
      ...config.ui?.phone_input,
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: ConfigFormState,
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
    config.ui.phone_input ??= {};

    const keys = new Map(config.identity.login_id.keys.map((k) => [k.type, k]));
    config.identity.login_id.keys = currentState.types
      .filter((t) => t.isEnabled)
      .map((t) => keys.get(t.type) ?? { type: t.type, key: t.type });

    if (currentState.types.find((t) => t.type === "email")?.isEnabled) {
      const emailConfig = config.identity.login_id.types.email;

      const keyConfig = config.identity.login_id.keys.find(
        (a) => a.type === "email"
      );
      if (keyConfig != null) {
        keyConfig.modify_disabled = currentState.email.modify_disabled;
      }

      emailConfig.block_plus_sign = currentState.email.block_plus_sign;
      emailConfig.case_sensitive = currentState.email.case_sensitive;
      emailConfig.ignore_dot_sign = currentState.email.ignore_dot_sign;
      emailConfig.domain_blocklist_enabled =
        currentState.email.domain_blocklist_enabled;
      emailConfig.block_free_email_provider_domains =
        currentState.email.block_free_email_provider_domains;
      emailConfig.domain_allowlist_enabled =
        currentState.email.domain_allowlist_enabled;
    }

    if (currentState.types.find((t) => t.type === "username")?.isEnabled) {
      const usernameConfig = config.identity.login_id.types.username;

      const keyConfig = config.identity.login_id.keys.find(
        (a) => a.type === "username"
      );
      if (keyConfig != null) {
        keyConfig.modify_disabled = currentState.username.modify_disabled;
      }

      usernameConfig.block_reserved_usernames =
        currentState.username.block_reserved_usernames;
      usernameConfig.exclude_keywords_enabled =
        currentState.username.exclude_keywords_enabled;
      usernameConfig.ascii_only = currentState.username.ascii_only;
      usernameConfig.case_sensitive = currentState.username.case_sensitive;
    }

    if (currentState.types.find((t) => t.type === "phone")?.isEnabled) {
      const keyConfig = config.identity.login_id.keys.find(
        (a) => a.type === "phone"
      );
      if (keyConfig != null) {
        keyConfig.modify_disabled = currentState.phone.modify_disabled;
      }

      const phoneConfig = config.ui.phone_input;
      phoneConfig.preselect_by_ip_disabled =
        currentState.phone.preselect_by_ip_disabled;

      // If the allowlist is the original one, we instead reset it to undefined.
      // This avoids the config storing the default value, and also
      // enable us to add more countries.
      if (currentState.phone.allowlist.length === ALL_COUNTRIES.length) {
        phoneConfig.allowlist = undefined;
      } else {
        phoneConfig.allowlist = currentState.phone.allowlist;
      }

      // If the pinned list is empty, we instead reset it to undefined.
      if (currentState.phone.pinned_list.length === 0) {
        phoneConfig.pinned_list = undefined;
      } else {
        phoneConfig.pinned_list = currentState.phone.pinned_list;
      }
    }

    function hasLoginIDTypes(s: ConfigFormState) {
      return s.types.some((t) => t.isEnabled);
    }

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

const switchStyle = { root: { margin: "0" } };

interface LoginIDTypeEditProps {
  index: number;
  loginIDType: LoginIDKeyType;
  toggleLoginIDType: (type: LoginIDKeyType, isEnabled: boolean) => void;
  swapPosition: (index1: number, index2: number) => void;
  featureDisabled: boolean;
  isEnabled: boolean;
  disabled?: boolean;
  children?: React.ReactNode;
}

const LoginIDTypeEdit: React.VFC<LoginIDTypeEditProps> =
  function LoginIDTypeEdit(props) {
    const {
      index,
      loginIDType,
      toggleLoginIDType,
      swapPosition,
      featureDisabled,
      isEnabled,
      disabled = false,
    } = props;

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

    const widgetHeader = useMemo(
      () => (
        <Toggle
          label={<FormattedMessage id={titleId} />}
          inlineLabel={true}
          styles={switchStyle}
          checked={isEnabled}
          onChange={onToggleIsEnabled}
          disabled={featureDisabled || disabled}
        />
      ),
      [titleId, isEnabled, onToggleIsEnabled, featureDisabled, disabled]
    );

    const widgetMessageHeader = useMemo(
      () =>
        featureDisabled && (
          <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
        ),
      [featureDisabled]
    );

    return (
      <WidgetWithOrdering
        className={styles.widget}
        disabled={!isEnabled || featureDisabled || disabled}
        index={index}
        itemCount={loginIDKeyTypes.length}
        onSwapClicked={swapPosition}
        HeaderComponent={widgetHeader}
        HeaderMessageComponent={widgetMessageHeader}
      >
        {props.children}
      </WidgetWithOrdering>
    );
  };

interface AuthenticationLoginIDSettingsContentProps {
  form: FormModel;
  disabled?: boolean;
}

const AuthenticationLoginIDSettingsContent: React.VFC<AuthenticationLoginIDSettingsContentProps> =
  // eslint-disable-next-line complexity
  function AuthenticationLoginIDSettingsContent(props) {
    const { disabled = false } = props;
    const { state, setState } = props.form;

    const emailIsEnabled =
      state.types.find((t) => t.type === "email")?.isEnabled ?? false;
    const phoneIsEnabled =
      state.types.find((t) => t.type === "phone")?.isEnabled ?? false;
    const usernameIsEnabled =
      state.types.find((t) => t.type === "username")?.isEnabled ?? false;

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
            path: expandSpecifier(specifier),
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

    const updateEmailDomainAllowlist = useCallback(
      (value: string[]) => {
        setState((prev) => {
          const updatedResources = { ...prev.resources };
          const specifier = emailDomainAllowlistSpecifier;
          const newResource: Resource = {
            specifier,
            path: expandSpecifier(specifier),
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

    const {
      selectedItems: domainBlocklist,
      onChange: onDomainBlocklistChange,
      onResolveSuggestions: onDomainBlocklistSuggestions,
      onAdd: onDomainBlocklistAdd,
    } = useTagPickerWithNewTags(
      valueForDomainBlocklist,
      updateEmailDomainBlocklist
    );

    const {
      selectedItems: domainAllowlist,
      onChange: onDomainAllowlistChange,
      onResolveSuggestions: onDomainAllowlistSuggestions,
      onAdd: onDomainAllowlistAdd,
    } = useTagPickerWithNewTags(
      valueForDomainAllowlist,
      updateEmailDomainAllowlist
    );

    const emailSection = (
      <>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.email.caseSensitive"
          )}
          checked={state.email.case_sensitive}
          onChange={onEmailCaseSensitiveChange}
          disabled={!emailIsEnabled}
        />
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.email.ignoreDotLocal"
          )}
          checked={state.email.ignore_dot_sign}
          onChange={onEmailIgnoreDotLocalChange}
          disabled={!emailIsEnabled}
        />
        <CheckboxWithTooltip
          label={renderToString("LoginIDConfigurationScreen.email.blockPlus")}
          checked={state.email.block_plus_sign}
          onChange={onEmailBlockPlusChange}
          tooltipMessageId="LoginIDConfigurationScreen.email.blockPlusTooltipMessage"
          disabled={!emailIsEnabled}
        />
        <CheckboxWithContentLayout>
          <CheckboxWithTooltip
            label={renderToString(
              "LoginIDConfigurationScreen.email.domainBlocklist"
            )}
            checked={state.email.domain_blocklist_enabled}
            onChange={onEmailDomainBlocklistEnabledChange}
            disabled={!emailIsEnabled || state.email.domain_allowlist_enabled}
            tooltipMessageId="LoginIDConfigurationScreen.email.domainBlocklistTooltipMessage"
          />
          <CustomTagPicker
            styles={fixTagPickerStyles}
            inputProps={{
              "aria-label": renderToString(
                "LoginIDConfigurationScreen.email.domainBlocklist"
              ),
            }}
            className={styles.widgetInputField}
            disabled={!emailIsEnabled || !state.email.domain_blocklist_enabled}
            selectedItems={domainBlocklist}
            onChange={onDomainBlocklistChange}
            onResolveSuggestions={onDomainBlocklistSuggestions}
            onAdd={onDomainBlocklistAdd}
          />
        </CheckboxWithContentLayout>
        <CheckboxWithTooltip
          label={renderToString(
            "LoginIDConfigurationScreen.email.blockFreeEmailProviderDomains"
          )}
          checked={state.email.block_free_email_provider_domains}
          onChange={onEmailBlockFreeEmailProviderDomainsChange}
          disabled={!emailIsEnabled || state.email.domain_allowlist_enabled}
          tooltipMessageId="LoginIDConfigurationScreen.email.blockFreeEmailProviderDomainsTooltipMessage"
        />
        <CheckboxWithContentLayout>
          <CheckboxWithTooltip
            label={renderToString(
              "LoginIDConfigurationScreen.email.domainAllowlist"
            )}
            checked={state.email.domain_allowlist_enabled}
            onChange={onEmailDomainAllowlistEnabledChange}
            disabled={!emailIsEnabled || state.email.domain_blocklist_enabled}
            tooltipMessageId="LoginIDConfigurationScreen.email.domainAllowlistTooltipMessage"
          />
          <CustomTagPicker
            styles={fixTagPickerStyles}
            inputProps={{
              "aria-label": renderToString(
                "LoginIDConfigurationScreen.email.domainAllowlist"
              ),
            }}
            className={styles.widgetInputField}
            disabled={!emailIsEnabled || !state.email.domain_allowlist_enabled}
            selectedItems={domainAllowlist}
            onChange={onDomainAllowlistChange}
            onResolveSuggestions={onDomainAllowlistSuggestions}
            onAdd={onDomainAllowlistAdd}
          />
        </CheckboxWithContentLayout>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.email.modify-disabled"
          )}
          checked={state.email.modify_disabled}
          onChange={onEmailModifyDisabledChange}
          disabled={!emailIsEnabled}
        />
      </>
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
            path: expandSpecifier(specifier),
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

    const {
      selectedItems: excludedKeywordItems,
      onChange: onExcludedKeywordsChange,
      onResolveSuggestions: onResolveExcludedKeywordSuggestions,
      onAdd: onExcludedKeywordsAdd,
    } = useTagPickerWithNewTags(
      valueForUsernameExcludedKeywords,
      updateUsernameExcludeKeywords
    );
    const usernameSection = (
      <>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.username.blockReservedUsername"
          )}
          checked={state.username.block_reserved_usernames}
          onChange={onUsernameBlockReservedUsernameChange}
          disabled={!usernameIsEnabled}
        />
        <CheckboxWithContentLayout>
          <CheckboxWithTooltip
            label={renderToString(
              "LoginIDConfigurationScreen.username.excludeKeywords"
            )}
            checked={state.username.exclude_keywords_enabled}
            onChange={onUsernameIsExcludedKeywordsEnabledChange}
            tooltipMessageId="LoginIDConfigurationScreen.username.excludeKeywordsTooltipMessage"
            disabled={!usernameIsEnabled}
          />
          <CustomTagPicker
            styles={fixTagPickerStyles}
            inputProps={{
              "aria-label": renderToString(
                "LoginIDConfigurationScreen.username.excludeKeywords"
              ),
            }}
            className={styles.widgetInputField}
            disabled={
              !usernameIsEnabled || !state.username.exclude_keywords_enabled
            }
            selectedItems={excludedKeywordItems}
            onChange={onExcludedKeywordsChange}
            onResolveSuggestions={onResolveExcludedKeywordSuggestions}
            onAdd={onExcludedKeywordsAdd}
          />
        </CheckboxWithContentLayout>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.username.caseSensitive"
          )}
          checked={state.username.case_sensitive}
          onChange={onUsernameCaseSensitiveChange}
          disabled={!usernameIsEnabled}
        />
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.username.asciiOnly"
          )}
          checked={state.username.ascii_only}
          onChange={onUsernameASCIIOnlyChange}
          disabled={!usernameIsEnabled}
        />
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.username.modify-disabled"
          )}
          checked={state.username.modify_disabled}
          onChange={onUsernameModifyDisabledChange}
          disabled={!usernameIsEnabled}
        />
      </>
    );

    const onPhonePreselectByIPDisabledChange = useCallback(
      (_, value?: boolean) => {
        change((state) => {
          state.phone.preselect_by_ip_disabled = !value;
        });
      },
      [change]
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
      (allowlist: string[], pinnedList: string[]) => {
        change((state) => {
          state.phone.allowlist = allowlist;
          state.phone.pinned_list = pinnedList;
        });
      },
      [change]
    );
    const phoneSection = (
      <>
        <Widget>
          <PhoneInputListWidget
            disabled={!phoneIsEnabled || state.loginIDPhoneDisabled}
            allowedAlpha2={state.phone.allowlist}
            pinnedAlpha2={state.phone.pinned_list}
            onChange={onPhoneListChange}
          />
        </Widget>
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.phone.preselect-by-ip"
          )}
          checked={state.phone.preselect_by_ip_disabled !== true}
          onChange={onPhonePreselectByIPDisabledChange}
          disabled={!phoneIsEnabled || state.loginIDPhoneDisabled}
        />
        <Checkbox
          label={renderToString(
            "LoginIDConfigurationScreen.phone.modify-disabled"
          )}
          checked={state.phone.modify_disabled}
          onChange={onPhoneModifyDisabledChange}
          disabled={!phoneIsEnabled || state.loginIDPhoneDisabled}
        />
      </>
    );

    const sections = {
      email: emailSection,
      username: usernameSection,
      phone: phoneSection,
    };

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="LoginIDConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="LoginIDConfigurationScreen.columns.orderTooltipMessage" />
        </ScreenDescription>
        {state.types.map(({ type, isEnabled }, index) => (
          <LoginIDTypeEdit
            key={type}
            index={index}
            loginIDType={type}
            toggleLoginIDType={toggleLoginIDType}
            swapPosition={swapPosition}
            isEnabled={isEnabled}
            disabled={disabled}
            featureDisabled={Boolean(
              type === "phone" && state.loginIDPhoneDisabled
            )}
          >
            {sections[type]}
          </LoginIDTypeEdit>
        ))}
      </ScreenContent>
    );
  };

const LoginIDConfigurationScreen: React.VFC =
  // eslint-disable-next-line complexity
  function LoginIDConfigurationScreen() {
    const { appID } = useParams() as { appID: string };

    const config = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

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

    const isSIWEEnabled = useMemo(() => {
      return (
        config.effectiveConfig.authentication?.identities?.includes("siwe") ??
        false
      );
    }, [config.effectiveConfig.authentication?.identities]);

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
      <FormContainer form={form} errorRules={errorRules}>
        {isSIWEEnabled ? (
          <MessageBar
            messageBarType={MessageBarType.warning}
            className={styles.widget}
          >
            <Text>
              <FormattedMessage id="LoginIDConfigurationScreen.siwe-enabled-warning.description" />
            </Text>
          </MessageBar>
        ) : null}
        <AuthenticationLoginIDSettingsContent
          form={form}
          disabled={isSIWEEnabled}
        />
      </FormContainer>
    );
  };

export default LoginIDConfigurationScreen;
