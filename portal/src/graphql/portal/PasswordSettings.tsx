import React, { useMemo, useContext, useCallback, ReactElement } from "react";
import { produce } from "immer";
import { Checkbox, Dropdown, Label } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  AuthenticatorPasswordConfig,
  ForgotPasswordConfig,
  PasswordPolicyFeatureConfig,
  PasswordPolicyConfig,
  isPasswordPolicyGuessableLevel,
  passwordPolicyGuessableLevels,
} from "../../types";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetSubtitle from "../../WidgetSubtitle";
import WidgetDescription from "../../WidgetDescription";
import HorizontalDivider from "../../HorizontalDivider";
import TextField from "../../TextField";
import Toggle from "../../Toggle";
import CustomTagPicker from "../../CustomTagPicker";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { useTagPickerWithNewTags } from "../../hook/useInput";
import { fixTagPickerStyles } from "../../bugs";
import { parseIntegerAllowLeadingZeros } from "../../util/input";

export interface State {
  forgotPasswordConfig: ForgotPasswordConfig;
  authenticatorPasswordConfig: AuthenticatorPasswordConfig;
  passwordPolicyFeatureConfig?: PasswordPolicyFeatureConfig;
}

export interface PasswordSettingsProps<T extends State> extends State {
  className?: string;
  setState: (fn: (state: T) => T) => void;
}

function usePasswordNumberOnChange<T extends State>(
  setState: PasswordSettingsProps<T>["setState"],
  key: "min_length" | "history_days" | "history_size"
) {
  return useCallback(
    (_, value) => {
      if (value == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy[key] =
            parseIntegerAllowLeadingZeros(value);
        })
      );
    },
    [setState, key]
  );
}

function usePasswordCheckboxOnChange<T extends State>(
  setState: PasswordSettingsProps<T>["setState"],
  key: keyof PasswordPolicyConfig
) {
  return useCallback(
    (_, value) => {
      if (value == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy[key] = value;
        })
      );
    },
    [setState, key]
  );
}

