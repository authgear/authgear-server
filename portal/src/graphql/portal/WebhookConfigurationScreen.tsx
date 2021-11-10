import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";
import { useParams } from "react-router-dom";
import {
  Dropdown,
  IDropdownOption,
  ISelectableOption,
  Label,
  TextField,
  MessageBar,
  PrimaryButton,
} from "@fluentui/react";
import produce from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import {
  HookFeatureConfig,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../types";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import { useFormField } from "../../form";
import FieldList from "../../FieldList";
import FormContainer from "../../FormContainer";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import styles from "./WebhookConfigurationScreen.module.scss";
import { renderErrors } from "../../error/parse";
import WidgetDescription from "../../WidgetDescription";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { startReauthentication } from "./Authenticated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useErrorMessage } from "../../formbinding";

interface BlockingEventHandler {
  event: string;
  url: string;
}

interface NonBlockingEventHandler {
  events: string[];
  url: string;
}
interface FormState {
  timeout: number | undefined;
  totalTimeout: number | undefined;
  blocking_handlers: BlockingEventHandler[];
  non_blocking_handlers: NonBlockingEventHandler[];
  secret: string | null;
}

const MASKED_SECRET = "***************";

const WEBHOOK_SIGNATURE_ID = "webhook-signature";

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  return {
    timeout: config.hook?.sync_hook_timeout_seconds,
    totalTimeout: config.hook?.sync_hook_total_timeout_seconds,
    blocking_handlers: config.hook?.blocking_handlers ?? [],
    non_blocking_handlers: config.hook?.non_blocking_handlers ?? [],
    secret: secrets.webhookSecret?.secret ?? null,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const newConfig = produce(config, (config) => {
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
  return [newConfig, secrets];
}

const blockingEventTypes: IDropdownOption[] = [
  "user.pre_create",
  "user.profile.pre_update",
].map((type): IDropdownOption => ({ key: type, text: type }));

interface BlockingHandlerItemEditProps {
  index: number;
  value: BlockingEventHandler;
  onChange: (newValue: BlockingEventHandler) => void;
}
const BlockingHandlerItemEdit: React.FC<BlockingHandlerItemEditProps> =
  function BlockingHandlerItemEdit(props) {
    const { index, value, onChange } = props;

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
    const eventFieldProps = useErrorMessage(eventField);
    const urlFieldProps = useErrorMessage(urlField);

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
          {...eventFieldProps}
        />
        <TextField
          className={styles.handlerURLField}
          value={value.url}
          onChange={onURLChange}
          placeholder="https://example.com/callback"
          {...urlFieldProps}
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
    // FIXME: use formbinding
    const urlErrorMessage = useMemo(
      () => renderErrors(urlErrors, renderToString),
      [urlErrors, renderToString]
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
  form: AppSecretConfigFormModel<FormState>;
  hookFeatureConfig?: HookFeatureConfig;
}

interface LocationState {
  isOAuthRedirect: boolean;
}

const WebhookConfigurationScreenContent: React.FC<WebhookConfigurationScreenContentProps> =
  // eslint-disable-next-line complexity
  function WebhookConfigurationScreenContent(props) {
    const { renderToString } = useContext(Context);
    const { hookFeatureConfig } = props;
    const { state, setState } = props.form;

    const locationState = useLocationEffect((state: LocationState) => {
      if (state.isOAuthRedirect) {
        window.location.hash = "";
        window.location.hash = "#" + WEBHOOK_SIGNATURE_ID;
      }
    });

    const [revealed, setRevealed] = useState(
      locationState?.isOAuthRedirect ?? false
    );

    const onTimeoutChange = useCallback(
      (_, value?: string) => {
        setState((state) => ({
          ...state,
          timeout: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const onTotalTimeoutChange = useCallback(
      (_, value?: string) => {
        setState((state) => ({
          ...state,
          totalTimeout: parseIntegerAllowLeadingZeros(value),
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

    const onClickReveal = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        if (state.secret != null) {
          setRevealed(true);
          return;
        }

        const locationState: LocationState = {
          isOAuthRedirect: true,
        };

        startReauthentication(locationState).catch((e) => {
          // Normally there should not be any error.
          console.error(e);
        });
      },
      [state.secret]
    );

    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: state.secret ?? "",
    });

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
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="WebhookConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="WebhookConfigurationScreen.description" />
        </ScreenDescription>

        <Widget className={styles.widget}>
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
                    planPagePath: "../../billing",
                  }}
                />
              </MessageBar>
            ) : (
              <MessageBar>
                <FormattedMessage
                  id="FeatureConfig.webhook.blocking-events.maximum"
                  values={{
                    planPagePath: "../../billing",
                    maximum: blockingHandlerMax,
                  }}
                />
              </MessageBar>
            ))}
          {!hideBlockingHandlerList && (
            <FieldList
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

        <Widget className={styles.widget}>
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
                    planPagePath: "../../billing",
                  }}
                />
              </MessageBar>
            ) : (
              <MessageBar>
                <FormattedMessage
                  id="FeatureConfig.webhook.non-blocking-events.maximum"
                  values={{
                    planPagePath: "../../billing",
                    maximum: nonBlockingHandlerMax,
                  }}
                />
              </MessageBar>
            ))}
          {!hideNonBlockingHandlerList && (
            <FieldList
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

        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="WebhookConfigurationScreen.webhook-settings" />
          </WidgetTitle>
          <TextField
            type="text"
            label={renderToString(
              "WebhookConfigurationScreen.total-timeout.label"
            )}
            value={state.totalTimeout?.toFixed(0) ?? ""}
            onChange={onTotalTimeoutChange}
          />
          <TextField
            type="text"
            label={renderToString("WebhookConfigurationScreen.timeout.label")}
            value={state.timeout?.toFixed(0) ?? ""}
            onChange={onTimeoutChange}
          />
        </Widget>

        <Widget className={cn(styles.widget, styles.grid)}>
          <WidgetTitle className={styles.columnFull} id={WEBHOOK_SIGNATURE_ID}>
            <FormattedMessage id="WebhookConfigurationScreen.signature.title" />
          </WidgetTitle>
          <WidgetDescription className={styles.columnFull}>
            <FormattedMessage id="WebhookConfigurationScreen.signature.description" />
          </WidgetDescription>
          <TextField
            className={styles.secretInput}
            type="text"
            label={renderToString(
              "WebhookConfigurationScreen.signature.secret-key"
            )}
            value={
              revealed && state.secret != null ? state.secret : MASKED_SECRET
            }
            readOnly={true}
          />
          <PrimaryButton
            className={styles.secretButton}
            id={copyButtonProps.id}
            onClick={revealed ? copyButtonProps.onClick : onClickReveal}
            onMouseLeave={revealed ? copyButtonProps.onMouseLeave : undefined}
          >
            {revealed ? (
              <FormattedMessage id="copy" />
            ) : (
              <FormattedMessage id="reveal" />
            )}
          </PrimaryButton>
          <Feedback />
        </Widget>
      </ScreenContent>
    );
  };

const WebhookConfigurationScreen: React.FC =
  function WebhookConfigurationScreen() {
    const { appID } = useParams();
    const form = useAppSecretConfigForm(
      appID,
      constructFormState,
      constructConfig
    );
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
