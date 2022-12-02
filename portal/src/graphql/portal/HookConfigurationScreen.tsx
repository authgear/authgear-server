import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useParams } from "react-router-dom";
import { Dropdown, IDropdownOption, Label, Modal } from "@fluentui/react";
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
import { useResourceForm } from "../../hook/useResourceForm";
import { resourcePath, ResourceSpecifier, Resource } from "../../util/resource";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import FieldList, { ListItemProps } from "../../FieldList";
import FormContainer from "../../FormContainer";
import FormTextField from "../../FormTextField";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { genRandomHexadecimalString } from "../../util/random";
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
import CodeEditor from "../../CodeEditor";
import DefaultButton from "../../DefaultButton";
import HorizontalDivider from "../../HorizontalDivider";

const DENOHOOK_BLOCKING_DEFAULT = `import { HookEvent, HookResponse } from "https://deno.land/x/authgear_deno_hook@v0.2.0/mod.ts";

export default async function(e: HookEvent): Promise<HookResponse> {
  // Write your hook with the help of the type definition.
  return { is_allowed: true };
}
`;

const DENOHOOK_NONBLOCKING_DEFAULT = `import { HookEvent } from "https://deno.land/x/authgear_deno_hook@v0.2.0/mod.ts";

export default async function(e: HookEvent): Promise<void> {
  // Write your hook with the help of the type definition.
}
`;

type HookKind = "webhook" | "denohook";

type EventKind = "blocking" | "nonblocking";

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

interface ConfigFormState {
  timeout: number | undefined;
  totalTimeout: number | undefined;
  blocking_handlers: BlockingEventHandler[];
  non_blocking_handlers: NonBlockingEventHandler[];
  secret: string | null;
}

interface FormState extends ConfigFormState {
  resources: Resource[];
}

interface FormModel {
  isLoading: boolean;
  isUpdating: boolean;
  isDirty: boolean;
  loadError: unknown;
  updateError: unknown;
  state: FormState;
  setState: (fn: (state: FormState) => FormState) => void;
  reload: () => void;
  reset: () => void;
  save: () => Promise<void>;
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

function constructConfigFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): ConfigFormState {
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
  initialState: ConfigFormState,
  currentState: ConfigFormState,
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

function getPathFromURL(url: string): string {
  const path = url.slice("authgeardeno:///".length);
  return path;
}

function makeNewURL(eventKind: EventKind): string {
  const rand = genRandomHexadecimalString();
  return `authgeardeno:///deno/${eventKind}.${rand}.ts`;
}

function makeSpecifier(url: string): ResourceSpecifier {
  const path = getPathFromURL(url);
  return {
    def: {
      resourcePath: resourcePath([path]),
      type: "text" as const,
      extensions: [],
    },
    locale: null,
    extension: null,
  };
}

function makeSpecifiersFromState(state: ConfigFormState): ResourceSpecifier[] {
  const specifiers = [];
  for (const h of state.blocking_handlers) {
    if (h.kind === "denohook") {
      specifiers.push(makeSpecifier(h.url));
    }
  }
  for (const h of state.non_blocking_handlers) {
    if (h.kind === "denohook") {
      specifiers.push(makeSpecifier(h.url));
    }
  }
  return specifiers;
}

function addMissingResources(state: FormState) {
  for (let i = 0; i < state.blocking_handlers.length; ++i) {
    const h = state.blocking_handlers[i];
    if (h.kind === "denohook") {
      const path = getPathFromURL(h.url);
      const specifier = makeSpecifier(h.url);
      const r = state.resources.find((r) => r.path === path);
      if (r == null) {
        state.resources.push({
          path,
          specifier,
          nullableValue: DENOHOOK_BLOCKING_DEFAULT,
        });
      }
    }
  }
  for (let i = 0; i < state.non_blocking_handlers.length; ++i) {
    const h = state.non_blocking_handlers[i];
    if (h.kind === "denohook") {
      const path = getPathFromURL(h.url);
      const specifier = makeSpecifier(h.url);
      const r = state.resources.find((r) => r.path === path);
      if (r == null) {
        state.resources.push({
          path,
          specifier,
          nullableValue: DENOHOOK_NONBLOCKING_DEFAULT,
        });
      }
    }
  }
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
  onEdit: (index: number, value: BlockingEventHandler) => void;
}
const BlockingHandlerItemEdit: React.VFC<BlockingHandlerItemEditProps> =
  function BlockingHandlerItemEdit(props) {
    const { index, value, onChange, onEdit } = props;

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
        const key = event?.key;
        if (key != null) {
          switch (key) {
            case "webhook":
              onChange({ ...value, kind: "webhook", url: "" });
              break;
            case "denohook":
              onChange({
                ...value,
                kind: "denohook",
                url: makeNewURL("blocking"),
              });
              break;
            default:
              break;
          }
        }
      },
      [onChange, value]
    );
    const onClickEdit = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        onEdit(index, value);
      },
      [onEdit, index, value]
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
            onClick={onClickEdit}
          />
        ) : null}
      </div>
    );
  };

