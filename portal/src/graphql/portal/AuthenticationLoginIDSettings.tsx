import React from "react";
import { Checkbox, Toggle, TextField } from "@fluentui/react";

import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import CheckboxWithTooltip from "../../CheckboxWithTooltip";
import CheckboxWithContent from "../../CheckboxWithContent";
import { useTextField, useCheckbox } from "../../hook/useInput";
import { LoginIDKeyType } from "../../types";

import styles from "./AuthenticationLoginIDSettings.module.scss";

interface Props {
  appConfig: Record<string, unknown> | null;
}

interface WidgetHeaderProps {
  enabled: boolean;
  setEnabled: (enabled: boolean) => void;
  titleId: string;
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
      <span className={styles.widgetHeaderText}>
        <FormattedMessage id={props.titleId} />
      </span>
    </div>
  );
};

const AuthenticationLoginIDSettings: React.FC<Props> = function AuthenticationLoginIDSettings(
  props: Props
) {
  const { appConfig } = props;
  console.log(appConfig);
  const { renderToString } = React.useContext(Context);
  const loginIdKeys = (appConfig?.identity as any)?.login_id?.keys ?? [];
  const [usernameEnabled, setUsernameEnabled] = React.useState(
    loginIdKeys.find((key: any) => key.type === "username")?.verification
      ?.enabled ?? false
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
    value: isCaseSensitive,
    onChange: onIsCaseSensitiveChange,
  } = useCheckbox(!!usernameConfig?.case_sensitive);
  const { value: isAsciiOnly, onChange: onIsAsciiOnlyChange } = useCheckbox(
    !!usernameConfig?.ascii_only
  );

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
              checked={isCaseSensitive}
              onChange={onIsCaseSensitiveChange}
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
    </div>
  );
};

export default AuthenticationLoginIDSettings;
