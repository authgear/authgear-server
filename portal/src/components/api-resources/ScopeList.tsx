import React, { useContext, useCallback } from "react";
import cn from "classnames";
import { Scope } from "../../graphql/adminapi/globalTypes.generated";
import {
  DetailsListLayoutMode,
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
  IDetailsRowProps,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
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

function getScopeListColumns(
  onEdit: (scope: Scope) => void,
  onDelete: (scope: Scope) => void,
  renderToString: (id: string) => string
): IColumn[] {
  return [
    {
      key: "scope",
      name: renderToString("ScopeList.columns.scope"),
      minWidth: 200,
      maxWidth: 400,
      isResizable: true,
      fieldName: "scope",
    },
    {
      key: "description",
      name: renderToString("ScopeList.columns.description"),
      minWidth: 200,
      isResizable: true,
      fieldName: "description",
    },
    {
      key: "actions",
      name: "",
      minWidth: 100,
      maxWidth: 100,
      isResizable: false,
      onRender: (item?: Scope, _0?: number, _1?: IColumn) => {
        if (item == null) {
          return null;
        }
        return (
          <ActionButtonsColumn
            scope={item}
            onDelete={onDelete}
            onEdit={onEdit}
          />
        );
      },
    },
  ];
}

export const ScopeList: React.VFC<ScopeListProps> = function ScopeList(props) {
  const { className, scopes, loading, pagination, onEdit, onDelete } = props;
  const { renderToString } = useContext(Context);

  const columns = getScopeListColumns(onEdit, onDelete, renderToString);

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
