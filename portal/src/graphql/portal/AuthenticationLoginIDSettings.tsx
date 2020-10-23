import React, { useCallback, useContext, useMemo, useState } from "react";
import produce from "immer";
import { Checkbox, Toggle, TagPicker, Label, Text } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ErrorDialog from "../../error/ErrorDialog";
import WidgetWithOrdering from "../../WidgetWithOrdering";
import { swap } from "../../OrderButtons";
import CheckboxWithContent from "../../CheckboxWithContent";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import CountryCallingCodeList from "./AuthenticationCountryCallingCodeList";
import { useCheckbox, useTagPickerWithNewTags } from "../../hook/useInput";
import { ModifiedIndicatorPortal } from "../../ModifiedIndicatorPortal";
import {
  LoginIDKeyType,
  LoginIDKeyConfig,
  PortalAPIAppConfig,
  PortalAPIApp,
  LoginIDEmailConfig,
  LoginIDUsernameConfig,
} from "../../types";
import {
  setFieldIfChanged,
  setFieldIfListNonEmpty,
  isArrayEqualInOrder,
  clearEmptyObject,
} from "../../util/misc";
import { countryCallingCodes as supportedCountryCallingCodes } from "../../data/countryCallingCode.json";

import styles from "./AuthenticationLoginIDSettings.module.scss";

interface Props {
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
  updateAppConfigError: unknown;
}

interface WidgetHeaderProps {
  enabled: boolean;
  setEnabled: (enabled: boolean) => void;
  titleId: string;
}

type LoginIDKeyState = Record<LoginIDKeyType, boolean>;

interface AuthenticationLoginIDSettingsState {
  loginIdKeyState: LoginIDKeyState;
  loginIdKeyTypes: LoginIDKeyType[];

  excludedKeywords: string[];
  isBlockReservedUsername: boolean;
  isExcludeKeywords: boolean;
  isUsernameCaseSensitive: boolean;
  isAsciiOnly: boolean;

  isEmailCaseSensitive: boolean;
  isIgnoreDotLocal: boolean;
  isAllowPlus: boolean;

  selectedCallingCodes: Set<string>;
  pinnedCallingCodes: string[];
}

const ALL_LOGIN_ID_KEYS: LoginIDKeyType[] = ["username", "email", "phone"];
const switchStyle = { root: { margin: "0" } };

const widgetTitleMessageId: Record<LoginIDKeyType, string> = {
  username: "AuthenticationWidget.usernameTitle",
  email: "AuthenticationWidget.emailAddressTitle",
  phone: "AuthenticationWidget.phoneNumberTitle",
};

const WidgetHeader: React.FC<WidgetHeaderProps> = function (
  props: WidgetHeaderProps
) {
  const { titleId, enabled, setEnabled } = props;
  const onChange = React.useCallback(
    (_event, checked?: boolean) => {
      setEnabled(!!checked);
    },
    [setEnabled]
  );
  return (
    <div className={styles.widgetHeader}>
      <Toggle
        label={<FormattedMessage id={titleId} />}
        inlineLabel={true}
        styles={switchStyle}
        checked={enabled}
        onChange={onChange}
      />
    </div>
  );
};

function extractConfigFromLoginIdKeys(
  configLoginIdKeys: LoginIDKeyConfig[]
): {
  loginIdKeyState: LoginIDKeyState;
  loginIdKeyTypes: LoginIDKeyType[];
} {
  const configLoginIdKeyTypes = configLoginIdKeys.map((key) => key.type);
  const enabledLoginIdKeySet = new Set(configLoginIdKeyTypes);
  const loginIdKeyState = ALL_LOGIN_ID_KEYS.reduce<Partial<LoginIDKeyState>>(
    (map, key) => {
      map[key] = enabledLoginIdKeySet.has(key);
      return map;
    },
    {}
  ) as LoginIDKeyState;

  const disabledLoginKeyTypes = ALL_LOGIN_ID_KEYS.filter(
    (key) => !enabledLoginIdKeySet.has(key)
  );
  const loginIdKeyTypes = configLoginIdKeyTypes.concat(disabledLoginKeyTypes);

  return {
    loginIdKeyState,
    loginIdKeyTypes,
  };
}

