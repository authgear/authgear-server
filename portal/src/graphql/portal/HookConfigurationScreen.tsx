import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";
import { Dropdown, IDropdownOption, Label } from "@fluentui/react";
import produce from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import {
  BlockingHookHandlerConfig,
  HookFeatureConfig,
  NonBlockingHookHandlerConfig,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../types";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import FieldList, { ListItemProps } from "../../FieldList";
import FormContainer from "../../FormContainer";
import FormTextField from "../../FormTextField";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import styles from "./HookConfigurationScreen.module.css";
import WidgetDescription from "../../WidgetDescription";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { startReauthentication } from "./Authenticated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useErrorMessage, useErrorMessageString } from "../../formbinding";
import TextField from "../../TextField";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import PrimaryButton from "../../PrimaryButton";
import ActionButton from "../../ActionButton";

type HookKind = "webhook" | "denohook";

interface BlockingEventHandler {
  event: string;
  kind: HookKind;
  url: string;
}

interface NonBlockingEventHandler {
  events: string[];
  kind: HookKind;
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

const EDIT_BUTTON_ICON_PROPS = {
  iconName: "Edit",
};

const EDIT_BUTTON_STYLES = {
  root: {
    // The native height is 40px.
    // But we want to make sure everything in the same row has the same height,
    // So we force it to 32px.
    height: "32px",
  },
};

function blockingConfigToState(
  c: BlockingHookHandlerConfig
): BlockingEventHandler {
  const kind = c.url.startsWith("authgeardeno:") ? "denohook" : "webhook";
  return {
    kind,
    ...c,
  };
}

function blockingStateToConfig(
  s: BlockingEventHandler
): BlockingHookHandlerConfig {
  return {
    event: s.event,
    url: s.url,
  };
}

function nonBlockingConfigToState(
  c: NonBlockingHookHandlerConfig
): NonBlockingEventHandler {
  const kind = c.url.startsWith("authgeardeno:") ? "denohook" : "webhook";
  return {
    kind,
    ...c,
  };
}

function nonBlockingStateToConfig(
  s: NonBlockingEventHandler
): NonBlockingHookHandlerConfig {
  return {
    events: s.events,
    url: s.url,
  };
}

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  return {
    timeout: config.hook?.sync_hook_timeout_seconds,
    totalTimeout: config.hook?.sync_hook_total_timeout_seconds,
    blocking_handlers: (config.hook?.blocking_handlers ?? []).map(
      blockingConfigToState
    ),
    non_blocking_handlers: (config.hook?.non_blocking_handlers ?? []).map(
      nonBlockingConfigToState
    ),
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
    config.hook.blocking_handlers = currentState.blocking_handlers.map(
      blockingStateToConfig
    );
    config.hook.non_blocking_handlers = currentState.non_blocking_handlers.map(
      nonBlockingStateToConfig
    );
    clearEmptyObject(config);
  });
  return [newConfig, secrets];
}

const BLOCK_EVENT_TYPES: string[] = [
  "user.pre_create",
  "user.profile.pre_update",
  "user.pre_schedule_deletion",
];

interface BlockingHandlerItemEditProps {
  index: number;
  value: BlockingEventHandler;
  onChange: (newValue: BlockingEventHandler) => void;
}
const BlockingHandlerItemEdit: React.VFC<BlockingHandlerItemEditProps> =
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
    const eventFieldProps = useErrorMessageString(eventField);
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
    const onChangeHookKind = useCallback(
      (_, event?: IDropdownOption) => {
        if (event?.key != null) {
          onChange({ ...value, kind: event.key as HookKind, url: "" });
        }
      },
      [onChange, value]
    );

    const eventOptions = useMemo(() => {
      return BLOCK_EVENT_TYPES.map((t) => ({
        key: t,
        text: renderToString(
          "HookConfigurationScreen.blocking-event-type." + t
        ),
      }));
    }, [renderToString]);

    const kindOptions = useMemo(() => {
      return [
        {
          key: "webhook",
          text: renderToString("HookConfigurationScreen.hook-kind.webhook"),
        },
        {
          key: "denohook",
          text: renderToString("HookConfigurationScreen.hook-kind.denohook"),
        },
      ];
    }, [renderToString]);

    return (
      <div className={styles.handlerEdit}>
        <Dropdown
          className={styles.handlerEventField}
          options={eventOptions}
          selectedKey={value.event}
          onChange={onBlockingEventChange}
          ariaLabel={"HookConfigurationScreen.blocking-events.label"}
          {...eventFieldProps}
        />
        <Dropdown
          className={styles.handlerKindField}
          options={kindOptions}
          selectedKey={value.kind}
          onChange={onChangeHookKind}
          ariaLabel={"HookConfigurationScreen.hook-kind.label"}
        />
        {value.kind === "webhook" ? (
          <TextField
            className={styles.handlerURLField}
            value={value.url}
            onChange={onURLChange}
            placeholder="https://example.com/callback"
            {...urlFieldProps}
          />
        ) : null}
        {value.kind === "denohook" ? (
          <ActionButton
            iconProps={EDIT_BUTTON_ICON_PROPS}
            styles={EDIT_BUTTON_STYLES}
            text={
              <FormattedMessage id="HookConfigurationScreen.edit-hook.label" />
            }
          />
        ) : null}
      </div>
    );
  };

