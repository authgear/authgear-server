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
  MessageBar,
  Link,
} from "@fluentui/react";
import produce from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import { HookFeatureConfig, PortalAPIAppConfig } from "../../types";
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
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

interface BlockingEventHandler {
  event: string;
  url: string;
}

interface NonBlockingEventHandler {
  events: string[];
  url: string;
}
interface FormState {
  timeout: number;
  totalTimeout: number;
  blocking_handlers: BlockingEventHandler[];
  non_blocking_handlers: NonBlockingEventHandler[];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    timeout: config.hook?.sync_hook_timeout_seconds ?? 0,
    totalTimeout: config.hook?.sync_hook_total_timeout_seconds ?? 0,
    blocking_handlers: config.hook?.blocking_handlers ?? [],
    non_blocking_handlers: config.hook?.non_blocking_handlers ?? [],
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
    config.hook.non_blocking_handlers = currentState.non_blocking_handlers;
    clearEmptyObject(config);
  });
}

const blockingEventTypes: IDropdownOption[] = ["user.pre_create"].map(
  (type): IDropdownOption => ({ key: type, text: type })
);

interface BlockingHandlerItemEditProps {
  index: number;
  value: BlockingEventHandler;
  onChange: (newValue: BlockingEventHandler) => void;
}
const BlockingHandlerItemEdit: React.FC<BlockingHandlerItemEditProps> =
  function BlockingHandlerItemEdit(props) {
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
          placeholder="https://example.com/callback"
        />
      </div>
    );
  };

interface NonBlockingHandlerItemEditProps {
  index: number;
  value: NonBlockingEventHandler;
  onChange: (newValue: NonBlockingEventHandler) => void;
}
const NonBlockingHandlerItemEdit: React.FC<NonBlockingHandlerItemEditProps> =
  function NonBlockingHandlerItemEdit(props) {
    const { index, value, onChange } = props;
    const { renderToString } = useContext(Context);

    const urlField = useMemo(
      () => ({
        parentJSONPointer: `/hook/non_blocking_handlers/${index}`,
        fieldName: "url",
      }),
      [index]
    );
    const { errors: urlErrors } = useFormField(urlField);
    const urlErrorMessage = useMemo(
      () => renderErrors(urlField, urlErrors, renderToString),
      [urlField, urlErrors, renderToString]
    );

    const onURLChange = useCallback(
      (_, url?: string) => {
        onChange({ ...value, url: url ?? "" });
      },
      [onChange, value]
    );

    return (
      <div className={styles.handlerEdit}>
        <TextField
          className={styles.handlerURLField}
          value={value.url}
          onChange={onURLChange}
          errorMessage={urlErrorMessage}
          placeholder="https://example.com/callback"
        />
      </div>
    );
  };
interface WebhookConfigurationScreenContentProps {
  form: AppConfigFormModel<FormState>;
  hookFeatureConfig?: HookFeatureConfig;
}

