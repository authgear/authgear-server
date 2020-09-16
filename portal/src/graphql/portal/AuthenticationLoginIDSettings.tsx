import React, { useContext, useEffect, useMemo, useState } from "react";
import produce from "immer";
import { Checkbox, Toggle, TagPicker, Label } from "@fluentui/react";
import deepEqual from "deep-equal";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import CheckboxWithContent from "../../CheckboxWithContent";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { useCheckbox, useTagPickerWithNewTags } from "../../hook/useInput";
import {
  LoginIDKeyType,
  LoginIDKeyConfig,
  PortalAPIAppConfig,
  PortalAPIApp,
} from "../../types";
import {
  setFieldIfChanged,
  isArrayEqualInOrder,
  clearEmptyObject,
} from "../../util/misc";

import styles from "./AuthenticationLoginIDSettings.module.scss";

interface Props {
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
}

interface WidgetHeaderProps {
  enabled: boolean;
  setEnabled: (enabled: boolean) => void;
  titleId: string;
}

interface AuthenticationLoginIDSettingsState {
  usernameEnabled: boolean;
  emailEnabled: boolean;
  phoneNumberEnabled: boolean;

  excludedKeywords: string[];
  isBlockReservedUsername: boolean;
  isExcludeKeywords: boolean;
  isUsernameCaseSensitive: boolean;
  isAsciiOnly: boolean;

  isEmailCaseSensitive: boolean;
  isIgnoreDotLocal: boolean;
  isAllowPlus: boolean;
}

