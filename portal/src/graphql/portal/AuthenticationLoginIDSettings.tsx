import React from "react";
import produce from "immer";
import { Checkbox, Toggle, PrimaryButton, TagPicker } from "@fluentui/react";

import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import CheckboxWithTooltip from "../../CheckboxWithTooltip";
import CheckboxWithContent from "../../CheckboxWithContent";
import { useCheckbox, useTagPickerWithNewTags } from "../../hook/useInput";
import {
  LoginIDKeyType,
  LoginIDKeyConfig,
  PortalAPIAppConfig,
} from "../../types";
import {
  ValidationRule,
  isValidEmailDomain,
  validateInput,
} from "../../util/validation";

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

  reservedUsernames: string[];
  excludedKeywords: string[];
  isBlockReservedUsername: boolean;
  isIncludeDefaultReservedUsernameList: boolean;
  isExcludeKeywords: boolean;
  isIncludeDefaultKeywordList: boolean;
  isUsernameCaseSensitive: boolean;
  isAsciiOnly: boolean;

  reservedKeywords: string[];
  blockedDomains: string[];
  isBlockReservedKeywords: boolean;
  isBlockDomains: boolean;
  isIncludeFreeEmailDomains: boolean;
  isEmailCaseSensitive: boolean;
  isIgnoreDotLocal: boolean;
  isAllowPlus: boolean;
}

interface AuthenticationLoginIDSettingErrorState {
  reservedUsernames?: string;
  excludedKeywords?: string;
  reservedKeywords?: string;
  blockedDomains?: string;
}

interface ValidationData {
  reservedUsernames: string[];
  excludedKeywords: string[];
  reservedKeywords: string[];
  blockedDomains: string[];
}

const validationRules: ValidationRule<
  ValidationData,
  AuthenticationLoginIDSettingErrorState
>[] = [
  {
    inputKey: "blockedDomains",
    errorKey: "blockedDomains",
    errorMessageId: "AuthenticationWidget.error.blockDomains",
    condition: (blockedDomains: string[]) => {
      const isValid = blockedDomains
        .map((domain) => domain.trim())
        .map((domain) => isValidEmailDomain(domain));
      return isValid.every(Boolean);
    },
  },
];

const switchStyle = { root: { margin: "0" } };

const WidgetHeader: React.FC<WidgetHeaderProps> = function (
  props: WidgetHeaderProps
) {
  return (
    <div className={styles.widgetHeader}>
      <Toggle
        styles={switchStyle}
        checked={props.enabled}
        onChanged={props.setEnabled}
      />
      <header className={styles.widgetHeaderText}>
        <FormattedMessage id={props.titleId} />
      </header>
    </div>
  );
};

