import React from "react";
import { Checkbox, Toggle, TextField, PrimaryButton } from "@fluentui/react";

import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import CheckboxWithTooltip from "../../CheckboxWithTooltip";
import CheckboxWithContent from "../../CheckboxWithContent";
import { useTextField, useCheckbox } from "../../hook/useInput";

import styles from "./AuthenticationLoginIDSettings.module.scss";

interface Props {
  appConfig: Record<string, unknown> | null;
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

  reservedUsername: string;
  excludedKeywords: string;
  isBlockReservedUsername: boolean;
  isIncludeDefaultReservedUsernameList: boolean;
  isExcludeKeywords: boolean;
  isIncludeDefaultKeywordList: boolean;
  isUsernameCaseSensitive: boolean;
  isAsciiOnly: boolean;

  reservedKeywords: string;
  blockedDomains: string;
  isBlockReservedKeywords: boolean;
  isBlockDomains: boolean;
  isIncludeFreeEmailDomains: boolean;
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
  loginIdKeys: any[]
): { [key: string]: boolean } {
  const usernameEnabledConfig =
    loginIdKeys.find((key: any) => key.type === "username")?.verification
      ?.enabled ?? false;
  const emailEnabledConfig =
    loginIdKeys.find((key: any) => key.type === "email")?.verification
      ?.enabled ?? false;
  const phoneNumberEnabledConfig =
    loginIdKeys.find((key: any) => key.type === "phone")?.verification
      ?.enabled ?? false;

  return {
    usernameEnabledConfig,
    emailEnabledConfig,
    phoneNumberEnabledConfig,
  };
}

function constructAppConfigFromState(
  appConfig: Record<string, unknown>,
  screenState: AuthenticationLoginIDSettingsState
): Record<string, unknown> {
  // TODO: to be implemented
  console.log(appConfig, screenState);
  return {};
}

const AuthenticationLoginIDSettings: React.FC<Props> = function AuthenticationLoginIDSettings(
  props: Props
) {
  const { appConfig } = props;
  const { renderToString } = React.useContext(Context);
  const loginIdKeys = (appConfig?.identity as any)?.login_id?.keys ?? [];
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
  const usernameConfig = (appConfig?.identity as any)?.login_id?.types
    ?.username;
  const reservedUsernamesConfig = usernameConfig?.reserved_usernames ?? [];
  const excludedKeywordsConfig = usernameConfig?.excluded_keywords ?? [];

  const {
    value: reservedUsername,
    onChange: onReservedUsernameChange,
  } = useTextField(reservedUsernamesConfig.join(", "));
  const {
    value: excludedKeywords,
    onChange: onExcludedKeywordsChange,
  } = useTextField(excludedKeywordsConfig.join(", "));
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
  const emailConfig = (appConfig?.identity as any)?.login_id?.types?.email;
  const reservedKeywordsConfig = emailConfig?.reserved_keywords ?? [];
  const blockedDomainsConfig = emailConfig?.blocked_domains ?? [];

  const {
    value: reservedKeywords,
    onChange: onReservedKeywordsChange,
  } = useTextField(reservedKeywordsConfig.join(", "));
  const {
    value: blockedDomains,
    onChange: onBlockedDomainsChange,
  } = useTextField(blockedDomainsConfig.join(", "));
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
    const screenState = {
      usernameEnabled,
      emailEnabled,
      phoneNumberEnabled,

      reservedUsername,
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
    constructAppConfigFromState(props.appConfig, screenState);
  }, [
    props.appConfig,

    usernameEnabled,
    emailEnabled,
    phoneNumberEnabled,

    reservedUsername,
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
          extendable={usernameEnabled}
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
              <TextField
                className={styles.widgetInputField}
                disabled={!isBlockReservedUsername}
                value={reservedUsername}
                onChange={onReservedUsernameChange}
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
              <TextField
                className={styles.widgetInputField}
                disabled={!isExcludeKeywords}
                value={excludedKeywords}
                onChange={onExcludedKeywordsChange}
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
          extendable={emailEnabled}
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
            <TextField
              className={styles.widgetInputField}
              disabled={!isBlockReservedKeywords}
              value={reservedKeywords}
              onChange={onReservedKeywordsChange}
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
            <TextField
              className={styles.widgetInputField}
              disabled={!isBlockDomains}
              value={blockedDomains}
              onChange={onBlockedDomainsChange}
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
          extendable={phoneNumberEnabled}
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
