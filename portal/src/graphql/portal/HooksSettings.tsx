import { Context } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import { TextField } from "@fluentui/react";
import produce from "immer";
import deepEqual from "deep-equal";

import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { useAppConfigQuery } from "./query/appConfigQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { ModifiedIndicatorPortal } from "../../ModifiedIndicatorPortal";
import { PortalAPIAppConfig } from "../../types";

import styles from "./HooksSettings.module.scss";

interface HookEventHandler {
  event: string;
  url: string;
}

interface FormState {
  timeout: number;
  totalTimeout: number;
  handlers: HookEventHandler[];
}

const emptyFormState: FormState = {
  timeout: 0,
  totalTimeout: 0,
  handlers: [],
};

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    timeout: config.hook?.sync_hook_timeout_seconds ?? 0,
    totalTimeout: config.hook?.sync_hook_total_timeout_seconds ?? 0,
    handlers: config.hook?.handlers ?? [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.hook = config.hook ?? {};
    if (initialState.timeout !== currentState.timeout) {
      config.hook.sync_hook_timeout_seconds = currentState.timeout;
    }
    if (initialState.totalTimeout !== currentState.totalTimeout) {
      config.hook.sync_hook_total_timeout_seconds = currentState.totalTimeout;
    }
    config.hook.handlers = currentState.handlers;
  });
}

interface HooksSettingsContentProps {
  form: FormState;
  isDirty: boolean;
  isSaving: boolean;
  update: (fn: (state: FormState) => FormState) => void;
  reset: () => void;
  save: () => void;
}

const HooksSettingsContent: React.FC<HooksSettingsContentProps> = function HooksSettingsContent(
  props
) {
  const { form, isDirty, isSaving, update, reset, save } = props;

  const { renderToString } = useContext(Context);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      save();
    },
    [save]
  );

  const onTimeoutChange = useCallback(
    (_, value?: string) => {
      update((state) => ({
        ...state,
        timeout: Number(value),
      }));
    },
    [update]
  );

  const onTotalTimeoutChange = useCallback(
    (_, value?: string) => {
      update((state) => ({
        ...state,
        totalTimeout: Number(value),
      }));
    },
    [update]
  );

  return (
    <form onSubmit={onFormSubmit}>
      <ModifiedIndicatorPortal resetForm={reset} isModified={isDirty} />
      <TextField
        className={styles.textField}
        type="number"
        min="1"
        step="1"
        label={renderToString("HooksSettings.total-timeout.label")}
        value={String(form.totalTimeout)}
        onChange={onTotalTimeoutChange}
      />
      <TextField
        className={styles.textField}
        type="number"
        min="1"
        step="1"
        label={renderToString("HooksSettings.timeout.label")}
        value={String(form.timeout)}
        onChange={onTimeoutChange}
      />

      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          type="submit"
          disabled={!isDirty}
          loading={isSaving}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>
      <NavigationBlockerDialog blockNavigation={isDirty} />
    </form>
  );
};

const HooksSettings: React.FC = function HooksSettings() {
  const { appID } = useParams();

  // TODO: extract app config form logic as hook
  const {
    loading,
    error,
    effectiveAppConfig,
    rawAppConfig,
    refetch,
  } = useAppConfigQuery(appID);
  const {
    loading: updatingAppConfig,
    error: updateAppConfigError,
    updateAppConfig,
    resetError: resetUpdateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const initialFormState = useMemo(
    () => effectiveAppConfig && constructFormState(effectiveAppConfig),
    [effectiveAppConfig]
  );
  const [currentFormState, setCurrentFormState] = useState<FormState | null>(
    null
  );

  const isDirty = useMemo(
    () =>
      Boolean(
        rawAppConfig &&
          initialFormState &&
          currentFormState &&
          !deepEqual(
            constructConfig(rawAppConfig, initialFormState, initialFormState),
            constructConfig(rawAppConfig, initialFormState, currentFormState),
            { strict: true }
          )
      ),
    [rawAppConfig, initialFormState, currentFormState]
  );

  const reset = useCallback(() => {
    resetUpdateAppConfigError();
    setCurrentFormState(initialFormState);
  }, [resetUpdateAppConfigError, initialFormState]);

  const save = useCallback(() => {
    if (!rawAppConfig || !initialFormState || !currentFormState) {
      return;
    }

    const newConfig = constructConfig(
      rawAppConfig,
      initialFormState,
      currentFormState
    );
    updateAppConfig(newConfig)
      .then(
        (app) =>
          app?.effectiveAppConfig &&
          setCurrentFormState(constructFormState(app.effectiveAppConfig))
      )
      .catch(() => {});
  }, [rawAppConfig, initialFormState, currentFormState, updateAppConfig]);

  const form = currentFormState ?? initialFormState ?? emptyFormState;
  const update = useCallback(
    (fn: (state: FormState) => FormState) => {
      setCurrentFormState(fn(form));
    },
    [form]
  );

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={styles.root}>
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <HooksSettingsContent
        form={form}
        isDirty={isDirty}
        isSaving={updatingAppConfig}
        update={update}
        reset={reset}
        save={save}
      />
    </main>
  );
};

export default HooksSettings;
