import React, { useContext, useCallback, useMemo } from "react";
import cn from "classnames";
import { Scope } from "../../graphql/adminapi/globalTypes.generated";
import {
  DetailsListLayoutMode,
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
  IDetailsRowProps,
  Text,
} from "@fluentui/react";
import { Context } from "../../intl";
import { Badge } from "../v2/Badge/Badge";
import { Tooltip } from "../v2/Tooltip/Tooltip";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import styles from "./ScopeList.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ActionButton from "../../ActionButton";

interface ScopeListProps {
  className?: string;
  scopes: Scope[];
  loading: boolean;
  pagination: PaginationProps;
  onEdit: (scope: Scope) => void;
  onDelete: (scope: Scope) => void;
}

interface ActionButtonsColumnProps {
  scope: Scope;
  onEdit: (scope: Scope) => void;
  onDelete: (scope: Scope) => void;
}

function ActionButtonsColumn({
  scope,
  onEdit,
  onDelete,
}: ActionButtonsColumnProps) {
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  return (
    <div className="flex items-center">
      <ActionButton
        text={renderToString("edit")}
        styles={{ label: { fontWeight: 600 } }}
        theme={themes.actionButton}
        onClick={useCallback(() => {
          onEdit(scope);
        }, [onEdit, scope])}
      />
      <ActionButton
        text={renderToString("delete")}
        styles={{ label: { fontWeight: 600 } }}
        theme={themes.destructive}
        onClick={useCallback(() => {
          onDelete(scope);
        }, [onDelete, scope])}
      />
    </div>
  );
}

// One enabled entry of a Scope's accessPolicy: a short badge label for the
// table plus the full sentence shown on hover.
interface AccessPolicyEntry {
  key: string;
  shortLabelID: string;
  descriptionID: string;
}

export const ScopeList: React.VFC<ScopeListProps> = function ScopeList(props) {
  const { className, scopes, loading, pagination, onEdit, onDelete } = props;
  const { renderToString } = useContext(Context);

  const onRenderActions = useCallback(
    (item?: Scope, _0?: number, _1?: IColumn) => {
      if (item == null) {
        return null;
      }
      return (
        <ActionButtonsColumn scope={item} onDelete={onDelete} onEdit={onEdit} />
      );
    },
    [onEdit, onDelete]
  );

  const onRenderScope = useCallback((item?: Scope) => {
    if (item == null) {
      return null;
    }
    return (
      <div className="py-0.5 px-1 bg-[#F3F2F1] rounded">
        <Text variant="smallPlus">{item.scope}</Text>
      </div>
    );
  }, []);

  // accessPolicy is an object the spec expects to grow more keys (see
  // docs/specs/api-resource.md, "Current keys: ..."), so build the enabled
  // policies as a list: a future policy becomes another badge here rather
  // than another column. The badge stays short enough to scan; the full
  // sentence lives in its tooltip, reusing the very string the scope form's
  // checkbox carries so the table and the form describe the policy in the
  // same words.
  const onRenderAccessPolicy = useCallback(
    (item?: Scope) => {
      if (item == null) {
        return null;
      }
      const enabled: AccessPolicyEntry[] = [];
      if (item.accessPolicy.allowDynamicThirdPartyClientAccess) {
        enabled.push({
          key: "allowDynamicThirdPartyClientAccess",
          shortLabelID:
            "ScopeList.access-policy.allow-dynamic-third-party-client-access",
          descriptionID: "ScopeForm.allow-dynamic-access.label",
        });
      }
      if (enabled.length === 0) {
        return (
          <Text
            variant="smallPlus"
            styles={{ root: { color: "var(--gray-11)" } }}
          >
            {renderToString("ScopeList.access-policy.none")}
          </Text>
        );
      }
      return (
        <div className="flex flex-row flex-wrap gap-1">
          {enabled.map((entry) => (
            <Tooltip
              key={entry.key}
              content={renderToString(entry.descriptionID)}
            >
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
    },
    [renderToString]
  );

  const columns = useMemo(
    (): IColumn[] => [
      {
        key: "scope",
        name: renderToString("ScopeList.columns.scope"),
        minWidth: 200,
        maxWidth: 400,
        isResizable: true,
        fieldName: "scope",
        onRender: onRenderScope,
      },
      {
        key: "description",
        name: renderToString("ScopeList.columns.description"),
        minWidth: 200,
        isResizable: true,
        fieldName: "description",
      },
      {
        key: "accessPolicy",
        name: renderToString("ScopeList.columns.access-policy"),
        minWidth: 160,
        maxWidth: 220,
        isResizable: true,
        onRender: onRenderAccessPolicy,
      },
      {
        key: "actions",
        name: "",
        minWidth: 100,
        maxWidth: 100,
        isResizable: false,
        onRender: onRenderActions,
      },
    ],
    [onRenderScope, onRenderAccessPolicy, onRenderActions, renderToString]
  );

  return (
    <div className={cn(className, styles.listRoot)}>
      <div data-is-scrollable="true" className={styles.listWrapper}>
        <ShimmeredDetailsList
          items={scopes}
          enableShimmer={loading}
          columns={columns}
          layoutMode={DetailsListLayoutMode.justified}
          selectionMode={SelectionMode.none}
          onRenderRow={rowRenderer}
        />
      </div>
      <PaginationWidget className={styles.paginator} {...pagination} />
    </div>
  );
};

function rowRenderer(
  props?: IDetailsRowProps,
  defaultRender?: (props?: IDetailsRowProps) => JSX.Element | null
) {
  if (props == null) {
    return defaultRender?.(props) ?? null;
  }
  props.styles = {
    cell: { display: "flex", alignItems: "center" },
  };
  return defaultRender?.(props) ?? null;
}
