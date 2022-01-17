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
  Toggle,
} from "@fluentui/react";
import deepEqual from "deep-equal";
import produce from "immer";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
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
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import FormContainer from "../../FormContainer";
import { fixTagPickerStyles } from "../../bugs";

import styles from "./PasswordConfigurationScreen.module.scss";

interface FormState {
  primaryAuthenticatorEnabled: boolean;
  forceChange: boolean;
  policy: PasswordPolicyConfig;
  isPreventPasswordReuseEnabled: boolean;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const policy: PasswordPolicyConfig = {
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
    primaryAuthenticatorEnabled:
      config.authentication?.primary_authenticators?.includes("password") ??
      false,
    forceChange: config.authenticator?.password?.force_change ?? true,
    policy,
    isPreventPasswordReuseEnabled:
      (policy.history_days != null && policy.history_days > 0) ||
      (policy.history_size != null && policy.history_size > 0),
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.authenticator ??= {};
    config.authenticator.password ??= {};
    config.authenticator.password.policy ??= {};
    const policy = config.authenticator.password.policy;
    const initial = initialState.policy;
    const current = currentState.policy;

    if (
      initialState.primaryAuthenticatorEnabled !==
      currentState.primaryAuthenticatorEnabled
    ) {
      config.authentication ??= {};
      config.authentication.primary_authenticators ??= [];

      if (config.authentication.primary_authenticators.includes("password")) {
        if (!currentState.primaryAuthenticatorEnabled) {
          config.authentication.primary_authenticators =
            config.authentication.primary_authenticators.filter(
              (p) => p !== "password"
            );
        }
      } else {
        if (currentState.primaryAuthenticatorEnabled) {
          config.authentication.primary_authenticators.push("password");
        }
      }
    }

    if (initialState.forceChange !== currentState.forceChange) {
      if (currentState.forceChange) {
        config.authenticator.password.force_change = undefined;
      } else {
        config.authenticator.password.force_change = false;
      }
    }

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

interface PasswordConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const PasswordConfigurationScreenContent: React.FC<PasswordConfigurationScreenContentProps> =
  function PasswordConfigurationScreenContent(props) {
    const { state, setState } = props.form;

    const { renderToString } = useContext(Context);

    const minGuessableLevelOptions: IDropdownOption[] = useMemo(() => {
      return passwordPolicyGuessableLevels.map((level) => ({
        key: level,
        text: renderToString(
          `PasswordConfigurationScreen.min-guessable-level.${level}`
        ),
        isSelected: level === state.policy.minimum_guessable_level,
      }));
    }, [state.policy.minimum_guessable_level, renderToString]);

    const defaultSelectedExcludedKeywordItems: ITag[] = useMemo(() => {
      return (
        state.policy.excluded_keywords?.map((keyword) => ({
          key: keyword,
          name: keyword,
        })) ?? []
      );
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
      (_, value?: string) => {
        setPolicy({
          min_length: parseIntegerAllowLeadingZeros(value),
        });
      },
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
            ...state,
            isPreventPasswordReuseEnabled: true,
            policy: {
              ...state.policy,
              history_days:
                state.policy.history_days === 0
                  ? 90
                  : state.policy.history_days,
              history_size:
                state.policy.history_size === 0 ? 3 : state.policy.history_size,
            },
          }));
        } else {
          setState((state) => ({
            ...state,
            isPreventPasswordReuseEnabled: false,
            policy: state.policy,
          }));
        }
      },
      [setState]
    );

    const onHistoryDaysChange = useCallback(
      (_, value?: string) => {
        setPolicy({
          history_days: parseIntegerAllowLeadingZeros(value),
        });
      },
      [setPolicy]
    );

    const onHistorySizeChange = useCallback(
      (_, value?: string) => {
        setPolicy({
          history_size: parseIntegerAllowLeadingZeros(value),
        });
      },
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

    const onPrimaryAuthenticatorChange = useCallback(
      (_, checked?: boolean) => {
        if (checked == null) {
          return;
        }
        setState((state) => ({
          ...state,
          primaryAuthenticatorEnabled: checked,
        }));
      },
      [setState]
    );

    const onForceChangeChange = useCallback(
      (_, checked?: boolean) => {
        if (checked == null) {
          return;
        }
        setState((state) => ({
          ...state,
          forceChange: checked,
        }));
      },
      [setState]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="PasswordConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="PasswordConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <Toggle
            checked={state.primaryAuthenticatorEnabled}
            inlineLabel={true}
            label={
              <FormattedMessage id="PasswordConfigurationScreen.primary-authenticator-enabled.label" />
            }
            onChange={onPrimaryAuthenticatorChange}
          />
        </Widget>
        <Widget className={styles.widget}>
          <Toggle
            checked={state.forceChange}
            inlineLabel={true}
            label={
              <FormattedMessage id="PasswordConfigurationScreen.force-change.label" />
            }
            onChange={onForceChangeChange}
          />
        </Widget>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="PasswordConfigurationScreen.basic-policies" />
          </WidgetTitle>
          <TextField
            type="text"
            label={renderToString(
              "PasswordConfigurationScreen.min-length.label"
            )}
            value={state.policy.min_length?.toFixed(0) ?? ""}
            onChange={onMinLengthChange}
          />
          <Checkbox
            label={renderToString(
              "PasswordConfigurationScreen.require-digit.label"
            )}
            checked={state.policy.digit_required}
            onChange={onDigitRequiredChange}
          />
          <Checkbox
            label={renderToString(
              "PasswordConfigurationScreen.require-lowercase.label"
            )}
            checked={state.policy.lowercase_required}
            onChange={onLowercaseRequiredChange}
          />
          <Checkbox
            label={renderToString(
              "PasswordConfigurationScreen.require-uppercase.label"
            )}
            checked={state.policy.uppercase_required}
            onChange={onUppercaseRequiredChange}
          />
          <Checkbox
            label={renderToString(
              "PasswordConfigurationScreen.require-symbol.label"
            )}
            checked={state.policy.symbol_required}
            onChange={onSymbolRequiredChange}
          />
        </Widget>

        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="PasswordConfigurationScreen.advanced-policies" />
          </WidgetTitle>
          <Dropdown
            label={renderToString(
              "PasswordConfigurationScreen.min-guessable-level.label"
            )}
            options={minGuessableLevelOptions}
            selectedKey={state.policy.minimum_guessable_level}
            onChange={onMinimumGuessableLevelChange}
          />
          <Toggle
            checked={state.isPreventPasswordReuseEnabled}
            inlineLabel={true}
            label={
              <FormattedMessage id="PasswordConfigurationScreen.prevent-reuse.label" />
            }
            onChange={onPreventReuseChange}
          />
          <TextField
            type="text"
            disabled={!state.isPreventPasswordReuseEnabled}
            label={renderToString(
              "PasswordConfigurationScreen.history-days.label"
            )}
            value={state.policy.history_days?.toFixed(0) ?? ""}
            onChange={onHistoryDaysChange}
          />
          <TextField
            type="text"
            disabled={!state.isPreventPasswordReuseEnabled}
            label={renderToString(
              "PasswordConfigurationScreen.history-size.label"
            )}
            value={state.policy.history_size?.toFixed(0) ?? ""}
            onChange={onHistorySizeChange}
          />
          <div>
            <Label>
              <FormattedMessage id="PasswordConfigurationScreen.excluded-keywords.label" />
            </Label>
            <TagPicker
              styles={fixTagPickerStyles}
              inputProps={{
                "aria-label": renderToString(
                  "PasswordConfigurationScreen.excluded-keywords.label"
                ),
              }}
              defaultSelectedItems={defaultSelectedExcludedKeywordItems}
              onResolveSuggestions={onResolveExcludedKeywordSuggestions}
              onChange={onExcludedKeywordsChange}
            />
          </div>
        </Widget>
      </ScreenContent>
    );
  };

const PasswordConfigurationScreen: React.FC =
  function PasswordConfigurationScreen() {
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
        <PasswordConfigurationScreenContent form={form} />
      </FormContainer>
    );
  };

export default PasswordConfigurationScreen;
