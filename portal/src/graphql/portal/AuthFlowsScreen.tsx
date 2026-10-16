import React, { useCallback, useContext, useMemo, useState } from "react";
import { Button, Flex, Heading, Select, Text } from "@radix-ui/themes";
import { useParams, useSearchParams } from "react-router-dom";
import Link from "../../Link";
import { Context, FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { Badge } from "../../components/v2/Badge/Badge";
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
  DEFAULT_FLOW_NAME,
  FlowSelection,
  GeneratedFlows,
  accountRecoveryDestinations,
  flowSelectionSearch,
  flowTypes,
  generatedFlowsConfig,
  isFlowType,
  listFlowNames,
  parseFlowSelection,
  resolveFlow,
} from "./authflows/model";
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
  EditorState,
  flowPointer,
  rawFlow,
  replaceFlow,
  stepErrors,
} from "./authflows/edit/save";
import styles from "./AuthFlowsScreen.module.css";

type EditorForm = AppConfigFormModel<EditorState>;

function constructFormState(): EditorState {
  return { draft: null };
}

function constructConfig(
  raw: PortalAPIAppConfig,
  _initial: EditorState,
  current: EditorState
): PortalAPIAppConfig {
  const { draft } = current;
  return draft != null
    ? replaceFlow(raw, draft.type, draft.name, draft.flow)
    : raw;
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
  const [searchParams, setSearchParams] = useSearchParams();

  const appConfig = form.effectiveConfig;
  const custom = appConfig.authentication_flow;
  const draft = form.state.draft;

  const urlSelection = useMemo(
    () => parseFlowSelection(searchParams, custom),
    [searchParams, custom]
  );
  // While editing, the draft decides which flow is shown, even if the URL
  // changes under it.
  const selection: FlowSelection = useMemo(
    () =>
      draft != null ? { type: draft.type, name: draft.name } : urlSelection,
    [draft, urlSelection]
  );
  const names = useMemo(
    () => listFlowNames(custom, selection.type),
    [custom, selection.type]
  );
  const resolved = useMemo(
    () =>
      generated != null
        ? resolveFlow(custom, generated, selection.type, selection.name)
        : null,
    [custom, generated, selection]
  );
  const shownFlow: AuthenticationFlowConfig | null =
    draft?.flow ?? resolved?.flow ?? null;

  const flowKey = `${selection.type}/${selection.name}`;
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
            flowType: selection.type,
            oauthProviders: appConfig.identity?.oauth?.providers ?? [],
            customAttributes:
              appConfig.user_profile?.custom_attributes?.attributes ?? [],
            phoneOTPMode: appConfig.authenticator?.oob_otp?.sms?.phone_otp_mode,
            recoveryDestinations:
              shownFlow != null ? accountRecoveryDestinations(shownFlow) : [],
            custom,
            generated,
          }
        : null,
    [selection.type, appConfig, custom, generated, shownFlow]
  );

  // Editing.
  const { setState } = form;
  const apply = useCallback(
    (result: EditResult, notice?: string) => {
      setState((s) =>
        s.draft != null ? { draft: { ...s.draft, flow: result.flow } } : s
      );
      setSelected(
        result.selected != null
          ? { flowKey, stepID: result.selected, notice }
          : null
      );
    },
    [setState, flowKey]
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
    flowType: selection.type,
    flow: draft?.flow ?? { name: selection.name },
    errorStepIDs,
    apply: applyAndClearNotice,
    requestRemove,
    notify: setDropNotice,
  });

  const [confirmGenerated, setConfirmGenerated] = useState(false);
  const startEditing = useCallback(
    (flow: AuthenticationFlowConfig) => {
      setState(() => ({
        draft: { type: selection.type, name: selection.name, flow },
      }));
    },
    [setState, selection]
  );
  const onEdit = useCallback(() => {
    if (resolved == null) {
      return;
    }
    if (resolved.source === "generated") {
      setConfirmGenerated(true);
      return;
    }
    // Edit what authgear.yaml says, not the flow after defaults.
    const raw =
      form.rawConfig != null
        ? rawFlow(form.rawConfig, selection.type, selection.name)
        : null;
    startEditing(raw ?? resolved.flow);
  }, [resolved, form.rawConfig, selection, startEditing]);
  const onConfirmGenerated = useCallback(() => {
    setConfirmGenerated(false);
    if (resolved != null) {
      startEditing({ ...resolved.flow, name: selection.name });
    }
  }, [resolved, selection.name, startEditing]);
  const onDiscard = useCallback(() => {
    form.reset();
    setSelected(null);
  }, [form]);

  const select = useCallback(
    (next: FlowSelection) => {
      setSearchParams(new URLSearchParams(flowSelectionSearch(next)));
    },
    [setSearchParams]
  );
  const onTypeChange = useCallback(
    (value: string) => {
      if (isFlowType(value)) {
        select({ type: value, name: DEFAULT_FLOW_NAME });
      }
    },
    [select]
  );
  const onNameChange = useCallback(
    (value: string) => select({ type: selection.type, name: value }),
    [select, selection.type]
  );

  const editing = draft != null;
  // Edits made while a save is in flight would be lost when it lands.
  const canEdit = editing && !isUpdating;
  const flowTypeName = renderToString(flowTypeMessageIDs[selection.type]);

  return (
    <div className={styles.root}>
      <header className={styles.header}>
        <Heading size="6" as="h1">
          <FormattedMessage id="AuthFlowsScreen.title" />
        </Heading>
        <Text as="p" size="2" color="gray">
          <FormattedMessage id="AuthFlowsScreen.description" />
        </Text>
        <Flex gap="3" align="center" wrap="wrap" className={styles.toolbar}>
          <Select.Root
            value={selection.type}
            onValueChange={onTypeChange}
            disabled={editing}
          >
            <Select.Trigger
              aria-label={renderToString("AuthFlowsScreen.picker.type")}
            />
            <Select.Content>
              {flowTypes.map((t) => (
                <Select.Item key={t} value={t}>
                  {renderToString(flowTypeMessageIDs[t])}
                </Select.Item>
              ))}
            </Select.Content>
          </Select.Root>
          {names.length > 1 ? (
            <Select.Root
              value={selection.name}
              onValueChange={onNameChange}
              disabled={editing}
            >
              <Select.Trigger
                aria-label={renderToString("AuthFlowsScreen.picker.name")}
              />
              <Select.Content>
                {names.map((n) => (
                  <Select.Item key={n} value={n}>
                    {n}
                  </Select.Item>
                ))}
              </Select.Content>
            </Select.Root>
          ) : null}
          {editing ? (
            <Badge
              size="2"
              variant="info"
              text={
                <FormattedMessage
                  id="AuthFlowsScreen.edit.editing"
                  values={{ name: selection.name, type: flowTypeName }}
                />
              }
            />
          ) : null}
          {!editing && resolved?.source === "custom" ? (
            <Badge
              size="2"
              variant="info"
              text={<FormattedMessage id="AuthFlowsScreen.badge.custom" />}
            />
          ) : null}
          {!editing && resolved?.source === "generated" ? (
            <Link
              to={`/project/${appID}/configuration/authentication/login-methods`}
            >
              <Badge
                size="2"
                variant="neutral"
                text={<FormattedMessage id="AuthFlowsScreen.badge.generated" />}
              />
            </Link>
          ) : null}
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
            <Button type="button" variant="soft" onClick={onEdit}>
              <FormattedMessage id="AuthFlowsScreen.edit.edit" />
            </Button>
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
        {tree != null && sketchContext != null ? (
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
            <FormattedMessage id="AuthFlowsScreen.flow-not-found" />
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
        open={confirmGenerated}
        onOpenChange={setConfirmGenerated}
        title={<FormattedMessage id="AuthFlowsScreen.edit.generated.title" />}
        description={
          <FormattedMessage
            id="AuthFlowsScreen.edit.generated.description"
            values={{ type: flowTypeName }}
          />
        }
        confirmText={
          <FormattedMessage id="AuthFlowsScreen.edit.generated.confirm" />
        }
        cancelText={<FormattedMessage id="AuthFlowsScreen.edit.cancel" />}
        confirmColor="indigo"
        onConfirm={onConfirmGenerated}
        onCancel={() => setConfirmGenerated(false)}
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
