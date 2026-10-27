import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import deepEqual from "deep-equal";
import { Button, Flex, Heading, IconButton, Text } from "@radix-ui/themes";
import { Pencil1Icon } from "@radix-ui/react-icons";
import { useNavigate, useParams } from "react-router-dom";
import Link from "../../Link";
import { Context, FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { Callout, ToastCallout } from "../../components/v2/Callout/Callout";
import { PersistentToast } from "../../components/v2/Toast/Toast";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import {
  FormContainerBase,
  useFormContainerBaseContext,
} from "../../FormContainerBase";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useAuthFlowsScreenQueryQuery } from "./query/authFlowsScreenQuery";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowStepConfig,
  AuthenticationFlowsConfig,
  PortalAPIAppConfig,
} from "../../types";
import {
  FlowType,
  GeneratedFlows,
  accountRecoveryDestinations,
  flowListPath,
  flowPath,
  generatedFlowsConfig,
  isFlowType,
  listFlowNames,
  DEFAULT_FLOW_NAME,
  resolveFlow,
} from "./authflows/model";
import { NewFlows } from "./authflows/list/starters";
import { FlowNameDialog } from "./authflows/list/FlowNameDialog";
import { renameNotice } from "./authflows/list/usage";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useLeavePageWarning } from "../../hook/useLeavePageWarning";
import { flowTypeMessageIDs, optionLabel } from "./authflows/labels";
import { formFieldLabel } from "./authflows/sketches";
import { FlowCanvas } from "./authflows/FlowCanvas";
import { FlowStepItem, flowToTree } from "./authflows/tree";
import { findStep } from "./authflows/findStep";
import { StepDetailsPanel } from "./authflows/StepDetailsPanel";
import {
  FlowSketchContext,
  FlowSketchContextValue,
} from "./authflows/FlowSketchContext";
import {
  FlowEditorContext,
  StepProblem,
  StepProblemsContext,
} from "./authflows/edit/FlowEditorContext";
import { ConflictDialog } from "./authflows/edit/ConflictDialog";
import {
  StoredDraft,
  readStoredDraft,
  removeStoredDraft,
  storeDraft,
} from "./authflows/edit/draftStorage";
import { useFlowEditorValue } from "./authflows/edit/useFlowEditorValue";
import {
  EditResult,
  countNestedSteps,
  countOptionSteps,
  insertStep,
  removeOption,
  removeStep,
} from "./authflows/edit/edits";
import { getAt } from "./authflows/edit/pointer";
import {
  FlowCheck,
  StepCheck,
  checkAccountStatusAt,
  fillablePointers,
  flowChecks,
  knownPointers,
  stepChecks,
} from "./authflows/edit/checks";
import { describeCause, flowErrorRule } from "./authflows/edit/causes";
import { ValidationFailedErrorInfoCause } from "../../error/validation";
import {
  EditorDraft,
  EditorState,
  flowPointer,
  rawFlow,
  removeFlows,
  renameFlow,
  replaceFlows,
  stepErrors,
} from "./authflows/edit/save";
import styles from "./AuthFlowsScreen.module.css";

type EditorForm = AppConfigFormModel<EditorState>;

function constructFormState(): EditorState {
  return { drafts: [], removals: [] };
}

function constructConfig(
  raw: PortalAPIAppConfig,
  _initial: EditorState,
  current: EditorState
): PortalAPIAppConfig {
  const next = removeFlows(replaceFlows(raw, current.drafts), current.removals);
  const { rename } = current;
  return rename != null
    ? renameFlow(next, rename.type, rename.name, rename.newName)
    : next;
}

const checkMessageIDs: Record<StepCheck["kind"], string> = {
  form_before_identify: "AuthFlowsScreen.check.form-before-identify",
  form_before_authentication:
    "AuthFlowsScreen.check.form-before-authentication",
  field_not_fillable: "AuthFlowsScreen.error.not-fillable",
  field_duplicate: "AuthFlowsScreen.error.duplicate-field",
};

const optionCheckMessageIDs: Record<
  "field_not_fillable" | "field_duplicate",
  string
