import React, { useContext, useCallback, useMemo, useState } from "react";
import cn from "classnames";
import { Cross2Icon } from "@radix-ui/react-icons";
import { Checkbox, Text } from "@radix-ui/themes";
import { Context, FormattedMessage } from "../../intl";
import styles from "./EditApplicationScopesList.module.css";
import { TextField, TextFieldIcon } from "../v2/TextField/TextField";

export interface EditApplicationScopesListItem {
  scope: string;
  isAssigned: boolean;
}

interface EditApplicationScopesListProps {
  className?: string;
  scopes: EditApplicationScopesListItem[];
  /** Reserve space at the bottom of the scroll area so the floating
   * SaveFunctionBar does not cover the last rows. */
  bottomInset?: boolean;
  onToggleAssignedScopes: (
    items: EditApplicationScopesListItem[],
    isAssigned: boolean
  ) => void;
}

export const EditApplicationScopesList: React.VFC<EditApplicationScopesListProps> =
  function EditApplicationScopesList(props: EditApplicationScopesListProps) {
    const { className, scopes, bottomInset, onToggleAssignedScopes } = props;
    const { renderToString } = useContext(Context);
    const [searchKeyword, setSearchKeyword] = useState("");

    const handleSearchChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchKeyword(e.target.value);
      },
      []
    );

    const onClearSearchKeyword = useCallback(() => {
      setSearchKeyword("");
    }, []);

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
        <div className={styles.toolbar}>
          <div className={styles.searchField}>
            <TextField
              size="2"
              type="search"
              placeholder={renderToString("search")}
              value={searchKeyword}
              iconStart={TextFieldIcon.MagnifyingGlass}
              onChange={handleSearchChange}
              suffixPlain={true}
              suffix={
                searchKeyword !== "" ? (
                  <button
                    type="button"
                    className={styles.searchClearButton}
                    aria-label={renderToString(
                      "APIResourcesScreen.clear-search"
                    )}
                    onClick={onClearSearchKeyword}
                  >
                    <Cross2Icon className={styles.searchClearIcon} />
                  </button>
                ) : undefined
              }
            />
          </div>
          <div className={styles.selectActions}>
            <Text size="2">
              <FormattedMessage id="EditApplicationScopesList.select" />
            </Text>
            <button
              type="button"
              className={styles.textAction}
              onClick={handleToggleAllScopes}
            >
              <FormattedMessage id="EditApplicationScopesList.buttons.all" />
            </button>
            <span className={styles.selectDivider} />
            <button
              type="button"
              className={styles.textAction}
              onClick={handleToggleNoneScopes}
            >
              <FormattedMessage id="EditApplicationScopesList.buttons.none" />
            </button>
          </div>
        </div>
        <div
          className={cn(
            styles.listWrapper,
            bottomInset ? styles.listWrapperInset : null
          )}
        >
          <div className={styles.table}>
            <div className={styles.tableHeader}>
              <div className={styles.tableHeaderCell}>
                <FormattedMessage id="EditApplicationScopesList.columns.scope" />
              </div>
            </div>
            {filteredScopes.map((item) => (
              <div key={item.scope} className={styles.tableRow}>
                <label className={styles.tableCell}>
                  <Checkbox
                    checked={item.isAssigned}
                    onCheckedChange={(checked) => {
                      if (checked === "indeterminate") {
                        return;
                      }
                      onToggleAssignedScopes([item], checked);
                    }}
                  />
                  <Text size="2">{item.scope}</Text>
                </label>
              </div>
            ))}
          </div>
          {filteredScopes.length === 0 ? (
            <Text as="p" size="2" color="gray" className={styles.empty}>
              <FormattedMessage id="EditApplicationScopesList.empty" />
            </Text>
          ) : null}
        </div>
      </div>
    );
  };
