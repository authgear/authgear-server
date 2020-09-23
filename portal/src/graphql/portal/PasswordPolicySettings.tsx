import React, { useMemo, useContext, useState, useCallback, useEffect, useRef } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { TextField, Checkbox, Dropdown, IDropdownOption, TagPicker, Label, ITag } from "@fluentui/react";
import cn from "classnames";
import deepEqual from "deep-equal";
import produce from "immer";

import ToggleWithContent from "../../ToggleWithContent";
import ButtonWithLoading from "../../ButtonWithLoading";
import {
  setFieldIfChanged,
  setFieldIfListNonEmpty,
  isArrayEqualInOrder,
  clearEmptyObject,
} from "../../util/misc";
import {
  PortalAPIAppConfig,
  PortalAPIApp,
  PasswordPolicyGuessableLevel,
  passwordPolicyGuessableLevels,
  isPasswordPolicyGuessableLevel
} from "../../types";

import styles from "./PasswordPolicySettings.module.scss";

interface PasswordPolicySettingsProps {
  className?: string;
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
  onIsFormModifiedChange: (modified: boolean) => void;
}

interface PasswordPolicySettingsState {
  minLength: number;
  isDigitRequired: boolean;
  isLowercaseRequired: boolean;
  isUppercaseRequired: boolean;
  isSymbolRequired: boolean;
  minGuessableLevel: PasswordPolicyGuessableLevel;
  preventReuse: boolean;
  historyDays: number;
  historySize: number;
  excludedKeywords: string[];
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): PasswordPolicySettingsState {
  const passwordPolicy = appConfig?.authenticator?.password?.policy;

  return {
    minLength: passwordPolicy?.min_length ?? 0,
    isDigitRequired: !!passwordPolicy?.digit_required,
    isLowercaseRequired: !!passwordPolicy?.lowercase_required,
    isUppercaseRequired: !!passwordPolicy?.uppercase_required,
    isSymbolRequired: !!passwordPolicy?.symbol_required,
    minGuessableLevel: passwordPolicy?.minimum_guessable_level ?? 1,
    preventReuse: passwordPolicy?.history_days !== 0 || passwordPolicy.history_size !== 0,
    historyDays: passwordPolicy?.history_days ?? 0,
    historySize: passwordPolicy?.history_size ?? 0,
    excludedKeywords: passwordPolicy?.excluded_keywords ?? [],
  };
}

function constructAppConfigFromState(
  rawAppConfig: PortalAPIAppConfig,
  initialScreenState: PasswordPolicySettingsState,
  screenState: PasswordPolicySettingsState
): PortalAPIAppConfig {
  const newAppConfig = produce(rawAppConfig, (draftConfig) => {
    draftConfig.authenticator = draftConfig.authenticator ?? {};
    draftConfig.authenticator.password = draftConfig.authenticator.password ?? {};
    draftConfig.authenticator.password.policy = draftConfig.authenticator.password.policy ?? {};

    const passwordPolicy = draftConfig.authenticator.password.policy;

    setFieldIfChanged(
      passwordPolicy,
      "min_length",
      initialScreenState.minLength,
      screenState.minLength
    );

    setFieldIfChanged(
      passwordPolicy,
      "digit_required",
      initialScreenState.isDigitRequired,
      screenState.isDigitRequired
    );

    setFieldIfChanged(
      passwordPolicy,
      "lowercase_required",
      initialScreenState.isLowercaseRequired,
      screenState.isLowercaseRequired
    );

    setFieldIfChanged(
      passwordPolicy,
      "uppercase_required",
      initialScreenState.isUppercaseRequired,
      screenState.isUppercaseRequired
    );

    setFieldIfChanged(
      passwordPolicy,
      "symbol_required",
      initialScreenState.isSymbolRequired,
      screenState.isSymbolRequired
    );

    setFieldIfChanged(
      passwordPolicy,
      "minimum_guessable_level",
      initialScreenState.minGuessableLevel,
      screenState.minGuessableLevel
    );

    setFieldIfChanged(
      passwordPolicy,
      "history_days",
      initialScreenState.historyDays,
      screenState.historyDays
    );

    setFieldIfChanged(
      passwordPolicy,
      "history_size",
      initialScreenState.historySize,
      screenState.historySize
    );

    if (
      !isArrayEqualInOrder(
        initialScreenState.excludedKeywords,
        screenState.excludedKeywords
      )
    ) {
      setFieldIfListNonEmpty(
        passwordPolicy,
        "excluded_keywords",
        screenState.excludedKeywords
      );
    }

    clearEmptyObject(draftConfig);
  });

  return newAppConfig;
}

