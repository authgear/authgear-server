import cn from "classnames";
import React, { useCallback, useContext, useEffect, useMemo, useRef, useState } from "react";
import { Context, FormattedMessage } from "../../intl";
import { useLocation, useParams, useNavigate } from "react-router-dom";
import { produce } from "immer";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import {
  ChevronDownIcon,
  CopyIcon,
  EyeOpenIcon,
  InfoCircledIcon,
  Pencil1Icon,
  PlusIcon,
  TrashIcon,
} from "@radix-ui/react-icons";
import {
  Callout as RadixCallout,
  Flex,
  IconButton as RadixIconButton,
  RadioGroup,
  Select,
  Tabs,
  Text as RadixText,
  Tooltip as RadixTooltip,
} from "@radix-ui/themes";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import ScreenContent from "../../ScreenContent";
import WidgetTitle from "../../WidgetTitle";
import {
  BlockingHookHandlerConfig,
  HookFeatureConfig,
  HookKind,
  NonBlockingHookHandlerConfig,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  getHookKind,
} from "../../types";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useResourceForm } from "../../hook/useResourceForm";
import {
  ResourceSpecifier,
  Resource,
  ResourcesDiffResult,
  getDenoScriptPathFromURL,
  makeDenoScriptSpecifier,
} from "../../util/resource";
import { copyToClipboard } from "../../util/clipboard";
import FormContainer from "../../FormContainer";
import { clearEmptyObject } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { genRandomHexadecimalString } from "../../util/random";
import styles from "./HookConfigurationScreen.module.css";
import WidgetDescription from "../../WidgetDescription";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { useCheckDenoHookMutation } from "./mutations/checkDenoHook";
import { startReauthentication } from "./Authenticated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useLoading, useIsLoading } from "../../hook/loading";
import { useProvideError } from "../../hook/error";
import { TextField as RadixTextField } from "../../components/v2/TextField/TextField";
import ExternalLink from "../../ExternalLink";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import PrimaryButton from "../../PrimaryButton";
import CodeEditor from "../../CodeEditor";
import DefaultButton from "../../DefaultButton";
import { AppSecretKey } from "./globalTypes.generated";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { DENO_TYPES_URL } from "../../util/deno";

const CODE_EDITOR_OPTIONS = {
  minimap: {
    enabled: false,
  },
};

const BLOCK_EVENT_TYPES = [
  "user.pre_create",
  "user.profile.pre_update",
  "user.pre_schedule_deletion",
  "user.pre_schedule_anonymization",
  "oidc.jwt.pre_create",
  "oidc.id_token.pre_create",
  "authentication.pre_initialize",
  "authentication.post_identified",
  "authentication.pre_authenticated",
] as const;

type BlockingEvent = (typeof BLOCK_EVENT_TYPES)[number];

const BLOCK_EVENT_CATEGORIES: Array<{
  labelId: string;
  events: BlockingEvent[];
}> = [
  {
    labelId: "HookConfigurationScreen.event-category.user",
    events: [
      "user.pre_create",
      "user.profile.pre_update",
      "user.pre_schedule_deletion",
      "user.pre_schedule_anonymization",
    ],
  },
  {
    labelId: "HookConfigurationScreen.event-category.oidc",
    events: ["oidc.jwt.pre_create", "oidc.id_token.pre_create"],
  },
  {
    labelId: "HookConfigurationScreen.event-category.authentication",
    events: [
      "authentication.pre_initialize",
      "authentication.post_identified",
      "authentication.pre_authenticated",
    ],
  },
];

const BLOCKING_EVENT_DESCRIPTION_MESSAGE_IDS: Record<BlockingEvent, string> = {
  "user.pre_create":
    "HookConfigurationScreen.blocking-event.user.pre_create.description",
  "user.profile.pre_update":
    "HookConfigurationScreen.blocking-event.user.profile.pre_update.description",
  "user.pre_schedule_deletion":
    "HookConfigurationScreen.blocking-event.user.pre_schedule_deletion.description",
  "user.pre_schedule_anonymization":
    "HookConfigurationScreen.blocking-event.user.pre_schedule_anonymization.description",
  "oidc.jwt.pre_create":
    "HookConfigurationScreen.blocking-event.oidc.jwt.pre_create.description",
  "oidc.id_token.pre_create":
    "HookConfigurationScreen.blocking-event.oidc.id_token.pre_create.description",
  "authentication.pre_initialize":
    "HookConfigurationScreen.blocking-event.authentication.pre_initialize.description",
  "authentication.post_identified":
    "HookConfigurationScreen.blocking-event.authentication.post_identified.description",
  "authentication.pre_authenticated":
    "HookConfigurationScreen.blocking-event.authentication.pre_authenticated.description",
};

const BLOCKING_EVENT_NAME_TO_PAYLOAD_TYPE_NAME: Record<BlockingEvent, string> =
  {
    "user.pre_create": "EventUserPreCreate",
    "user.profile.pre_update": "EventUserProfilePreUpdate",
    "user.pre_schedule_deletion": "EventUserPreScheduleDeletion",
    "user.pre_schedule_anonymization": "EventUserPreScheduleAnonymization",
    "oidc.jwt.pre_create": "EventOIDCJWTPreCreate",
    "oidc.id_token.pre_create": "EventOIDCIDTokenPreCreate",
    "authentication.pre_initialize": "EventAuthenticationPreInitialize",
    "authentication.post_identified": "EventAuthenticationPostIdentified",
    "authentication.pre_authenticated": "EventAuthenticationPreAuthenticated",
  };