> = {
  field_not_fillable: "AuthFlowsScreen.check.option-field-not-fillable",
  field_duplicate: "AuthFlowsScreen.check.option-field-duplicate",
};

// A save error and the editor's check on it say the same thing once.
function uniqueProblems<T extends StepProblem>(problems: T[]): T[] {
  return problems.filter(
    (p, i) => problems.findIndex((q) => q.text === p.text) === i
  );
}

const flowCheckMessageIDs: Record<FlowCheck, string> = {
  no_check_account_status: "AuthFlowsScreen.check.no-check-account-status",
};

function allStepIDs(tree: FlowStepItem[]): string[] {
  return tree.flatMap((item) => [
    item.id,
    ...(item.branches ?? []).flatMap((b) => allStepIDs(b.sequence)),
  ]);
}

const AuthFlowsScreen: React.VFC = function AuthFlowsScreen() {
  const { appID } = useParams() as { appID: string };
  // Settings edited elsewhere in the portal change the generated flows, and
  // nothing refetches them, so a cached copy here would be stale.
  const { data, loading, error, refetch } = useAuthFlowsScreenQueryQuery({
    variables: { id: appID },
    fetchPolicy: "network-only",
  });
  const form = useAppConfigForm<EditorState>({
    appID,
    constructFormState,
    constructConfig,
  });
  const app = data?.node?.__typename === "App" ? data.node : null;
  const generatedFlows: GeneratedFlows | undefined =
    app?.generatedAuthenticationFlows;
  const generated = useMemo(
    () =>
      generatedFlows != null ? generatedFlowsConfig(generatedFlows) : undefined,
    [generatedFlows]
  );
  // While the shown flow is being edited, save errors inside it are shown on
  // its steps, and the error bar only counts them. A rename or reset has no
  // draft to show them on, so the bar shows them as usual.
  const params = useParams() as { type?: string; name?: string };
  const { rawConfig } = form;
  const editingRoute = form.state.drafts.some(
    (d) => d.type === params.type && d.name === params.name
  );
  const errorRules = useMemo(
    () =>
      rawConfig != null && isFlowType(params.type) && editingRoute
        ? [
            flowErrorRule(
              flowPointer(rawConfig, params.type, params.name ?? "")
            ),
          ]
        : undefined,
    [rawConfig, params.type, params.name, editingRoute]
  );

  if (loading || form.isLoading) {
    return <ShowLoading />;
  }
  if (error != null) {
    return (
      <ShowError
        error={error}
        onRetry={() => {
          refetch();
        }}
      />
    );
  }
  if (form.loadError != null) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }
  return (
    <FormContainerBase
      form={form}
      ownConflictDialog={true}
      errorRules={errorRules}
    >
      <AuthFlowsContent appID={appID} form={form} generated={generated} />
    </FormContainerBase>
  );
};

interface PendingRemove {
  stepID: string;
  // Set when removing one option of the step, not the step.
  optionIndex: number | null;
  nested: number;
}

