import React, { useCallback, useContext, useMemo, useState } from "react";
import { produce } from "immer";
import {
  Button,
  DropdownMenu,
  Heading,
  IconButton,
  Text,
} from "@radix-ui/themes";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import { useNavigate, useParams } from "react-router-dom";
import Link from "../../Link";
import { Context, FormattedMessage } from "../../intl";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { Badge } from "../../components/v2/Badge/Badge";
import { CardTable } from "../../components/v2/CardTable/CardTable";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { FormContainerBase } from "../../FormContainerBase";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { AuthenticationFlowsConfig, PortalAPIAppConfig } from "../../types";
import { useAuthFlowsScreenQueryQuery } from "./query/authFlowsScreenQuery";
import { DEFAULT_FLOW_NAME, flowPath, resolveFlow } from "./authflows/model";
import { flowTypeMessageIDs } from "./authflows/labels";
import {
  FlowRow,
  SectionKey,
  bundleTypes,
  flowSections,
} from "./authflows/list/rows";
import {
  FlowAction,
  applyFlowAction,
  copyAs,
  deleteBlockers,
  nameError,
} from "./authflows/list/actions";
import { FlowNameDialog } from "./authflows/list/FlowNameDialog";
import { NewFlowDialog } from "./authflows/list/NewFlowDialog";
import styles from "./AuthFlowsListScreen.module.css";

// The one action being saved; null otherwise.
interface ListState {
  action: FlowAction | null;
}

type ListForm = AppConfigFormModel<ListState>;

function constructFormState(): ListState {
  return { action: null };
}

// Applied to authgear.yaml as written, so defaults (rate limits) are not
// written back with it.
function constructConfig(
  raw: PortalAPIAppConfig,
  _initial: ListState,
  current: ListState
): PortalAPIAppConfig {
  const { action } = current;
  if (action == null) {
    return raw;
  }
  return produce(raw, (d) => {
    d.authentication_flow = applyFlowAction(raw.authentication_flow, action);
  });
}

const sectionTitleMessageIDs: Record<SectionKey, string> = {
  signup_login: "AuthFlowsScreen.list.section.signup_login",
  signup: "AuthFlowsScreen.list.section.signup",
  login: "AuthFlowsScreen.list.section.login",
  other: "AuthFlowsScreen.list.section.other",
};

type Pending = {
  kind: "rename" | "duplicate" | "set_default" | "delete";
  row: FlowRow;
} | null;

const AuthFlowsListScreen: React.VFC = function AuthFlowsListScreen() {
  const { appID } = useParams() as { appID: string };
  // Settings edited elsewhere change the generated flows; never use a cache.
  const { data, loading, error, refetch } = useAuthFlowsScreenQueryQuery({
    variables: { id: appID },
    fetchPolicy: "network-only",
  });
  const form = useAppConfigForm<ListState>({
    appID,
    constructFormState,
    constructConfig,
  });
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
  const app = data?.node?.__typename === "App" ? data.node : null;
  const generated = (app?.generatedAuthenticationFlows ??
    {}) as AuthenticationFlowsConfig;
  return (
    <FormContainerBase form={form}>
      <ListContent appID={appID} form={form} generated={generated} />
    </FormContainerBase>
  );
};

