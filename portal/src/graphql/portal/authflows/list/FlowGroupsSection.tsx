import React, { useContext, useState } from "react";
import { Button, DropdownMenu, IconButton, Text } from "@radix-ui/themes";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../../../intl";
import Link from "../../../../Link";
import { Badge } from "../../../../components/v2/Badge/Badge";
import { CardTable } from "../../../../components/v2/CardTable/CardTable";
import { ConfirmationDialog } from "../../../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { Tooltip } from "../../../../components/v2/Tooltip/Tooltip";
import { PortalAPIAppConfig } from "../../../../types";
import { FlowGroupDialog, groupNameErrorMessageIDs } from "./FlowGroupDialog";
import { FlowNameDialog } from "./FlowNameDialog";
import {
  DEFAULT_GROUP_NAME,
  GroupAction,
  flowGroups,
  groupDeleteBlockers,
  groupNameError,
  groupUsers,
  standardGroupUsers,
} from "./groups";
import styles from "../../AuthFlowsListScreen.module.css";
import groupStyles from "./FlowGroupsSection.module.css";

type Pending =
  | { kind: "new" }
  | { kind: "edit" | "rename" | "delete"; name: string }
  | null;

export interface FlowGroupsSectionProps {
  appID: string;
  config: PortalAPIAppConfig | null;
  // Saves at once, like the flow list's actions.
  onRun: (action: GroupAction) => void;
}

