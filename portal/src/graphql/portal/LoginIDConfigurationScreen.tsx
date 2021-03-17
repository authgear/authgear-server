import React, { useCallback, useContext, useMemo } from "react";
import produce from "immer";
import { Checkbox, Label, TagPicker, Toggle } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import WidgetWithOrdering from "../../WidgetWithOrdering";
import CheckboxWithContent from "../../CheckboxWithContent";
import CountryCallingCodeList from "./AuthenticationCountryCallingCodeList";
import { useTagPickerWithNewTags } from "../../hook/useInput";
import { clearEmptyObject } from "../../util/misc";
import { countryCallingCodes as supportedCountryCallingCodes } from "../../data/countryCallingCode.json";
import { useParams } from "react-router-dom";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
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

import styles from "./LoginIDConfigurationScreen.module.scss";
import CheckboxWithTooltip from "../../CheckboxWithTooltip";

interface LoginIDKeyTypeFormState {
  isEnabled: boolean;
  type: LoginIDKeyType;
}

interface FormState {
  types: LoginIDKeyTypeFormState[];
  email: Required<LoginIDEmailConfig>;
  username: Required<LoginIDUsernameConfig>;
  phone: Required<UICountryCallingCodeConfig>;

  isUsernameExcludedKeywordEnabled: boolean;
}

function effectiveExcludedKeywords(state: FormState) {
  if (!state.isUsernameExcludedKeywordEnabled) {
    return [];
  }
  return state.username.excluded_keywords;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
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
      ...config.identity?.login_id?.types?.email,
    },
    username: {
      block_reserved_usernames: true,
      excluded_keywords: [],
      ascii_only: true,
      case_sensitive: false,
      ...config.identity?.login_id?.types?.username,
    },
    phone: {
      allowlist: [],
      pinned_list: [],
      ...config.ui?.country_calling_code,
    },
    isUsernameExcludedKeywordEnabled:
      (config.identity?.login_id?.types?.username?.excluded_keywords ?? [])
        .length > 0,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState,
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
    }

    if (currentState.types.find((t) => t.type === "username")?.isEnabled) {
      const usernameConfig = config.identity.login_id.types.username;
      if (
        initialState.username.block_reserved_usernames !==
        currentState.username.block_reserved_usernames
      ) {
        usernameConfig.block_reserved_usernames =
          currentState.username.block_reserved_usernames;
      }
      if (
        !deepEqual(
          effectiveExcludedKeywords(initialState),
          effectiveExcludedKeywords(currentState),
          { strict: true }
        )
      ) {
        usernameConfig.excluded_keywords = effectiveExcludedKeywords(
          currentState
        );
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

    function hasLoginIDTypes(s: FormState) {
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

function validateForm(
  state: FormState,
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
  state: FormState;
  index: number;
  loginIDType: LoginIDKeyType;
  toggleLoginIDType: (type: LoginIDKeyType, isEnabled: boolean) => void;
  swapPosition: (index1: number, index2: number) => void;
}

const LoginIDTypeEdit: React.FC<LoginIDTypeEditProps> = function LoginIDTypeEdit(
  props
) {
  const { index, loginIDType, toggleLoginIDType, swapPosition, state } = props;
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

  const renderAriaLabel = useCallback(() => renderToString(titleId), [
    renderToString,
    titleId,
  ]);

  const widgetHeader = useMemo(
    () => (
      <Toggle
        label={<FormattedMessage id={titleId} />}
        inlineLabel={true}
        styles={switchStyle}
        checked={isEnabled}
        onChange={onToggleIsEnabled}
      />
    ),
    [titleId, isEnabled, onToggleIsEnabled]
  );

  return (
    <WidgetWithOrdering
      className={styles.widget}
      index={index}
      itemCount={loginIDKeyTypes.length}
      onSwapClicked={swapPosition}
      readOnly={!isEnabled}
      renderAriaLabel={renderAriaLabel}
      HeaderComponent={widgetHeader}
    >
      {props.children}
    </WidgetWithOrdering>
  );
};

interface AuthenticationLoginIDSettingsContentProps {
  form: AppConfigFormModel<FormState>;
}

const AuthenticationLoginIDSettingsContent: React.FC<AuthenticationLoginIDSettingsContentProps> = function AuthenticationLoginIDSettingsContent(
  props
) {
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
    (fn: (state: FormState) => void) =>
      setState((state) =>
        produce(state, (state) => {
          fn(state);
        })
      ),
    [setState]
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
  const emailSection = (
    <div className={styles.widgetContent}>
      <Checkbox
        label={renderToString("LoginIDConfigurationScreen.email.caseSensitive")}
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
        helpText={renderToString(
          "LoginIDConfigurationScreen.email.blockPlusTooltipMessage"
        )}
      />
    </div>
  );

  const onUsernameBlockReservedUsernameChange = useCallback(
    (_, value?: boolean) =>
      change((state) => {
        state.username.block_reserved_usernames = value ?? false;
      }),
    [change]
  );
  const onUsernameExcludedKeywordsChange = useCallback(
    (value: string[]) =>
      change((state) => {
        state.username.excluded_keywords = value;
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
        state.isUsernameExcludedKeywordEnabled = value ?? false;
      }),
    [change]
  );
  const {
    selectedItems: excludedKeywordItems,
    onChange: onExcludedKeywordsChange,
    onResolveSuggestions: onResolveExcludedKeywordSuggestions,
  } = useTagPickerWithNewTags(
    state.username.excluded_keywords,
    onUsernameExcludedKeywordsChange
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
      <CheckboxWithContent
        ariaLabel={renderToString(
          "LoginIDConfigurationScreen.username.excludeKeywords"
        )}
        checked={state.isUsernameExcludedKeywordEnabled}
        onChange={onUsernameIsExcludedKeywordsEnabledChange}
        className={styles.control}
      >
        <Label className={styles.checkboxLabel}>
          <FormattedMessage id="LoginIDConfigurationScreen.username.excludeKeywords" />
        </Label>
        <TagPicker
          inputProps={{
            "aria-label": renderToString(
              "LoginIDConfigurationScreen.username.excludeKeywords"
            ),
          }}
          className={styles.widgetInputField}
          disabled={!state.isUsernameExcludedKeywordEnabled}
          selectedItems={excludedKeywordItems}
          onChange={onExcludedKeywordsChange}
          onResolveSuggestions={onResolveExcludedKeywordSuggestions}
        />
      </CheckboxWithContent>
      <Checkbox
        label={renderToString(
          "LoginIDConfigurationScreen.username.caseSensitive"
        )}
        className={styles.control}
        checked={state.username.case_sensitive}
        onChange={onUsernameCaseSensitiveChange}
      />
      <Checkbox
        label={renderToString("LoginIDConfigurationScreen.username.asciiOnly")}
        className={styles.control}
        checked={state.username.ascii_only}
        onChange={onUsernameASCIIOnlyChange}
      />
    </div>
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
      <CountryCallingCodeList
        className={styles.control}
        allCountryCallingCodes={supportedCountryCallingCodes}
        selectedCountryCallingCodes={state.phone.allowlist}
        pinnedCountryCallingCodes={state.phone.pinned_list}
        onChange={onPhoneListChange}
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
        >
          {sections[type]}
        </LoginIDTypeEdit>
      ))}
    </ScreenContent>
  );
};

const LoginIDConfigurationScreen: React.FC = function LoginIDConfigurationScreen() {
  const { appID } = useParams();
  const { renderToString } = useContext(Context);

  const form = useAppConfigForm(appID, constructFormState, constructConfig);
  const localValidationError = validateForm(form.state, renderToString);

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
