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
import {
  GeneratedFlows,
  flowPath,
  generatedFlowsConfig,
  resolveFlow,
} from "./authflows/model";
import { FlowRow, SectionKey, flowSections } from "./authflows/list/rows";
import {
  FlowAction,
  applyFlowAction,
  deleteBlockers,
} from "./authflows/list/actions";
import { FlowNameDialog } from "./authflows/list/FlowNameDialog";
import { NewFlowDialog } from "./authflows/list/NewFlowDialog";
import { flowUses, removeFlowUses } from "./authflows/list/usage";
import { renameFlow } from "./authflows/edit/save";
import { removeStoredDraft } from "./authflows/edit/draftStorage";
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
  if (action.kind === "rename") {
    return renameFlow(raw, action.type, action.name, action.newName);
  }
  const next = produce(raw, (d) => {
    d.authentication_flow = applyFlowAction(raw.authentication_flow, action);
  });
  // Groups and applications let go of a deleted flow. A deleted default
  // needs nothing: the generated one answers to it.
  if (action.kind === "delete") {
    return removeFlowUses(next, action.type, action.name);
  }
  return next;
}

const sectionTitleMessageIDs: Record<SectionKey, string> = {
  signup_login: "AuthFlowsScreen.list.section.signup_login",
  signup: "AuthFlowsScreen.list.section.signup",
  login: "AuthFlowsScreen.list.section.login",
};

