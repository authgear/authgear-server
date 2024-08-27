import cn from "classnames";
import React, { useCallback, useContext, useMemo, useState } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useLocation, useParams, useNavigate } from "react-router-dom";
import {
  Dropdown,
  IDropdownOption,
  Label,
  FontIcon,
  Text,
  Dialog,
  useTheme,
  DialogFooter,
} from "@fluentui/react";
import { produce } from "immer";
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
import {
  resourcePath,
  ResourceSpecifier,
  Resource,
  ResourcesDiffResult,
} from "../../util/resource";
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
import { useCheckDenoHookMutation } from "./mutations/checkDenoHook";
import { startReauthentication } from "./Authenticated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useErrorMessage, useErrorMessageString } from "../../formbinding";
import { useLoading, useIsLoading } from "../../hook/loading";
import { useProvideError } from "../../hook/error";
import TextField from "../../TextField";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import PrimaryButton from "../../PrimaryButton";
import ActionButton from "../../ActionButton";
import CodeEditor from "../../CodeEditor";
import DefaultButton from "../../DefaultButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { AppSecretKey } from "./globalTypes.generated";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import HorizontalDivider from "../../HorizontalDivider";

const CODE_EDITOR_OPTIONS = {
  minimap: {
    enabled: false,
  },
};

const BLOCKING_EVENT_NAME_TO_TYPE_NAME: Record<string, string | undefined> = {
  "user.pre_create": "EventUserPreCreate",
  "user.profile.pre_update": "EventUserProfilePreUpdate",
  "user.pre_schedule_deletion": "EventUserPreScheduleDeletion",
  "user.pre_schedule_anonymization": "EventUserPreScheduleAnonymization",
  "oidc.jwt.pre_create": "EventOIDCJWTPreCreate",
};

const DENOHOOK_NONBLOCKING_DEFAULT = `import { HookEvent } from "https://deno.land/x/authgear_deno_hook@v1.3.0/mod.ts";

export default async function(e: HookEvent): Promise<void> {
  // Write your hook with the help of the type definition.
  //
  // Since this hook will receive all events,
  // you usually want to differentiate the exact event type,
  // and handle the events accordingly.
  // This can be done by using a switch statement as shown below.
  switch (e.type) {
  case "user.created":
    // Thanks to TypeScript compiler, e is now of type EventUserCreated.
    break;
  default:
    // Add a default case to catch the rest.
    // You can add more case to match other events.
    break;
  }
}
`;

function makeDefaultDenoHookBlockingScript(event: string): string {
  const typeName = BLOCKING_EVENT_NAME_TO_TYPE_NAME[event] ?? "HookEvent";
  return `import { ${typeName}, HookResponse } from "https://deno.land/x/authgear_deno_hook@v1.3.0/mod.ts";

export default async function(e: ${typeName}): Promise<HookResponse> {
  // Write your hook with the help of the type definition.
  return { is_allowed: true };
}
`;
}

type HookKind = "webhook" | "denohook";

type EventKind = "blocking" | "nonblocking";

interface BlockingEventHandler {
  event: string;
  kind: HookKind;
  url: string;
  isDirty: boolean;
}

interface NonBlockingEventHandler {
  events: string[];
  kind: HookKind;
  url: string;
  isDirty: boolean;
}

interface ConfigFormState {
  timeout: number | undefined;
  totalTimeout: number | undefined;
  blocking_handlers: BlockingHookHandlerConfig[];
  non_blocking_handlers: NonBlockingHookHandlerConfig[];
  secret: string | null;
}

function getKind(url: string): HookKind {
  if (url.startsWith("authgeardeno:")) {
    return "denohook";
  }
  return "webhook";
}

function checkDirty(diff: ResourcesDiffResult | null, url: string): boolean {
  if (diff == null) {
    return false;
  }

  const kind = getKind(url);
  if (kind !== "denohook") {
    return false;
  }

  const path = url.slice("authgeardeno:///".length);
  for (const a of diff.newResources) {
    if (a.path === path) {
      return true;
    }
  }
  for (const a of diff.editedResources) {
    if (a.path === path) {
      return true;
    }
  }

  return false;
}

interface FormState extends ConfigFormState {
  resources: Resource[];
  diff: ResourcesDiffResult | null;
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
  label: {
    display: "flex",
    flexDirection: "row",
    alignItems: "center",
  },
};

function constructConfigFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): ConfigFormState {
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
    config.hook.blocking_handlers = currentState.blocking_handlers;
    config.hook.non_blocking_handlers = currentState.non_blocking_handlers;
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
    if (getKind(h.url) === "denohook") {
      specifiers.push(makeSpecifier(h.url));
    }
  }
  for (const h of state.non_blocking_handlers) {
    if (getKind(h.url) === "denohook") {
      specifiers.push(makeSpecifier(h.url));
    }
  }
  return specifiers;
}

function addMissingResources(state: FormState) {
  for (let i = 0; i < state.blocking_handlers.length; ++i) {
    const h = state.blocking_handlers[i];
    if (getKind(h.url) === "denohook") {
      const path = getPathFromURL(h.url);
      const specifier = makeSpecifier(h.url);
      const r = state.resources.find((r) => r.path === path);
      if (r == null) {
        state.resources.push({
          path,
          specifier,
          nullableValue: makeDefaultDenoHookBlockingScript(h.event),
        });
      }
    }
  }
  for (let i = 0; i < state.non_blocking_handlers.length; ++i) {
    const h = state.non_blocking_handlers[i];
    if (getKind(h.url) === "denohook") {
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
  "user.pre_schedule_anonymization",
  "oidc.jwt.pre_create",
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
    const [newEventName, setNewEventName] = useState<string | null>(null);

    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const theme = useTheme();

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

    const onDismissDialog = useCallback((e) => {
      e.preventDefault();
      e.stopPropagation();
      setNewEventName(null);
    }, []);
    const onConfirmChangeEvent = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        if (newEventName != null) {
          onChange({ ...value, event: newEventName });
          setNewEventName(null);
        }
      },
      [onChange, value, newEventName]
    );
    const onBlockingEventChange = useCallback(
      (_, event?: IDropdownOption) => {
        // Show the dialog to confirm overwriting the script if
        // the kind is denohook.
        if (value.kind === "denohook") {
          const key = event?.key ?? null;
          if (typeof key === "string") {
            setNewEventName(key);
          }
        } else {
          onChange({ ...value, event: String(event?.key ?? "") });
        }
      },
      [value, onChange]
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

    const dialogContentProps = useMemo(() => {
      return {
        title: renderToString("HookConfigurationScreen.change-event.title"),
        subText: renderToString(
          "HookConfigurationScreen.change-event.description"
        ),
      };
    }, [renderToString]);

    return (
      <>
        <Dialog
          hidden={newEventName == null}
          onDismiss={onDismissDialog}
          dialogContentProps={dialogContentProps}
        >
          <DialogFooter>
            <PrimaryButton
              theme={themes.destructive}
              text={
                <FormattedMessage id="HookConfigurationScreen.change-event.label" />
              }
              onClick={onConfirmChangeEvent}
            />
            <DefaultButton
              text={<FormattedMessage id="cancel" />}
              onClick={onDismissDialog}
            />
          </DialogFooter>
        </Dialog>
        <div className={styles.hookContainer}>
          <Dropdown
            className={styles.blockingHookKind}
            options={kindOptions}
            selectedKey={value.kind}
            onChange={onChangeHookKind}
            ariaLabel={"HookConfigurationScreen.hook-kind.label"}
          />
          <Dropdown
            className={styles.blockingHookEvent}
            options={eventOptions}
            selectedKey={value.event}
            onChange={onBlockingEventChange}
            ariaLabel={"HookConfigurationScreen.blocking-events.label"}
            {...eventFieldProps}
          />
          {value.kind === "webhook" ? (
            <div className={cn(styles.blockingHookConfig, styles.hookConfig)}>
              <Label>
                <FormattedMessage id="HookConfigurationScreen.action.endpoint.label" />
              </Label>
              <TextField
                className={styles.hookConfigConfig}
                value={value.url}
                onChange={onURLChange}
                placeholder="https://example.com/callback"
                {...urlFieldProps}
              />
            </div>
          ) : null}
          {value.kind === "denohook" ? (
            <div className={cn(styles.blockingHookConfig, styles.hookConfig)}>
              <Label>
                <FormattedMessage id="HookConfigurationScreen.action.script.label" />
              </Label>
              <ActionButton
                className={styles.hookConfigConfig}
                iconProps={EDIT_BUTTON_ICON_PROPS}
                styles={EDIT_BUTTON_STYLES}
                text={
                  <>
                    <FormattedMessage id="HookConfigurationScreen.edit-hook.label" />
                    {value.isDirty ? (
                      <FontIcon
                        iconName="LocationDot"
                        className={styles.dot}
                        style={{
                          color: theme.palette.themePrimary,
                        }}
                      />
                    ) : null}
                  </>
                }
                onClick={onClickEdit}
              />
            </div>
          ) : null}
        </div>
      </>
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

    const theme = useTheme();

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
      <div className={styles.hookContainer}>
        <Dropdown
          className={styles.nonblockingHookEvent}
          options={kindOptions}
          selectedKey={value.kind}
          onChange={onChangeHookKind}
          ariaLabel={"HookConfigurationScreen.hook-kind.label"}
        />
        {value.kind === "webhook" ? (
          <div className={cn(styles.nonblockingHookConfig, styles.hookConfig)}>
            <Label>
              <FormattedMessage id="HookConfigurationScreen.action.endpoint.label" />
            </Label>
            <FormTextField
              className={styles.hookConfigConfig}
              parentJSONPointer={`/hook/non_blocking_handlers/${index}`}
              fieldName="url"
              value={value.url}
              onChange={onURLChange}
              placeholder="https://example.com/callback"
            />
          </div>
        ) : null}
        {value.kind === "denohook" ? (
          <div className={cn(styles.nonblockingHookConfig, styles.hookConfig)}>
            <Label>
              <FormattedMessage id="HookConfigurationScreen.action.script.label" />
            </Label>
            <ActionButton
              className={styles.hookConfigConfig}
              iconProps={EDIT_BUTTON_ICON_PROPS}
              styles={EDIT_BUTTON_STYLES}
              text={
                <>
                  <FormattedMessage id="HookConfigurationScreen.edit-hook.label" />
                  {value.isDirty ? (
                    <FontIcon
                      iconName="LocationDot"
                      className={styles.dot}
                      style={{
                        color: theme.palette.themePrimary,
                      }}
                    />
                  ) : null}
                </>
              }
              onClick={onClickEdit}
            />
          </div>
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
function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isOAuthRedirect != null
  );
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

    const theme = useTheme();

    const [codeEditorState, setCodeEditorState] =
      useState<CodeEditorState | null>(null);

    const isLoading = useIsLoading();

    const specifiers = useMemo(() => {
      return makeSpecifiersFromState(config.state);
    }, [config.state]);

    const resources = useResourceForm(
      appID,
      specifiers,
      (resources) => resources,
      (resources) => resources
    );

    const {
      checkDenoHook,
      loading: checkDenoHookLoading,
      error: checkDenoHookError,
      reset: checkDenoHookReset,
    } = useCheckDenoHookMutation(appID);
    useLoading(checkDenoHookLoading);
    useProvideError(codeEditorState != null ? checkDenoHookError : null);

    const state = useMemo<FormState>(() => {
      return {
        ...config.state,
        resources: resources.state,
        diff: resources.diff,
      };
    }, [config.state, resources.state, resources.diff]);

    const form: FormModel = {
      isLoading: config.isLoading || resources.isLoading,
      isUpdating: config.isUpdating || resources.isUpdating,
      isDirty:
        config.isDirty || resources.isDirty || codeEditorState?.value != null,
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
      save: async (ignoreConflict: boolean = false) => {
        await resources.save(ignoreConflict);
        await config.save(ignoreConflict);
      },
    };

    const { setState } = form;

    const onClickCancelEditing = useCallback(
      (e) => {
        if (e.nativeEvent instanceof KeyboardEvent && e.key === "Escape") {
          return;
        }
        setCodeEditorState(null);
        checkDenoHookReset();
      },
      [checkDenoHookReset]
    );

    const onClickFinishEditing = useCallback(
      async (e) => {
        e.preventDefault();
        e.stopPropagation();
        if (codeEditorState != null) {
          const { eventKind, index, value } = codeEditorState;

          if (value != null) {
            try {
              await checkDenoHook(value);
            } catch {
              // error is handled in the hook.
              return;
            }
          }

          setState((prev) =>
            // eslint-disable-next-line complexity
            produce(prev, (prev) => {
              switch (eventKind) {
                case "blocking": {
                  const h = state.blocking_handlers[index];
                  const path = getPathFromURL(h.url);
                  for (const r of prev.resources) {
                    if (r.path === path) {
                      // value is nullable because onEditBlocking and onEditNonBlocking cannot have deps.
                      // If they had deps, they would change when deps change, causing the ListItemComponent to change as well.
                      // If ListItemComponent changes on every key stroke, the DOM will unmount, result in losing focus on every key stroke.
                      // We encountered this bug before.
                      r.nullableValue = value ?? r.nullableValue ?? "";
                    }
                  }
                  break;
                }
                case "nonblocking": {
                  const h = state.non_blocking_handlers[index];
                  const path = getPathFromURL(h.url);
                  for (const r of prev.resources) {
                    if (r.path === path) {
                      // value is nullable because onEditBlocking and onEditNonBlocking cannot have deps.
                      // If they had deps, they would change when deps change, causing the ListItemComponent to change as well.
                      // If ListItemComponent changes on every key stroke, the DOM will unmount, result in losing focus on every key stroke.
                      // We encountered this bug before.
                      r.nullableValue = value ?? r.nullableValue ?? "";
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
        checkDenoHook,
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

      return "";
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
        isDirty: false,
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
            const newValue: BlockingHookHandlerConfig[] = value.map((h) => {
              return {
                event: h.event,
                url: h.url,
              };
            });
            state.blocking_handlers = newValue;
            addMissingResources(state);
          })
        );
      },
      [setState]
    );
    const onBlockingHandlersChangeItemChange = useCallback(
      (
        value: BlockingEventHandler[],
        _index: number,
        item: BlockingEventHandler
      ) => {
        setState((state) =>
          produce(state, (state) => {
            const newValue: BlockingHookHandlerConfig[] = value.map((h) => {
              return {
                event: h.event,
                url: h.url,
              };
            });
            state.blocking_handlers = newValue;
            addMissingResources(state);
            for (const r of state.resources) {
              if (r.path === getPathFromURL(item.url)) {
                r.nullableValue = makeDefaultDenoHookBlockingScript(item.event);
              }
            }
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
        isDirty: false,
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
            const newValue = value.map((h) => {
              return {
                events: h.events,
                url: h.url,
              };
            });
            state.non_blocking_handlers = newValue;
            addMissingResources(state);
          })
        );
      },
      [setState]
    );

    const navigate = useNavigate();

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

        startReauthentication(navigate, locationState).catch((e) => {
          // Normally there should not be any error.
          console.error(e);
        });
      },
      [navigate, state.secret]
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

    const blockingHandlers: BlockingEventHandler[] = useMemo(() => {
      const diff = state.diff;
      const cfgs = state.blocking_handlers;
      const out = [];
      for (const c of cfgs) {
        out.push({
          ...c,
          kind: getKind(c.url),
          isDirty: checkDirty(diff, c.url),
        });
      }
      return out;
    }, [state.diff, state.blocking_handlers]);

    const nonBlockingHandlers: NonBlockingEventHandler[] = useMemo(() => {
      const diff = state.diff;
      const cfgs = state.non_blocking_handlers;
      const out = [];
      for (const c of cfgs) {
        out.push({
          ...c,
          kind: getKind(c.url),
          isDirty: checkDirty(diff, c.url),
        });
      }
      return out;
    }, [state.diff, state.non_blocking_handlers]);

    return (
      <FormContainer
        form={form}
        hideFooterComponent={codeEditorState != null}
        stickyFooterComponent={true}
        showDiscardButton={true}
      >
        <ScreenContent>
          {codeEditorState != null ? (
            <div className={cn(styles.codeEditorContainer)}>
              <WidgetTitle>
                <FormattedMessage id="HookConfigurationScreen.edit-hook.label" />
              </WidgetTitle>
              <WidgetDescription>
                <FormattedMessage id="HookConfigurationScreen.edit-hook.description" />
              </WidgetDescription>
              <CodeEditor
                className={styles.codeEditor}
                language="typescript"
                value={code}
                onChange={onChangeCode}
                options={CODE_EDITOR_OPTIONS}
              />
              <div className={styles.codeEditorFooter}>
                <PrimaryButton
                  text="Finish Editing"
                  onClick={onClickFinishEditing}
                  disabled={isLoading}
                />
                <DefaultButton
                  text="Cancel"
                  onClick={onClickCancelEditing}
                  disabled={isLoading}
                />
              </div>
            </div>
          ) : (
            <>
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
                    listClassName={styles.hookList}
                    listItemClassName={styles.hookListItem}
                    listItemStyle={{
                      borderBottomColor: theme.semanticColors.bodyDivider,
                    }}
                    label={
                      <>
                        <Label>
                          <FormattedMessage id="HookConfigurationScreen.blocking-handlers.label" />
                        </Label>
                        <div
                          className={styles.hookHeader}
                          style={{
                            borderBottomColor: theme.semanticColors.bodyDivider,
                          }}
                        >
                          <Text
                            block={true}
                            className={styles.blockingHookKind}
                            styles={{
                              root: {
                                color: theme.semanticColors.bodySubtext,
                              },
                            }}
                          >
                            <FormattedMessage id="HookConfigurationScreen.header.type.label" />
                          </Text>
                          <Text
                            block={true}
                            className={styles.blockingHookEvent}
                            styles={{
                              root: {
                                color: theme.semanticColors.bodySubtext,
                              },
                            }}
                          >
                            <FormattedMessage id="HookConfigurationScreen.header.event.label" />
                          </Text>
                          <Text
                            block={true}
                            className={styles.blockingHookConfig}
                            styles={{
                              root: {
                                color: theme.semanticColors.bodySubtext,
                              },
                            }}
                          >
                            <FormattedMessage id="HookConfigurationScreen.header.config.label" />
                          </Text>
                        </div>
                      </>
                    }
                    parentJSONPointer="/hook"
                    fieldName="blocking_handlers"
                    list={blockingHandlers}
                    onListItemAdd={onBlockingHandlersChange}
                    onListItemChange={onBlockingHandlersChangeItemChange}
                    onListItemDelete={onBlockingHandlersChange}
                    makeDefaultItem={makeDefaultHandler}
                    ListItemComponent={BlockingHandlerListItem}
                    addButtonLabelMessageID="add"
                    addDisabled={blockingHandlerLimitReached}
                  />
                ) : null}
              </Widget>
              <Widget className={styles.widget}>
                <HorizontalDivider />
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
                    listClassName={styles.hookList}
                    listItemClassName={styles.hookListItem}
                    listItemStyle={{
                      borderBottomColor: theme.semanticColors.bodyDivider,
                    }}
                    label={
                      <>
                        <Label>
                          <FormattedMessage id="HookConfigurationScreen.non-blocking-events-endpoints.label" />
                        </Label>
                        <div
                          className={styles.hookHeader}
                          style={{
                            borderBottomColor: theme.semanticColors.bodyDivider,
                          }}
                        >
                          <Text
                            block={true}
                            className={styles.nonblockingHookEvent}
                            styles={{
                              root: {
                                color: theme.semanticColors.bodySubtext,
                              },
                            }}
                          >
                            <FormattedMessage id="HookConfigurationScreen.header.event.label" />
                          </Text>
                          <Text
                            block={true}
                            className={styles.nonblockingHookConfig}
                            styles={{
                              root: {
                                color: theme.semanticColors.bodySubtext,
                              },
                            }}
                          >
                            <FormattedMessage id="HookConfigurationScreen.header.config.label" />
                          </Text>
                        </div>
                      </>
                    }
                    parentJSONPointer="/hook"
                    fieldName="non_blocking_handlers"
                    list={nonBlockingHandlers}
                    onListItemAdd={onNonBlockingHandlersChange}
                    onListItemChange={onNonBlockingHandlersChange}
                    onListItemDelete={onNonBlockingHandlersChange}
                    makeDefaultItem={makeDefaultNonBlockingHandler}
                    ListItemComponent={NonBlockingHandlerListItem}
                    addButtonLabelMessageID="add"
                    addDisabled={nonBlockingHandlerLimitReached}
                  />
                ) : null}
              </Widget>
              <HorizontalDivider className={styles.separator} />
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
                  label={renderToString(
                    "HookConfigurationScreen.timeout.label"
                  )}
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
            </>
          )}
        </ScreenContent>
      </FormContainer>
    );
  };

const HookConfigurationScreen1: React.VFC<{
  appID: string;
  secretToken: string | null;
}> = function HookConfigurationScreen1({ appID, secretToken }) {
  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: secretToken,
    constructFormState: constructConfigFormState,
    constructConfig,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);

  if (featureConfig.loading) {
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

const SECRETS = [AppSecretKey.WebhookSecret];

const HookConfigurationScreen: React.VFC = function HookConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const location = useLocation();
  const [shouldRefreshToken] = useState<boolean>(() => {
    const { state } = location;
    if (isLocationState(state) && state.isOAuthRedirect) {
      return true;
    }
    return false;
  });
  const { token, error, retry } = useAppSecretVisitToken(
    appID,
    SECRETS,
    shouldRefreshToken
  );

  if (error) {
    return <ShowError error={error} onRetry={retry} />;
  }

  if (token === undefined) {
    return <ShowLoading />;
  }

  return <HookConfigurationScreen1 appID={appID} secretToken={token} />;
};

export default HookConfigurationScreen;
