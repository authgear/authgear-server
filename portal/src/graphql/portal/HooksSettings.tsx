import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useCallback, useContext } from "react";
import { useParams } from "react-router-dom";
import {
  Dropdown,
  IDropdownOption,
  ISelectableOption,
  Label,
  Text,
  TextField,
} from "@fluentui/react";
import produce from "immer";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { PortalAPIAppConfig } from "../../types";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useFormField } from "../../error/FormFieldContext";
import FieldList from "../../FieldList";
import FormContainer from "../../FormContainer";
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
  form: AppConfigFormModel<FormState>;
}

const HooksSettingsContent: React.FC<HooksSettingsContentProps> = function HooksSettingsContent(
  props
) {
  const { state, setState } = props.form;

  const { renderToString } = useContext(Context);

  const onTimeoutChange = useCallback(
    (_, value?: string) => {
      setState((state) => ({
        ...state,
        timeout: Number(value),
      }));
    },
    [setState]
  );

  const onTotalTimeoutChange = useCallback(
    (_, value?: string) => {
      setState((state) => ({
        ...state,
        totalTimeout: Number(value),
      }));
    },
    [setState]
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
      setState((state) => ({ ...state, handlers: value }));
    },
    [setState]
  );

  return (
    <div className={styles.root}>
      <Text as="h1" variant="xLarge" block={true}>
        <FormattedMessage id="HooksSettings.title" />
      </Text>
      <TextField
        className={styles.textField}
        type="number"
        min="1"
        step="1"
        label={renderToString("HooksSettings.total-timeout.label")}
        value={String(state.totalTimeout)}
        onChange={onTotalTimeoutChange}
      />
      <TextField
        className={styles.textField}
        type="number"
        min="1"
        step="1"
        label={renderToString("HooksSettings.timeout.label")}
        value={String(state.timeout)}
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
        list={state.handlers}
        onListChange={onHandlersChange}
        makeDefaultItem={makeDefaultHandler}
        renderListItem={renderHandlerItem}
        addButtonLabelMessageID="add"
      />
    </div>
  );
};

const HooksSettings: React.FC = function HooksSettings() {
  const { appID } = useParams();
  const form = useAppConfigForm(
    appID,
    emptyFormState,
    constructFormState,
    constructConfig
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <HooksSettingsContent form={form} />
    </FormContainer>
  );
};

export default HooksSettings;
