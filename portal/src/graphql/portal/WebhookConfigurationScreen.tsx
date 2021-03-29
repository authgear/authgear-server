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
import WidgetDescription from "../../WidgetDescription";

interface BlockingEventHandler {
  event: string;
  url: string;
}
interface FormState {
  timeout: number;
  totalTimeout: number;
  blocking_handlers: BlockingEventHandler[];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    timeout: config.hook?.sync_hook_timeout_seconds ?? 0,
    totalTimeout: config.hook?.sync_hook_total_timeout_seconds ?? 0,
    blocking_handlers: config.hook?.blocking_handlers ?? [],
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
    config.hook.blocking_handlers = currentState.blocking_handlers;
    clearEmptyObject(config);
  });
}

const blockingEventTypes: IDropdownOption[] = [
  "pre_signup",
  "admin_api_create_user",
].map((type): IDropdownOption => ({ key: type, text: type }));

interface BlockingHandlerItemEditProps {
  index: number;
  value: BlockingEventHandler;
  onChange: (newValue: BlockingEventHandler) => void;
}
const BlockingHandlerItemEdit: React.FC<BlockingHandlerItemEditProps> = function BlockingHandlerItemEdit(
  props
) {
  const { index, value, onChange } = props;
  const { renderToString } = useContext(Context);

  const eventField = useMemo(
    () => ({
      parentJSONPointer: `/hook/blocking_handlers/${index}`,
      fieldName: "event",
    }),
    [index]
  );
  const urlField = useMemo(
    () => ({
      parentJSONPointer: `/hook/blocking_handlers/${index}`,
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

  const onBlockingEventChange = useCallback(
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

  const renderBlockingEventDropdownItem = useCallback(
    (item?: ISelectableOption) => {
      return (
        <span>
          <FormattedMessage
            id={`WebhookConfigurationScreen.blocking-event-type.${item?.key}`}
          />
        </span>
      );
    },
    []
  );
  const renderBlockingEventDropdownTitle = useCallback(
    (items?: IDropdownOption[]) => {
      return (
        <span>
          <FormattedMessage
            id={`WebhookConfigurationScreen.blocking-event-type.${items?.[0].key}`}
          />
        </span>
      );
    },
    []
  );

  return (
    <div className={styles.handlerEdit}>
      <Dropdown
        className={styles.handlerEventField}
        options={blockingEventTypes}
        selectedKey={value.event}
        onChange={onBlockingEventChange}
        onRenderOption={renderBlockingEventDropdownItem}
        onRenderTitle={renderBlockingEventDropdownTitle}
        ariaLabel={"WebhookConfigurationScreen.blocking-events.label"}
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
    (): BlockingEventHandler => ({
      event: String(blockingEventTypes[0].key),
      url: "",
    }),
    []
  );
  const renderBlockingHandlerItem = useCallback(
    (
      index: number,
      value: BlockingEventHandler,
      onChange: (newValue: BlockingEventHandler) => void
    ) => (
      <BlockingHandlerItemEdit
        index={index}
        value={value}
        onChange={onChange}
      />
    ),
    []
  );
  const onBlockingHandlersChange = useCallback(
    (value: BlockingEventHandler[]) => {
      setState((state) => ({ ...state, blocking_handlers: value }));
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
      </Widget>
      <Widget className={cn(styles.widget, styles.controlGroup)}>
        <WidgetTitle>
          <FormattedMessage id="WebhookConfigurationScreen.blocking-events" />
        </WidgetTitle>
        <WidgetDescription>
          <FormattedMessage id="WebhookConfigurationScreen.blocking-events.description" />
        </WidgetDescription>
        <FieldList
          className={styles.control}
          label={
            <Label>
              <FormattedMessage id="WebhookConfigurationScreen.blocking-handlers.label" />
            </Label>
          }
          parentJSONPointer="/hook"
          fieldName="blocking_handlers"
          list={state.blocking_handlers}
          onListChange={onBlockingHandlersChange}
          makeDefaultItem={makeDefaultHandler}
          renderListItem={renderBlockingHandlerItem}
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