export function FlowGroupsSection({
  appID,
  config,
  onRun,
}: FlowGroupsSectionProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const [pending, setPending] = useState<Pending>(null);
  const close = () => setPending(null);
  const run = (action: GroupAction) => {
    setPending(null);
    onRun(action);
  };
  const groups = flowGroups(config);
  // The default group also runs for every app that allows no group.
  const users = (name: string) =>
    config == null
      ? []
      : name === DEFAULT_GROUP_NAME
      ? standardGroupUsers(config)
      : groupUsers(config, name);
  const usedByLine = (name: string) => {
    const apps = users(name);
    return apps.length > 0
      ? renderToString("AuthFlowsScreen.groups.used-by", {
          names: apps.join(" · "),
        })
      : renderToString("AuthFlowsScreen.groups.unused");
  };
  // Undefined, the default group is built in: shown first, not editable.
  const hasBuiltIn = !groups.some((g) => g.name === DEFAULT_GROUP_NAME);
  const deleteBlocked = (name: string): string | null => {
    const apps = config != null ? groupDeleteBlockers(config, name) : [];
    return apps.length > 0
      ? renderToString("AuthFlowsScreen.groups.delete.blocked", {
          names: apps.join(", "),
        })
      : null;
  };
  const pendingGroup =
    pending != null && pending.kind !== "new"
      ? groups.find((g) => g.name === pending.name) ?? null
      : null;

  return (
    <section className={styles.section}>
      <Text as="p" size="2" className={groupStyles.about}>
        <FormattedMessage
          id="AuthFlowsScreen.groups.description"
          values={{
            // eslint-disable-next-line react/no-unstable-nested-components
            apps: (chunks: React.ReactNode) => (
              <Link
                to={`/project/${encodeURIComponent(appID)}/configuration/apps`}
              >
                {chunks}
              </Link>
            ),
          }}
        />
      </Text>
      <div className={groupStyles.actions}>
        <Button
          type="button"
          size="1"
          variant="soft"
          onClick={() => setPending({ kind: "new" })}
        >
          <FormattedMessage id="AuthFlowsScreen.groups.new" />
        </Button>
      </div>
      <CardTable>
        {hasBuiltIn ? (
          <CardTable.Row>
            <CardTable.Cell className={styles.colName}>
              <span className={styles.nameLine}>
                {renderToString("AuthFlowsScreen.groups.standard")}
                <Badge
                  size="1"
                  variant="neutral"
                  text={renderToString("AuthFlowsScreen.groups.built-in")}
                />
              </span>
              <Text size="1" color="gray">
                <FormattedMessage
                  id="AuthFlowsScreen.groups.standard.description"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    flows: (chunks: React.ReactNode) => (
                      <Link
                        to={`/project/${encodeURIComponent(
                          appID
                        )}/configuration/authentication/flows?tab=flows`}
                      >
                        {chunks}
                      </Link>
                    ),
                  }}
                />
              </Text>
              <Text size="1" color="gray">
                {usedByLine(DEFAULT_GROUP_NAME)}
              </Text>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colActions} />
          </CardTable.Row>
        ) : null}
        {groups.map((g) => {
          return (
            <CardTable.Row key={g.name}>
              <CardTable.Cell className={styles.colName}>
                <span className={styles.nameLine}>{g.name}</span>
                {g.name === DEFAULT_GROUP_NAME ? (
                  <Text size="1" color="gray">
                    <FormattedMessage id="AuthFlowsScreen.groups.default-note" />
                  </Text>
                ) : null}
                <Text size="1" color="gray">
                  {usedByLine(g.name)}
                </Text>
              </CardTable.Cell>
              <CardTable.Cell className={styles.colActions}>
                <GroupMenu
                  name={g.name}
                  deleteBlockedReason={deleteBlocked(g.name)}
                  onAction={(kind) => setPending({ kind, name: g.name })}
                />
              </CardTable.Cell>
            </CardTable.Row>
          );
        })}
      </CardTable>
      {pending?.kind === "new" || pending?.kind === "edit" ? (
        <FlowGroupDialog
          group={pending.kind === "edit" ? pendingGroup : null}
          flows={config?.authentication_flow}
          validateName={(n) => groupNameError(config, n, null)}
          onSubmit={(name, selection) =>
            run(
              pending.kind === "new"
                ? { kind: "add", name, selection }
                : { kind: "set-flows", name: pending.name, selection }
            )
          }
          onCancel={close}
        />
      ) : null}
      {pending?.kind === "rename" ? (
        <FlowNameDialog
          open={true}
          title={renderToString("AuthFlowsScreen.list.rename.title", {
            name: pending.name,
          })}
          confirmText={renderToString("AuthFlowsScreen.list.rename")}
          initialName={pending.name}
          validate={(n) => groupNameError(config, n, pending.name)}
          errorMessageIDs={groupNameErrorMessageIDs}
          notice={renderToString("AuthFlowsScreen.groups.rename.notice", {
            name: pending.name,
          })}
          onSubmit={(newName) =>
            run({ kind: "rename", name: pending.name, newName })
          }
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
        title={
          <FormattedMessage
            id="AuthFlowsScreen.list.delete.title"
            values={{
              name: pending?.kind === "delete" ? pending.name : "",
            }}
          />
        }
        description={
          <>
            <FormattedMessage id="AuthFlowsScreen.list.delete.description" />
            {/* Only apps that list the group lose it; the rest fall back to
                the built-in default group. */}
            {pending?.kind === "delete" &&
            config != null &&
            groupUsers(config, pending.name).length > 0
              ? ` ${renderToString("AuthFlowsScreen.groups.delete.effect", {
                  names: groupUsers(config, pending.name).join(", "),
                })}`
              : null}
          </>
        }
        confirmText={<FormattedMessage id="AuthFlowsScreen.list.delete" />}
        cancelText={<FormattedMessage id="AuthFlowsScreen.edit.cancel" />}
        onConfirm={() => {
          if (pending?.kind === "delete") {
            run({ kind: "delete", name: pending.name });
          }
        }}
        onCancel={close}
      />
    </section>
  );
}

function GroupMenu({
  name,
  deleteBlockedReason,
  onAction,
}: {
  name: string;
  // Why the group cannot be deleted, as rendered text; null when it can.
  deleteBlockedReason: string | null;
  onAction: (kind: "edit" | "rename" | "delete") => void;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const blocked = deleteBlockedReason != null;
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger>
        <IconButton
          variant="ghost"
          color="gray"
          aria-label={renderToString("AuthFlowsScreen.list.actions", { name })}
        >
          <DotsHorizontalIcon />
        </IconButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content size="1">
        <DropdownMenu.Item onSelect={() => onAction("edit")}>
          {renderToString("AuthFlowsScreen.groups.edit")}
        </DropdownMenu.Item>
        {/* A default group in authgear.yaml keeps its name: it is the one
            applications get when they ask for none. */}
        {name !== DEFAULT_GROUP_NAME ? (
          <DropdownMenu.Item onSelect={() => onAction("rename")}>
            {renderToString("AuthFlowsScreen.list.rename")}
          </DropdownMenu.Item>
        ) : null}
        <DropdownMenu.Separator />
        <Tooltip content={deleteBlockedReason} disabled={!blocked}>
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
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  );
}