const PasswordPolicySettings: React.FC<PasswordPolicySettingsProps> = function PasswordPolicySettings(props) {
  const {
    className,
    effectiveAppConfig,
    rawAppConfig,
    updateAppConfig,
    updatingAppConfig,
    onIsFormModifiedChange,
  } = props;

  const { renderToString } = useContext(Context);

  const onIsFormModifiedChangeRef = useRef(onIsFormModifiedChange);
  onIsFormModifiedChangeRef.current = onIsFormModifiedChange;

  const initialState = useMemo(() => {
    return constructStateFromAppConfig(effectiveAppConfig);
  }, [effectiveAppConfig]);

  const [state, setState] = useState(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const minGuessableLevelOptions: IDropdownOption[] = useMemo(() => {
    return passwordPolicyGuessableLevels.map((level) => ({
      key: level,
      text: renderToString(`PasswordsScreen.password-policy.min-guessable-level.${level}`),
      isSelected: level === state.minGuessableLevel,
    }));
  }, [state.minGuessableLevel, renderToString]);

  const defaultSelectedExcludedKeywordItems: ITag[] = useMemo(() => {
    return state.excludedKeywords.map((keyword) => ({
      key: keyword,
      name: keyword,
    }));
  }, [state.excludedKeywords]);

  const onMinLengthChange = useCallback(
    (_event, value?: string) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        minLength: parseInt(value, 10),
      }));
    },
    []
  );

  const onIsDigitRequiredChange = useCallback(
    (_event, checked?: boolean) => {
      setState((state) => ({
        ...state,
        isDigitRequired: !!checked,
      }));
    },
    []
  );

  const onIsLowercaseRequiredChange = useCallback(
    (_event, checked?: boolean) => {
      setState((state) => ({
        ...state,
        isLowercaseRequired: !!checked,
      }));
    },
    []
  );

  const onIsUppercaseRequiredChange = useCallback(
    (_event, checked?: boolean) => {
      setState((state) => ({
        ...state,
        isUppercaseRequired: !!checked,
      }));
    },
    []
  );

  const onIsSymbolRequiredChange = useCallback(
    (_event, checked?: boolean) => {
      setState((state) => ({
        ...state,
        isSymbolRequired: !!checked,
      }));
    },
    []
  );

  const onMinGuessableLevelOptionChange = useCallback(
    (_event, option?: IDropdownOption) => {
      if (option != null && isPasswordPolicyGuessableLevel(option.key)) {
        setState((state) => ({
          ...state,
          minGuessableLevel: option.key as PasswordPolicyGuessableLevel,
        }));
      }
    },
    []
  );

  const onPreventReuseChange = useCallback(
    (_event, checked?: boolean) => {
      if (checked === undefined) {
        return;
      }
      if (checked) {
        setState((state) => ({
          ...state,
          preventReuse: true,
          historyDays: 90,
          historySize: 3,
        }));
      } else {
        setState((state) => ({
          ...state,
          preventReuse: false,
          historyDays: 0,
          historySize: 0,
        }));
      }
    },
    []
  );

  const onHistoryDaysChange = useCallback(
    (_event, value?: string) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        historyDays: parseInt(value, 10),
      }));
    },
    []
  );

  const onHistorySizeChange = useCallback(
    (_event, value?: string) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        historySize: parseInt(value, 10),
      }));
    },
    []
  );

  const onSaveButtonClicked = useCallback(() => {
    if (rawAppConfig == null) {
      return;
    }

    const newAppConfig = constructAppConfigFromState(
      rawAppConfig,
      initialState,
      state
    );

    // TODO: handle error
    updateAppConfig(newAppConfig).catch(() => {});
  }, [state, rawAppConfig, updateAppConfig, initialState]);

  const onResolveExcludedKeywordSuggestions = useCallback(
    (filterText: string, _tagList?: ITag[]): ITag[] => {
      return [{ key: filterText, name: filterText }];
    },
    []
  );

  const onExcludedKeywordsChange = useCallback(
    (items?: ITag[]) => {
      if (items == null) {
        return;
      }
      setState((state) => ({
        ...state,
        excludedKeywords: items.map((item) => item.name),
      }));
    },
    []
  );

  useEffect(() => {
    onIsFormModifiedChangeRef.current(isFormModified);
  }, [isFormModified]);

  return (
    <div className={cn(styles.root, className)}>
      <TextField
        className={styles.textField}
        type="number"
        min="1"
        step="1"
        label={renderToString("PasswordsScreen.password-policy.min-length.label")}
        value={`${state.minLength}`}
        onChange={onMinLengthChange}
      />
      <Checkbox
        className={styles.checkbox}
        label={renderToString("PasswordsScreen.password-policy.require-digit.label")}
        checked={state.isDigitRequired}
        onChange={onIsDigitRequiredChange}
      />
      <Checkbox
        className={styles.checkbox}
        label={renderToString("PasswordsScreen.password-policy.require-lowercase.label")}
        checked={state.isLowercaseRequired}
        onChange={onIsLowercaseRequiredChange}
      />
      <Checkbox
        className={styles.checkbox}
        label={renderToString("PasswordsScreen.password-policy.require-uppercase.label")}
        checked={state.isUppercaseRequired}
        onChange={onIsUppercaseRequiredChange}
      />
      <Checkbox
        className={styles.checkbox}
        label={renderToString("PasswordsScreen.password-policy.require-symbol.label")}
        checked={state.isSymbolRequired}
        onChange={onIsSymbolRequiredChange}
      />

      <Dropdown
        className={styles.dropdown}
        label={renderToString("PasswordsScreen.password-policy.min-guessable-level.label")}
        options={minGuessableLevelOptions}
        onChange={onMinGuessableLevelOptionChange}
      />

      <ToggleWithContent
        className={styles.toggleWithContent}
        checked={state.preventReuse}
        inlineLabel={true}
        onChange={onPreventReuseChange}
      >
        <Label className={styles.toggleLabel}>
          <FormattedMessage id="PasswordsScreen.password-policy.prevent-reuse.label" />
        </Label>
        <TextField
          className={styles.textField}
          type="number"
          min="0"
          step="1"
          disabled={!state.preventReuse}
          label={renderToString("PasswordsScreen.password-policy.history-days.label")}
          value={`${state.historyDays}`}
          onChange={onHistoryDaysChange}
        />
        <TextField
          className={styles.textField}
          type="number"
          min="0"
          step="1"
          disabled={!state.preventReuse}
          label={renderToString("PasswordsScreen.password-policy.history-size.label")}
          value={`${state.historySize}`}
          onChange={onHistorySizeChange}
        />
      </ToggleWithContent>

      <Label className={styles.tagPickerLabel}>
        <FormattedMessage id="PasswordsScreen.password-policy.excluded-keywords.label" />
      </Label>
      <TagPicker
        className={styles.tagPicker}
        inputProps={{
          "aria-label": renderToString(
            "PasswordsScreen.password-policy.excluded-keywords.label"
          ),
        }}
        defaultSelectedItems={defaultSelectedExcludedKeywordItems}
        onResolveSuggestions={onResolveExcludedKeywordSuggestions}
        onChange={onExcludedKeywordsChange}
      />

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

export default PasswordPolicySettings;
