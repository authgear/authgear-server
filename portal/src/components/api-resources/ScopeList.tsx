import React, { useContext } from "react";
import cn from "classnames";
import {
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import {
  DotsVerticalIcon,
  Pencil1Icon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { Scope } from "../../graphql/adminapi/globalTypes.generated";
import { Context, FormattedMessage } from "../../intl";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import { Badge } from "../v2/Badge/Badge";
import { CardTable } from "../v2/CardTable/CardTable";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import styles from "./ScopeList.module.css";

interface ScopeListProps {
  className?: string;
  scopes: Scope[];
  loading: boolean;
  pagination: PaginationProps;
  onEdit: (scope: Scope) => void;
  onDelete: (scope: Scope) => void;
}

// One enabled entry of a Scope's accessPolicy: a short badge label for the
// table plus the full sentence shown on hover.
interface AccessPolicyEntry {
  key: string;
  shortLabelID: string;
  descriptionID: string;
}

// accessPolicy is an object the spec expects to grow more keys (see
// docs/specs/api-resource.md, "Current keys: ..."), so build the enabled
// policies as a list: a future policy becomes another badge here rather
// than another column. The badge stays short enough to scan; the full
// sentence lives in its tooltip, reusing the very string the scope form's
// checkbox carries so the table and the form describe the policy in the
// same words.
function AccessPolicyCell({ scope }: { scope: Scope }): React.ReactElement {
  const { renderToString } = useContext(Context);
  const enabled: AccessPolicyEntry[] = [];
  if (scope.accessPolicy.allowDynamicThirdPartyClientAccess) {
    enabled.push({
      key: "allowDynamicThirdPartyClientAccess",
      shortLabelID:
        "ScopeList.access-policy.allow-dynamic-third-party-client-access",
      descriptionID: "ScopeForm.allow-dynamic-access.label",
    });
  }
  if (enabled.length === 0) {
    return (
      <Text size="1" color="gray">
        <FormattedMessage id="ScopeList.access-policy.none" />
      </Text>
    );
  }
  return (
    <div className={styles.accessPolicyBadges}>
      {enabled.map((entry) => (
        <Tooltip key={entry.key} content={renderToString(entry.descriptionID)}>
          <span>
            <Badge
              size="1"
              variant="neutral"
              text={renderToString(entry.shortLabelID)}
            />
          </span>
        </Tooltip>
      ))}
    </div>
  );
}

export const ScopeList: React.VFC<ScopeListProps> = function ScopeList(props) {
  const { className, scopes, pagination, onEdit, onDelete } = props;
  const { renderToString } = useContext(Context);

  return (
    <div className={cn(className, styles.listRoot)}>
      <CardTable>
        <CardTable.Header>
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
            <CardTable.Cell className={styles.colScope}>
              <span className={styles.scopeChip}>{scope.scope}</span>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colDescription}>
              <Text size="2" className={styles.description}>
                {scope.description}
              </Text>
            </CardTable.Cell>
            <CardTable.Cell className={styles.colAccessPolicy}>
              <AccessPolicyCell scope={scope} />
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