// eslint-disable-next-line complexity
export default function PasswordSettings<T extends State>(
  props: PasswordSettingsProps<T>
): ReactElement {
  const {
    className,
    authenticatorPasswordConfig,
    forgotPasswordConfig,
    passwordPolicyFeatureConfig,
    setState,
  } = props;

  const { renderToString } = useContext(Context);

  const anyAdvancedPolicyDisabled =
    (passwordPolicyFeatureConfig?.minimum_guessable_level?.disabled ?? false) ||
    (passwordPolicyFeatureConfig?.history?.disabled ?? false) ||
    (passwordPolicyFeatureConfig?.excluded_keywords?.disabled ?? false);

  const isPreventPasswordReuseEnabled =
    (authenticatorPasswordConfig.policy?.history_days != null &&
      authenticatorPasswordConfig.policy.history_days > 0) ||
    (authenticatorPasswordConfig.policy?.history_size != null &&
      authenticatorPasswordConfig.policy.history_size > 0);

  const onChangeForceChange = useCallback(
    (_e, checked) => {
      if (checked == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.force_change = checked;
        })
      );
    },
    [setState]
  );

  const onChangeCodeExpirySeconds = useCallback(
    (_e, value) => {
      if (value == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.forgotPasswordConfig.reset_code_expiry_seconds =
            parseIntegerAllowLeadingZeros(value);
        })
      );
    },
    [setState]
  );

  const onChangeMinLength = usePasswordNumberOnChange(setState, "min_length");
  const onChangeDigitRequired = usePasswordCheckboxOnChange(
    setState,
    "digit_required"
  );
  const onChangeLowercaseRequired = usePasswordCheckboxOnChange(
    setState,
    "lowercase_required"
  );
  const onChangeUppercaseRequired = usePasswordCheckboxOnChange(
    setState,
    "uppercase_required"
  );
  const onChangeSymbolRequired = usePasswordCheckboxOnChange(
    setState,
    "symbol_required"
  );
  const onChangeHistoryDays = usePasswordNumberOnChange(
    setState,
    "history_days"
  );
  const onChangeHistorySize = usePasswordNumberOnChange(
    setState,
    "history_size"
  );

  const minGuessableLevelOptions = useMemo(() => {
    return passwordPolicyGuessableLevels.map((level) => ({
      key: level,
      text: renderToString(
        `PasswordPolicyConfigurationScreen.min-guessable-level.${level}`
      ),
    }));
  }, [renderToString]);
  const onChangeMinimumGuessableLevel = useCallback(
    (_, option) => {
      const key = option?.key;
      if (!isPasswordPolicyGuessableLevel(key)) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy.minimum_guessable_level = key;
        })
      );
    },
    [setState]
  );

  const onChangePreventReuseEnabled = useCallback(
    (_, checked) => {
      if (checked == null) {
        return;
      }
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          if (checked) {
            prev.authenticatorPasswordConfig.policy.history_days = 90;
            prev.authenticatorPasswordConfig.policy.history_size = 3;
          } else {
            prev.authenticatorPasswordConfig.policy.history_days = 0;
            prev.authenticatorPasswordConfig.policy.history_size = 0;
          }
        })
      );
    },
    [setState]
  );

  const valueForExcludedKeywords = useMemo(() => {
    return authenticatorPasswordConfig.policy?.excluded_keywords ?? [];
  }, [authenticatorPasswordConfig.policy?.excluded_keywords]);
  const updateExcludedKeywords = useCallback(
    (value: string[]) => {
      setState((prev) =>
        produce(prev, (prev) => {
          prev.authenticatorPasswordConfig.policy ??= {};
          prev.authenticatorPasswordConfig.policy.excluded_keywords = value;
        })
      );
    },
    [setState]
  );
  const {
    selectedItems: excludedKeywords,
    onChange: onChangeExcludedKeywords,
    onResolveSuggestions: onResolveSuggestionsExcludedKeywords,
    onAdd: onAddExcludedKeywords,
  } = useTagPickerWithNewTags(valueForExcludedKeywords, updateExcludedKeywords);

  return (
    <Widget className={className}>
      <WidgetTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.title" />
      </WidgetTitle>
      <WidgetDescription>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.description" />
      </WidgetDescription>
      <Toggle
        checked={authenticatorPasswordConfig.force_change}
        inlineLabel={true}
        label={
          <FormattedMessage id="PasswordPolicyConfigurationScreen.force-change.label" />
        }
        onChange={onChangeForceChange}
      />
      <TextField
        type="text"
        label={renderToString(
          "ForgotPasswordConfigurationScreen.reset-code-valid-duration.label"
        )}
        value={forgotPasswordConfig.reset_code_expiry_seconds?.toFixed(0) ?? ""}
        onChange={onChangeCodeExpirySeconds}
      />
      <HorizontalDivider />
      <WidgetSubtitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.basic" />
      </WidgetSubtitle>
      <TextField
        type="text"
        label={renderToString(
          "PasswordPolicyConfigurationScreen.min-length.label"
        )}
        value={authenticatorPasswordConfig.policy?.min_length?.toFixed(0) ?? ""}
        onChange={onChangeMinLength}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-digit.label"
        )}
        checked={authenticatorPasswordConfig.policy?.digit_required}
        onChange={onChangeDigitRequired}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-lowercase.label"
        )}
        checked={authenticatorPasswordConfig.policy?.lowercase_required}
        onChange={onChangeLowercaseRequired}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-uppercase.label"
        )}
        checked={authenticatorPasswordConfig.policy?.uppercase_required}
        onChange={onChangeUppercaseRequired}
      />
      <Checkbox
        label={renderToString(
          "PasswordPolicyConfigurationScreen.require-symbol.label"
        )}
        checked={authenticatorPasswordConfig.policy?.symbol_required}
        onChange={onChangeSymbolRequired}
      />
      <HorizontalDivider />
      <WidgetSubtitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.password.advanced" />
      </WidgetSubtitle>
      {anyAdvancedPolicyDisabled ? (
        <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
      ) : null}
      <Dropdown
        label={renderToString(
          "PasswordPolicyConfigurationScreen.min-guessable-level.label"
        )}
        disabled={
          passwordPolicyFeatureConfig?.minimum_guessable_level?.disabled
        }
        options={minGuessableLevelOptions}
        selectedKey={
          authenticatorPasswordConfig.policy?.minimum_guessable_level
        }
        onChange={onChangeMinimumGuessableLevel}
      />
      <Toggle
        disabled={passwordPolicyFeatureConfig?.history?.disabled}
        checked={isPreventPasswordReuseEnabled}
        inlineLabel={true}
        label={
          <FormattedMessage id="PasswordPolicyConfigurationScreen.prevent-reuse.label" />
        }
        onChange={onChangePreventReuseEnabled}
      />
      <TextField
        type="text"
        disabled={
          (passwordPolicyFeatureConfig?.history?.disabled ?? false) ||
          !isPreventPasswordReuseEnabled
        }
        label={renderToString(
          "PasswordPolicyConfigurationScreen.history-days.label"
        )}
        value={
          authenticatorPasswordConfig.policy?.history_days?.toFixed(0) ?? ""
        }
        onChange={onChangeHistoryDays}
      />
      <TextField
        type="text"
        disabled={
          (passwordPolicyFeatureConfig?.history?.disabled ?? false) ||
          !isPreventPasswordReuseEnabled
        }
        label={renderToString(
          "PasswordPolicyConfigurationScreen.history-size.label"
        )}
        value={
          authenticatorPasswordConfig.policy?.history_size?.toFixed(0) ?? ""
        }
        onChange={onChangeHistorySize}
      />
      <div>
        <Label
          disabled={passwordPolicyFeatureConfig?.excluded_keywords?.disabled}
        >
          <FormattedMessage id="PasswordPolicyConfigurationScreen.excluded-keywords.label" />
        </Label>
        <CustomTagPicker
          styles={fixTagPickerStyles}
          inputProps={{
            "aria-label": renderToString(
              "PasswordPolicyConfigurationScreen.excluded-keywords.label"
            ),
          }}
          disabled={passwordPolicyFeatureConfig?.excluded_keywords?.disabled}
          selectedItems={excludedKeywords}
          onChange={onChangeExcludedKeywords}
          onResolveSuggestions={onResolveSuggestionsExcludedKeywords}
          onAdd={onAddExcludedKeywords}
        />
      </div>
    </Widget>
  );
}
