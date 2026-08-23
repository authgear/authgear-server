import React, { useMemo, useContext, useCallback } from "react";
import cn from "classnames";
import {
  DetailsListLayoutMode,
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
  IDetailsRowProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import {
  OAuthClientKind,
  OAuthClientSource,
} from "../../graphql/adminapi/globalTypes.generated";
import PaginationWidget, { PaginationProps } from "../../PaginationWidget";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { TextWithCopyButton } from "../common/TextWithCopyButton";
import { formatDatetime } from "../../util/formatDatetime";
import styles from "./DynamicClientList.module.css";

export interface DynamicClientListItem {
  id: string;
  clientID: string;
  clientName: string | null;
  name: string;
  kind: OAuthClientKind;
  source: OAuthClientSource;
  registeredAt: string | null;
  applicationType: string | null;
  redirectURIs: string[];
  grantTypes: string[];
  responseTypes: string[];
  logoURI: string | null;
  clientURI: string | null;
  tosURI: string | null;
  policyURI: string | null;
}

interface DynamicClientListProps {
  className?: string;
  clients: DynamicClientListItem[];
  loading: boolean;
  pagination: PaginationProps;
  onDelete: (client: DynamicClientListItem) => void;
  onItemClicked: (item: DynamicClientListItem) => void;
}

export const DynamicClientList: React.VFC<DynamicClientListProps> =
  function DynamicClientList(props) {
    const { className, clients, loading, pagination, onDelete, onItemClicked } =
      props;
    const { renderToString, locale } = useContext(Context);

    const columns: IColumn[] = useMemo(
      () => [
        {
          key: "name",
          name: renderToString("DynamicClientList.columns.name"),
          minWidth: 150,
          maxWidth: 300,
          isResizable: true,
          fieldName: "name",
        },
        {
          key: "clientID",
          name: renderToString("DynamicClientList.columns.client-id"),
          minWidth: 200,
          isResizable: true,
          // eslint-disable-next-line react/no-unstable-nested-components
          onRender: (item?: DynamicClientListItem) => {
            if (item == null) {
              return null;
            }
            return <TextWithCopyButton text={item.clientID} />;
          },
        },
        {
          key: "kind",
          name: renderToString("DynamicClientList.columns.kind"),
          minWidth: 100,
          maxWidth: 120,
          isResizable: true,
          // eslint-disable-next-line react/no-unstable-nested-components
          onRender: (item?: DynamicClientListItem) => {
            if (item == null) {
              return null;
            }
            return (
              <span>
                {item.kind === OAuthClientKind.FirstParty ? (
                  <FormattedMessage id="DynamicClientList.kind.first-party" />
                ) : (
                  <FormattedMessage id="DynamicClientList.kind.third-party" />
                )}
              </span>
            );
          },
        },
        {
          key: "registeredAt",
          name: renderToString("DynamicClientList.columns.registered-at"),
          minWidth: 150,
          maxWidth: 220,
          isResizable: true,
          // eslint-disable-next-line react/no-unstable-nested-components
          onRender: (item?: DynamicClientListItem) => {
            if (item == null) {
              return null;
            }
            return <span>{formatDatetime(locale, item.registeredAt)}</span>;
          },
        },
        {
          key: "actions",
          name: "",
          minWidth: 100,
          maxWidth: 100,
          isResizable: false,
          // eslint-disable-next-line react/no-unstable-nested-components
          onRender: (item?: DynamicClientListItem) => {
            if (item == null) {
              return null;
            }
            return <ActionButtonsColumn client={item} onDelete={onDelete} />;
          },
        },
      ],
      [onDelete, renderToString, locale]
    );

    const rowRenderer = useCallback(
      (
        props?: IDetailsRowProps,
        defaultRender?: (props?: IDetailsRowProps) => JSX.Element | null
      ): JSX.Element | null => {
        if (props == null) {
          return defaultRender?.(props) ?? null;
        }
        const item = props.item as DynamicClientListItem | undefined;
        props.styles = {
          cell: { display: "flex", alignItems: "center" },
        };

        return (
          <div
            onClick={() => {
              if (item != null) {
                onItemClicked(item);
              }
            }}
            className="contents cursor-pointer"
          >
            {defaultRender?.(props)}
          </div>
        );
      },
      [onItemClicked]
    );

    return (
      <div className={cn(className, styles.listRoot)}>
        <div
          // For DetailList to correctly know what to display
          // https://developer.microsoft.com/en-us/fluentui#/controls/web/detailslist
          data-is-scrollable="true"
          className={styles.listWrapper}
        >
          <ShimmeredDetailsList
            items={clients}
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
  client: DynamicClientListItem;
  onDelete: (client: DynamicClientListItem) => void;
}

function ActionButtonsColumn({ client, onDelete }: ActionButtonsColumnProps) {
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  return (
    <div className="flex items-center justify-end flex-1">
      <ActionButton
        text={renderToString("delete")}
        styles={{
          label: { fontWeight: 600 },
        }}
        theme={themes.destructive}
        onClick={useCallback(
          (e: React.MouseEvent<HTMLButtonElement>) => {
            e.stopPropagation();
            onDelete(client);
          },
          [onDelete, client]
        )}
      />
    </div>
  );
}
