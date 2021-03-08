import { Context, FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";
import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import {
  Dropdown,
  IDropdownOption,
  ISelectableOption,
  Label,
  TextField,
} from "@fluentui/react";
import produce from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import { PortalAPIAppConfig } from "../../types";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useFormField } from "../../form";
import FieldList from "../../FieldList";
import FormContainer from "../../FormContainer";
import { clearEmptyObject } from "../../util/misc";
import styles from "./WebhookConfigurationScreen.module.scss";
import { renderErrors } from "../../error/parse";

interface HookEventHandler {
  event: string;
  url: string;
}

interface FormState {
  timeout: number;
  totalTimeout: number;
  handlers: HookEventHandler[];
}

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
    clearEmptyObject(config);
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
  const { renderToString } = useContext(Context);

  const eventField = useMemo(
    () => ({
      parentJSONPointer: `/hook/handlers/${index}`,
      fieldName: "event",
    }),
    [index]
  );
  const urlField = useMemo(
    () => ({
      parentJSONPointer: `/hook/handlers/${index}`,
      fieldName: "url",
    }),
    [index]
  );
  const { errors: eventErrors } = useFormField(eventField);
  const { errors: urlErrors } = useFormField(urlField);
  const eventErrorMessage = useMemo(
    () => renderErrors(eventField, eventErrors, renderToString),
    [eventField, eventErrors, renderToString]
  );
  const urlErrorMessage = useMemo(
    () => renderErrors(urlField, urlErrors, renderToString),
    [urlField, urlErrors, renderToString]
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
        <FormattedMessage
          id={`WebhookConfigurationScreen.event-type.${item?.key}`}
        />
      </span>
    );
  }, []);
  const renderEventDropdownTitle = useCallback((items?: IDropdownOption[]) => {
    return (
      <span>
        <FormattedMessage
          id={`WebhookConfigurationScreen.event-type.${items?.[0].key}`}
        />
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
        ariaLabel={"WebhookConfigurationScreen.events.label"}
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

interface WebhookConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
}

const WebhookConfigurationScreenContent: React.FC<WebhookConfigurationScreenContentProps> = function WebhookConfigurationScreenContent(
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
    <ScreenContent className={styles.root}>
      <ScreenTitle>
        <FormattedMessage id="WebhookConfigurationScreen.title" />
      </ScreenTitle>
      <ScreenDescription className={styles.widget}>
        <FormattedMessage id="WebhookConfigurationScreen.description" />
      </ScreenDescription>
      <Widget className={cn(styles.widget, styles.controlGroup)}>
        <WidgetTitle>
          <FormattedMessage id="WebhookConfigurationScreen.webhook-settings" />
        </WidgetTitle>
        <TextField
          className={styles.control}
          type="number"
          min="1"
          step="1"
          label={renderToString(
            "WebhookConfigurationScreen.total-timeout.label"
          )}
          value={String(state.totalTimeout)}
          onChange={onTotalTimeoutChange}
        />
        <TextField
          className={styles.control}
          type="number"
          min="1"
          step="1"
          label={renderToString("WebhookConfigurationScreen.timeout.label")}
          value={String(state.timeout)}
          onChange={onTimeoutChange}
        />
        <FieldList
          className={styles.control}
          label={
            <Label>
              <FormattedMessage id="WebhookConfigurationScreen.handlers.label" />
            </Label>
          }
          parentJSONPointer="/hook"
          fieldName="handlers"
          list={state.handlers}
          onListChange={onHandlersChange}
          makeDefaultItem={makeDefaultHandler}
          renderListItem={renderHandlerItem}
          addButtonLabelMessageID="add"
        />
      </Widget>
    </ScreenContent>
  );
};

const WebhookConfigurationScreen: React.FC = function WebhookConfigurationScreen() {
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
      <WebhookConfigurationScreenContent form={form} />
    </FormContainer>
  );
};

export default WebhookConfigurationScreen;