function handleStringListInput(
  stringList: string[],
  options = {
    optionEnabled: true,
    useDefaultList: false,
    defaultList: [] as string[],
  }
) {
  if (!options.optionEnabled) {
    return [];
  }
  const sanitizedList = stringList.map((item) => item.trim()).filter(Boolean);
  return options.useDefaultList
    ? [...sanitizedList, ...options.defaultList]
    : sanitizedList;
}

function constructLoginIdKeyConfig(
  loginIdKeyTypes: LoginIDKeyType[],
  loginIdKeyState: LoginIDKeyState
): LoginIDKeyConfig[] {
  const enabledKeyTypes = loginIdKeyTypes.filter((key) => loginIdKeyState[key]);
  return enabledKeyTypes.map((key) => {
    return { key, type: key };
  });
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): AuthenticationLoginIDSettingsState {
  const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
  const { loginIdKeyTypes, loginIdKeyState } = extractConfigFromLoginIdKeys(
    loginIdKeys
  );

  // username widget
  const usernameConfig = appConfig?.identity?.login_id?.types?.username;
  const excludedKeywords = usernameConfig?.excluded_keywords ?? [];

  // email widget
  const emailConfig = appConfig?.identity?.login_id?.types?.email;

  // phone widget
  const selectedCallingCodes = new Set(
    appConfig?.ui?.country_calling_code?.allowlist ?? []
  );
  const pinnedCallingCodes =
    appConfig?.ui?.country_calling_code?.pinned_list ?? [];

  return {
    loginIdKeyState,
    loginIdKeyTypes,

    excludedKeywords,
    isBlockReservedUsername: !!usernameConfig?.block_reserved_usernames,
    isExcludeKeywords: excludedKeywords.length > 0,
    isUsernameCaseSensitive: !!usernameConfig?.case_sensitive,
    isAsciiOnly: !!usernameConfig?.ascii_only,

    isEmailCaseSensitive: !!emailConfig?.case_sensitive,
    isIgnoreDotLocal: !!emailConfig?.ignore_dot_sign,
    isAllowPlus: !emailConfig?.block_plus_sign,

    selectedCallingCodes,
    pinnedCallingCodes,
  };
}

function localValidate(
  renderToString: (messageID: string) => string,
  state: AuthenticationLoginIDSettingsState
): string | undefined {
  if (
    state.pinnedCallingCodes.length === 0 ||
    state.selectedCallingCodes.size === 0
  ) {
    return renderToString(
      "AuthenticationScreen.login-id.error.calling-code-min-items"
    );
  }

  return undefined;
}

function mutateUsernameConfig(
  usernameConfig: LoginIDUsernameConfig,
  initialScreenState: AuthenticationLoginIDSettingsState,
  screenState: AuthenticationLoginIDSettingsState
) {
  if (
    !isArrayEqualInOrder(
      initialScreenState.excludedKeywords,
      screenState.excludedKeywords
    )
  ) {
    const excludedKeywordList = handleStringListInput(
      screenState.excludedKeywords,
      {
        optionEnabled: screenState.isExcludeKeywords,
        useDefaultList: false,
        defaultList: [],
      }
    );

    setFieldIfListNonEmpty(
      usernameConfig,
      "excluded_keywords",
      excludedKeywordList
    );
  }
  setFieldIfChanged(
    usernameConfig,
    "case_sensitive",
    initialScreenState.isUsernameCaseSensitive,
    screenState.isUsernameCaseSensitive
  );
  setFieldIfChanged(
    usernameConfig,
    "ascii_only",
    initialScreenState.isAsciiOnly,
    screenState.isAsciiOnly
  );
}