function AuthFlowsContent({
  appID,
  form,
  generated,
}: {
  appID: string;
  form: EditorForm;
  generated: AuthenticationFlowsConfig | undefined;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { canSave, isUpdating } = useFormContainerBaseContext();
  const params = useParams() as { type?: string; name?: string };
  const navigate = useNavigate();

  const appConfig = form.effectiveConfig;
  const custom = appConfig.authentication_flow;
  const drafts = form.state.drafts;

  const routeType: FlowType | null = isFlowType(params.type)
    ? params.type
    : null;
  const name = params.name ?? "";
  const shownType = routeType;
  const selection = useMemo(
    () => (shownType != null ? { type: shownType, name } : null),
    [shownType, name]
  );

  const draft =
    drafts.find((d) => d.type === shownType && d.name === name) ?? null;
  // A custom flow, or for default with none, the generated one.
  const resolved = useMemo(
    () =>
      generated != null && selection != null
        ? resolveFlow(custom, generated, selection.type, selection.name)
        : null,
    [custom, generated, selection]
  );
  const shownFlow: AuthenticationFlowConfig | null =
    draft?.flow ?? resolved?.flow ?? null;
  // An edited default can go back to the one built from Login methods.
  const canReset = resolved?.source === "custom" && name === DEFAULT_FLOW_NAME;

  // A flow created in the New flow dialog arrives as unsaved drafts, once:
  // the state is popped from history, so a reload or Back cannot bring the
  // starter back over later edits.
  const { setState } = form;
  const incoming = useLocationEffect<{ newFlows?: NewFlows } | null>(
    (state) => {
      const newFlows = state?.newFlows;
      if (newFlows != null) {
        setState(() => ({
          drafts: newFlows.map(
            ({ type, flow }): EditorDraft => ({ type, name: flow.name, flow })
          ),
          removals: [],
        }));
      }
    }
  );
  const arriving = incoming?.newFlows != null && drafts.length === 0;

  const flowKey = `${shownType ?? ""}/${name}`;
  const tree = useMemo(
    () => (shownFlow != null ? flowToTree(shownFlow) : null),
    [shownFlow]
  );
  // Remember the selection per flow, so switching flows clears it.
  const [selected, setSelected] = useState<{
    flowKey: string;
    stepID: string;
    notice?: string;
  } | null>(null);
  const selectedStepID = selected?.flowKey === flowKey ? selected.stepID : null;
  const onSelectStep = useCallback(
    (stepID: string | null) =>
      setSelected(stepID != null ? { flowKey, stepID } : null),
    [flowKey]
  );
  const selectedItem = useMemo(
    () =>
      tree != null && selectedStepID != null
        ? findStep(tree, selectedStepID)
        : null,
    [tree, selectedStepID]
  );
  const onClosePanel = useCallback(() => setSelected(null), []);

  const sketchContext = useMemo<FlowSketchContextValue | null>(
    () =>
      generated != null
        ? {
            flowType: shownType ?? "signup_login",
            oauthProviders: appConfig.identity?.oauth?.providers ?? [],
            customAttributes:
              appConfig.user_profile?.custom_attributes?.attributes ?? [],
            phoneOTPMode: appConfig.authenticator?.oob_otp?.sms?.phone_otp_mode,
            recoveryDestinations:
              shownFlow != null ? accountRecoveryDestinations(shownFlow) : [],
            custom,
            generated,
            flowLink: (ref) => flowPath(appID, ref.type, ref.name),
          }
        : null,
    [shownType, appConfig, custom, generated, shownFlow, appID]
  );

  // Editing.
  const apply = useCallback(
    (result: EditResult, notice?: string) => {
      setState((s) => ({
        ...s,
        drafts: s.drafts.map((d) =>
          d.type === shownType && d.name === name
            ? { ...d, flow: result.flow }
            : d
        ),
      }));
      setSelected(
        result.selected != null
          ? { flowKey, stepID: result.selected, notice }
          : null
      );
    },
    [setState, flowKey, shownType, name]
  );
  const [pendingRemove, setPendingRemove] = useState<PendingRemove | null>(
    null
  );
  const requestRemove = useCallback(
    (stepID: string) => {
      if (draft == null) {
        return;
      }
      const nested = countNestedSteps(getAt(draft.flow, stepID) as never);
      if (nested > 0) {
        setPendingRemove({ stepID, optionIndex: null, nested });
      } else {
        apply(removeStep(draft.flow, stepID));
      }
    },
    [draft, apply]
  );
  const requestRemoveOption = useCallback(
    (stepID: string, optionIndex: number) => {
      if (draft == null) {
        return;
      }
      // What goes with the option: its nested steps and theirs.
      const nested = countOptionSteps(
        getAt(draft.flow, `${stepID}/one_of/${optionIndex}`) as never
      );
      if (nested > 0) {
        setPendingRemove({ stepID, optionIndex, nested });
      } else {
        apply(removeOption(draft.flow, stepID, optionIndex));
      }
    },
    [draft, apply]
  );
  const confirmRemove = useCallback(() => {
    if (draft != null && pendingRemove != null) {
      apply(
        pendingRemove.optionIndex != null
          ? removeOption(
              draft.flow,
              pendingRemove.stepID,
              pendingRemove.optionIndex
            )
          : removeStep(draft.flow, pendingRemove.stepID)
      );
    }
    setPendingRemove(null);
  }, [draft, pendingRemove, apply]);

  const errorsByStep = useMemo(
    () =>
      draft != null && tree != null && form.rawConfig != null
        ? stepErrors(
            form.updateError,
            flowPointer(form.rawConfig, draft.type, draft.name),
            allStepIDs(tree)
          )
        : new Map<string, ValidationFailedErrorInfoCause[]>(),
    [draft, tree, form.rawConfig, form.updateError]
  );
  const errorStepIDs = useMemo(
    () => new Set<string>([...errorsByStep.keys()].filter((k) => k !== "")),
    [errorsByStep]
  );
  const errorPrefix =
    draft != null && form.rawConfig != null
      ? flowPointer(form.rawConfig, draft.type, draft.name)
      : "";
  const causeText = useCallback(
    (cause: ValidationFailedErrorInfoCause, stepID: string) => {
      const m = describeCause(cause, `${errorPrefix}${stepID}`);
      return renderToString(m.messageID, m.values);
    },
    [errorPrefix, renderToString]
  );

  // The editor's own checks: forms the server would reject, and warnings.
  const fillable = useMemo(() => fillablePointers(appConfig), [appConfig]);
  const checksByStep = useMemo(
    () =>
      shownType != null && shownFlow != null
        ? stepChecks(shownType, shownFlow, new Set(fillable))
        : new Map<string, StepCheck[]>(),
    [shownType, shownFlow, fillable]
  );
  const flowLevelChecks = useMemo(
    () =>
      shownType != null && shownFlow != null
        ? flowChecks(shownType, shownFlow)
        : [],
    [shownType, shownFlow]
  );
  const customAttributes =
    appConfig.user_profile?.custom_attributes?.attributes;
  const checkProblem = useCallback(
    (c: StepCheck, stepID: string): StepProblem => {
      const tone =
        c.kind === "form_before_authentication" ? "warning" : "error";
      // An option's form has no fields in the panel, so say which option
      // and attribute.
      if (
        (c.kind === "field_not_fillable" || c.kind === "field_duplicate") &&
        c.option != null &&
        shownFlow != null
      ) {
        const step = getAt(shownFlow, stepID) as
          | AuthenticationFlowStepConfig
          | undefined;
        const option = step?.one_of?.[c.option];
        if (option != null) {
          return {
            tone,
            text: renderToString(optionCheckMessageIDs[c.kind], {
              option: optionLabel(option, renderToString),
              attribute: formFieldLabel(
                c.pointer,
                customAttributes ?? [],
                renderToString
              ),
            }),
          };
        }
      }
      return { tone, text: renderToString(checkMessageIDs[c.kind]) };
    },
    [renderToString, shownFlow, customAttributes]
  );
  // Each step's first problem, for its card.
  const problems = useMemo(() => {
    const out = new Map<string, StepProblem>();
    for (const [id, causes] of errorsByStep) {
      if (id !== "" && causes.length > 0) {
        out.set(id, { tone: "error", text: causeText(causes[0], id) });
      }
    }
    for (const [id, checks] of checksByStep) {
      if (!out.has(id) && checks.length > 0) {
        out.set(id, checkProblem(checks[0], id));
      }
    }
    return out;
  }, [errorsByStep, checksByStep, causeText, checkProblem]);
  // The selected step's problems, for the panel. Field checks are marked: a
  // form step's fields show their own while editing.
  const selectedProblems = useMemo(() => {
    if (selectedItem == null) {
      return [];
    }
    const id = selectedItem.id;
    return uniqueProblems([
      ...(errorsByStep.get(id) ?? []).map((c) => ({
        tone: "error" as const,
        text: causeText(c, id),
        fieldCheck: false,
      })),
      ...(checksByStep.get(id) ?? []).map((c) => ({
        ...checkProblem(c, id),
        fieldCheck:
          selectedItem.step.type === "fill_form" &&
          (c.kind === "field_not_fillable" || c.kind === "field_duplicate"),
      })),
    ]);
  }, [selectedItem, errorsByStep, checksByStep, causeText, checkProblem]);
  // Why the last dropped card did not move; cleared by the next edit.
  const [dropNotice, setDropNotice] = useState<string | null>(null);
  // The flow whose flow-level warning was closed; shown again for another.
  const [closedFlowCheckFor, setClosedFlowCheckFor] = useState<string | null>(
    null
  );
  const applyAndClearNotice = useCallback(
    (result: EditResult, notice?: string) => {
      setDropNotice(null);
      apply(result, notice);
    },
    [apply]
  );
  const editorValue = useFlowEditorValue({
    flowType: shownType ?? "signup_login",
    flow: shownFlow ?? { name },
    errorStepIDs,
    problems,
    fillablePointers: fillable,
    apply: applyAndClearNotice,
    requestRemove,
    notify: setDropNotice,
  });
  const { hasErrors, problem } = editorValue;
  const stepProblemsValue = useMemo(
    () => ({ hasErrors, problem }),
    [hasErrors, problem]
  );

  // The flow as authgear.yaml says it, not after defaults; a generated
  // default is edited straight away and becomes custom when saved.
  const onEdit = useCallback(() => {
    if (routeType == null) {
      return;
    }
    const flow =
      (form.rawConfig != null
        ? rawFlow(form.rawConfig, routeType, name)
        : null) ?? resolved?.flow;
    if (flow != null) {
      setState(() => ({
        drafts: [{ type: routeType, name, flow }],
        removals: [],
      }));
    }
  }, [routeType, form.rawConfig, name, resolved, setState]);
  // A refresh or a closed tab asks first; in-app navigation is asked about
  // by FormContainerBase.
  const { getIsDirty, save, saveWith, reload, reset } = form;
  useLeavePageWarning(getIsDirty);

  // An unsaved draft kept in this browser, offered back on return. Read once
  // per flow; written as the draft changes; dropped once saved or discarded.
  const storedKey = `${appID}/${flowKey}`;
  const readStored = () =>
    routeType != null ? readStoredDraft(appID, routeType, name) : null;
  const [stored, setStored] = useState<{
    key: string;
    draft: StoredDraft | null;
  }>(() => ({ key: storedKey, draft: readStored() }));
  if (stored.key !== storedKey) {
    setStored({ key: storedKey, draft: readStored() });
  }
  const forgetStored = useCallback(
    (flows: { type: FlowType; name: string }[]) => {
      for (const f of flows) {
        removeStoredDraft(appID, f.type, f.name);
      }
      setStored({ key: storedKey, draft: null });
    },
    [appID, storedKey]
  );
  // What the draft started from: an unchanged draft is not worth keeping.
  const baseline =
    (form.rawConfig != null && routeType != null
      ? rawFlow(form.rawConfig, routeType, name)
      : null) ??
    resolved?.flow ??
    null;
  useEffect(() => {
    if (routeType == null || draft == null) {
      return;
    }
    if (baseline != null && deepEqual(baseline, draft.flow, { strict: true })) {
      return;
    }
    storeDraft(appID, routeType, name, draft.flow, Date.now());
  }, [appID, routeType, name, draft, baseline]);
  // Leaving through "Discard and leave" resets the form before the next
  // page shows; the drafts it threw away are not offered back later. A
  // refresh or closed tab runs no cleanup, so its draft is kept.
  const leftDrafts = useRef<{ type: FlowType; name: string }[]>([]);
  const getIsDirtyRef = useRef(getIsDirty);
  useEffect(() => {
    if (drafts.length > 0) {
      leftDrafts.current = drafts.map((d) => ({ type: d.type, name: d.name }));
    }
    getIsDirtyRef.current = getIsDirty;
  }, [drafts, getIsDirty]);
  useEffect(() => {
    return () => {
      if (!getIsDirtyRef.current()) {
        for (const d of leftDrafts.current) {
          removeStoredDraft(appID, d.type, d.name);
        }
      }
      leftDrafts.current = [];
    };
  }, [appID, flowKey]);
  const onRestoreStored = useCallback(() => {
    const flow = stored.draft?.flow;
    if (routeType != null && flow != null) {
      setState(() => ({
        drafts: [{ type: routeType, name, flow }],
        removals: [],
      }));
    }
  }, [stored, routeType, name, setState]);

  // The last save tried, so Overwrite in the conflict dialog can try it
  // again without the checksum check.
  const lastSave = useRef<((ignoreConflict: boolean) => Promise<void>) | null>(
    null
  );
  const attemptSave = useCallback(
    (op: (ignoreConflict: boolean) => Promise<void>) => {
      lastSave.current = op;
      // A failure shows in the error bar, or the conflict dialog.
      op(false).catch(() => {});
    },
    []
  );
  const onSaveDrafts = useCallback(() => {
    const saving = drafts;
    attemptSave(async (ignoreConflict) => {
      await save(ignoreConflict);
      // save() can return without saving, e.g. while another save runs.
      if (!getIsDirty()) {
        forgetStored(saving);
      }
    });
  }, [drafts, attemptSave, save, getIsDirty, forgetStored]);
  // The latest config, with the flow being edited kept on top, to check and
  // save again.
  const onReloadKeepingDraft = useCallback(() => {
    const keep = form.state;
    reset();
    setState(() => keep);
    reload();
  }, [form.state, reset, setState, reload]);
  const onOverwrite = useCallback(() => {
    lastSave.current?.(true).catch(() => {});
  }, []);

  const [confirmReset, setConfirmReset] = useState(false);
  const onReset = useCallback(() => {
    setConfirmReset(false);
    if (routeType == null) {
      return;
    }
    attemptSave(async (ignoreConflict) => {
      await saveWith(
        () => ({ drafts: [], removals: [{ type: routeType, name }] }),
        ignoreConflict
      );
      forgetStored([{ type: routeType, name }]);
    });
  }, [attemptSave, saveWith, forgetStored, routeType, name]);
  // Renaming saves at once, like on the list, then moves to the new name's
  // page; until then the old page would find no flow.
  const [renaming, setRenaming] = useState<"dialog" | "saving" | null>(null);
  const onRename = useCallback(
    (newName: string) => {
      if (routeType == null) {
        return;
      }
      attemptSave(async (ignoreConflict) => {
        setRenaming("saving");
        try {
          await saveWith(
            () => ({
              drafts: [],
              removals: [],
              rename: { type: routeType, name, newName },
            }),
            ignoreConflict
          );
        } catch (e: unknown) {
          // The failed rename is not left pending; Overwrite retries it.
          setState(() => constructFormState());
          setRenaming(null);
          throw e;
        }
        navigate(flowPath(appID, routeType, newName), { replace: true });
        setRenaming(null);
      });
    },
    [attemptSave, saveWith, setState, navigate, appID, routeType, name]
  );
  // The default flow keeps its name: AuthUI runs it by that name. A rename
  // saves straight away, so it waits for pending edits to be settled.
  const renameDisabledReason =
    name === DEFAULT_FLOW_NAME
      ? renderToString("AuthFlowsScreen.rename.disabled.default")
      : draft != null
      ? renderToString("AuthFlowsScreen.rename.disabled.editing")
      : null;

  // Discarding a flow that was never saved leaves its page: there is
  // nothing behind the draft. Navigating waits for the reset to render, so
  // the unsaved-changes prompt does not stop it.
  const [leaving, setLeaving] = useState(false);
  const onDiscard = useCallback(() => {
    const written = drafts.some(
      (d) =>
        form.rawConfig != null &&
        rawFlow(form.rawConfig, d.type, d.name) != null
    );
    forgetStored(drafts);
    reset();
    setSelected(null);
    if (!written && resolved == null) {
      setLeaving(true);
    }
  }, [drafts, form.rawConfig, forgetStored, reset, resolved]);
  useEffect(() => {
    if (leaving && drafts.length === 0) {
      navigate(flowListPath(appID));
    }
  }, [leaving, drafts.length, navigate, appID]);

  const editing = draft != null;
  // Edits made while a save is in flight would be lost when it lands.
  const canEdit = editing && !isUpdating;
  const flowTypeName =
    routeType != null ? renderToString(flowTypeMessageIDs[routeType]) : "";
  const listPath = flowListPath(appID);

  return (
    <div className={styles.root}>
      <header className={styles.header}>
        <Link to={listPath} className={styles.back}>
          <FormattedMessage id="AuthFlowsScreen.back" />
        </Link>
        <Flex gap="3" align="center" wrap="wrap" className={styles.toolbar}>
          <Heading size="6" as="h1">
            {name}
          </Heading>
          {resolved != null ? (
            <Tooltip
              content={
                renameDisabledReason ??
                renderToString("AuthFlowsScreen.list.rename")
              }
            >
              <span>
                <IconButton
                  type="button"
                  variant="ghost"
                  color="gray"
                  aria-label={renderToString("AuthFlowsScreen.list.rename")}
                  disabled={renameDisabledReason != null || isUpdating}
                  onClick={() => setRenaming("dialog")}
                >
                  <Pencil1Icon />
                </IconButton>
              </span>
            </Tooltip>
          ) : null}
          <Text size="2" color="gray">
            {flowTypeName}
          </Text>
          <span className={styles.toolbarSpacer} />
          {editing ? (
            <>
              <Button
                type="button"
                variant="soft"
                color="gray"
                disabled={isUpdating}
                onClick={onDiscard}
              >
                <FormattedMessage id="AuthFlowsScreen.edit.discard" />
              </Button>
              <Button
                type="button"
                disabled={!canSave}
                loading={isUpdating}
                onClick={onSaveDrafts}
              >
                <FormattedMessage id="AuthFlowsScreen.edit.save" />
              </Button>
            </>
          ) : resolved != null ? (
            <>
              {canReset ? (
                <Button
                  type="button"
                  variant="soft"
                  color="gray"
                  disabled={isUpdating}
                  onClick={() => setConfirmReset(true)}
                >
                  <FormattedMessage id="AuthFlowsScreen.reset" />
                </Button>
              ) : null}
              <Button type="button" variant="soft" onClick={onEdit}>
                <FormattedMessage id="AuthFlowsScreen.edit.edit" />
              </Button>
            </>
          ) : null}
        </Flex>
      </header>
      <FormErrorMessageBar />
      {!editing && !arriving && stored.draft != null ? (
        <div className={styles.storedDraft} role="status">
          <Text size="2" className={styles.storedDraftText}>
            <FormattedMessage id="AuthFlowsScreen.stored-draft.message" />
          </Text>
          <Button
            type="button"
            size="1"
            variant="soft"
            color="gray"
            onClick={() =>
              forgetStored(routeType != null ? [{ type: routeType, name }] : [])
            }
          >
            <FormattedMessage id="AuthFlowsScreen.edit.discard" />
          </Button>
          <Button type="button" size="1" onClick={onRestoreStored}>
            <FormattedMessage id="AuthFlowsScreen.stored-draft.restore" />
          </Button>
        </div>
      ) : null}
      {(errorsByStep.get("") ?? []).length > 0 ? (
        <div className={styles.flowProblems}>
          {(errorsByStep.get("") ?? []).map((c, i) => (
            <Callout
              key={`e${i}`}
              type="error"
              size="1"
              text={causeText(c, "")}
            />
          ))}
        </div>
      ) : null}
      {/* Only while editing, where it can be acted on. */}
      {draft != null
        ? flowLevelChecks.map((c) => (
            <PersistentToast
              key={c}
              open={closedFlowCheckFor !== flowKey}
              onOpenChange={(open) => {
                if (!open) {
                  setClosedFlowCheckFor(flowKey);
                }
              }}
            >
              <ToastCallout
                type="warning"
                text={
                  <Flex direction="column" align="start" gap="2">
                    {renderToString(flowCheckMessageIDs[c])}
                    {/* The one flow check; it can add the missing step. */}
                    <Button
                      type="button"
                      disabled={!canEdit}
                      size="1"
                      variant="soft"
                      onClick={() =>
                        applyAndClearNotice(
                          insertStep(
                            draft.flow,
                            checkAccountStatusAt(draft.flow),
                            {
                              type: "check_account_status",
                            }
                          )
                        )
                      }
                    >
                      <FormattedMessage id="AuthFlowsScreen.check.add-check-account-status" />
                    </Button>
                  </Flex>
                }
              />
            </PersistentToast>
          ))
        : null}
      <div className={styles.canvasArea}>
        {canEdit && dropNotice != null ? (
          <div className={styles.dropNotice}>
            <Callout type="warning" size="1" text={dropNotice} />
          </div>
        ) : null}
        {arriving || renaming === "saving" ? null : tree != null &&
          sketchContext != null ? (
          <FlowSketchContext.Provider value={sketchContext}>
            <StepProblemsContext.Provider value={stepProblemsValue}>
              <FlowEditorContext.Provider value={canEdit ? editorValue : null}>
                <FlowCanvas
                  key={flowKey}
                  tree={tree}
                  selectedStepID={selectedStepID}
                  onSelectStep={onSelectStep}
                />
              </FlowEditorContext.Provider>
            </StepProblemsContext.Provider>
          </FlowSketchContext.Provider>
        ) : (
          <Text as="p" size="2" className={styles.notFound}>
            <FormattedMessage id="AuthFlowsScreen.flow-not-found" />{" "}
            <Link to={listPath}>
              <FormattedMessage id="AuthFlowsScreen.back" />
            </Link>
          </Text>
        )}
        {selectedItem != null ? (
          <StepDetailsPanel
            item={selectedItem}
            phoneOTPMode={appConfig.authenticator?.oob_otp?.sms?.phone_otp_mode}
            onClose={onClosePanel}
            problems={selectedProblems}
            edit={
              draft != null && canEdit
                ? {
                    flowType: draft.type,
                    flow: draft.flow,
                    apply,
                    phoneOTPMode:
                      appConfig.authenticator?.oob_otp?.sms?.phone_otp_mode,
                    customAttributes:
                      appConfig.user_profile?.custom_attributes?.attributes ??
                      [],
                    appID,
                    fillablePointers: fillable,
                    knownPointers: knownPointers(appConfig),
                    loginFlowNames: listFlowNames(custom, "login"),
                    signupFlowNames: listFlowNames(custom, "signup"),
                    requestRemoveOption,
                    removeOptionBlockedReason:
                      editorValue.removeOptionBlockedReason,
                    // A form step's fields show their own problems.
                    problems: selectedProblems.filter((p) => !p.fieldCheck),
                    checks: checksByStep.get(selectedItem.id) ?? [],
                    notice: selected?.notice,
                  }
                : undefined
            }
          />
        ) : null}
      </div>
      {renaming === "dialog" && routeType != null ? (
        <FlowNameDialog
          open={true}
          title={renderToString("AuthFlowsScreen.list.rename.title", { name })}
          confirmText={renderToString("AuthFlowsScreen.list.rename")}
          type={routeType}
          initialName={name}
          flows={form.rawConfig?.authentication_flow}
          notice={renameNotice(renderToString, form.rawConfig, routeType, name)}
          onSubmit={onRename}
          onCancel={() => setRenaming(null)}
        />
      ) : null}
      <ConflictDialog
        onReload={onReloadKeepingDraft}
        onOverwrite={onOverwrite}
      />
      <ConfirmationDialog
        open={confirmReset}
        onOpenChange={setConfirmReset}
        title={
          <FormattedMessage
            id="AuthFlowsScreen.reset.title"
            values={{ name }}
          />
        }
        description={
          <FormattedMessage id="AuthFlowsScreen.reset.description" />
        }
        confirmText={<FormattedMessage id="AuthFlowsScreen.reset" />}
        cancelText={<FormattedMessage id="AuthFlowsScreen.edit.cancel" />}
        onConfirm={onReset}
        onCancel={() => setConfirmReset(false)}
      />
      <ConfirmationDialog
        open={pendingRemove != null}
        onOpenChange={(open) => {
          if (!open) {
            setPendingRemove(null);
          }
        }}
        title={<FormattedMessage id="AuthFlowsScreen.edit.remove.title" />}
        description={
          <FormattedMessage
            id="AuthFlowsScreen.edit.remove.description"
            values={{ count: pendingRemove?.nested ?? 0 }}
          />
        }
        confirmText={<FormattedMessage id="AuthFlowsScreen.edit.remove" />}
        cancelText={<FormattedMessage id="AuthFlowsScreen.edit.cancel" />}
        onConfirm={confirmRemove}
        onCancel={() => setPendingRemove(null)}
      />
    </div>
  );
}

export default AuthFlowsScreen;