function extractConfigFromLoginIdKeys(
  loginIdKeys: LoginIDKeyConfig[]
): { [key: string]: boolean } {
  const usernameEnabledConfig =
    loginIdKeys.find((key) => key.type === "username")?.verification?.enabled ??
    false;
  const emailEnabledConfig =
    loginIdKeys.find((key) => key.type === "email")?.verification?.enabled ??
    false;
  const phoneNumberEnabledConfig =
    loginIdKeys.find((key) => key.type === "phone")?.verification?.enabled ??
    false;

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

    const reservedUsernameList = handleStringListInput(
      screenState.reservedUsernames,
      {
        optionEnabled: screenState.isBlockReservedUsername,
        useDefaultList: screenState.isIncludeDefaultReservedUsernameList,
        defaultList: [],
      }
    );
    const excludedKeywordList = handleStringListInput(
      screenState.excludedKeywords,
      {
        optionEnabled: screenState.isExcludeKeywords,
        useDefaultList: screenState.isIncludeDefaultKeywordList,
        defaultList: [],
      }
    );

    setFieldIfListNonEmpty(
      usernameConfig,
      "reserved_usernames",
      reservedUsernameList
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

    const reservedKeywordList = handleStringListInput(
      screenState.reservedKeywords,
      {
        optionEnabled: screenState.isBlockReservedKeywords,
        useDefaultList: false,
        defaultList: [],
      }
    );
    const blockedDomainList = handleStringListInput(
      screenState.blockedDomains,
      {
        optionEnabled: screenState.isBlockDomains,
        useDefaultList: screenState.isIncludeFreeEmailDomains,
        defaultList: [],
      }
    );

    setFieldIfListNonEmpty(
      emailConfig,
      "reserved_keywords",
      reservedKeywordList
    );
    setFieldIfListNonEmpty(emailConfig, "blocked_domains", blockedDomainList);
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
  const [errorState, setErrorState] = React.useState<
    AuthenticationLoginIDSettingErrorState
  >({});
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
  const reservedUsernamesConfig = usernameConfig?.reserved_usernames ?? [];
  const excludedKeywordsConfig = usernameConfig?.excluded_keywords ?? [];

  const {
    list: reservedUsernames,
    onChange: onReservedUsernamesChange,
    defaultSelectedItems: defaultSelectedReservedUsernames,
    onResolveSuggestions: onResolveReservedUsernameSuggestions,
  } = useTagPickerWithNewTags(reservedUsernamesConfig);
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
    value: isIncludeDefaultReservedUsernameList,
    onChange: onIsIncludeDefaultReservedUsernameListChange,
  } = useCheckbox(false);
  const {
    value: isExcludeKeywords,
    onChange: onIsExcludeKeywordsChange,
  } = useCheckbox(excludedKeywordsConfig.length > 0);
  const {
    value: isIncludeDefaultKeywordList,
    onChange: onIsIncludeDefaultKeywordListChange,
  } = useCheckbox(false);
  const {
    value: isUsernameCaseSensitive,
    onChange: onIsUsernameCaseSensitiveChange,
  } = useCheckbox(!!usernameConfig?.case_sensitive);
  const { value: isAsciiOnly, onChange: onIsAsciiOnlyChange } = useCheckbox(
    !!usernameConfig?.ascii_only
  );

  // email widget
  const emailConfig = appConfig?.identity?.login_id?.types.email;
  const reservedKeywordsConfig = emailConfig?.reserved_keywords ?? [];
  const blockedDomainsConfig = emailConfig?.blocked_domains ?? [];

  const {
    list: reservedKeywords,
    onChange: onReservedKeywordsChange,
    defaultSelectedItems: defaultSelectedReservedKeywords,
    onResolveSuggestions: onResolveReservedKeywordSuggestions,
  } = useTagPickerWithNewTags(reservedKeywordsConfig);
  const {
    list: blockedDomains,
    onChange: onBlockedDomainsChange,
    defaultSelectedItems: defaultSelectedBlockedDomains,
    onResolveSuggestions: onResolveBlockedDomainSuggestions,
  } = useTagPickerWithNewTags(blockedDomainsConfig);
  const {
    value: isBlockReservedKeywords,
    onChange: onIsBlockReservedKeywordsChange,
  } = useCheckbox(reservedKeywordsConfig.length > 0);
  const {
    value: isBlockDomains,
    onChange: onIsBlockDomainsChange,
  } = useCheckbox(blockedDomainsConfig.length > 0);
  const {
    value: isIncludeFreeEmailDomains,
    onChange: onIsIncludeFreeEmailDomainsChange,
  } = useCheckbox(false);
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

    const validationData = {
      reservedUsernames,
      excludedKeywords,
      reservedKeywords,
      blockedDomains,
    };

    const { errorResult, isValid } = validateInput(
      validationData,
      validationRules
    );
    setErrorState(errorResult);
    if (!isValid) {
      return;
    }

    const screenState = {
      usernameEnabled,
      emailEnabled,
      phoneNumberEnabled,

      reservedUsernames,
      excludedKeywords,
      isBlockReservedUsername,
      isIncludeDefaultReservedUsernameList,
      isExcludeKeywords,
      isIncludeDefaultKeywordList,
      isUsernameCaseSensitive,
      isAsciiOnly,

      reservedKeywords,
      blockedDomains,
      isBlockReservedKeywords,
      isBlockDomains,
      isIncludeFreeEmailDomains,
      isEmailCaseSensitive,
      isIgnoreDotLocal,
      isAllowPlus,
    };

    const newAppConfig = constructAppConfigFromState(
      props.appConfig,
      screenState
    );
    // TODO: call mutation to save config
    console.log(newAppConfig);
  }, [
    props.appConfig,

    usernameEnabled,
    emailEnabled,
    phoneNumberEnabled,

    reservedUsernames,
    excludedKeywords,
    isBlockReservedUsername,
    isIncludeDefaultReservedUsernameList,
    isExcludeKeywords,
    isIncludeDefaultKeywordList,
    isUsernameCaseSensitive,
    isAsciiOnly,

    reservedKeywords,
    blockedDomains,
    isBlockReservedKeywords,
    isBlockDomains,
    isIncludeFreeEmailDomains,
    isEmailCaseSensitive,
    isIgnoreDotLocal,
    isAllowPlus,
  ]);

  return (
    <div className={styles.root}>
      <div className={styles.widgetContainer}>
        <ExtendableWidget
          extendable={true}
          readOnly={!usernameEnabled}
          HeaderComponent={
            <WidgetHeader
              enabled={usernameEnabled}
              setEnabled={setUsernameEnabled}
              titleId={"AuthenticationWidget.usernameTitle"}
            />
          }
        >
          <div className={styles.usernameWidgetContent}>
            <CheckboxWithContent
              checked={isBlockReservedUsername}
              onChange={onIsBlockReservedUsernameChange}
              className={styles.checkboxWithContent}
            >
              <div className={styles.checkboxLabel}>
                <FormattedMessage id="AuthenticationWidget.blockReservedUsername" />
              </div>
              <TagPicker
                className={styles.widgetInputField}
                disabled={!isBlockReservedUsername}
                onChange={onReservedUsernamesChange}
                defaultSelectedItems={defaultSelectedReservedUsernames}
                onResolveSuggestions={onResolveReservedUsernameSuggestions}
              />
              <CheckboxWithTooltip
                label={renderToString(
                  "AuthenticationWidget.includeDefaultList"
                )}
                helpText={renderToString(
                  "AuthenticationWidget.includeDefaultList.reservedUsernameHelp"
                )}
                disabled={!isBlockReservedUsername}
                checked={isIncludeDefaultReservedUsernameList}
                onChange={onIsIncludeDefaultReservedUsernameListChange}
              />
            </CheckboxWithContent>

            <CheckboxWithContent
              checked={isExcludeKeywords}
              onChange={onIsExcludeKeywordsChange}
              className={styles.checkboxWithContent}
            >
              <div className={styles.checkboxLabel}>
                <FormattedMessage id="AuthenticationWidget.excludeKeywords" />
              </div>
              <TagPicker
                className={styles.widgetInputField}
                disabled={!isExcludeKeywords}
                onChange={onExcludedKeywordsChange}
                defaultSelectedItems={defaultSelectedExcludedKeywords}
                onResolveSuggestions={onResolveExcludedKeywordSuggestions}
              />
              <CheckboxWithTooltip
                label={renderToString(
                  "AuthenticationWidget.includeDefaultList"
                )}
                helpText={renderToString(
                  "AuthenticationWidget.includeDefaultList.excludeKeywordsHelp"
                )}
                disabled={!isExcludeKeywords}
                checked={isIncludeDefaultKeywordList}
                onChange={onIsIncludeDefaultKeywordListChange}
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
          extendable={true}
          readOnly={!emailEnabled}
          HeaderComponent={
            <WidgetHeader
              enabled={emailEnabled}
              setEnabled={setEmailEnabled}
              titleId={"AuthenticationWidget.emailAddressTitle"}
            />
          }
        >
          <CheckboxWithContent
            checked={isBlockReservedKeywords}
            onChange={onIsBlockReservedKeywordsChange}
            className={styles.checkboxWithContent}
          >
            <div className={styles.checkboxLabel}>
              <FormattedMessage id="AuthenticationWidget.blockReservedKeywords" />
            </div>
            <TagPicker
              className={styles.widgetInputField}
              disabled={!isBlockReservedKeywords}
              onChange={onReservedKeywordsChange}
              defaultSelectedItems={defaultSelectedReservedKeywords}
              onResolveSuggestions={onResolveReservedKeywordSuggestions}
            />
          </CheckboxWithContent>

          <CheckboxWithContent
            checked={isBlockDomains}
            onChange={onIsBlockDomainsChange}
            className={styles.checkboxWithContent}
          >
            <div className={styles.checkboxLabel}>
              <FormattedMessage id="AuthenticationWidget.blockDomains" />
            </div>
            <TagPicker
              className={styles.widgetInputField}
              disabled={!isBlockDomains}
              onChange={onBlockedDomainsChange}
              defaultSelectedItems={defaultSelectedBlockedDomains}
              onResolveSuggestions={onResolveBlockedDomainSuggestions}
            />
            <CheckboxWithTooltip
              label={renderToString(
                "AuthenticationWidget.includeFreeEmailDomains"
              )}
              helpText={renderToString(
                "AuthenticationWidget.includeFreeEmailDomainsHelp"
              )}
              disabled={!isBlockDomains}
              checked={isIncludeFreeEmailDomains}
              onChange={onIsIncludeFreeEmailDomainsChange}
            />
          </CheckboxWithContent>

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
          extendable={true}
          readOnly={!phoneNumberEnabled}
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