function ListContent({
  appID,
  form,
  generated,
}: {
  appID: string;
  form: ListForm;
  generated: AuthenticationFlowsConfig;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const flows = form.rawConfig?.authentication_flow;
  const sections = useMemo(
    () => flowSections(flows, generated),
    [flows, generated]
  );
  const [pending, setPending] = useState<Pending>(null);
  const [creating, setCreating] = useState(false);
  const close = useCallback(() => setPending(null), []);

  const { saveWith } = form;
  const run = useCallback(
    (action: FlowAction) => {
      setPending(null);
      // A failure shows in the error bar, from the form's updateError.
      saveWith(() => ({ action })).catch(() => {});
    },
    [saveWith]
  );

  const typesOf = (row: FlowRow) => (row.bundle ? bundleTypes() : [row.type]);
  const duplicate = (row: FlowRow, newName: string) => {
    const sources = typesOf(row).flatMap((type) => {
      const flow = resolveFlow(flows, generated, type, row.name)?.flow;
      return flow != null ? [{ type, flow }] : [];
    });
    run({ kind: "add", flows: copyAs(sources, newName) });
  };

  return (
    <div className={styles.root}>
      <header className={styles.header}>
        <div className={styles.titleRow}>
          <Heading size="6" as="h1">
            <FormattedMessage id="AuthFlowsScreen.title" />
          </Heading>
          <Button type="button" onClick={() => setCreating(true)}>
            <FormattedMessage id="AuthFlowsScreen.new.button" />
          </Button>
        </div>
        <Text as="p" size="2" color="gray">
          <FormattedMessage id="AuthFlowsScreen.list.description" />
        </Text>
      </header>
      {creating ? (
        <NewFlowDialog
          appID={appID}
          open={true}
          onOpenChange={setCreating}
          flows={flows}
          generated={generated}
          config={form.effectiveConfig}
        />
      ) : null}
      <FormErrorMessageBar />
      <div className={styles.sections}>
        {sections.map((section) => (
          <section key={section.key} className={styles.section}>
            <Heading size="3" as="h2">
              {renderToString(sectionTitleMessageIDs[section.key])}
            </Heading>
            <CardTable>
              {section.rows.map((row) => (
                <CardTable.Row key={`${row.type}/${row.name}`}>
                  <CardTable.Cell className={styles.colName}>
                    <Link to={flowPath(appID, row.type, row.name)}>
                      {row.name}
                    </Link>
                    {section.key === "other" ? (
                      <Text size="1" color="gray">
                        {renderToString(flowTypeMessageIDs[row.type])}
                      </Text>
                    ) : null}
                    {row.bundle ? (
                      <Text size="1" color="gray">
                        <FormattedMessage
                          id="AuthFlowsScreen.list.bundle"
                          values={{ name: row.name }}
                        />
                      </Text>
                    ) : null}
                    {row.partOf != null ? (
                      <Text size="1" color="gray">
                        <FormattedMessage
                          id="AuthFlowsScreen.list.part-of"
                          values={{ name: row.partOf }}
                        />
                      </Text>
                    ) : null}
                  </CardTable.Cell>
                  <CardTable.Cell className={styles.colStatus}>
                    <Badge
                      size="1"
                      variant={row.inUse ? "success" : "neutral"}
                      text={renderToString(
                        row.inUse
                          ? "AuthFlowsScreen.badge.in-use"
                          : "AuthFlowsScreen.badge.not-in-use"
                      )}
                    />
                    {row.generated ? (
                      <Badge
                        size="1"
                        variant="neutral"
                        text={renderToString("AuthFlowsScreen.badge.generated")}
                      />
                    ) : null}
                  </CardTable.Cell>
                  <CardTable.Cell className={styles.colActions}>
                    <RowMenu
                      row={row}
                      deleteBlockedBy={deleteBlockers(
                        flows,
                        row.type,
                        row.name,
                        row.bundle
                      )}
                      onOpen={() => {
                        navigate(flowPath(appID, row.type, row.name));
                      }}
                      onAction={(kind) => setPending({ kind, row })}
                    />
                  </CardTable.Cell>
                </CardTable.Row>
              ))}
            </CardTable>
          </section>
        ))}
      </div>
      {pending?.kind === "rename" || pending?.kind === "duplicate" ? (
        <FlowNameDialog
          open={true}
          title={renderToString(
            pending.kind === "rename"
              ? "AuthFlowsScreen.list.rename.title"
              : "AuthFlowsScreen.list.duplicate.title",
            { name: pending.row.name }
          )}
          confirmText={renderToString(
            pending.kind === "rename"
              ? "AuthFlowsScreen.list.rename"
              : "AuthFlowsScreen.list.duplicate"
          )}
          types={typesOf(pending.row)}
          initialName={
            pending.kind === "rename"
              ? pending.row.name
              : `${pending.row.name}_copy`
          }
          flows={flows}
          onSubmit={(newName) => {
            const { row } = pending;
            if (pending.kind === "rename") {
              run({
                kind: "rename",
                type: row.type,
                name: row.name,
                newName,
                bundle: row.bundle,
              });
            } else {
              duplicate(row, newName);
            }
          }}
          onCancel={close}
        />
      ) : null}
      <ConfirmationDialog
        open={pending?.kind === "set_default"}
        onOpenChange={(o) => {
          if (!o) {
            close();
          }
        }}
        title={<FormattedMessage id="AuthFlowsScreen.list.set-default.title" />}
        description={
          <FormattedMessage
            id={
              pending != null &&
              nameError(flows, typesOf(pending.row), DEFAULT_FLOW_NAME, true) ==
                null
                ? "AuthFlowsScreen.list.set-default.description.generated"
                : "AuthFlowsScreen.list.set-default.description"
            }
            values={{ name: pending?.row.name ?? "" }}
          />
        }
        confirmText={<FormattedMessage id="AuthFlowsScreen.list.set-default" />}
        cancelText={<FormattedMessage id="AuthFlowsScreen.edit.cancel" />}
        confirmColor="indigo"
        onConfirm={() => {
          if (pending != null) {
            run({
              kind: "set_default",
              type: pending.row.type,
              name: pending.row.name,
              bundle: pending.row.bundle,
            });
          }
        }}
        onCancel={close}
      />
      <ConfirmationDialog
        open={pending?.kind === "delete"}
        onOpenChange={(o) => {
          if (!o) {
            close();
          }
        }}
        title={
          <FormattedMessage
            id="AuthFlowsScreen.list.delete.title"
            values={{ name: pending?.row.name ?? "" }}
          />
        }
        description={
          <FormattedMessage
            id={
              pending?.row.inUse === true
                ? "AuthFlowsScreen.list.delete.description.in-use"
                : "AuthFlowsScreen.list.delete.description"
            }
            values={{ bundle: pending?.row.bundle === true ? "yes" : "no" }}
          />
        }
        confirmText={<FormattedMessage id="AuthFlowsScreen.list.delete" />}
        cancelText={<FormattedMessage id="AuthFlowsScreen.edit.cancel" />}
        onConfirm={() => {
          if (pending != null) {
            run({
              kind: "delete",
              type: pending.row.type,
              name: pending.row.name,
              bundle: pending.row.bundle,
            });
          }
        }}
        onCancel={close}
      />
    </div>
  );
}

function RowMenu({
  row,
  deleteBlockedBy,
  onOpen,
  onAction,
}: {
  row: FlowRow;
  deleteBlockedBy: string[];
  onOpen: () => void;
  onAction: (kind: "rename" | "duplicate" | "set_default" | "delete") => void;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const blocked = deleteBlockedBy.length > 0;
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger>
        <IconButton
          variant="ghost"
          color="gray"
          aria-label={renderToString("AuthFlowsScreen.list.actions", {
            name: row.name,
          })}
        >
          <DotsHorizontalIcon />
        </IconButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content size="1">
        <DropdownMenu.Item onSelect={onOpen}>
          {renderToString("AuthFlowsScreen.list.open")}
        </DropdownMenu.Item>
        {!row.inUse && !row.generated ? (
          <DropdownMenu.Item onSelect={() => onAction("set_default")}>
            {renderToString("AuthFlowsScreen.list.set-default")}
          </DropdownMenu.Item>
        ) : null}
        {!row.generated ? (
          <DropdownMenu.Item onSelect={() => onAction("rename")}>
            {renderToString("AuthFlowsScreen.list.rename")}
          </DropdownMenu.Item>
        ) : null}
        <DropdownMenu.Item onSelect={() => onAction("duplicate")}>
          {renderToString("AuthFlowsScreen.list.duplicate")}
        </DropdownMenu.Item>
        {!row.generated ? (
          <>
            <DropdownMenu.Separator />
            <Tooltip
              content={renderToString("AuthFlowsScreen.list.delete.blocked", {
                names: deleteBlockedBy.join(", "),
              })}
              disabled={!blocked}
            >
              <span>
                <DropdownMenu.Item
                  color="red"
                  disabled={blocked}
                  onSelect={() => onAction("delete")}
                >
                  {renderToString("AuthFlowsScreen.list.delete")}
                </DropdownMenu.Item>
              </span>
            </Tooltip>
          </>
        ) : null}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  );
}

export default AuthFlowsListScreen;
