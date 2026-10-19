import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { Button, Flex, Heading, Text } from "@radix-ui/themes";
import { useNavigate, useParams } from "react-router-dom";
import Link from "../../Link";
import { Context, FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { Callout } from "../../components/v2/Callout/Callout";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
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
import { useLocationEffect } from "../../hook/useLocationEffect";
import { flowTypeMessageIDs } from "./authflows/labels";
import { FlowCanvas } from "./authflows/FlowCanvas";
import { FlowStepItem, flowToTree } from "./authflows/tree";
import { findStep } from "./authflows/findStep";
import { StepDetailsPanel } from "./authflows/StepDetailsPanel";
import {
  FlowSketchContext,
  FlowSketchContextValue,
} from "./authflows/FlowSketchContext";
import { FlowEditorContext } from "./authflows/edit/FlowEditorContext";
import { useFlowEditorValue } from "./authflows/edit/useFlowEditorValue";
import {
  EditResult,
  countNestedSteps,
  countOptionSteps,
  removeOption,
  removeStep,
} from "./authflows/edit/edits";
import { getAt } from "./authflows/edit/pointer";
import {
  EditorDraft,
  EditorState,
  flowPointer,
  rawFlow,
  removeFlows,
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
  return removeFlows(replaceFlows(raw, current.drafts), current.removals);
}

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
    <FormContainerBase form={form}>
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
  const { onSave, canSave, isUpdating } = useFormContainerBaseContext();
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
        : new Map(),
    [draft, tree, form.rawConfig, form.updateError]
  );
  const errorStepIDs = useMemo(
    () => new Set<string>(errorsByStep.keys()),
    [errorsByStep]
  );
  // Why the last dropped card did not move; cleared by the next edit.
  const [dropNotice, setDropNotice] = useState<string | null>(null);
  const applyAndClearNotice = useCallback(
    (result: EditResult, notice?: string) => {
      setDropNotice(null);
      apply(result, notice);
    },
    [apply]
  );
  const editorValue = useFlowEditorValue({
    flowType: shownType ?? "signup_login",
    flow: draft?.flow ?? { name },
    errorStepIDs,
    apply: applyAndClearNotice,
    requestRemove,
    notify: setDropNotice,
  });

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
  const [confirmReset, setConfirmReset] = useState(false);
  const { saveWith } = form;
  const onReset = useCallback(() => {
    setConfirmReset(false);
    if (routeType == null) {
      return;
    }
    // A failure shows in the error bar, from the form's updateError.
    saveWith(() => ({
      drafts: [],
      removals: [{ type: routeType, name }],
    })).catch(() => {});
  }, [saveWith, routeType, name]);
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
    form.reset();
    setSelected(null);
    if (!written && resolved == null) {
      setLeaving(true);
    }
  }, [drafts, form, resolved]);
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
                onClick={onSave}
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
      <div className={styles.canvasArea}>
        {canEdit && dropNotice != null ? (
          <div className={styles.dropNotice}>
            <Callout type="warning" size="1" text={dropNotice} />
          </div>
        ) : null}
        {arriving ? null : tree != null && sketchContext != null ? (
          <FlowSketchContext.Provider value={sketchContext}>
            <FlowEditorContext.Provider value={canEdit ? editorValue : null}>
              <FlowCanvas
                key={flowKey}
                tree={tree}
                selectedStepID={selectedStepID}
                onSelectStep={onSelectStep}
              />
            </FlowEditorContext.Provider>
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
                    standardAttributePointers: (
                      appConfig.user_profile?.standard_attributes
                        ?.access_control ?? []
                    ).map((a) => a.pointer),
                    loginFlowNames: listFlowNames(custom, "login"),
                    signupFlowNames: listFlowNames(custom, "signup"),
                    requestRemoveOption,
                    removeOptionBlockedReason:
                      editorValue.removeOptionBlockedReason,
                    causes: errorsByStep.get(selectedItem.id) ?? [],
                    notice: selected?.notice,
                  }
                : undefined
            }
          />
        ) : null}
      </div>
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
