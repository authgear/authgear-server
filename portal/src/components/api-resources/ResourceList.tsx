import React, { useMemo, useContext } from "react";
import cn from "classnames";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { ResourceListEmptyView } from "./ResourceListEmptyView";
import {
  DetailsListLayoutMode,
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
} from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import styles from "./ResourceList.module.css";

interface ResourceListProps {
  className?: string;
  resources: Resource[];
  loading: boolean;
  pagination: PaginationProps;
}

export const ResourceList: React.VFC<ResourceListProps> = function ResourceList(
  props
) {
  const { className, resources, loading, pagination } = props;
  const { renderToString } = useContext(Context);

  const columns: IColumn[] = useMemo(
    () => [
      {
        key: "name",
        name: renderToString("ResourceList.columns.name"),
        minWidth: 200,
        maxWidth: 300,
        isResizable: true,
        fieldName: "name",
      },
      {
        key: "identifier",
        name: renderToString("ResourceList.columns.identifier"),
        minWidth: 200,
        maxWidth: 300,
        isResizable: true,
        fieldName: "resourceURI",
      },
    ],
    [renderToString]
  );

  if (resources.length === 0) {
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
        />
      </div>
      <PaginationWidget className={styles.paginator} {...pagination} />
    </div>
  );
};