function mutateEmailConfig(
  emailConfig: LoginIDEmailConfig,
  initialScreenState: AuthenticationLoginIDSettingsState,
  screenState: AuthenticationLoginIDSettingsState
) {
  setFieldIfChanged(
    emailConfig,
    "case_sensitive",
    initialScreenState.isEmailCaseSensitive,
    screenState.isEmailCaseSensitive
  );
  setFieldIfChanged(
    emailConfig,
    "ignore_dot_sign",
    initialScreenState.isIgnoreDotLocal,
    screenState.isIgnoreDotLocal
  );
  setFieldIfChanged(
    emailConfig,
    "block_plus_sign",
    !initialScreenState.isAllowPlus,
    !screenState.isAllowPlus
  );
}

function mutatePhoneConfig(
  appConfig: PortalAPIAppConfig,
  initialScreenState: AuthenticationLoginIDSettingsState,
  screenState: AuthenticationLoginIDSettingsState
) {
  appConfig.ui = appConfig.ui ?? {};
  appConfig.ui.country_calling_code = appConfig.ui.country_calling_code ?? {};

  if (
    !deepEqual(
      initialScreenState.selectedCallingCodes,
      screenState.selectedCallingCodes
    )
  ) {
    appConfig.ui.country_calling_code.allowlist = Array.from(
      screenState.selectedCallingCodes
    );
  }

  if (
    !deepEqual(
      initialScreenState.pinnedCallingCodes,
      screenState.pinnedCallingCodes
    )
  ) {
    appConfig.ui.country_calling_code.pinned_list =
      screenState.pinnedCallingCodes;
  }
}

function constructAppConfigFromState(
  rawAppConfig: PortalAPIAppConfig,
  initialScreenState: AuthenticationLoginIDSettingsState,
  screenState: AuthenticationLoginIDSettingsState
): PortalAPIAppConfig {
  const newAppConfig = produce(rawAppConfig, (draftConfig) => {
    draftConfig.identity = draftConfig.identity ?? {};
    draftConfig.identity.login_id = draftConfig.identity.login_id ?? {};
    draftConfig.identity.login_id.types =
      draftConfig.identity.login_id.types ?? {};

    draftConfig.identity.login_id.keys =
      draftConfig.identity.login_id.keys ?? [];

    draftConfig.identity.login_id.keys = constructLoginIdKeyConfig(
      screenState.loginIdKeyTypes,
      screenState.loginIdKeyState
    );

    const loginIdTypes = draftConfig.identity.login_id.types;

    // username config
    if (screenState.loginIdKeyState["username"]) {
      loginIdTypes.username = loginIdTypes.username ?? {};
      const usernameConfig = loginIdTypes.username;
      mutateUsernameConfig(usernameConfig, initialScreenState, screenState);
    }

    // email config
    if (screenState.loginIdKeyState["email"]) {
      loginIdTypes.email = loginIdTypes.email ?? {};
      const emailConfig = loginIdTypes.email;
      mutateEmailConfig(emailConfig, initialScreenState, screenState);
    }

    // phone config
    if (screenState.loginIdKeyState["phone"]) {
      mutatePhoneConfig(draftConfig, initialScreenState, screenState);
    }

    clearEmptyObject(draftConfig);
  });

  return newAppConfig;
}