const BLOCKING_EVENT_NAME_TO_RESPONSE_TYPE_NAME: Record<BlockingEvent, string> =
  {
    "user.pre_create": "EventUserPreCreateHookResponse",
    "user.profile.pre_update": "EventUserProfilePreUpdateHookResponse",
    "user.pre_schedule_deletion": "EventUserPreScheduleDeletionHookResponse",
    "user.pre_schedule_anonymization":
      "EventUserPreScheduleAnonymizationHookResponse",
    "oidc.jwt.pre_create": "EventOIDCJWTPreCreateHookResponse",
    "oidc.id_token.pre_create": "EventOIDCIDTokenPreCreateHookResponse",
    "authentication.pre_initialize":
      "EventAuthenticationPreInitializeHookResponse",
    "authentication.post_identified":
      "EventAuthenticationPostIdentifiedHookResponse",
    "authentication.pre_authenticated":
      "EventAuthenticationPreAuthenticatedHookResponse",
  };

const BLOCKING_HOOK_EXAMPLES: Record<BlockingEvent, string> = {
  "user.pre_create": ``,
  "user.profile.pre_update": ``,
  "user.pre_schedule_deletion": ``,
  "user.pre_schedule_anonymization": ``,
  "oidc.jwt.pre_create": ``,
  "oidc.id_token.pre_create": ``,
  "authentication.pre_initialize": `
// This event is triggered right before any authentication, such as login. 
//
// For example, if your business only operate during weekdays, therefore you do not want any user login during weekends:
// 
// const today = new Date();
// // 0 is sunday, and 6 is saturday
// if (today.getDay() === 0 || today.getDay() === 6) {
//   return {
//     is_allowed: false,
//   };
// }
// return {
//   is_allowed: true,
// };`,
  "authentication.post_identified": `
// This event is triggered after the identification step during signup/login.
// For example, block login based on email address:
//
// const email = e.payload.identification.identity?.claims?.email
// if (typeof email === "string" && email.endsWith("@authgear.com")) {
//   return {
//     is_allowed: true,
//   };
// }
// return {
//   is_allowed: false,
//   reason: "Email address not allowed"
// };`,
  "authentication.pre_authenticated": `
// This event is triggered right before any authentication completes, such as login. 
//
// For example, logins from outside \`HK\` are considered at a higher risk, 
// and MFA is enforced.
// if (e.context.geo_location_code !== "HK") {
//   return {
//     is_allowed: true,
//     constraints:{
//       amr: ["mfa"]
//     }
//   };
// }
// return {
//   is_allowed: true,
// };`,
};

const DENOHOOK_NONBLOCKING_DEFAULT = `import { HookNonBlockingEvent } from "${DENO_TYPES_URL}";

export default async function(e: HookNonBlockingEvent): Promise<void> {
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

function makeDefaultDenoHookBlockingScript(event: BlockingEvent): string {
  const payloadTypeName = BLOCKING_EVENT_NAME_TO_PAYLOAD_TYPE_NAME[event];
  const responseTypeName = BLOCKING_EVENT_NAME_TO_RESPONSE_TYPE_NAME[event];
  const exampleCode = BLOCKING_HOOK_EXAMPLES[event];
  return `import { ${payloadTypeName}, ${responseTypeName} } from "${DENO_TYPES_URL}";

export default async function(e: ${payloadTypeName}): Promise<${responseTypeName}> {
  // Write your hook with the help of the type definition.
${exampleCode.replace(/^/gm, "  ")}
  return { is_allowed: true };
}
`;
}

type EventKind = "blocking" | "nonblocking";

interface BlockingEventHandler {
  event: string;
  kind: HookKind;
  url: string;
  isDirty: boolean;
}

interface NonBlockingEventHandler {
  name: string;
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

function checkDirty(diff: ResourcesDiffResult | null, url: string): boolean {
  if (diff == null) {
    return false;
  }

  const kind = getHookKind(url);
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

function BlockingEventInfoIcon({
  event,
}: {
  event: string;
}): React.ReactElement | null {
  const messageID =
    BLOCKING_EVENT_DESCRIPTION_MESSAGE_IDS[event as BlockingEvent];
  if (messageID == null) {
    return null;
  }

  return (
    <RadixTooltip content={<FormattedMessage id={messageID} />}>
      <InfoCircledIcon
        className={styles.hookEventInfoIcon}
        width="1rem"
        height="1rem"
        aria-hidden
      />
    </RadixTooltip>
  );
}

function CopyIconButton({
  textToCopy,
}: {
  textToCopy: string;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleCopy = useCallback(() => {
    copyToClipboard(textToCopy);
    setCopied(true);
    if (timerRef.current != null) {
      clearTimeout(timerRef.current);
    }
    timerRef.current = setTimeout(() => {
      setCopied(false);
    }, 2000);
  }, [textToCopy]);

  return (
    <RadixTooltip
      content={
        copied
          ? renderToString("copied-to-clipboard")
          : renderToString("copy")
      }
      open={copied ? true : undefined}
    >
      <RadixIconButton
        type="button"
        variant="ghost"
        color="gray"
        size="1"
        aria-label={renderToString("copy")}
        onClick={handleCopy}
        className={styles.copyIconButton}
      >
        <CopyIcon width="1rem" height="1rem" />
      </RadixIconButton>
    </RadixTooltip>
  );
}

function RevealIconButton({
  onClick,
}: {
  onClick: () => void;
}): React.ReactElement {
  const { renderToString } = useContext(Context);

  return (
    <RadixTooltip content={renderToString("reveal")}>
      <RadixIconButton
        type="button"
        variant="ghost"
        color="gray"
        size="1"
        aria-label={renderToString("reveal")}
        onClick={onClick}
        className={styles.copyIconButton}
      >
        <EyeOpenIcon width="1rem" height="1rem" />
      </RadixIconButton>
    </RadixTooltip>
  );
}


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

function makeNewURL(eventKind: EventKind): string {
  const rand = genRandomHexadecimalString();
  return `authgeardeno:///deno/${eventKind}.${rand}.ts`;
}