type Pending = {
  kind: "rename" | "duplicate" | "delete";
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
  const app = data?.node?.__typename === "App" ? data.node : null;
  const generatedFlows: GeneratedFlows | undefined =
    app?.generatedAuthenticationFlows;
  const generated = useMemo(
    () => (generatedFlows != null ? generatedFlowsConfig(generatedFlows) : {}),
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
  // The section whose New flow was clicked: it sets the flow's type.
  const [creating, setCreating] = useState<SectionKey | null>(null);
  const close = useCallback(() => setPending(null), []);

  const { saveWith } = form;
  const { setState } = form;
  const run = useCallback(
    (action: FlowAction) => {
      setPending(null);
      // A failure shows in the error bar, from the form's updateError. The
      // failed action is dropped so the list is not left with unsaved
      // changes it does not have.
      saveWith(() => ({ action })).then(
        () => {
          // A draft kept in this browser for a deleted flow is not offered
          // back if the name is used again.
          if (action.kind === "delete") {
            removeStoredDraft(appID, action.type, action.name);
          }
        },
        () => {
          setState(() => ({ action: null }));
        }
      );
    },
    [saveWith, setState, appID]
  );

  // Every place that still names the flow: deleting it would leave them
  // pointing at nothing, which the server refuses.
  // What names the flow: hand-offs from signup_login flows, flow groups,
  // and applications that allow it.
  const usesOf = (row: FlowRow) => {
    const uses =
      form.rawConfig != null
        ? flowUses(form.rawConfig, row.type, row.name)
        : [];
    return {
      handoffs: deleteBlockers(flows, row.type, row.name),
      groups: uses.filter((u) => u.kind === "group").map((u) => u.name),
      apps: uses.filter((u) => u.kind === "application").map((u) => u.name),
    };
  };
  // A hand-off has no safe stand-in, so it blocks deleting; groups and
  // applications are cleaned up by the delete itself.
  const deleteBlockedReason = (row: FlowRow): string | null => {
    const { handoffs } = usesOf(row);
    return handoffs.length > 0
      ? renderToString("AuthFlowsScreen.list.delete.blocked", {
          names: handoffs.join(", "),
        })
      : null;
  };
  const usedByLine = (row: FlowRow): string | null => {
    const { handoffs, groups, apps } = usesOf(row);
    const parts = [
      ...groups.map((n) =>
        renderToString("AuthFlowsScreen.list.used-by.group", { name: n })
      ),
      ...apps.map((n) =>
        renderToString("AuthFlowsScreen.list.used-by.application", { name: n })
      ),
      ...handoffs.map((n) =>
        renderToString("AuthFlowsScreen.list.used-by.handoff", { name: n })
      ),
    ];
    return parts.length > 0
      ? renderToString("AuthFlowsScreen.list.used-by", {
          uses: parts.join(" · "),
        })
      : null;
  };
  // What deleting changes elsewhere, for the confirmation.
  const deleteEffects = (row: FlowRow): string => {
    const { groups, apps } = usesOf(row);
    return [
      groups.length > 0
        ? renderToString("AuthFlowsScreen.list.delete.effect.groups", {
            names: groups.join(", "),
          })
        : null,
      apps.length > 0
        ? renderToString("AuthFlowsScreen.list.delete.effect.applications", {
            names: apps.join(", "),
          })
        : null,
    ]
      .filter((t): t is string => t != null)
      .join(" ");
  };
  const duplicate = (row: FlowRow, newName: string) => {
    // A generated row duplicates the generated flow.
    const flow = resolveFlow(flows, generated, row.type, row.name)?.flow;
    if (flow != null) {
      run({ kind: "add", type: row.type, flow: { ...flow, name: newName } });
    }
  };

  return (
    <div className={styles.root}>
      <header className={styles.header}>
        <Heading size="6" as="h1">
          <FormattedMessage id="AuthFlowsScreen.title" />
        </Heading>
        <Text as="p" size="2" color="gray">
          <FormattedMessage id="AuthFlowsScreen.list.description" />
        </Text>
      </header>
      {creating != null ? (
        <NewFlowDialog
          appID={appID}
          type={creating}
          open={true}
          onOpenChange={(o) => {
            if (!o) {
              setCreating(null);
            }
          }}
          flows={flows}
          generated={generated}
          config={form.effectiveConfig}
        />
      ) : null}
      <FormErrorMessageBar />
      <div className={styles.sections}>
        {sections.map((section) => (
          <section key={section.key} className={styles.section}>
            <div className={styles.sectionHeader}>
              <Heading size="3" as="h2">
                {renderToString(sectionTitleMessageIDs[section.key])}
              </Heading>
              <Button
                type="button"
                size="1"
                variant="soft"
                onClick={() => setCreating(section.key)}
              >
                <FormattedMessage id="AuthFlowsScreen.new.button" />
              </Button>
            </div>
            <CardTable>
              {section.rows.map((row) => (
                <CardTable.Row key={`${row.type}/${row.name}`}>
                  <CardTable.Cell className={styles.colName}>
                    <span className={styles.nameLine}>
                      <Link to={flowPath(appID, row.type, row.name)}>
                        {row.name}
                      </Link>
                      {/* A default flow someone changed: Reset brings back
                          the one built from Login methods. */}
                      {row.isDefault && !row.generated ? (
                        <Tooltip
                          content={renderToString(
                            "AuthFlowsScreen.list.edited.tooltip"
                          )}
                        >
                          <span>
                            <Badge
                              size="1"
                              variant="neutral"
                              text={renderToString(
                                "AuthFlowsScreen.list.edited"
                              )}
                            />
                          </span>
                        </Tooltip>
                      ) : null}
                    </span>
                    {usedByLine(row) != null ? (
                      <Text size="1" color="gray">
                        {usedByLine(row)}
                      </Text>
                    ) : null}
                  </CardTable.Cell>
                  <CardTable.Cell className={styles.colActions}>
                    <RowMenu
                      row={row}
                      deleteBlockedReason={deleteBlockedReason(row)}
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
          type={pending.row.type}
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
              });
            } else {
              duplicate(row, newName);
            }
          }}
          onCancel={close}
        />
      ) : null}
      <ConfirmationDialog
        open={pending?.kind === "delete"}
        onOpenChange={(o) => {
          if (!o) {
            close();
          }
        }}
        // Removing an edited default resets it to the one built from Login
        // methods, so it reads as Reset.
        title={
          <FormattedMessage
            id={
              pending?.row.isDefault === true
                ? "AuthFlowsScreen.reset.title"
                : "AuthFlowsScreen.list.delete.title"
            }
            values={{ name: pending?.row.name ?? "" }}
          />
        }
        description={
          <>
            <FormattedMessage
              id={
                pending?.row.isDefault === true
                  ? "AuthFlowsScreen.reset.description"
                  : "AuthFlowsScreen.list.delete.description"
              }
            />
            {pending != null && !pending.row.isDefault
              ? ` ${deleteEffects(pending.row)}`
              : null}
          </>
        }
        confirmText={
          <FormattedMessage
            id={
              pending?.row.isDefault === true
                ? "AuthFlowsScreen.reset"
                : "AuthFlowsScreen.list.delete"
            }
          />
        }
        cancelText={<FormattedMessage id="AuthFlowsScreen.edit.cancel" />}
        onConfirm={() => {
          if (pending != null) {
            run({
              kind: "delete",
              type: pending.row.type,
              name: pending.row.name,
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
  deleteBlockedReason,
  onOpen,
  onAction,
}: {
  row: FlowRow;
  // Why the flow cannot be deleted, as rendered text; null when it can.
  deleteBlockedReason: string | null;
  onOpen: () => void;
  onAction: (kind: "rename" | "duplicate" | "delete") => void;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const blocked = deleteBlockedReason != null;
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
        {/* A generated flow is not in authgear.yaml: nothing to rename or
            delete until it is edited and saved. */}
        {/* The default flow keeps its name: AuthUI runs it by that name. */}
        {!row.generated && !row.isDefault ? (
          <DropdownMenu.Item onSelect={() => onAction("rename")}>
            {renderToString("AuthFlowsScreen.list.rename")}
          </DropdownMenu.Item>
        ) : null}
        <DropdownMenu.Item onSelect={() => onAction("duplicate")}>
          {renderToString("AuthFlowsScreen.list.duplicate")}
        </DropdownMenu.Item>
        {!row.generated ? <DropdownMenu.Separator /> : null}
        {!row.generated ? (
          <Tooltip content={deleteBlockedReason} disabled={!blocked}>
            <span>
              <DropdownMenu.Item
                color="red"
                disabled={blocked}
                onSelect={() => onAction("delete")}
              >
                {renderToString(
                  row.isDefault
                    ? "AuthFlowsScreen.reset"
                    : "AuthFlowsScreen.list.delete"
                )}
              </DropdownMenu.Item>
            </span>
          </Tooltip>
        ) : null}
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  );
}

export default AuthFlowsListScreen;
