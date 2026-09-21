import React, { useContext } from "react";
import cn from "classnames";
import {
  Checkbox,
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import {
  DotsVerticalIcon,
  ExclamationTriangleIcon,
  Pencil1Icon,
  TrashIcon,
} from "@radix-ui/react-icons";
import {
  AccessPolicy,
  Scope,
} from "../../graphql/adminapi/globalTypes.generated";
import { Context, FormattedMessage } from "../../intl";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import { Badge } from "../v2/Badge/Badge";
import { CardTable } from "../v2/CardTable/CardTable";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import {
  VISIBLE_ACCESS_POLICY_KEYS,
  accessPolicyLabelIDs,
  accessPolicyShortLabelIDs,
} from "./accessPolicy";
import styles from "./ScopeList.module.css";

interface ScopeListProps {
  className?: string;
  scopes: Scope[];
  // The parent resource's policy: a category the scope allows but the
  // resource does not is unreachable, and is flagged as such.
  resourceAccessPolicy: AccessPolicy;
  // Row selection for the bulk access bar.
  selectedIDs: ReadonlySet<string>;
  onToggleSelected: (scope: Scope, checked: boolean) => void;
  onToggleAllVisible: (checked: boolean) => void;
  loading: boolean;
  pagination: PaginationProps;
  onEdit: (scope: Scope) => void;
  onDelete: (scope: Scope) => void;
}

// One badge per category the scope allows. A category the resource itself
// does not allow is marked with a warning icon and dimmed, since the
// resource-level check fails first and the scope is unreachable for it
// (docs/specs/api-resource.md). Colour alone read as arbitrary highlighting.
// Static third-party access has no control in the portal (see
// VisibleAccessPolicyKey) but can be set through the Admin API, so it still
// gets a read-only badge rather than being reported as "None".
function AccessPolicyCell({
  scope,
  resourceAccessPolicy,
}: {
  scope: Scope;
  resourceAccessPolicy: AccessPolicy;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const enabled = VISIBLE_ACCESS_POLICY_KEYS.filter(
    (key) => scope.accessPolicy[key]
  );
  const staticThirdParty = scope.accessPolicy.allowStaticThirdPartyClientAccess;
  if (enabled.length === 0 && !staticThirdParty) {
    return (
      <Text size="1" color="gray">
        <FormattedMessage id="ScopeList.access-policy.none" />
      </Text>
    );
  }
  return (
    <div className={styles.accessPolicyBadges}>
      {enabled.map((key) => {
        const reachable = resourceAccessPolicy[key];
        const category = renderToString(accessPolicyLabelIDs[key]);
        return (
          <Tooltip
            key={key}
            content={
              reachable
                ? category
                : renderToString("ScopeList.access-policy.unreachable", {
                    category,
                  })
            }
          >
            <span>
              <Badge
                size="1"
                variant="neutral"
                className={cn(!reachable && styles.badgeUnreachable)}
                text={
                  <span className={styles.badgeText}>
                    {!reachable ? (
                      <ExclamationTriangleIcon
                        className={styles.badgeWarnIcon}
                        aria-hidden={true}
                      />
                    ) : null}
                    {renderToString(accessPolicyShortLabelIDs[key])}
                  </span>
                }
              />
            </span>
          </Tooltip>
        );
      })}
      {staticThirdParty ? (
        <StaticThirdPartyBadge
          reachable={resourceAccessPolicy.allowStaticThirdPartyClientAccess}
        />
      ) : null}
    </div>
  );
}

function StaticThirdPartyBadge({
  reachable,
}: {
  reachable: boolean;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const category = renderToString("AccessPolicy.static-third-party.label");
  const readOnly = renderToString("ScopeList.access-policy.read-only", {
    category,
  });
  return (
    <Tooltip
      content={
        reachable
          ? readOnly
          : `${renderToString("ScopeList.access-policy.unreachable", {
              category,
            })} ${readOnly}`
      }
    >
      <span>
        <Badge
          size="1"
          variant="neutral"
          className={cn(!reachable && styles.badgeUnreachable)}
          text={
            <span className={styles.badgeText}>
              {!reachable ? (
                <ExclamationTriangleIcon
                  className={styles.badgeWarnIcon}
                  aria-hidden={true}
                />
              ) : null}
              {renderToString("AccessPolicy.static-third-party.short")}
            </span>
          }
        />
      </span>
    </Tooltip>
  );
}

export const ScopeList: React.VFC<ScopeListProps> = function ScopeList(props) {
  const {
    className,
    scopes,
    resourceAccessPolicy,
    selectedIDs,
    onToggleSelected,
    onToggleAllVisible,
    pagination,
    onEdit,
    onDelete,
  } = props;
  const { renderToString } = useContext(Context);

  const selectedVisibleCount = scopes.filter((scope) =>
    selectedIDs.has(scope.id)
  ).length;
  const allVisibleChecked: boolean | "indeterminate" =
    scopes.length > 0 && selectedVisibleCount === scopes.length
      ? true
      : selectedVisibleCount > 0
      ? "indeterminate"
      : false;
  const onAllCheckedChange = (checked: boolean | "indeterminate") => {
    onToggleAllVisible(checked === true);
  };

  return (
    <div className={cn(className, styles.listRoot)}>
      <CardTable className={styles.scopeTable}>
        <CardTable.Header>
          <CardTable.HeaderCell className={styles.colSelect}>
            <Checkbox
              checked={allVisibleChecked}
              disabled={scopes.length === 0}
              aria-label={renderToString("ScopeList.select-all")}
              onCheckedChange={onAllCheckedChange}
            />
          </CardTable.HeaderCell>
          <CardTable.HeaderCell className={styles.colScope}>
            <FormattedMessage id="ScopeList.columns.scope" />
          </CardTable.HeaderCell>
          <CardTable.HeaderCell className={styles.colDescription}>
            <FormattedMessage id="ScopeList.columns.description" />
          </CardTable.HeaderCell>
          <CardTable.HeaderCell className={styles.colAccessPolicy}>
            <FormattedMessage id="ScopeList.columns.access-policy" />
          </CardTable.HeaderCell>
          <CardTable.HeaderCell className={styles.colActions} />
        </CardTable.Header>
        {scopes.map((scope) => (
          <CardTable.Row key={scope.id}>
            <CardTable.Cell className={styles.colSelect}>
              <Checkbox
                checked={selectedIDs.has(scope.id)}
                aria-label={renderToString("ScopeList.select-scope", {
                  scope: scope.scope,
                })}
                onCheckedChange={(checked) =>
                  onToggleSelected(scope, checked === true)
                }
              />
            </CardTable.Cell>
            <CardTable.Cell className={styles.colScope}>
              <span className={styles.scopeChip}>{scope.scope}</span>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colDescription}>
              <Text size="2" className={styles.description}>
                {scope.description}
              </Text>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colAccessPolicy}>
              <AccessPolicyCell
                scope={scope}
                resourceAccessPolicy={resourceAccessPolicy}
              />
            </CardTable.Cell>
            <CardTable.Cell className={styles.colActions}>
              <DropdownMenu.Root>
                <DropdownMenu.Trigger>
                  <RadixIconButton
                    className={styles.rowActionsButton}
                    variant="soft"
                    color="gray"
                    size="2"
                    aria-label={renderToString("ScopeList.row-actions")}
                  >
                    <DotsVerticalIcon width="1rem" height="1rem" />
                  </RadixIconButton>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="end">
                  <DropdownMenu.Item onSelect={() => onEdit(scope)}>
                    <Pencil1Icon />
                    <FormattedMessage id="edit" />
                  </DropdownMenu.Item>
                  <DropdownMenu.Item
                    color="red"
                    onSelect={() => onDelete(scope)}
                  >
                    <TrashIcon />
                    <FormattedMessage id="delete" />
                  </DropdownMenu.Item>
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            </CardTable.Cell>
          </CardTable.Row>
        ))}
      </CardTable>
      <PaginationWidget className={styles.paginator} {...pagination} />
    </div>
  );
};
