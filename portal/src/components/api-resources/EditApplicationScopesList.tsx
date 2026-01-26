import React, { useContext, useCallback, useMemo, useState } from "react";
import cn from "classnames";
import {
  DetailsListLayoutMode,
  IColumn,
  ShimmeredDetailsList,
  SelectionMode,
  Checkbox,
  SearchBox,
  Text,
} from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import styles from "./EditApplicationScopesList.module.css";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";

export interface EditApplicationScopesListItem {
  scope: string;
  isAssigned: boolean;
}

interface EditApplicationScopesListProps {
  className?: string;
  scopes: EditApplicationScopesListItem[];
  onToggleAssignedScopes: (
    items: EditApplicationScopesListItem[],
    isAssigned: boolean
  ) => void;
}

const TextActionButton: React.VFC<{
  text: React.ReactNode;
  onClick?: () => void;
}> = ({ text, onClick }) => {
  const { themes } = useSystemConfig();

  return (
    <ActionButton
      text={text}
      styles={{
        root: { padding: "none" },
        label: { margin: "0", padding: "0" },
      }}
      theme={themes.actionButton}
      onClick={onClick}
    />
  );
};

export const EditApplicationScopesList: React.VFC<EditApplicationScopesListProps> =
  function EditApplicationScopesList(props: EditApplicationScopesListProps) {
    const { className, scopes, onToggleAssignedScopes } = props;
    const { renderToString } = useContext(Context);
    const [searchKeyword, setSearchKeyword] = useState("");
    const { themes } = useSystemConfig();

    const handleSearchChange = useCallback((_: unknown, newValue?: string) => {
      setSearchKeyword(newValue ?? "");
    }, []);

    const onRenderScope = useCallback(
      (item?: EditApplicationScopesListItem) => {
        if (item == null) {
          return null;
        }
        return (
          <Checkbox
            label={item.scope}
            checked={item.isAssigned}
            onChange={(_, checked?: boolean) => {
              if (checked == null) {
                return;
              }
              onToggleAssignedScopes([item], checked);
            }}
          />
        );
      },
      [onToggleAssignedScopes]
    );

    const columns = useMemo(
      (): IColumn[] => [
        {
          key: "scope",
          name: renderToString("EditApplicationScopesList.columns.scope"),
          minWidth: 200,
          maxWidth: 400,
          isResizable: true,
          fieldName: "scope",
          onRender: onRenderScope,
        },
      ],
      [onRenderScope, renderToString]
    );

    const filteredScopes = useMemo(() => {
      if (!searchKeyword) {
        return scopes;
      }
      const lowerCaseSearchKeyword = searchKeyword.toLowerCase();
      return scopes.filter((item) =>
        item.scope.toLowerCase().includes(lowerCaseSearchKeyword)
      );
    }, [scopes, searchKeyword]);

    const handleToggleAllScopes = useCallback(() => {
      onToggleAssignedScopes(filteredScopes, true);
    }, [filteredScopes, onToggleAssignedScopes]);

    const handleToggleNoneScopes = useCallback(() => {
      onToggleAssignedScopes(filteredScopes, false);
    }, [filteredScopes, onToggleAssignedScopes]);

    return (
      <div className={cn(className, styles.listRoot)}>
        <div className="flex items-center space-x-8">
          <SearchBox
            placeholder={renderToString("search")}
            styles={{
              root: {
                width: 300,
              },
            }}
            onChange={handleSearchChange}
          />
          <div className="flex items-center space-x-2">
            <Text>
              <FormattedMessage id="EditApplicationScopesList.select" />
            </Text>
            <TextActionButton
              text={
                <FormattedMessage id="EditApplicationScopesList.buttons.all" />
              }
              onClick={handleToggleAllScopes}
            />
            <hr className="w-px bg-[#C8C6C4] h-4" />
            <TextActionButton
              text={
                <FormattedMessage id="EditApplicationScopesList.buttons.none" />
              }
              onClick={handleToggleNoneScopes}
            />
          </div>
        </div>
        <div data-is-scrollable="true" className={styles.listWrapper}>
          <ShimmeredDetailsList
            items={filteredScopes}
            columns={columns}
            layoutMode={DetailsListLayoutMode.justified}
            selectionMode={SelectionMode.none}
          />
          {filteredScopes.length === 0 ? (
            <Text
              styles={{ root: { color: themes.main.palette.neutralTertiary } }}
            >
              <FormattedMessage id="EditApplicationScopesList.empty" />
            </Text>
          ) : null}
        </div>
      </div>
    );
  };
