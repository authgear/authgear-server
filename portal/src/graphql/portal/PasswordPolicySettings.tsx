import React, { useCallback, useContext, useMemo } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Checkbox,
  Dropdown,
  IDropdownOption,
  ITag,
  Label,
  TagPicker,
  TextField,
} from "@fluentui/react";
import deepEqual from "deep-equal";
import produce from "immer";

import ToggleWithContent from "../../ToggleWithContent";
import { clearEmptyObject } from "../../util/misc";
import {
  isPasswordPolicyGuessableLevel,
  PasswordPolicyConfig,
  passwordPolicyGuessableLevels,
  PortalAPIAppConfig,
} from "../../types";
import { useParams } from "react-router-dom";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import FormContainer from "../../FormContainer";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";

import styles from "./PasswordPolicySettings.module.scss";

interface FormState {
  policy: Required<PasswordPolicyConfig>;
  isPreventPasswordReuseEnabled: boolean;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const policy: Required<PasswordPolicyConfig> = {
    min_length: 1,
    uppercase_required: false,
    lowercase_required: false,
    digit_required: false,
    symbol_required: false,
    minimum_guessable_level: 0,
    excluded_keywords: [],
    history_size: 0,
    history_days: 0,
    ...config.authenticator?.password?.policy,
  };
  return {
    policy,
    isPreventPasswordReuseEnabled:
      policy.history_days > 0 || policy.history_size > 0,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.authenticator ??= {};
    config.authenticator.password ??= {};
    config.authenticator.password.policy ??= {};
    const policy = config.authenticator.password.policy;
    const initial = initialState.policy;
    const current = currentState.policy;

    if (initial.min_length !== current.min_length) {
      policy.min_length = current.min_length;
    }
    if (initial.uppercase_required !== current.uppercase_required) {
      policy.uppercase_required = current.uppercase_required;
    }
    if (initial.lowercase_required !== current.lowercase_required) {
      policy.lowercase_required = current.lowercase_required;
    }
    if (initial.digit_required !== current.digit_required) {
      policy.digit_required = current.digit_required;
    }
    if (initial.symbol_required !== current.symbol_required) {
      policy.symbol_required = current.symbol_required;
    }
    if (initial.minimum_guessable_level !== current.minimum_guessable_level) {
      policy.minimum_guessable_level = current.minimum_guessable_level;
    }
    if (
      !deepEqual(initial.excluded_keywords, current.excluded_keywords, {
        strict: true,
      })
    ) {
      policy.excluded_keywords = current.excluded_keywords;
    }

    function effectiveHistorySize(s: FormState) {
      return s.isPreventPasswordReuseEnabled ? s.policy.history_size : 0;
    }

    function effectiveHistoryDays(s: FormState) {
      return s.isPreventPasswordReuseEnabled ? s.policy.history_days : 0;
    }

    if (
      effectiveHistorySize(initialState) !== effectiveHistorySize(currentState)
    ) {
      policy.history_size = effectiveHistorySize(currentState);
    }
    if (
      effectiveHistoryDays(initialState) !== effectiveHistoryDays(currentState)
    ) {
      policy.history_days = effectiveHistoryDays(currentState);
    }

    clearEmptyObject(config);
  });
}

interface PasswordPolicySettingsContentProps {
  form: AppConfigFormModel<FormState>;
}