const switchStyle = { root: { margin: "0" } };

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
  loginIdKeys: LoginIDKeyConfig[]
): { [key: string]: boolean } {
  // We consider them as enabled if they are listed as allowed login ID keys.
  const usernameEnabled =
    loginIdKeys.find((key) => key.type === "username") != null;
  const emailEnabled = loginIdKeys.find((key) => key.type === "email") != null;
  const phoneNumberEnabled =
    loginIdKeys.find((key) => key.type === "phone") != null;

  return {
    usernameEnabled,
    emailEnabled,
    phoneNumberEnabled,
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

function setFieldIfListNonEmpty(
  map: Record<string, unknown>,
  field: string,
  list: (string | number | boolean)[]
): void {
  if (list.length === 0) {
    delete map[field];
  } else {
    map[field] = list;
  }
}

function getLoginIdKeyIndex(
  loginIdKeys: LoginIDKeyConfig[],
  keyType: LoginIDKeyType
): number {
  return loginIdKeys.findIndex((key: any) => key.type === keyType);
}

function updateLoginIdKey(
  loginIdKeys: LoginIDKeyConfig[],
  keyType: LoginIDKeyType,
  enabled: boolean,
  initialEnabled: boolean
) {
  if (enabled === initialEnabled) {
    return;
  }
  const loginIdKeyIndex = getLoginIdKeyIndex(loginIdKeys, keyType);
  if (enabled) {
    if (loginIdKeyIndex >= 0) {
      return;
    }
    const newLoginIdKey = { type: keyType, key: keyType };
    loginIdKeys.push(newLoginIdKey);
  }

  if (!enabled) {
    if (loginIdKeyIndex < 0) {
      return;
    }
    loginIdKeys.splice(loginIdKeyIndex, 1);
  }
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): AuthenticationLoginIDSettingsState {
  const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
  const {
    usernameEnabled,
    emailEnabled,
    phoneNumberEnabled,
  } = extractConfigFromLoginIdKeys(loginIdKeys);

  // username widget
  const usernameConfig = appConfig?.identity?.login_id?.types?.username;
  const excludedKeywords = usernameConfig?.excluded_keywords ?? [];

  // email widget
  const emailConfig = appConfig?.identity?.login_id?.types?.email;

  return {
    usernameEnabled,
    emailEnabled,
    phoneNumberEnabled,

    excludedKeywords,
    isBlockReservedUsername: !!usernameConfig?.block_reserved_usernames,
    isExcludeKeywords: excludedKeywords.length > 0,
    isUsernameCaseSensitive: !!usernameConfig?.case_sensitive,
    isAsciiOnly: !!usernameConfig?.ascii_only,

    isEmailCaseSensitive: !!emailConfig?.case_sensitive,
    isIgnoreDotLocal: !!emailConfig?.ignore_dot_sign,
    isAllowPlus: !emailConfig?.block_plus_sign,
  };
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

    const loginIdKeys = draftConfig.identity.login_id.keys;

    updateLoginIdKey(
      loginIdKeys,
      "username",
      screenState.usernameEnabled,
      initialScreenState.usernameEnabled
    );
    updateLoginIdKey(
      loginIdKeys,
      "email",
      screenState.emailEnabled,
      initialScreenState.emailEnabled
    );
    updateLoginIdKey(
      loginIdKeys,
      "phone",
      screenState.phoneNumberEnabled,
      initialScreenState.phoneNumberEnabled
    );

    const loginIdTypes = draftConfig.identity.login_id.types;

    // username config
    loginIdTypes.username = loginIdTypes.username ?? {};
    const usernameConfig = loginIdTypes.username;

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

    // email config
    loginIdTypes.email = loginIdTypes.email ?? {};
    const emailConfig = loginIdTypes.email;

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
  } = props;
  const { renderToString } = useContext(Context);

  const initialState = useMemo(() => {
    return constructStateFromAppConfig(effectiveAppConfig);
  }, [effectiveAppConfig]);

  const [isFormModified, setIsFormModified] = useState(false);

  const [usernameEnabled, setUsernameEnabled] = useState(
    initialState.usernameEnabled
  );
  const [emailEnabled, setEmailEnabled] = useState(initialState.emailEnabled);
  const [phoneNumberEnabled, setPhoneNumberEnabled] = useState(
    initialState.phoneNumberEnabled
  );

  // username widget
  const {
    list: excludedKeywords,
    onChange: onExcludedKeywordsChange,
    defaultSelectedItems: defaultSelectedExcludedKeywords,
    onResolveSuggestions: onResolveExcludedKeywordSuggestions,
  } = useTagPickerWithNewTags(initialState.excludedKeywords);
  const {
    value: isBlockReservedUsername,
    onChange: onIsBlockReservedUsernameChange,
  } = useCheckbox(initialState.isBlockReservedUsername);
  const {
    value: isExcludeKeywords,
    onChange: onIsExcludeKeywordsChange,
  } = useCheckbox(initialState.isExcludeKeywords);
  const {
    value: isUsernameCaseSensitive,
    onChange: onIsUsernameCaseSensitiveChange,
  } = useCheckbox(initialState.isUsernameCaseSensitive);
  const { value: isAsciiOnly, onChange: onIsAsciiOnlyChange } = useCheckbox(
    initialState.isAsciiOnly
  );

  // email widget
  const {
    value: isEmailCaseSensitive,
    onChange: onIsEmailCaseSensitiveChange,
  } = useCheckbox(initialState.isEmailCaseSensitive);
  const {
    value: isIgnoreDotLocal,
    onChange: onIsIgnoreDotLocalChange,
  } = useCheckbox(initialState.isIgnoreDotLocal);
  const { value: isAllowPlus, onChange: onIsAllowPlusChange } = useCheckbox(
    initialState.isAllowPlus
  );

  const screenState = useMemo(
    () => ({
      usernameEnabled,
      emailEnabled,
      phoneNumberEnabled,

      excludedKeywords,
      isBlockReservedUsername,
      isExcludeKeywords,
      isUsernameCaseSensitive,
      isAsciiOnly,

      isEmailCaseSensitive,
      isIgnoreDotLocal,
      isAllowPlus,
    }),
    [
      usernameEnabled,
      emailEnabled,
      phoneNumberEnabled,

      excludedKeywords,
      isBlockReservedUsername,
      isExcludeKeywords,
      isUsernameCaseSensitive,
      isAsciiOnly,

      isEmailCaseSensitive,
      isIgnoreDotLocal,
      isAllowPlus,
    ]
  );

  useEffect(() => {
    const equalInitial = deepEqual(initialState, screenState, { strict: true });
    setIsFormModified(!equalInitial);
  }, [initialState, screenState]);

  // on save
  const onSaveButtonClicked = React.useCallback(() => {
    if (rawAppConfig == null) {
      return;
    }

    const newAppConfig = constructAppConfigFromState(
      rawAppConfig,
      initialState,
      screenState
    );

    // TODO: handle error
    updateAppConfig(newAppConfig).catch(() => {});
  }, [screenState, rawAppConfig, updateAppConfig, initialState]);

  return (
    <div className={styles.root}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <div className={styles.widgetContainer}>
        <ExtendableWidget
          initiallyExtended={usernameEnabled}
          extendable={true}
          readOnly={!usernameEnabled}
          extendButtonAriaLabelId={"AuthenticationWidget.usernameExtend"}
          HeaderComponent={
            <WidgetHeader
              enabled={usernameEnabled}
              setEnabled={setUsernameEnabled}
              titleId={"AuthenticationWidget.usernameTitle"}
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
                onChange={onExcludedKeywordsChange}
                defaultSelectedItems={defaultSelectedExcludedKeywords}
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
        </ExtendableWidget>
      </div>

      <div className={styles.widgetContainer}>
        <ExtendableWidget
          initiallyExtended={emailEnabled}
          extendable={true}
          readOnly={!emailEnabled}
          extendButtonAriaLabelId={"AuthenticationWidget.emailAddressExtend"}
          HeaderComponent={
            <WidgetHeader
              enabled={emailEnabled}
              setEnabled={setEmailEnabled}
              titleId={"AuthenticationWidget.emailAddressTitle"}
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
        </ExtendableWidget>
      </div>

      <div className={styles.widgetContainer}>
        <ExtendableWidget
          initiallyExtended={phoneNumberEnabled}
          extendable={true}
          readOnly={!phoneNumberEnabled}
          extendButtonAriaLabelId={"AuthenticationWidget.phoneNumberExtend"}
          HeaderComponent={
            <WidgetHeader
              enabled={phoneNumberEnabled}
              setEnabled={setPhoneNumberEnabled}
              titleId={"AuthenticationWidget.phoneNumberTitle"}
            />
          }
        >
          <div>TODO: To be implemented</div>
        </ExtendableWidget>
      </div>
      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          disabled={!isFormModified}
          onClick={onSaveButtonClicked}
          loading={updatingAppConfig}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>
    </div>
  );
};

export default AuthenticationLoginIDSettings;
