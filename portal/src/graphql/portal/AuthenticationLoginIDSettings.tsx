import React from "react";
import produce from "immer";
import { Checkbox, Toggle, PrimaryButton, TagPicker } from "@fluentui/react";

import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import CheckboxWithContent from "../../CheckboxWithContent";
import { useCheckbox, useTagPickerWithNewTags } from "../../hook/useInput";
import {
  LoginIDKeyType,
  LoginIDKeyConfig,
  PortalAPIAppConfig,
} from "../../types";

import styles from "./AuthenticationLoginIDSettings.module.scss";

interface Props {
  appConfig: PortalAPIAppConfig | null;
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
  return (
    <div className={styles.widgetHeader}>
      <Toggle
        label={<FormattedMessage id={props.titleId} />}
        inlineLabel={true}
        styles={switchStyle}
        checked={props.enabled}
        onChanged={props.setEnabled}
      />
    </div>
  );
};

function extractConfigFromLoginIdKeys(
  loginIdKeys: LoginIDKeyConfig[]
): { [key: string]: boolean } {
  // We consider them as enabled if they are listed as allowed login ID keys.
  const usernameEnabledConfig =
    loginIdKeys.find((key) => key.type === "username") != null;
  const emailEnabledConfig =
    loginIdKeys.find((key) => key.type === "email") != null;
  const phoneNumberEnabledConfig =
    loginIdKeys.find((key) => key.type === "phone") != null;

  return {
    usernameEnabledConfig,
    emailEnabledConfig,
    phoneNumberEnabledConfig,
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
function getOrCreateLoginIdKey(
  loginIdKeys: LoginIDKeyConfig[],
  keyType: LoginIDKeyType
): LoginIDKeyConfig {
  const loginIdKey = loginIdKeys.find((key: any) => key.type === keyType);
  if (loginIdKey != null) {
    return loginIdKey;
  }
  const newLoginIdKey = { type: keyType };
  loginIdKeys.push(newLoginIdKey);
  return newLoginIdKey;
}

function setLoginIdKeyEnabled(loginIdKey: LoginIDKeyConfig, enabled: boolean) {
  loginIdKey.verification = loginIdKey.verification ?? { enabled: false };
  loginIdKey.verification.enabled = enabled;
}

function constructAppConfigFromState(
  appConfig: PortalAPIAppConfig,
  screenState: AuthenticationLoginIDSettingsState
): PortalAPIAppConfig {
  const newAppConfig = produce(appConfig, (draftConfig) => {
    const loginIdKeys = draftConfig.identity?.login_id?.keys ?? [];
    const loginIdUsernameKey = getOrCreateLoginIdKey(loginIdKeys, "username");
    const loginIdEmailKey = getOrCreateLoginIdKey(loginIdKeys, "email");
    const loginIdPhoneNumberKey = getOrCreateLoginIdKey(loginIdKeys, "phone");

    setLoginIdKeyEnabled(loginIdUsernameKey, screenState.usernameEnabled);
    setLoginIdKeyEnabled(loginIdEmailKey, screenState.emailEnabled);
    setLoginIdKeyEnabled(loginIdPhoneNumberKey, screenState.phoneNumberEnabled);

    const loginIdTypes = draftConfig.identity?.login_id?.types;

    if (loginIdTypes == null) {
      return;
    }

    // username config
    loginIdTypes.username = loginIdTypes.username ?? {};
    const usernameConfig = loginIdTypes.username;

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
    usernameConfig.case_sensitive = screenState.isUsernameCaseSensitive;
    usernameConfig.ascii_only = screenState.isAsciiOnly;

    // email config
    loginIdTypes.email = loginIdTypes.email ?? {};
    const emailConfig = loginIdTypes.email;

    emailConfig.case_sensitive = screenState.isEmailCaseSensitive;
    emailConfig.ignore_dot_sign = screenState.isIgnoreDotLocal;
    emailConfig.block_plus_sign = !screenState.isAllowPlus;
  });

  return newAppConfig;
}

const AuthenticationLoginIDSettings: React.FC<Props> = function AuthenticationLoginIDSettings(
  props: Props
) {
  const { appConfig } = props;
  const { renderToString } = React.useContext(Context);
  const loginIdKeys = appConfig?.identity?.login_id?.keys ?? [];
  const {
    usernameEnabledConfig,
    emailEnabledConfig,
    phoneNumberEnabledConfig,
  } = extractConfigFromLoginIdKeys(loginIdKeys);
  const [usernameEnabled, setUsernameEnabled] = React.useState(
    usernameEnabledConfig
  );
  const [emailEnabled, setEmailEnabled] = React.useState(emailEnabledConfig);
  const [phoneNumberEnabled, setPhoneNumberEnabled] = React.useState(
    phoneNumberEnabledConfig
  );

  // username widget
  const usernameConfig = appConfig?.identity?.login_id?.types.username;
  const excludedKeywordsConfig = usernameConfig?.excluded_keywords ?? [];

  const {
    list: excludedKeywords,
    onChange: onExcludedKeywordsChange,
    defaultSelectedItems: defaultSelectedExcludedKeywords,
    onResolveSuggestions: onResolveExcludedKeywordSuggestions,
  } = useTagPickerWithNewTags(excludedKeywordsConfig);
  const {
    value: isBlockReservedUsername,
    onChange: onIsBlockReservedUsernameChange,
  } = useCheckbox(!!usernameConfig?.block_reserved_usernames);
  const {
    value: isExcludeKeywords,
    onChange: onIsExcludeKeywordsChange,
  } = useCheckbox(excludedKeywordsConfig.length > 0);
  const {
    value: isUsernameCaseSensitive,
    onChange: onIsUsernameCaseSensitiveChange,
  } = useCheckbox(!!usernameConfig?.case_sensitive);
  const { value: isAsciiOnly, onChange: onIsAsciiOnlyChange } = useCheckbox(
    !!usernameConfig?.ascii_only
  );

  // email widget
  const emailConfig = appConfig?.identity?.login_id?.types.email;

  const {
    value: isEmailCaseSensitive,
    onChange: onIsEmailCaseSensitiveChange,
  } = useCheckbox(!!emailConfig?.case_sensitive);
  const {
    value: isIgnoreDotLocal,
    onChange: onIsIgnoreDotLocalChange,
  } = useCheckbox(!!emailConfig?.ignore_dot_sign);
  const { value: isAllowPlus, onChange: onIsAllowPlusChange } = useCheckbox(
    !emailConfig?.block_plus_sign
  );

  const onSaveButtonClicked = React.useCallback(() => {
    if (props.appConfig == null) {
      return;
    }

    const screenState = {
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
    };

    constructAppConfigFromState(props.appConfig, screenState);
    // TODO: call mutation to save config
  }, [
    props.appConfig,

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
  ]);

  return (
    <div className={styles.root}>
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
              <div className={styles.checkboxLabel}>
                <FormattedMessage id="AuthenticationWidget.excludeKeywords" />
              </div>
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
        <PrimaryButton onClick={onSaveButtonClicked}>
          <FormattedMessage id="save" />
        </PrimaryButton>
      </div>
    </div>
  );
};

export default AuthenticationLoginIDSettings;