const PasswordPolicySettingsContent: React.FC<PasswordPolicySettingsContentProps> = function PasswordPolicySettingsContent(
  props
) {
  const { state, setState } = props.form;

  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="PasswordPolicySettingsScreen.title" />,
      },
    ];
  }, []);

  const minGuessableLevelOptions: IDropdownOption[] = useMemo(() => {
    return passwordPolicyGuessableLevels.map((level) => ({
      key: level,
      text: renderToString(
        `PasswordPolicySettingsScreen.min-guessable-level.${level}`
      ),
      isSelected: level === state.policy.minimum_guessable_level,
    }));
  }, [state.policy.minimum_guessable_level, renderToString]);

  const defaultSelectedExcludedKeywordItems: ITag[] = useMemo(() => {
    return state.policy.excluded_keywords.map((keyword) => ({
      key: keyword,
      name: keyword,
    }));
  }, [state.policy.excluded_keywords]);

  const setPolicy = useCallback(
    (policy: PasswordPolicyConfig) =>
      setState((state) => ({
        ...state,
        policy: { ...state.policy, ...policy },
      })),
    [setState]
  );

  const onMinLengthChange = useCallback(
    (_, value?: string) =>
      setPolicy({
        min_length: Number(value),
      }),
    [setPolicy]
  );

  const onUppercaseRequiredChange = useCallback(
    (_, value?: boolean) =>
      setPolicy({
        uppercase_required: value ?? false,
      }),
    [setPolicy]
  );

  const onLowercaseRequiredChange = useCallback(
    (_, value?: boolean) =>
      setPolicy({
        lowercase_required: value ?? false,
      }),
    [setPolicy]
  );

  const onDigitRequiredChange = useCallback(
    (_, value?: boolean) =>
      setPolicy({
        digit_required: value ?? false,
      }),
    [setPolicy]
  );

  const onSymbolRequiredChange = useCallback(
    (_, value?: boolean) =>
      setPolicy({
        symbol_required: value ?? false,
      }),
    [setPolicy]
  );

  const onMinimumGuessableLevelChange = useCallback(
    (_, option?: IDropdownOption) => {
      const key = option?.key;
      if (!isPasswordPolicyGuessableLevel(key)) {
        return;
      }
      setPolicy({ minimum_guessable_level: key });
    },
    [setPolicy]
  );

  const onPreventReuseChange = useCallback(
    (_, checked?: boolean) => {
      if (checked == null) {
        return;
      }
      if (checked) {
        setState((state) => ({
          isPreventPasswordReuseEnabled: true,
          policy: {
            ...state.policy,
            history_days:
              state.policy.history_days === 0 ? 90 : state.policy.history_days,
            history_size:
              state.policy.history_size === 0 ? 3 : state.policy.history_size,
          },
        }));
      } else {
        setState((state) => ({
          isPreventPasswordReuseEnabled: false,
          policy: state.policy,
        }));
      }
    },
    [setState]
  );

  const onHistoryDaysChange = useCallback(
    (_, value?: string) =>
      setPolicy({
        history_days: Number(value),
      }),
    [setPolicy]
  );

  const onHistorySizeChange = useCallback(
    (_, value?: string) =>
      setPolicy({
        history_size: Number(value),
      }),
    [setPolicy]
  );

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
      setPolicy({
        excluded_keywords: items.map((item) => item.name),
      });
    },
    [setPolicy]
  );

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <TextField
        className={styles.textField}
        type="number"
        min="1"
        step="1"
        label={renderToString("PasswordPolicySettingsScreen.min-length.label")}
        value={String(state.policy.min_length)}
        onChange={onMinLengthChange}
      />
      <Checkbox
        className={styles.checkbox}
        label={renderToString(
          "PasswordPolicySettingsScreen.require-digit.label"
        )}
        checked={state.policy.digit_required}
        onChange={onDigitRequiredChange}
      />
      <Checkbox
        className={styles.checkbox}
        label={renderToString(
          "PasswordPolicySettingsScreen.require-lowercase.label"
        )}
        checked={state.policy.lowercase_required}
        onChange={onLowercaseRequiredChange}
      />
      <Checkbox
        className={styles.checkbox}
        label={renderToString(
          "PasswordPolicySettingsScreen.require-uppercase.label"
        )}
        checked={state.policy.uppercase_required}
        onChange={onUppercaseRequiredChange}
      />
      <Checkbox
        className={styles.checkbox}
        label={renderToString(
          "PasswordPolicySettingsScreen.require-symbol.label"
        )}
        checked={state.policy.symbol_required}
        onChange={onSymbolRequiredChange}
      />

      <Dropdown
        className={styles.dropdown}
        label={renderToString(
          "PasswordPolicySettingsScreen.min-guessable-level.label"
        )}
        options={minGuessableLevelOptions}
        selectedKey={state.policy.minimum_guessable_level}
        onChange={onMinimumGuessableLevelChange}
      />

      <ToggleWithContent
        className={styles.toggleWithContent}
        checked={state.isPreventPasswordReuseEnabled}
        inlineLabel={true}
        onChange={onPreventReuseChange}
      >
        <Label className={styles.toggleLabel}>
          <FormattedMessage id="PasswordPolicySettingsScreen.prevent-reuse.label" />
        </Label>
        <TextField
          className={styles.textField}
          type="number"
          min="0"
          step="1"
          disabled={!state.isPreventPasswordReuseEnabled}
          label={renderToString(
            "PasswordPolicySettingsScreen.history-days.label"
          )}
          value={String(state.policy.history_days)}
          onChange={onHistoryDaysChange}
        />
        <TextField
          className={styles.textField}
          type="number"
          min="0"
          step="1"
          disabled={!state.isPreventPasswordReuseEnabled}
          label={renderToString(
            "PasswordPolicySettingsScreen.history-size.label"
          )}
          value={String(state.policy.history_size)}
          onChange={onHistorySizeChange}
        />
      </ToggleWithContent>

      <Label className={styles.tagPickerLabel}>
        <FormattedMessage id="PasswordPolicySettingsScreen.excluded-keywords.label" />
      </Label>
      <TagPicker
        className={styles.tagPicker}
        inputProps={{
          "aria-label": renderToString(
            "PasswordPolicySettingsScreen.excluded-keywords.label"
          ),
        }}
        defaultSelectedItems={defaultSelectedExcludedKeywordItems}
        onResolveSuggestions={onResolveExcludedKeywordSuggestions}
        onChange={onExcludedKeywordsChange}
      />
    </div>
  );
};

const PasswordPolicySettingsScreen: React.FC = function PasswordPolicySettingsScreen() {
  const { appID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <PasswordPolicySettingsContent form={form} />
    </FormContainer>
  );
};

export default PasswordPolicySettingsScreen;
