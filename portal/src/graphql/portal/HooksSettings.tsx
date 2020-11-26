import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
  Dropdown,
  IDropdownOption,
  ISelectableOption,
  Label,
  TextField,
} from "@fluentui/react";
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
import { useFormField } from "../../error/FormFieldContext";
import FieldList from "../../FieldList";
import { useValidationError } from "../../error/useValidationError";
import { FormContext } from "../../error/FormContext";
import ShowUnhandledValidationErrorCause from "../../error/ShowUnhandledValidationErrorCauses";

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

const hookEventTypes: IDropdownOption[] = [
  "before_user_create",
  "after_user_create",
].map((type): IDropdownOption => ({ key: type, text: type }));

interface HookHandlerItemEditProps {
  index: number;
  value: HookEventHandler;
  onChange: (newValue: HookEventHandler) => void;
}
const HookHandlerItemEdit: React.FC<HookHandlerItemEditProps> = function HookHandlerItemEdit(
  props
) {
  const { index, value, onChange } = props;

  const parentJSONPointer = "/hook/handlers";
  const jsonPointer = `/hook/handlers/${index}`;

  const { errorMessage: eventErrorMessage } = useFormField(
    jsonPointer + "/event",
    parentJSONPointer,
    "event"
  );
  const { errorMessage: urlErrorMessage } = useFormField(
    jsonPointer + "/url",
    parentJSONPointer,
    "url"
  );

  const onEventChange = useCallback(
    (_, event?: IDropdownOption) => {
      onChange({ ...value, event: String(event?.key ?? "") });
    },
    [onChange, value]
  );
  const onURLChange = useCallback(
    (_, url?: string) => {
      onChange({ ...value, url: url ?? "" });
    },
    [onChange, value]
  );

  const renderEventDropdownItem = useCallback((item?: ISelectableOption) => {
    return (
      <span>
        <FormattedMessage id={`HooksSettings.event-type.${item?.key}`} />
      </span>
    );
  }, []);
  const renderEventDropdownTitle = useCallback((items?: IDropdownOption[]) => {
    return (
      <span>
        <FormattedMessage id={`HooksSettings.event-type.${items?.[0].key}`} />
      </span>
    );
  }, []);

  return (
    <div className={styles.handlerEdit}>
      <Dropdown
        className={styles.handlerEventField}
        options={hookEventTypes}
        selectedKey={value.event}
        onChange={onEventChange}
        onRenderOption={renderEventDropdownItem}
        onRenderTitle={renderEventDropdownTitle}
        ariaLabel={"HooksSettings.events.label"}
        errorMessage={eventErrorMessage}
      />
      <TextField
        className={styles.handlerURLField}
        value={value.url}
        onChange={onURLChange}
        errorMessage={urlErrorMessage}
      />
    </div>
  );
};

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

  const makeDefaultHandler = useCallback(
    (): HookEventHandler => ({ event: String(hookEventTypes[0].key), url: "" }),
    []
  );
  const renderHandlerItem = useCallback(
    (
      index: number,
      value: HookEventHandler,
      onChange: (newValue: HookEventHandler) => void
    ) => (
      <HookHandlerItemEdit index={index} value={value} onChange={onChange} />
    ),
    []
  );
  const onHandlersChange = useCallback(
    (value: HookEventHandler[]) => {
      update((state) => ({ ...state, handlers: value }));
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

      <FieldList
        className={styles.handlerList}
        label={
          <Label>
            <FormattedMessage id="HooksSettings.handlers.label" />
          </Label>
        }
        jsonPointer="/hook/handlers"
        parentJSONPointer="/hook"
        fieldName="handlers"
        list={form.handlers}
        onListChange={onHandlersChange}
        makeDefaultItem={makeDefaultHandler}
        renderListItem={renderHandlerItem}
        addButtonLabelMessageID="add"
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

  const {
    otherError,
    unhandledCauses,
    value: formContextValue,
  } = useValidationError(updateAppConfigError);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <FormContext.Provider value={formContextValue}>
      <main className={styles.root}>
        <ShowUnhandledValidationErrorCause causes={unhandledCauses} />
        {(unhandledCauses ?? []).length === 0 && otherError && (
          <ShowError error={otherError} />
        )}
        <HooksSettingsContent
          form={form}
          isDirty={isDirty}
          isSaving={updatingAppConfig}
          update={update}
          reset={reset}
          save={save}
        />
      </main>
    </FormContext.Provider>
  );
};

export default HooksSettings;