function makeSpecifiersFromState(state: ConfigFormState): ResourceSpecifier[] {
  const specifiers: ResourceSpecifier[] = [];
  for (const h of state.blocking_handlers) {
    if (getHookKind(h.url) === "denohook") {
      specifiers.push(makeDenoScriptSpecifier(h.url));
    }
  }
  for (const h of state.non_blocking_handlers) {
    if (getHookKind(h.url) === "denohook") {
      specifiers.push(makeDenoScriptSpecifier(h.url));
    }
  }
  return specifiers;
}

function addMissingResources(state: FormState) {
  for (let i = 0; i < state.blocking_handlers.length; ++i) {
    const h = state.blocking_handlers[i];
    if (getHookKind(h.url) === "denohook") {
      const path = getDenoScriptPathFromURL(h.url);
      const specifier = makeDenoScriptSpecifier(h.url);
      const r = state.resources.find((r) => r.path === path);
      if (r == null) {
        state.resources.push({
          path,
          specifier,
          nullableValue: makeDefaultDenoHookBlockingScript(
            h.event as BlockingEvent
          ),
        });
      }
    }
  }
  for (let i = 0; i < state.non_blocking_handlers.length; ++i) {
    const h = state.non_blocking_handlers[i];
    if (getHookKind(h.url) === "denohook") {
      const path = getDenoScriptPathFromURL(h.url);
      const specifier = makeDenoScriptSpecifier(h.url);
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

interface BlockingHooksTableProps {
  handlers: BlockingEventHandler[];
  onHandlersChange: (handlers: BlockingEventHandler[]) => void;
  onHandlerItemChange: (
    handlers: BlockingEventHandler[],
    index: number,
    item: BlockingEventHandler
  ) => void;
  makeDefaultHandler: () => BlockingEventHandler;
  onEditDeno: (index: number, value: BlockingEventHandler) => void;
  addDisabled: boolean;
  onEditingChange?: (editing: boolean) => void;
}

function BlockingHooksTable({
  handlers,
  onHandlersChange,
  onHandlerItemChange,
  makeDefaultHandler,
  onEditDeno,
  addDisabled,
  onEditingChange,
}: BlockingHooksTableProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { isDirty } = useFormContainerBaseContext();

  const [expandedIndex, setExpandedIndex] = useState<number | null>(null);
  const [draft, setDraft] = useState<BlockingEventHandler | null>(null);
  const [pendingEventName, setPendingEventName] = useState<string | null>(null);

  useEffect(() => {
    onEditingChange?.(expandedIndex != null);
  }, [expandedIndex, onEditingChange]);

  useEffect(() => {
    if (!isDirty) {
      setExpandedIndex(null);
      setDraft(null);
    }
  }, [isDirty]);

  const applyDraftChange = useCallback(
    (newDraft: BlockingEventHandler) => {
      if (expandedIndex == null) {
        return;
      }
      setDraft(newDraft);
      const newHandlers = handlers.map((h, i) =>
        i === expandedIndex ? newDraft : h
      );
      onHandlerItemChange(newHandlers, expandedIndex, newDraft);
    },
    [expandedIndex, handlers, onHandlerItemChange]
  );

  const kindOptions = useMemo(
    () => [
      {
        value: "webhook",
        label: renderToString("HookConfigurationScreen.hook-kind.webhook"),
      },
      {
        value: "denohook",
        label: renderToString("HookConfigurationScreen.hook-kind.denohook"),
      },
    ],
    [renderToString]
  );

  const onClickAdd = useCallback(() => {
    const newHandler = makeDefaultHandler();
    const newHandlers = [...handlers, newHandler];
    onHandlersChange(newHandlers);
    setExpandedIndex(newHandlers.length - 1);
    setDraft({ ...newHandler });
  }, [handlers, makeDefaultHandler, onHandlersChange]);

  const onClickEdit = useCallback(
    (index: number) => {
      if (expandedIndex === index) {
        setExpandedIndex(null);
        setDraft(null);
        return;
      }
      setExpandedIndex(index);
      setDraft({ ...handlers[index] });
    },
    [handlers, expandedIndex]
  );

  const onClickDelete = useCallback(
    (index: number) => {
      const newHandlers = handlers.filter((_, i) => i !== index);
      onHandlersChange(newHandlers);
      if (expandedIndex === index) {
        setExpandedIndex(null);
        setDraft(null);
      } else if (expandedIndex != null && expandedIndex > index) {
        setExpandedIndex(expandedIndex - 1);
      }
    },
    [handlers, onHandlersChange, expandedIndex]
  );

  const onDraftKindChange = useCallback(
    (kind: string) => {
      if (draft == null) {
        return;
      }
      if (kind === "webhook") {
        applyDraftChange({ ...draft, kind: "webhook", url: "" });
      } else if (kind === "denohook") {
        applyDraftChange({
          ...draft,
          kind: "denohook",
          url: makeNewURL("blocking"),
        });
      }
    },
    [draft, applyDraftChange]
  );

  const onDraftEventChange = useCallback(
    (event: string) => {
      if (draft == null) {
        return;
      }
      if (draft.kind === "denohook") {
        setPendingEventName(event);
      } else {
        applyDraftChange({ ...draft, event });
      }
    },
    [draft, applyDraftChange]
  );

  const onConfirmEventChange = useCallback(() => {
    if (draft == null || pendingEventName == null) {
      return;
    }
    applyDraftChange({ ...draft, event: pendingEventName });
    setPendingEventName(null);
  }, [draft, pendingEventName, applyDraftChange]);

  const onCancelEventChange = useCallback(() => {
    setPendingEventName(null);
  }, []);

  const onDraftURLChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (draft == null) {
        return;
      }
      applyDraftChange({ ...draft, url: e.target.value });
    },
    [draft, applyDraftChange]
  );

  const onClickEditScript = useCallback(() => {
    if (expandedIndex == null) {
      return;
    }
    const handler = handlers[expandedIndex];
    setExpandedIndex(null);
    setDraft(null);
    onEditDeno(expandedIndex, handler);
  }, [expandedIndex, handlers, onEditDeno]);

  return (
    <>
      <ConfirmationDialog
        open={pendingEventName != null}
        onOpenChange={(open) => {
          if (!open) onCancelEventChange();
        }}
        title={
          <FormattedMessage id="HookConfigurationScreen.change-event.title" />
        }
        description={
          <FormattedMessage id="HookConfigurationScreen.change-event.description" />
        }
        confirmText={
          <FormattedMessage id="HookConfigurationScreen.change-event.label" />
        }
        cancelText={<FormattedMessage id="cancel" />}
        onConfirm={onConfirmEventChange}
        onCancel={onCancelEventChange}
        confirmColor="red"
      />

      {handlers.length === 0 ? (
        <RadixCallout.Root color="gray" variant="surface" size="1">
          <RadixCallout.Icon>
            <InfoCircledIcon width="1rem" height="1rem" />
          </RadixCallout.Icon>
          <RadixCallout.Text>
            <FormattedMessage id="HookConfigurationScreen.blocking-handlers.empty" />
          </RadixCallout.Text>
        </RadixCallout.Root>
      ) : (
        <div className={styles.hookAccordionList}>
          {handlers.map((handler, index) => {
            const isOpen = expandedIndex === index;
            return (
              <div key={index} className={styles.hookAccordionItem}>
                {/* Accordion header */}
                <div className={styles.hookAccordionHeader}>
                  <button
                    type="button"
                    className={styles.hookAccordionToggle}
                    onClick={() => onClickEdit(index)}
                    aria-expanded={isOpen}
                  >
                    <RadixText size="2" weight="medium" className={styles.hookCellTruncate}>
                      {handler.event}
                    </RadixText>
                    {handler.isDirty ? (
                      <RadixText size="1" className={styles.hookDirtyDot}>
                        {"●"}
                      </RadixText>
                    ) : null}
                  </button>
                  <RadixIconButton
                    type="button"
                    variant="ghost"
                    color="red"
                    size="2"
                    onClick={() => onClickDelete(index)}
                  >
                    <TrashIcon width="1rem" height="1rem" />
                  </RadixIconButton>
                  <button
                    type="button"
                    className={styles.hookAccordionChevronButton}
                    onClick={() => onClickEdit(index)}
                    aria-expanded={isOpen}
                  >
                    <ChevronDownIcon
                      className={cn(
                        styles.hookAccordionChevron,
                        isOpen && styles.hookAccordionChevronOpen
                      )}
                      width="1rem"
                      height="1rem"
                    />
                  </button>
                </div>

                {/* Accordion body - shown when expanded */}
                {isOpen && draft != null ? (
                  <div className={styles.hookAccordionBody}>
                    <div className={styles.hookAccordionField}>
                      <RadixText as="label" size="1" weight="medium" color="gray">
                        <FormattedMessage id="HookConfigurationScreen.header.type.label" />
                      </RadixText>
                      <Select.Root
                        size="2"
                        value={draft.kind}
                        onValueChange={onDraftKindChange}
                      >
                        <Select.Trigger />
                        <Select.Content style={{ zIndex: 200 }}>
                          {kindOptions.map((opt) => (
                            <Select.Item key={opt.value} value={opt.value}>
                              {opt.label}
                            </Select.Item>
                          ))}
                        </Select.Content>
                      </Select.Root>
                    </div>
                    <div className={styles.hookAccordionField}>
                      <RadixText as="p" size="1" weight="medium" color="gray">
                        <FormattedMessage id="HookConfigurationScreen.header.event.label" />
                      </RadixText>
                      <RadioGroup.Root
                        value={draft.event}
                        onValueChange={onDraftEventChange}
                        className={styles.hookEventRadioContainer}
                      >
                        <Flex direction="column" gap="3">
                          {BLOCK_EVENT_CATEGORIES.map((category) => (
                            <Flex key={category.labelId} direction="column" gap="2">
                              <RadixText
                                as="p"
                                size="1"
                                weight="bold"
                                className={styles.hookEventCategoryLabel}
                              >
                                <FormattedMessage id={category.labelId} />
                              </RadixText>
                              {category.events.map((eventType) => (
                                <RadixText key={eventType} as="label" size="2">
                                  <Flex gap="2" align="center">
                                    <RadioGroup.Item value={eventType} />
                                    {eventType}
                                    <BlockingEventInfoIcon event={eventType} />
                                  </Flex>
                                </RadixText>
                              ))}
                            </Flex>
                          ))}
                        </Flex>
                      </RadioGroup.Root>
                    </div>
                    {draft.kind === "webhook" ? (
                      <div className={styles.hookAccordionField}>
                        <RadixText as="label" size="1" weight="medium" color="gray">
                          <FormattedMessage id="HookConfigurationScreen.action.endpoint.label" />
                        </RadixText>
                        <RadixTextField.Input
                          size="2"
                          value={draft.url}
                          onChange={onDraftURLChange}
                          placeholder="https://example.com/callback"
                        >
                          {null}
                        </RadixTextField.Input>
                      </div>
                    ) : null}
                    {draft.kind === "denohook" ? (
                      <div className={styles.hookAccordionField}>
                        <RadixText as="label" size="1" weight="medium" color="gray">
                          <FormattedMessage id="HookConfigurationScreen.action.script.label" />
                        </RadixText>
                        <button
                          type="button"
                          className={styles.editScriptButton}
                          onClick={onClickEditScript}
                        >
                          <Pencil1Icon />
                          <FormattedMessage id="HookConfigurationScreen.edit-hook.label" />
                        </button>
                      </div>
                    ) : null}
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>
      )}

      {/* Add button */}
      {!addDisabled ? (
        <button
          type="button"
          className={styles.hookAddButton}
          onClick={onClickAdd}
        >
          <PlusIcon />
          <FormattedMessage id="add" />
        </button>
      ) : null}
    </>
  );
}


interface NonBlockingHooksTableProps {
  handlers: NonBlockingEventHandler[];
  onHandlersChange: (handlers: NonBlockingEventHandler[]) => void;
  onHandlerItemChange: (
    handlers: NonBlockingEventHandler[],
    index: number,
    item: NonBlockingEventHandler
  ) => void;
  makeDefaultHandler: () => NonBlockingEventHandler;
  onEditDeno: (index: number, value: NonBlockingEventHandler) => void;
  addDisabled: boolean;
  onEditingChange?: (editing: boolean) => void;
}

function NonBlockingHooksTable({
  handlers,
  onHandlersChange,
  onHandlerItemChange,
  makeDefaultHandler,
  onEditDeno,
  addDisabled,
  onEditingChange,
}: NonBlockingHooksTableProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { isDirty } = useFormContainerBaseContext();

  const [expandedIndex, setExpandedIndex] = useState<number | null>(null);
  const [draft, setDraft] = useState<NonBlockingEventHandler | null>(null);

  useEffect(() => {
    onEditingChange?.(expandedIndex != null);
  }, [expandedIndex, onEditingChange]);

  useEffect(() => {
    if (!isDirty) {
      setExpandedIndex(null);
      setDraft(null);
    }
  }, [isDirty]);

  const applyDraftChange = useCallback(
    (newDraft: NonBlockingEventHandler) => {
      if (expandedIndex == null) {
        return;
      }
      setDraft(newDraft);
      const newHandlers = handlers.map((h, i) =>
        i === expandedIndex ? newDraft : h
      );
      onHandlerItemChange(newHandlers, expandedIndex, newDraft);
    },
    [expandedIndex, handlers, onHandlerItemChange]
  );

  const kindOptions = useMemo(
    () => [
      {
        value: "webhook",
        label: renderToString("HookConfigurationScreen.hook-kind.webhook"),
      },
      {
        value: "denohook",
        label: renderToString("HookConfigurationScreen.hook-kind.denohook"),
      },
    ],
    [renderToString]
  );

  const onClickAdd = useCallback(() => {
    const newHandler = makeDefaultHandler();
    const newHandlers = [...handlers, newHandler];
    onHandlersChange(newHandlers);
    setExpandedIndex(newHandlers.length - 1);
    setDraft({ ...newHandler });
  }, [handlers, makeDefaultHandler, onHandlersChange]);

  const onClickEdit = useCallback(
    (index: number) => {
      if (expandedIndex === index) {
        setExpandedIndex(null);
        setDraft(null);
        return;
      }
      setExpandedIndex(index);
      setDraft({ ...handlers[index] });
    },
    [handlers, expandedIndex]
  );

  const onClickDelete = useCallback(
    (index: number) => {
      const newHandlers = handlers.filter((_, i) => i !== index);
      onHandlersChange(newHandlers);
      if (expandedIndex === index) {
        setExpandedIndex(null);
        setDraft(null);
      } else if (expandedIndex != null && expandedIndex > index) {
        setExpandedIndex(expandedIndex - 1);
      }
    },
    [handlers, onHandlersChange, expandedIndex]
  );

  const onDraftKindChange = useCallback(
    (kind: string) => {
      if (draft == null) {
        return;
      }
      if (kind === "webhook") {
        applyDraftChange({ ...draft, kind: "webhook", url: "" });
      } else if (kind === "denohook") {
        applyDraftChange({
          ...draft,
          kind: "denohook",
          url: makeNewURL("nonblocking"),
        });
      }
    },
    [draft, applyDraftChange]
  );

  const onDraftNameChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (draft == null) {
        return;
      }
      applyDraftChange({ ...draft, name: e.target.value });
    },
    [draft, applyDraftChange]
  );

  const onDraftURLChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      if (draft == null) {
        return;
      }
      applyDraftChange({ ...draft, url: e.target.value });
    },
    [draft, applyDraftChange]
  );

  const onClickEditScript = useCallback(() => {
    if (expandedIndex == null) {
      return;
    }
    const handler = handlers[expandedIndex];
    setExpandedIndex(null);
    setDraft(null);
    onEditDeno(expandedIndex, handler);
  }, [expandedIndex, handlers, onEditDeno]);

  return (
    <>
      {handlers.length === 0 ? (
        <RadixCallout.Root color="gray" variant="surface" size="1">
          <RadixCallout.Icon>
            <InfoCircledIcon width="1rem" height="1rem" />
          </RadixCallout.Icon>
          <RadixCallout.Text>
            <FormattedMessage id="HookConfigurationScreen.non-blocking-handlers.empty" />
          </RadixCallout.Text>
        </RadixCallout.Root>
      ) : (
        <div className={styles.hookAccordionList}>
          {handlers.map((handler, index) => {
            const isOpen = expandedIndex === index;
            const fallbackLabel =
              handler.kind === "webhook"
                ? handler.url ||
                  renderToString("HookConfigurationScreen.hook-kind.webhook")
                : renderToString("HookConfigurationScreen.hook-kind.denohook");
            const headerLabel = handler.name || fallbackLabel;

            return (
              <div key={index} className={styles.hookAccordionItem}>
                <div className={styles.hookAccordionHeader}>
                  <button
                    type="button"
                    className={styles.hookAccordionToggle}
                    onClick={() => onClickEdit(index)}
                    aria-expanded={isOpen}
                  >
                    <RadixText
                      size="2"
                      weight="medium"
                      className={styles.hookCellTruncate}
                    >
                      {headerLabel}
                    </RadixText>
                    {handler.isDirty ? (
                      <RadixText size="1" className={styles.hookDirtyDot}>
                        {"●"}
                      </RadixText>
                    ) : null}
                  </button>
                  <RadixIconButton
                    type="button"
                    variant="ghost"
                    color="red"
                    size="2"
                    onClick={() => onClickDelete(index)}
                  >
                    <TrashIcon width="1rem" height="1rem" />
                  </RadixIconButton>
                  <button
                    type="button"
                    className={styles.hookAccordionChevronButton}
                    onClick={() => onClickEdit(index)}
                    aria-expanded={isOpen}
                  >
                    <ChevronDownIcon
                      className={cn(
                        styles.hookAccordionChevron,
                        isOpen && styles.hookAccordionChevronOpen
                      )}
                      width="1rem"
                      height="1rem"
                    />
                  </button>
                </div>

                {isOpen && draft != null ? (
                  <div className={styles.hookAccordionBody}>
                    <div className={styles.hookAccordionField}>
                      <RadixText as="label" size="1" weight="medium" color="gray">
                        <FormattedMessage id="HookConfigurationScreen.non-blocking-handler.name.label" />
                      </RadixText>
                      <RadixTextField.Input
                        size="2"
                        value={draft.name}
                        onChange={onDraftNameChange}
                        placeholder={renderToString("HookConfigurationScreen.non-blocking-handler.name.placeholder")}
                      >
                        {null}
                      </RadixTextField.Input>
                    </div>
                    <div className={styles.hookAccordionField}>
                      <RadixText as="label" size="1" weight="medium" color="gray">
                        <FormattedMessage id="HookConfigurationScreen.header.type.label" />
                      </RadixText>
                      <Select.Root
                        size="2"
                        value={draft.kind}
                        onValueChange={onDraftKindChange}
                      >
                        <Select.Trigger />
                        <Select.Content style={{ zIndex: 200 }}>
                          {kindOptions.map((opt) => (
                            <Select.Item key={opt.value} value={opt.value}>
                              {opt.label}
                            </Select.Item>
                          ))}
                        </Select.Content>
                      </Select.Root>
                    </div>
                    {draft.kind === "webhook" ? (
                      <div className={styles.hookAccordionField}>
                        <RadixText as="label" size="1" weight="medium" color="gray">
                          <FormattedMessage id="HookConfigurationScreen.action.endpoint.label" />
                        </RadixText>
                        <RadixTextField.Input
                          size="2"
                          value={draft.url}
                          onChange={onDraftURLChange}
                          placeholder="https://example.com/callback"
                        >
                          {null}
                        </RadixTextField.Input>
                      </div>
                    ) : null}
                    {draft.kind === "denohook" ? (
                      <div className={styles.hookAccordionField}>
                        <RadixText as="label" size="1" weight="medium" color="gray">
                          <FormattedMessage id="HookConfigurationScreen.action.script.label" />
                        </RadixText>
                        <button
                          type="button"
                          className={styles.editScriptButton}
                          onClick={onClickEditScript}
                        >
                          <Pencil1Icon />
                          <FormattedMessage id="HookConfigurationScreen.edit-hook.label" />
                        </button>
                      </div>
                    ) : null}
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>
      )}

      {!addDisabled ? (
        <button
          type="button"
          className={styles.hookAddButton}
          onClick={onClickAdd}
        >
          <PlusIcon />
          <FormattedMessage id="add" />
        </button>
      ) : null}
    </>
  );
}

function HookScreenWithSaveBar({
  codeEditorState,
  anchorRef,
  children,
}: {
  codeEditorState: CodeEditorState | null;
  anchorRef: React.RefObject<HTMLDivElement>;
  children: React.ReactNode;
}): React.ReactElement {
  const { isDirty } = useFormContainerBaseContext();
  return (
    <ScreenContent
      className={isDirty && codeEditorState == null ? styles.contentWithSaveBar : undefined}
    >
      {children}
      {codeEditorState == null ? (
        <SaveFunctionBar anchorRef={anchorRef} />
      ) : null}
    </ScreenContent>
  );
}

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
  function HookConfigurationScreenContent(props) {
    const { appID } = useParams() as { appID: string };
    const { hookFeatureConfig, form: config } = props;

    const [codeEditorState, setCodeEditorState] =
      useState<CodeEditorState | null>(null);
    const [activeTab, setActiveTab] = useState("blocking-events");
    const [blockingTableEditing, setBlockingTableEditing] = useState(false);
    const [nonBlockingTableEditing, setNonBlockingTableEditing] = useState(false);
    const hookTableEditing = blockingTableEditing || nonBlockingTableEditing;

    const clearHookTableEditing = useCallback(() => {
      setBlockingTableEditing(false);
      setNonBlockingTableEditing(false);
    }, []);

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
        config.isDirty ||
        resources.isDirty ||
        codeEditorState?.value != null ||
        hookTableEditing,
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
        clearHookTableEditing();
      },
      save: async (ignoreConflict: boolean = false) => {
        await resources.save(ignoreConflict);
        await config.save(ignoreConflict);
        clearHookTableEditing();
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
            produce(prev, (prev) => {
              switch (eventKind) {
                case "blocking": {
                  const h = state.blocking_handlers[index];
                  const path = getDenoScriptPathFromURL(h.url);
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
                  const path = getDenoScriptPathFromURL(h.url);
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
        setActiveTab("signing-secret");
        window.location.hash = "";
        window.location.hash = "#" + WEBHOOK_SIGNATURE_ID;
      }
    });

    const [revealed, setRevealed] = useState(
      locationState?.isOAuthRedirect ?? false
    );

    const onTimeoutChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setState((state) => ({
          ...state,
          timeout: parseIntegerAllowLeadingZeros(e.target.value),
        }));
      },
      [setState]
    );

    const onTotalTimeoutChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setState((state) => ({
          ...state,
          totalTimeout: parseIntegerAllowLeadingZeros(e.target.value),
        }));
      },
      [setState]
    );

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
          const path = getDenoScriptPathFromURL(h.url);
          for (const r of state.resources) {
            if (r.path === path && r.nullableValue != null) {
              return r.nullableValue;
            }
          }
          break;
        }
        case "nonblocking": {
          const h = state.non_blocking_handlers[index];
          const path = getDenoScriptPathFromURL(h.url);
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
        setActiveTab("blocking-events");
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
        setActiveTab("non-blocking-events");
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
              if (r.path === getDenoScriptPathFromURL(item.url)) {
                r.nullableValue = makeDefaultDenoHookBlockingScript(
                  item.event as BlockingEvent
                );
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
        name: "",
        events: ["*"],
        kind: "webhook",
        url: "",
        isDirty: false,
      }),
      []
    );

    const onNonBlockingHandlersChange = useCallback(
      (value: NonBlockingEventHandler[]) => {
        setState((state) =>
          produce(state, (state) => {
            const newValue = value.map((h) => {
              return {
                name: h.name || undefined,
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

    const onNonBlockingHandlersChangeItemChange = useCallback(
      (
        value: NonBlockingEventHandler[],
        _index: number,
        _item: NonBlockingEventHandler
      ) => {
        setState((state) =>
          produce(state, (state) => {
            const newValue = value.map((h) => {
              return {
                name: h.name || undefined,
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

    const onRevealSecret = useCallback(() => {
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
    }, [navigate, state.secret]);

    const isSecretMasked = !revealed || state.secret == null;
    const secretKeyValue = isSecretMasked ? MASKED_SECRET : (state.secret ?? "");

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
      const out: BlockingEventHandler[] = [];
      for (const c of cfgs) {
        out.push({
          ...c,
          kind: getHookKind(c.url),
          isDirty: checkDirty(diff, c.url),
        });
      }
      return out;
    }, [state.diff, state.blocking_handlers]);

    const nonBlockingHandlers: NonBlockingEventHandler[] = useMemo(() => {
      const diff = state.diff;
      const cfgs = state.non_blocking_handlers;
      const out: NonBlockingEventHandler[] = [];
      for (const c of cfgs) {
        out.push({
          ...c,
          name: c.name ?? "",
          kind: getHookKind(c.url),
          isDirty: checkDirty(diff, c.url),
        });
      }
      return out;
    }, [state.diff, state.non_blocking_handlers]);

    const contentWidthAnchorRef = React.useRef<HTMLDivElement>(null);

    return (
      <FormContainer
        form={form}
        hideFooterComponent={true}
      >
        <HookScreenWithSaveBar
          codeEditorState={codeEditorState}
          anchorRef={contentWidthAnchorRef}
        >
          {codeEditorState != null ? (
            <div className={cn(styles.codeEditorContainer)}>
              <WidgetTitle>
                <FormattedMessage id="HookConfigurationScreen.edit-hook.label" />
              </WidgetTitle>
              <WidgetDescription>
                <FormattedMessage
                  id="HookConfigurationScreen.edit-hook.description"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    docLink: (chunks: React.ReactNode) => (
                      <ExternalLink
                        href={
                          codeEditorState.eventKind === "blocking"
                            ? "https://docs.authgear.com/customization/events-hooks/blocking-events"
                            : "https://docs.authgear.com/customization/events-hooks/non-blocking-events"
                        }
                      >
                        {chunks}
                      </ExternalLink>
                    ),
                  }}
                />
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
              <div
                ref={contentWidthAnchorRef}
                className={cn(styles.widget, styles.pageHeader)}
              >
                <h1 className={styles.pageTitle}>
                  <FormattedMessage id="HookConfigurationScreen.title" />
                </h1>
                <RadixText
                  as="p"
                  size="2"
                  color="gray"
                  className={styles.pageDescription}
                >
                  <FormattedMessage id="HookConfigurationScreen.description" />
                </RadixText>
              </div>

              <Tabs.Root
                className={styles.tabsRoot}
                value={activeTab}
                onValueChange={setActiveTab}
              >
                <Tabs.List className={styles.tabsList}>
                  <Tabs.Trigger value="blocking-events">
                    <FormattedMessage id="HookConfigurationScreen.blocking-events" />
                  </Tabs.Trigger>
                  <Tabs.Trigger value="non-blocking-events">
                    <FormattedMessage id="HookConfigurationScreen.non-blocking-events" />
                  </Tabs.Trigger>
                  <Tabs.Trigger value="settings">
                    <FormattedMessage id="HookConfigurationScreen.settings" />
                  </Tabs.Trigger>
                  <Tabs.Trigger value="signing-secret">
                    <FormattedMessage id="HookConfigurationScreen.signing-secret" />
                  </Tabs.Trigger>
                </Tabs.List>

                <Tabs.Content value="blocking-events" className={styles.tabContent}>
                  <section className={styles.section}>
                    <div className={styles.sectionInner}>
                      <RadixText
                        as="p"
                        size="3"
                        weight="medium"
                        className={styles.sectionHeading}
                      >
                        <FormattedMessage id="HookConfigurationScreen.blocking-handlers.label" />
                      </RadixText>
                      <div className={styles.sectionContent}>
                        <RadixText
                          as="p"
                          size="1"
                          color="gray"
                          className={styles.sectionDescription}
                        >
                          <FormattedMessage
                            id="HookConfigurationScreen.blocking-events.description"
                            values={{
                              // eslint-disable-next-line react/no-unstable-nested-components
                              docLink: (chunks: React.ReactNode) => (
                                <ExternalLink href="https://docs.authgear.com/customization/events-hooks/blocking-events">
                                  {chunks}
                                </ExternalLink>
                              ),
                            }}
                          />
                        </RadixText>
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
                          <BlockingHooksTable
                            handlers={blockingHandlers}
                            onHandlersChange={onBlockingHandlersChange}
                            onHandlerItemChange={onBlockingHandlersChangeItemChange}
                            makeDefaultHandler={makeDefaultHandler}
                            onEditDeno={onEditBlocking}
                            addDisabled={blockingHandlerLimitReached}
                            onEditingChange={setBlockingTableEditing}
                          />
                        ) : null}
                      </div>
                    </div>
                  </section>
                </Tabs.Content>

                <Tabs.Content value="non-blocking-events" className={styles.tabContent}>
                  <section className={styles.section}>
                    <div className={styles.sectionInner}>
                      <RadixText
                        as="p"
                        size="3"
                        weight="medium"
                        className={styles.sectionHeading}
                      >
                        <FormattedMessage id="HookConfigurationScreen.non-blocking-handlers.label" />
                      </RadixText>
                      <div className={styles.sectionContent}>
                        <RadixText
                          as="p"
                          size="1"
                          color="gray"
                          className={styles.sectionDescription}
                        >
                          <FormattedMessage
                            id="HookConfigurationScreen.non-blocking-events.description"
                            values={{
                              // eslint-disable-next-line react/no-unstable-nested-components
                              docLink: (chunks: React.ReactNode) => (
                                <ExternalLink href="https://docs.authgear.com/customization/events-hooks/non-blocking-events">
                                  {chunks}
                                </ExternalLink>
                              ),
                            }}
                          />
                        </RadixText>
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
                          <NonBlockingHooksTable
                            handlers={nonBlockingHandlers}
                            onHandlersChange={onNonBlockingHandlersChange}
                            onHandlerItemChange={onNonBlockingHandlersChangeItemChange}
                            makeDefaultHandler={makeDefaultNonBlockingHandler}
                            onEditDeno={onEditNonBlocking}
                            addDisabled={nonBlockingHandlerLimitReached}
                            onEditingChange={setNonBlockingTableEditing}
                          />
                        ) : null}
                      </div>
                    </div>
                  </section>
                </Tabs.Content>

                <Tabs.Content value="settings" className={styles.tabContent}>
                  <section className={styles.section}>
                    <div className={styles.sectionInner}>
                      <RadixText
                        as="p"
                        size="3"
                        weight="medium"
                        className={styles.sectionHeading}
                      >
                        <FormattedMessage id="HookConfigurationScreen.hook-settings" />
                      </RadixText>
                      <div className={styles.sectionContent}>
                        <RadixTextField
                          size="2"
                          labelSize="2"
                          type="text"
                          label={
                            <FormattedMessage id="HookConfigurationScreen.total-timeout.label" />
                          }
                          value={state.totalTimeout?.toFixed(0) ?? ""}
                          onChange={onTotalTimeoutChange}
                        />
                        <RadixTextField
                          size="2"
                          labelSize="2"
                          type="text"
                          label={
                            <FormattedMessage id="HookConfigurationScreen.timeout.label" />
                          }
                          value={state.timeout?.toFixed(0) ?? ""}
                          onChange={onTimeoutChange}
                        />
                      </div>
                    </div>
                  </section>
                </Tabs.Content>

                <Tabs.Content value="signing-secret" className={styles.tabContent}>
                  <section className={styles.section}>
                    <div className={styles.sectionInner}>
                      <RadixText
                        as="p"
                        size="3"
                        weight="medium"
                        className={styles.sectionHeading}
                      >
                        <span id={WEBHOOK_SIGNATURE_ID}>
                          <FormattedMessage id="HookConfigurationScreen.signature.title" />
                        </span>
                      </RadixText>
                      <div className={styles.sectionContent}>
                        <RadixTextField
                          size="2"
                          labelSize="2"
                          type="text"
                          label={
                            <FormattedMessage id="HookConfigurationScreen.signature.secret-key" />
                          }
                          value={secretKeyValue}
                          readOnly={true}
                          suffixPlain={true}
                          suffix={
                            isSecretMasked ? (
                              <RevealIconButton onClick={onRevealSecret} />
                            ) : secretKeyValue.length > 0 ? (
                              <CopyIconButton textToCopy={secretKeyValue} />
                            ) : undefined
                          }
                          hint={
                            <FormattedMessage
                              id="HookConfigurationScreen.signature.description"
                              values={{
                                // eslint-disable-next-line react/no-unstable-nested-components
                                docLink: (chunks: React.ReactNode) => (
                                  <ExternalLink href="https://docs.authgear.com/customization/events-hooks/webhooks#verifying-signature">
                                    {chunks}
                                  </ExternalLink>
                                ),
                              }}
                            />
                          }
                        />
                      </div>
                    </div>
                  </section>
                </Tabs.Content>
              </Tabs.Root>
            </>
          )}
        </HookScreenWithSaveBar>
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

  if (featureConfig.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  if (featureConfig.loadError) {
    return (
      <ShowError
        error={featureConfig.loadError}
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