interface NonBlockingHandlerItemEditProps {
  index: number;
  value: NonBlockingEventHandler;
  onChange: (newValue: NonBlockingEventHandler) => void;
  onEdit: (index: number, value: NonBlockingEventHandler) => void;
}
const NonBlockingHandlerItemEdit: React.VFC<NonBlockingHandlerItemEditProps> =
  function NonBlockingHandlerItemEdit(props) {
    const { index, value, onChange, onEdit } = props;

    const { renderToString } = useContext(Context);

    const onURLChange = useCallback(
      (_, url?: string) => {
        onChange({ ...value, url: url ?? "" });
      },
      [onChange, value]
    );
    const onChangeHookKind = useCallback(
      (_, event?: IDropdownOption) => {
        const key = event?.key;
        if (key != null) {
          switch (key) {
            case "webhook":
              onChange({ ...value, kind: "webhook", url: "" });
              break;
            case "denohook":
              onChange({
                ...value,
                kind: "denohook",
                url: makeNewURL("nonblocking"),
              });
              break;
            default:
              break;
          }
        }
      },
      [onChange, value]
    );
    const onClickEdit = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        onEdit(index, value);
      },
      [onEdit, index, value]
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
            onClick={onClickEdit}
          />
        ) : null}
      </div>
    );
  };

interface HookConfigurationScreenContentProps {
  form: AppSecretConfigFormModel<ConfigFormState>;
  hookFeatureConfig?: HookFeatureConfig;
}

interface LocationState {
  isOAuthRedirect: boolean;
}

interface CodeEditorState {
  eventKind: EventKind;
  index: number;
  value: string | null;
}

