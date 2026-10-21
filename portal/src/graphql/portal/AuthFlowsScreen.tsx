import React, { useCallback, useContext, useMemo, useState } from "react";
import { Flex, Heading, Select, Text } from "@radix-ui/themes";
import { useParams, useSearchParams } from "react-router-dom";
import Link from "../../Link";
import { Context, FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { Badge } from "../../components/v2/Badge/Badge";
import { useAuthFlowsScreenQueryQuery } from "./query/authFlowsScreenQuery";
import { PortalAPIAppConfig } from "../../types";
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
import { flowToTree } from "./authflows/tree";
import { findStep } from "./authflows/findStep";
import { StepDetailsPanel } from "./authflows/StepDetailsPanel";
import {
  FlowSketchContext,
  FlowSketchContextValue,
} from "./authflows/FlowSketchContext";
import styles from "./AuthFlowsScreen.module.css";

const AuthFlowsScreen: React.VFC = function AuthFlowsScreen() {
  const { appID } = useParams() as { appID: string };
  const { renderToString } = useContext(Context);
  // Settings edited elsewhere in the portal change the generated flows, and
  // nothing refetches them, so a cached copy here would be stale.
  const { data, loading, error, refetch } = useAuthFlowsScreenQueryQuery({
    variables: { id: appID },
    fetchPolicy: "network-only",
  });
  const [searchParams, setSearchParams] = useSearchParams();

  const app = data?.node?.__typename === "App" ? data.node : null;
  const appConfig = app?.effectiveAppConfig as PortalAPIAppConfig | undefined;
  const custom = appConfig?.authentication_flow;
  const generatedFlows: GeneratedFlows | undefined =
    app?.generatedAuthenticationFlows;
  const generated = useMemo(
    () =>
      generatedFlows != null ? generatedFlowsConfig(generatedFlows) : undefined,
    [generatedFlows]
  );

  const selection = useMemo(
    () => parseFlowSelection(searchParams, custom),
    [searchParams, custom]
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

  const flowKey = `${selection.type}/${selection.name}`;
  // The canvas measures its nodes once, so it remounts whenever the flow's
  // content changes, not only when another flow is selected.
  const canvasKey = useMemo(
    () => `${flowKey}:${JSON.stringify(resolved?.flow ?? null)}`,
    [flowKey, resolved]
  );
  const tree = useMemo(
    () => (resolved != null ? flowToTree(resolved.flow) : null),
    [resolved]
  );
  // Remember the selection per flow, so switching flows clears it.
  const [selected, setSelected] = useState<{
    flowKey: string;
    stepID: string;
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
            oauthProviders: appConfig?.identity?.oauth?.providers ?? [],
            customAttributes:
              appConfig?.user_profile?.custom_attributes?.attributes ?? [],
            phoneOTPMode:
              appConfig?.authenticator?.oob_otp?.sms?.phone_otp_mode,
            recoveryDestinations:
              resolved != null
                ? accountRecoveryDestinations(resolved.flow)
                : [],
            custom,
            generated,
          }
        : null,
    [selection.type, appConfig, custom, generated, resolved]
  );

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

  if (loading) {
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
          <Select.Root value={selection.type} onValueChange={onTypeChange}>
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
            <Select.Root value={selection.name} onValueChange={onNameChange}>
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
          {resolved?.source === "custom" ? (
            <Badge
              size="2"
              variant="info"
              text={<FormattedMessage id="AuthFlowsScreen.badge.custom" />}
            />
          ) : null}
          {resolved?.source === "generated" ? (
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
        </Flex>
      </header>
      <div className={styles.canvasArea}>
        {tree != null && sketchContext != null ? (
          <FlowSketchContext.Provider value={sketchContext}>
            <FlowCanvas
              key={canvasKey}
              tree={tree}
              selectedStepID={selectedStepID}
              onSelectStep={onSelectStep}
            />
          </FlowSketchContext.Provider>
        ) : (
          <Text as="p" size="2" className={styles.notFound}>
            <FormattedMessage id="AuthFlowsScreen.flow-not-found" />
          </Text>
        )}
        {selectedItem != null ? (
          <StepDetailsPanel
            item={selectedItem}
            phoneOTPMode={
              appConfig?.authenticator?.oob_otp?.sms?.phone_otp_mode
            }
            onClose={onClosePanel}
          />
        ) : null}
      </div>
    </div>
  );
};

export default AuthFlowsScreen;