interface NonBlockingHandlerItemEditProps {
  index: number;
  value: NonBlockingEventHandler;
  onChange: (newValue: NonBlockingEventHandler) => void;
}
const NonBlockingHandlerItemEdit: React.VFC<NonBlockingHandlerItemEditProps> =
  function NonBlockingHandlerItemEdit(props) {
    const { index, value, onChange } = props;

    const { renderToString } = useContext(Context);

    const onURLChange = useCallback(
      (_, url?: string) => {
        onChange({ ...value, url: url ?? "" });
      },
      [onChange, value]
    );
    const onChangeHookKind = useCallback(
      (_, event?: IDropdownOption) => {
        if (event?.key != null) {
          onChange({ ...value, kind: event.key as HookKind, url: "" });
        }
      },
      [onChange, value]
    );

    const kindOptions = useMemo(() => {
      return [
        {
          key: "webhook",
          text: renderToString("HookConfigurationScreen.hook-kind.webhook"),
        },
        {
          key: "denohook",
          text: renderToString("HookConfigurationScreen.hook-kind.denohook"),
        },
      ];
    }, [renderToString]);

    return (
      <div className={styles.handlerEdit}>
        <Dropdown
          className={styles.handlerKindField}
          options={kindOptions}
          selectedKey={value.kind}
          onChange={onChangeHookKind}
          ariaLabel={"HookConfigurationScreen.hook-kind.label"}
        />
        {value.kind === "webhook" ? (
          <FormTextField
            parentJSONPointer={`/hook/non_blocking_handlers/${index}`}
            fieldName="url"
            className={styles.handlerURLField}
            value={value.url}
            onChange={onURLChange}
            placeholder="https://example.com/callback"
          />
        ) : null}
        {value.kind === "denohook" ? (
          <ActionButton
            iconProps={EDIT_BUTTON_ICON_PROPS}
            styles={EDIT_BUTTON_STYLES}
            text={
              <FormattedMessage id="HookConfigurationScreen.edit-hook.label" />
            }
          />
        ) : null}
      </div>
    );
  };

interface HookConfigurationScreenContentProps {
  form: AppSecretConfigFormModel<FormState>;
  hookFeatureConfig?: HookFeatureConfig;
}

interface LocationState {
  isOAuthRedirect: boolean;
}

