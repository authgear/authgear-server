import React, { useContext, useCallback, useMemo } from "react";
import cn from "classnames";
import {
  DetailsListLayoutMode,
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
} from "@fluentui/react";
import { Context as MessageContext } from "../../intl";
import styles from "./ApplicationList.module.css";
import Toggle from "../../Toggle";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";

export interface ApplicationListItem {
  clientID: string;
  name: string;
  authorized: boolean;
}

interface ApplicationListProps {
  className?: string;
  applications: ApplicationListItem[];
  loading: boolean;
  disabledToggleClientIDs: string[];
  onToggleAuthorized: (item: ApplicationListItem, checked: boolean) => void;
  onManageScopes: (item: ApplicationListItem) => void;
}

export const ApplicationList: React.VFC<ApplicationListProps> =
  function ApplicationList(props) {
    const {
      className,
      applications,
      loading,
      onToggleAuthorized,
      onManageScopes,
      disabledToggleClientIDs,
    } = props;
    const { renderToString } = useContext(MessageContext);
    const { themes } = useSystemConfig();

    const disabledToggleClientIDsSet = useMemo(() => {
      return new Set(disabledToggleClientIDs);
    }, [disabledToggleClientIDs]);

    const onRenderAuthorized = useCallback(
      (item?: ApplicationListItem) => {
        if (item == null) {
          return null;
        }
        return (
          <AuthorizedToggle
            item={item}
            onToggleAuthorized={onToggleAuthorized}
            disabled={disabledToggleClientIDsSet.has(item.clientID)}
          />
        );
      },
      [disabledToggleClientIDsSet, onToggleAuthorized]
    );

    const columns = useMemo(
      (): IColumn[] => [
        {
          key: "application",
          name: renderToString("ApplicationList.columns.application"),
          minWidth: 200,
          maxWidth: 400,
          isResizable: true,
          fieldName: "name",
        },
        {
          key: "authorized",
          name: renderToString("ApplicationList.columns.authorized"),
          minWidth: 100,
          maxWidth: 200,
          isResizable: true,
          fieldName: "authorized",
          onRender: onRenderAuthorized,
        },
        {
          key: "actions",
          name: "",
          minWidth: 100,
          maxWidth: 100,
          isResizable: false,
          // eslint-disable-next-line react/no-unstable-nested-components
          onRender: (item?: ApplicationListItem) => {
            if (
              !item?.authorized ||
              disabledToggleClientIDsSet.has(item.clientID)
            ) {
              return null;
            }
            const handleClick = () => {
              onManageScopes(item);
            };
            return (
              <ActionButton
                text={renderToString("ApplicationList.columns.manageScopes")}
                styles={{
                  label: { fontWeight: 600 },
                  root: { height: "auto" },
                }}
                theme={themes.actionButton}
                onClick={handleClick}
              />
            );
          },
        },
      ],
      [
        renderToString,
        onRenderAuthorized,
        disabledToggleClientIDsSet,
        themes.actionButton,
        onManageScopes,
      ]
    );

    return (
      <div className={cn(className, styles.listRoot)}>
        <div data-is-scrollable="true" className={styles.listWrapper}>
          <ShimmeredDetailsList
            items={applications}
            enableShimmer={loading}
            columns={columns}
            layoutMode={DetailsListLayoutMode.justified}
            selectionMode={SelectionMode.none}
          />
        </div>
      </div>
    );
  };

interface AuthorizedToggleProps {
  item: ApplicationListItem;
  onToggleAuthorized: (item: ApplicationListItem, checked: boolean) => void;
  disabled: boolean;
}

function AuthorizedToggle(props: AuthorizedToggleProps) {
  const { item, onToggleAuthorized, disabled } = props;
  const onChange = useCallback(
    (_?: React.MouseEvent<HTMLElement>, checked?: boolean) => {
      if (checked == null) {
        return;
      }
      onToggleAuthorized(item, checked);
    },
    [item, onToggleAuthorized]
  );
  return (
    <Toggle checked={item.authorized} onChange={onChange} disabled={disabled} />
  );
}