const HookConfigurationScreenContent: React.VFC<HookConfigurationScreenContentProps> =
  // eslint-disable-next-line complexity
  function HookConfigurationScreenContent(props) {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(Context);
    const { hookFeatureConfig, form: config } = props;

    const [codeEditorState, setCodeEditorState] =
      useState<CodeEditorState | null>(null);

    const specifiers = useMemo(() => {
      return makeSpecifiersFromState(config.state);
    }, [config.state]);

    const resources = useResourceForm(
      appID,
      specifiers,
      (resources) => resources,
      (resources) => resources
    );

    const state = useMemo<FormState>(() => {
      return {
        ...config.state,
        resources: resources.state,
      };
    }, [config.state, resources.state]);

    const form: FormModel = {
      isLoading: config.isLoading || resources.isLoading,
      isUpdating: config.isUpdating || resources.isUpdating,
      isDirty: config.isDirty || resources.isDirty,
      loadError: config.loadError ?? resources.loadError,
      updateError: config.updateError ?? resources.updateError,
      state,
      setState: (fn) => {
        const newState = fn(state);
        const { resources: newResources, ...configState } = newState;
        config.setState(() => ({
          ...configState,
        }));
        resources.setState(() => newResources);
      },
      reload: () => {
        resources.reload();
        config.reload();
      },
      reset: () => {
        resources.reset();
        config.reset();
      },
      save: async () => {
        await resources.save();
        await config.save();
      },
    };

    const { setState } = form;

    const onClickCancelEditing = useCallback((e) => {
      if (e.nativeEvent instanceof KeyboardEvent && e.key === "Escape") {
        return;
      }
      setCodeEditorState(null);
    }, []);

    const onClickFinishEditing = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        if (codeEditorState != null) {
          const { eventKind, index, value } = codeEditorState;
          setState((prev) =>
            produce(prev, (prev) => {
              switch (eventKind) {
                case "blocking": {
                  const h = state.blocking_handlers[index];
                  const path = getPathFromURL(h.url);
                  for (const r of prev.resources) {
                    if (r.path === path) {
                      r.nullableValue = value ?? DENOHOOK_BLOCKING_DEFAULT;
                    }
                  }
                  break;
                }
                case "nonblocking": {
                  const h = state.non_blocking_handlers[index];
                  const path = getPathFromURL(h.url);
                  for (const r of prev.resources) {
                    if (r.path === path) {
                      r.nullableValue = value ?? DENOHOOK_NONBLOCKING_DEFAULT;
                    }
                  }
                  break;
                }
              }
            })
          );
        }
        setCodeEditorState(null);
      },
      [
        codeEditorState,
        setState,
        state.blocking_handlers,
        state.non_blocking_handlers,
      ]
    );

    const onChangeCode = useCallback((value) => {
      if (value != null) {
        setCodeEditorState((prev) => {
          if (prev == null) {
            return prev;
          }

          return {
            ...prev,
            value,
          };
        });
      }
    }, []);

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

    // eslint-disable-next-line complexity
    const code = useMemo(() => {
      if (codeEditorState == null) {
        return "";
      }

      if (codeEditorState.value != null) {
        return codeEditorState.value;
      }

      const { eventKind, index } = codeEditorState;
      switch (eventKind) {
        case "blocking": {
          const h = state.blocking_handlers[index];
          const path = getPathFromURL(h.url);
          for (const r of state.resources) {
            if (r.path === path && r.nullableValue != null) {
              return r.nullableValue;
            }
          }
          break;
        }
        case "nonblocking": {
          const h = state.non_blocking_handlers[index];
          const path = getPathFromURL(h.url);
          for (const r of state.resources) {
            if (r.path === path && r.nullableValue != null) {
              return r.nullableValue;
            }
          }
          break;
        }
      }

      if (eventKind === "nonblocking") {
        return DENOHOOK_NONBLOCKING_DEFAULT;
      }
      return DENOHOOK_BLOCKING_DEFAULT;
    }, [
      codeEditorState,
      state.blocking_handlers,
      state.non_blocking_handlers,
      state.resources,
    ]);

    const onEditBlocking = useCallback(
      (index: number, _value: BlockingEventHandler) => {
        setCodeEditorState({
          eventKind: "blocking",
          index,
          value: null,
        });
      },
      []
    );

    const onEditNonBlocking = useCallback(
      (index: number, _value: NonBlockingEventHandler) => {
        setCodeEditorState({
          eventKind: "nonblocking",
          index,
          value: null,
        });
      },
      []
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
            onEdit={onEditBlocking}
          />
        );
      },
      [onEditBlocking]
    );
    const onBlockingHandlersChange = useCallback(
      (value: BlockingEventHandler[]) => {
        setState((state) =>
          produce(state, (state) => {
            state.blocking_handlers = value;
            addMissingResources(state);
          })
        );
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
            onEdit={onEditNonBlocking}
          />
        );
      },
      [onEditNonBlocking]
    );

    const onNonBlockingHandlersChange = useCallback(
      (value: NonBlockingEventHandler[]) => {
        setState((state) =>
          produce(state, (state) => {
            state.non_blocking_handlers = value;
            addMissingResources(state);
          })
        );
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
      <>
        <Modal
          isOpen={codeEditorState != null}
          onDismiss={onClickCancelEditing}
          isBlocking={true}
        >
          <div className={styles.codeEditorContainer}>
            <CodeEditor
              className={styles.codeEditor}
              language="typescript"
              value={code}
              onChange={onChangeCode}
            />
            <HorizontalDivider />
            <div className={styles.codeEditorFooter}>
              <PrimaryButton
                text="Finish Editing"
                onClick={onClickFinishEditing}
              />
              <DefaultButton text="Cancel" onClick={onClickCancelEditing} />
            </div>
          </div>
        </Modal>
        <FormContainer form={form}>
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
              <WidgetTitle
                className={styles.columnFull}
                id={WEBHOOK_SIGNATURE_ID}
              >
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
                  revealed && state.secret != null
                    ? state.secret
                    : MASKED_SECRET
                }
                readOnly={true}
              />
              <PrimaryButton
                className={styles.secretButton}
                id={copyButtonProps.id}
                onClick={revealed ? copyButtonProps.onClick : onClickReveal}
                onMouseLeave={
                  revealed ? copyButtonProps.onMouseLeave : undefined
                }
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
        </FormContainer>
      </>
    );
  };

const HookConfigurationScreen: React.VFC = function HookConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const form = useAppSecretConfigForm({
    appID,
    constructFormState: constructConfigFormState,
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
    <HookConfigurationScreenContent
      form={form}
      hookFeatureConfig={featureConfig.effectiveFeatureConfig?.hook}
    />
  );
};

export default HookConfigurationScreen;