const HookConfigurationScreenContent: React.VFC<HookConfigurationScreenContentProps> =
  // eslint-disable-next-line complexity
  function HookConfigurationScreenContent(props) {
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
        event: BLOCK_EVENT_TYPES[0],
        kind: "webhook",
        url: "",
      }),
      []
    );
    const BlockingHandlerListItem = useCallback(
      (props: ListItemProps<BlockingEventHandler>) => {
        const { index, value, onChange } = props;
        return (
          <BlockingHandlerItemEdit
            index={index}
            value={value}
            onChange={onChange}
          />
        );
      },
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
        kind: "webhook",
        url: "",
      }),
      []
    );

    const NonBlockingHandlerListItem = useCallback(
      (props: ListItemProps<NonBlockingEventHandler>) => {
        const { index, value, onChange } = props;
        return (
          <NonBlockingHandlerItemEdit
            index={index}
            value={value}
            onChange={onChange}
          />
        );
      },
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
          <FormattedMessage id="HookConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="HookConfigurationScreen.description" />
        </ScreenDescription>

        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="HookConfigurationScreen.blocking-events" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="HookConfigurationScreen.blocking-events.description" />
          </WidgetDescription>
          {blockingHandlerMax < 99 ? (
            blockingHandlerDisabled ? (
              <FeatureDisabledMessageBar messageID="FeatureConfig.webhook.blocking-events.disabled" />
            ) : (
              <FeatureDisabledMessageBar
                messageID="FeatureConfig.webhook.blocking-events.maximum"
                messageValues={{
                  maximum: blockingHandlerMax,
                }}
              />
            )
          ) : null}
          {!hideBlockingHandlerList ? (
            <FieldList
              label={
                <Label>
                  <FormattedMessage id="HookConfigurationScreen.blocking-handlers.label" />
                </Label>
              }
              parentJSONPointer="/hook"
              fieldName="blocking_handlers"
              list={state.blocking_handlers}
              onListChange={onBlockingHandlersChange}
              makeDefaultItem={makeDefaultHandler}
              ListItemComponent={BlockingHandlerListItem}
              addButtonLabelMessageID="add"
              addDisabled={blockingHandlerLimitReached}
            />
          ) : null}
        </Widget>

        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="HookConfigurationScreen.non-blocking-events" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage id="HookConfigurationScreen.non-blocking-events.description" />
          </WidgetDescription>
          {nonBlockingHandlerMax < 99 ? (
            nonBlockingHandlerDisabled ? (
              <FeatureDisabledMessageBar messageID="FeatureConfig.webhook.non-blocking-events.disabled" />
            ) : (
              <FeatureDisabledMessageBar
                messageID="FeatureConfig.webhook.non-blocking-events.maximum"
                messageValues={{
                  maximum: nonBlockingHandlerMax,
                }}
              />
            )
          ) : null}
          {!hideNonBlockingHandlerList ? (
            <FieldList
              label={
                <Label>
                  <FormattedMessage id="HookConfigurationScreen.non-blocking-events-endpoints.label" />
                </Label>
              }
              parentJSONPointer="/hook"
              fieldName="non_blocking_handlers"
              list={state.non_blocking_handlers}
              onListChange={onNonBlockingHandlersChange}
              makeDefaultItem={makeDefaultNonBlockingHandler}
              ListItemComponent={NonBlockingHandlerListItem}
              addButtonLabelMessageID="add"
              addDisabled={nonBlockingHandlerLimitReached}
            />
          ) : null}
        </Widget>

        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="HookConfigurationScreen.hook-settings" />
          </WidgetTitle>
          <TextField
            type="text"
            label={renderToString(
              "HookConfigurationScreen.total-timeout.label"
            )}
            value={state.totalTimeout?.toFixed(0) ?? ""}
            onChange={onTotalTimeoutChange}
          />
          <TextField
            type="text"
            label={renderToString("HookConfigurationScreen.timeout.label")}
            value={state.timeout?.toFixed(0) ?? ""}
            onChange={onTimeoutChange}
          />
        </Widget>

        <Widget className={styles.widget} contentLayout="grid">
          <WidgetTitle className={styles.columnFull} id={WEBHOOK_SIGNATURE_ID}>
            <FormattedMessage id="HookConfigurationScreen.signature.title" />
          </WidgetTitle>
          <WidgetDescription className={styles.columnFull}>
            <FormattedMessage id="HookConfigurationScreen.signature.description" />
          </WidgetDescription>
          <TextField
            className={styles.secretInput}
            type="text"
            label={renderToString(
              "HookConfigurationScreen.signature.secret-key"
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
            text={
              revealed ? (
                <FormattedMessage id="copy" />
              ) : (
                <FormattedMessage id="reveal" />
              )
            }
          />
          <Feedback />
        </Widget>
      </ScreenContent>
    );
  };

const HookConfigurationScreen: React.VFC = function HookConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const form = useAppSecretConfigForm({
    appID,
    constructFormState,
    constructConfig,
  });
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
      <HookConfigurationScreenContent
        form={form}
        hookFeatureConfig={featureConfig.effectiveFeatureConfig?.hook}
      />
    </FormContainer>
  );
};

export default HookConfigurationScreen;