const WebhookConfigurationScreenContent: React.FC<WebhookConfigurationScreenContentProps> =
  function WebhookConfigurationScreenContent(props) {
    const { hookFeatureConfig } = props;
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

    // non-blocking handlers
    const makeDefaultNonBlockingHandler = useCallback(
      (): NonBlockingEventHandler => ({
        events: ["*"],
        url: "",
      }),
      []
    );

    const renderNonBlockingHandlerItem = useCallback(
      (
        index: number,
        value: NonBlockingEventHandler,
        onChange: (newValue: NonBlockingEventHandler) => void
      ) => (
        <NonBlockingHandlerItemEdit
          index={index}
          value={value}
          onChange={onChange}
        />
      ),
      []
    );

    const onNonBlockingHandlersChange = useCallback(
      (value: NonBlockingEventHandler[]) => {
        setState((state) => ({ ...state, non_blocking_handlers: value }));
      },
      [setState]
    );

    const blockingHandlerMax = useMemo(() => {
      return hookFeatureConfig?.blocking_handler?.maximum ?? 99;
    }, [hookFeatureConfig?.blocking_handler?.maximum]);

    const nonBlockingHandlerMax = useMemo(() => {
      return hookFeatureConfig?.non_blocking_handler?.maximum ?? 99;
    }, [hookFeatureConfig?.non_blocking_handler?.maximum]);

    const blockingHandlerLimitReached = useMemo(() => {
      return state.blocking_handlers.length >= blockingHandlerMax;
    }, [state.blocking_handlers, blockingHandlerMax]);

    const nonBlockingHandlerLimitReached = useMemo(() => {
      return state.non_blocking_handlers.length >= nonBlockingHandlerMax;
    }, [state.non_blocking_handlers, nonBlockingHandlerMax]);

    const blockingHandlerDisabled = useMemo(() => {
      return blockingHandlerMax < 1;
    }, [blockingHandlerMax]);

    const nonBlockingHandlerDisabled = useMemo(() => {
      return nonBlockingHandlerMax < 1;
    }, [nonBlockingHandlerMax]);

    const hideBlockingHandlerList = useMemo(() => {
      return blockingHandlerDisabled && state.blocking_handlers.length === 0;
    }, [state.blocking_handlers.length, blockingHandlerDisabled]);

    const hideNonBlockingHandlerList = useMemo(() => {
      return (
        nonBlockingHandlerDisabled && state.non_blocking_handlers.length === 0
      );
    }, [state.non_blocking_handlers.length, nonBlockingHandlerDisabled]);

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
            <FormattedMessage id="WebhookConfigurationScreen.blocking-events" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="WebhookConfigurationScreen.blocking-events.description" />
          </WidgetDescription>
          {blockingHandlerMax < 99 &&
            (blockingHandlerDisabled ? (
              <MessageBar>
                <FormattedMessage
                  id="FeatureConfig.webhook.blocking-events.disabled"
                  values={{
                    HREF: "./settings/subscription",
                  }}
                  components={{
                    Link,
                  }}
                />
              </MessageBar>
            ) : (
              <MessageBar>
                <FormattedMessage
                  id="FeatureConfig.webhook.blocking-events.maximum"
                  values={{
                    HREF: "./settings/subscription",
                    maximum: blockingHandlerMax,
                  }}
                  components={{
                    Link,
                  }}
                />
              </MessageBar>
            ))}
          {!hideBlockingHandlerList && (
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
              addDisabled={blockingHandlerLimitReached}
            />
          )}
        </Widget>

        <Widget className={cn(styles.widget, styles.controlGroup)}>
          <WidgetTitle>
            <FormattedMessage id="WebhookConfigurationScreen.non-blocking-events" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="WebhookConfigurationScreen.non-blocking-events.description" />
          </WidgetDescription>
          {nonBlockingHandlerMax < 99 &&
            (nonBlockingHandlerDisabled ? (
              <MessageBar>
                <FormattedMessage
                  id="FeatureConfig.webhook.non-blocking-events.disabled"
                  values={{
                    HREF: "./settings/subscription",
                  }}
                  components={{
                    Link,
                  }}
                />
              </MessageBar>
            ) : (
              <MessageBar>
                <FormattedMessage
                  id="FeatureConfig.webhook.non-blocking-events.maximum"
                  values={{
                    HREF: "./settings/subscription",
                    maximum: nonBlockingHandlerMax,
                  }}
                  components={{
                    Link,
                  }}
                />
              </MessageBar>
            ))}
          {!hideNonBlockingHandlerList && (
            <FieldList
              className={styles.control}
              label={
                <Label>
                  <FormattedMessage id="WebhookConfigurationScreen.non-blocking-events-endpoints.label" />
                </Label>
              }
              parentJSONPointer="/hook"
              fieldName="non_blocking_handlers"
              list={state.non_blocking_handlers}
              onListChange={onNonBlockingHandlersChange}
              makeDefaultItem={makeDefaultNonBlockingHandler}
              renderListItem={renderNonBlockingHandlerItem}
              addButtonLabelMessageID="add"
              addDisabled={nonBlockingHandlerLimitReached}
            />
          )}
        </Widget>

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
      </ScreenContent>
    );
  };

const WebhookConfigurationScreen: React.FC =
  function WebhookConfigurationScreen() {
    const { appID } = useParams();
    const form = useAppConfigForm(appID, constructFormState, constructConfig);
    const featureConfig = useAppFeatureConfigQuery(appID);

    if (form.isLoading || featureConfig.loading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    if (featureConfig.error) {
      return (
        <ShowError
          error={featureConfig.error}
          onRetry={() => {
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }

    return (
      <FormContainer form={form}>
        <WebhookConfigurationScreenContent
          form={form}
          hookFeatureConfig={featureConfig.effectiveFeatureConfig?.hook}
        />
      </FormContainer>
    );
  };

export default WebhookConfigurationScreen;
