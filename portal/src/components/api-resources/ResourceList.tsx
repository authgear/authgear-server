import React, { useMemo, useContext, useCallback } from "react";
import cn from "classnames";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { ResourceListEmptyView } from "./ResourceListEmptyView";
import {
  DetailsListLayoutMode,
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
  IDetailsRowProps,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import styles from "./ResourceList.module.css";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ActionButton from "../../ActionButton";

interface ResourceListProps {
  className?: string;
  resources: Resource[];
  loading: boolean;
  pagination: PaginationProps;
  onEdit: (resource: Resource) => void;
  onDelete: (resource: Resource) => void;
}

export const ResourceList: React.VFC<ResourceListProps> = function ResourceList(
  props
) {
  const { className, resources, loading, pagination, onDelete, onEdit } = props;
  const { renderToString } = useContext(Context);

  const columns: IColumn[] = useMemo(
    () => [
      {
        key: "name",
        name: renderToString("ResourceList.columns.name"),
        minWidth: 200,
        maxWidth: 400,
        isResizable: true,
        fieldName: "name",
      },
      {
        key: "identifier",
        name: renderToString("ResourceList.columns.identifier"),
        minWidth: 200,
        isResizable: true,
        fieldName: "resourceURI",
      },
      {
        key: "actions",
        name: "",
        minWidth: 100,
        maxWidth: 100,
        isResizable: false,
        // eslint-disable-next-line react/no-unstable-nested-components
        onRender: (item?: Resource, _0?: number, _1?: IColumn) => {
          if (item == null) {
            return null;
          }
          return (
            <ActionButtonsColumn
              resource={item}
              onDelete={onDelete}
              onEdit={onEdit}
            />
          );
        },
      },
    ],
    [onDelete, onEdit, renderToString]
  );

  if (resources.length === 0 && !loading) {
    return <ResourceListEmptyView />;
  }

  return (
    <div className={cn(className, styles.listRoot)}>
      <div
        // For DetailList to correctly know what to display
        // https://developer.microsoft.com/en-us/fluentui#/controls/web/detailslist
        data-is-scrollable="true"
        className={styles.listWrapper}
      >
        <ShimmeredDetailsList
          items={resources}
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

interface ActionButtonsColumnProps {
  resource: Resource;
  onEdit: (resource: Resource) => void;
  onDelete: (resource: Resource) => void;
}

function ActionButtonsColumn({
  resource,
  onEdit,
  onDelete,
}: ActionButtonsColumnProps) {
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  return (
    <div className="flex items-center">
      <ActionButton
        text={renderToString("edit")}
        styles={{
          label: { fontWeight: 600 },
        }}
        theme={themes.actionButton}
        onClick={useCallback(() => {
          onEdit(resource);
        }, [onEdit, resource])}
      />
      <ActionButton
        text={renderToString("delete")}
        styles={{
          label: { fontWeight: 600 },
        }}
        theme={themes.destructive}
        onClick={useCallback(() => {
          onDelete(resource);
        }, [onDelete, resource])}
      />
    </div>
  );
}

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