const AuthenticationLoginIDSettings: React.FC<Props> = function AuthenticationLoginIDSettings(
  props: Props
) {
  const {
    effectiveAppConfig,
    rawAppConfig,
    updateAppConfig,
    updatingAppConfig,
    updateAppConfigError,
  } = props;
  const { renderToString } = useContext(Context);

  const initialState = useMemo(() => {
    return constructStateFromAppConfig(effectiveAppConfig);
  }, [effectiveAppConfig]);

  const [state, setState] = useState(initialState);
  const [localErrorMessage, setLocalErrorMessage] = useState<
    string | undefined
  >();

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const resetForm = useCallback(() => {
    setState(initialState);
  }, [initialState]);

  const {
    loginIdKeyState,
    loginIdKeyTypes,

    isBlockReservedUsername,
    isExcludeKeywords,
    isUsernameCaseSensitive,
    isAsciiOnly,

    isEmailCaseSensitive,
    isIgnoreDotLocal,
    isAllowPlus,

    pinnedCallingCodes,
    selectedCallingCodes,
  } = state;

  const setLoginIdKeTypeState = useCallback(
    (loginIdKeyType: LoginIDKeyType, enabled: boolean) => {
      setState((prev) => {
        return produce(prev, (draftState) => {
          draftState.loginIdKeyState[loginIdKeyType] = enabled;
        });
      });
    },
    []
  );

  // username widget
  const setUsernameEnabled = useCallback(
    (enabled: boolean) => {
      setLoginIdKeTypeState("username", enabled);
    },
    [setLoginIdKeTypeState]
  );

  const updateExcludedKeywords = useCallback((list: string[]) => {
    setState((prev) => ({
      ...prev,
      excludedKeywords: list,
    }));
  }, []);

  const {
    selectedItems: excludedKeywordItems,
    onChange: onExcludedKeywordsChange,
    onResolveSuggestions: onResolveExcludedKeywordSuggestions,
  } = useTagPickerWithNewTags(state.excludedKeywords, updateExcludedKeywords);

  const { onChange: onIsBlockReservedUsernameChange } = useCheckbox(
    (checked: boolean) => {
      setState((prev) => ({
        ...prev,
        isBlockReservedUsername: checked,
      }));
    }
  );

  const { onChange: onIsExcludeKeywordsChange } = useCheckbox(
    (checked: boolean) => {
      setState((prev) => ({
        ...prev,
        isExcludeKeywords: checked,
      }));
    }
  );

  const { onChange: onIsUsernameCaseSensitiveChange } = useCheckbox(
    (checked: boolean) => {
      setState((prev) => ({
        ...prev,
        isUsernameCaseSensitive: checked,
      }));
    }
  );

  const { onChange: onIsAsciiOnlyChange } = useCheckbox((checked: boolean) => {
    setState((prev) => ({
      ...prev,
      isAsciiOnly: checked,
    }));
  });

  // email widget
  const setEmailEnabled = useCallback(
    (enabled: boolean) => {
      setLoginIdKeTypeState("email", enabled);
    },
    [setLoginIdKeTypeState]
  );

  const { onChange: onIsEmailCaseSensitiveChange } = useCheckbox(
    (checked: boolean) => {
      setState((prev) => ({
        ...prev,
        isEmailCaseSensitive: checked,
      }));
    }
  );

  const { onChange: onIsIgnoreDotLocalChange } = useCheckbox(
    (checked: boolean) => {
      setState((prev) => ({
        ...prev,
        isIgnoreDotLocal: checked,
      }));
    }
  );

  const { onChange: onIsAllowPlusChange } = useCheckbox((checked: boolean) => {
    setState((prev) => ({
      ...prev,
      isAllowPlus: checked,
    }));
  });

  // phone widget
  const setPhoneNumberEnabled = useCallback(
    (enabled: boolean) => {
      setLoginIdKeTypeState("phone", enabled);
    },
    [setLoginIdKeTypeState]
  );

  const onSelectedCallingCodesChange = useCallback(
    (newSelectedCallingCodes: Set<string>) => {
      setState((prev) => ({
        ...prev,
        selectedCallingCodes: newSelectedCallingCodes,
      }));
    },
    []
  );

  const onPinnedCallingCodesChange = useCallback(
    (newPinnedCallingCodes: string[]) => {
      setState((prev) => ({
        ...prev,
        pinnedCallingCodes: newPinnedCallingCodes,
      }));
    },
    []
  );

  // widget order
  const renderWidgetOrderAriaLabel = useCallback(
    (index?: number) => {
      if (index == null) {
        return "";
      }
      const loginIdKeyType = loginIdKeyTypes[index];
      const messageID = widgetTitleMessageId[loginIdKeyType];
      return renderToString(messageID);
    },
    [renderToString, loginIdKeyTypes]
  );

  const onWidgetSwapClicked = useCallback((index1: number, index2: number) => {
    setState((prev) => ({
      ...prev,
      loginIdKeyTypes: swap(prev.loginIdKeyTypes, index1, index2),
    }));
  }, []);

  // on save
  const onFormSubmit = React.useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      if (rawAppConfig == null) {
        return;
      }

      const localErrorMessage = localValidate(renderToString, state);
      setLocalErrorMessage(localErrorMessage);

      if (localErrorMessage != null) {
        return;
      }

      const newAppConfig = constructAppConfigFromState(
        rawAppConfig,
        initialState,
        state
      );

      updateAppConfig(newAppConfig).catch(() => {});
    },
    [renderToString, state, rawAppConfig, updateAppConfig, initialState]
  );

  const renderUsernameWidget = useCallback(
    (index: number) => {
      return (
        <WidgetWithOrdering
          index={index}
          itemCount={ALL_LOGIN_ID_KEYS.length}
          onSwapClicked={onWidgetSwapClicked}
          readOnly={!loginIdKeyState["username"]}
          renderAriaLabel={renderWidgetOrderAriaLabel}
          HeaderComponent={
            <WidgetHeader
              enabled={loginIdKeyState["username"]}
              setEnabled={setUsernameEnabled}
              titleId={widgetTitleMessageId["username"]}
            />
          }
        >
          <div className={styles.usernameWidgetContent}>
            <Checkbox
              label={renderToString(
                "AuthenticationWidget.blockReservedUsername"
              )}
              checked={isBlockReservedUsername}
              onChange={onIsBlockReservedUsernameChange}
              className={styles.checkboxWithContent}
            />

            <CheckboxWithContent
              ariaLabel={renderToString("AuthenticationWidget.excludeKeywords")}
              checked={isExcludeKeywords}
              onChange={onIsExcludeKeywordsChange}
              className={styles.checkboxWithContent}
            >
              <Label className={styles.checkboxLabel}>
                <FormattedMessage id="AuthenticationWidget.excludeKeywords" />
              </Label>
              <TagPicker
                inputProps={{
                  "aria-label": renderToString(
                    "AuthenticationWidget.excludeKeywords"
                  ),
                }}
                className={styles.widgetInputField}
                disabled={!isExcludeKeywords}
                selectedItems={excludedKeywordItems}
                onChange={onExcludedKeywordsChange}
                onResolveSuggestions={onResolveExcludedKeywordSuggestions}
              />
            </CheckboxWithContent>

            <Checkbox
              label={renderToString("AuthenticationWidget.caseSensitive")}
              className={styles.widgetCheckbox}
              checked={isUsernameCaseSensitive}
              onChange={onIsUsernameCaseSensitiveChange}
            />

            <Checkbox
              label={renderToString("AuthenticationWidget.asciiOnly")}
              className={styles.widgetCheckbox}
              checked={isAsciiOnly}
              onChange={onIsAsciiOnlyChange}
            />
          </div>
        </WidgetWithOrdering>
      );
    },
    [
      renderToString,
      onWidgetSwapClicked,
      setUsernameEnabled,
      renderWidgetOrderAriaLabel,
      loginIdKeyState,

      excludedKeywordItems,
      isAsciiOnly,
      isUsernameCaseSensitive,
      isExcludeKeywords,
      isBlockReservedUsername,
      onExcludedKeywordsChange,
      onIsAsciiOnlyChange,
      onIsBlockReservedUsernameChange,
      onIsExcludeKeywordsChange,
      onIsUsernameCaseSensitiveChange,
      onResolveExcludedKeywordSuggestions,
    ]
  );

  const renderEmailWidget = useCallback(
    (index: number) => {
      return (
        <WidgetWithOrdering
          index={index}
          itemCount={ALL_LOGIN_ID_KEYS.length}
          onSwapClicked={onWidgetSwapClicked}
          readOnly={!loginIdKeyState["email"]}
          renderAriaLabel={renderWidgetOrderAriaLabel}
          HeaderComponent={
            <WidgetHeader
              enabled={loginIdKeyState["email"]}
              setEnabled={setEmailEnabled}
              titleId={widgetTitleMessageId["email"]}
            />
          }
        >
          <Checkbox
            label={renderToString("AuthenticationWidget.caseSensitive")}
            className={styles.widgetCheckbox}
            checked={isEmailCaseSensitive}
            onChange={onIsEmailCaseSensitiveChange}
          />

          <Checkbox
            label={renderToString("AuthenticationWidget.ignoreDotLocal")}
            className={styles.widgetCheckbox}
            checked={isIgnoreDotLocal}
            onChange={onIsIgnoreDotLocalChange}
          />

          <Checkbox
            label={renderToString("AuthenticationWidget.allowPlus")}
            className={styles.widgetCheckbox}
            checked={isAllowPlus}
            onChange={onIsAllowPlusChange}
          />
        </WidgetWithOrdering>
      );
    },
    [
      renderToString,
      onWidgetSwapClicked,
      setEmailEnabled,
      renderWidgetOrderAriaLabel,
      loginIdKeyState,

      isAllowPlus,
      isIgnoreDotLocal,
      isEmailCaseSensitive,
      onIsAllowPlusChange,
      onIsIgnoreDotLocalChange,
      onIsEmailCaseSensitiveChange,
    ]
  );

  const renderPhoneWidget = useCallback(
    (index: number) => {
      return (
        <WidgetWithOrdering
          index={index}
          itemCount={ALL_LOGIN_ID_KEYS.length}
          onSwapClicked={onWidgetSwapClicked}
          readOnly={!loginIdKeyState["phone"]}
          renderAriaLabel={renderWidgetOrderAriaLabel}
          HeaderComponent={
            <WidgetHeader
              enabled={loginIdKeyState["phone"]}
              setEnabled={setPhoneNumberEnabled}
              titleId={widgetTitleMessageId["phone"]}
            />
          }
        >
          <CountryCallingCodeList
            allCountryCallingCodes={supportedCountryCallingCodes}
            selectedCountryCallingCodes={selectedCallingCodes}
            pinnedCountryCallingCodes={pinnedCallingCodes}
            onSelectedCountryCallingCodesChange={onSelectedCallingCodesChange}
            onPinnedCountryCallingCodesChange={onPinnedCallingCodesChange}
          />
        </WidgetWithOrdering>
      );
    },
    [
      onWidgetSwapClicked,
      setPhoneNumberEnabled,
      renderWidgetOrderAriaLabel,
      loginIdKeyState,

      selectedCallingCodes,
      pinnedCallingCodes,
      onSelectedCallingCodesChange,
      onPinnedCallingCodesChange,
    ]
  );

  const loginIdWidgetRenderer: Record<
    LoginIDKeyType,
    (index: number) => React.ReactNode
  > = {
    username: renderUsernameWidget,
    email: renderEmailWidget,
    phone: renderPhoneWidget,
  };

  return (
    <form className={styles.root} onSubmit={onFormSubmit}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
      <ErrorDialog
        errorMessage={localErrorMessage}
        error={updateAppConfigError}
        rules={[]}
        fallbackErrorMessageID="AuthenticationScreen.login-id.error.generic"
      />
      <header className={styles.header}>
        <Text>
          <FormattedMessage id="AuthenticationScreen.login-id.title" />
        </Text>
        <Text>
          <FormattedMessage id="AuthenticationScreen.login-id.order" />
        </Text>
      </header>

      {loginIdKeyTypes.map((keyType, index) => (
        <div key={keyType} className={styles.widgetContainer}>
          {loginIdWidgetRenderer[keyType](index)}
        </div>
      ))}

      <ButtonWithLoading
        type="submit"
        className={styles.saveButton}
        disabled={!isFormModified}
        loading={updatingAppConfig}
        labelId="save"
        loadingLabelId="saving"
      />
    </form>
  );
};

export default AuthenticationLoginIDSettings;
