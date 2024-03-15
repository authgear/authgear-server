import {
  Dropdown,
  IDropdownProps,
  SearchBox,
  Spinner,
  SpinnerSize,
} from "@fluentui/react";
import React, { useCallback, useContext } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";

import styles from "./SearchableDropdown.module.css";

interface SearchableDropdownProps extends IDropdownProps {
  isLoadingOptions?: boolean;
  onSearchValueChange?: (value: string) => void;
  searchValue?: string;
  searchPlaceholder?: string;
}

function SearchableDropdownSearchBox(props: {
  onValueChange: ((value: string) => void) | undefined;
  value: string | undefined;
  placeholder: string | undefined;
}) {
  const { renderToString } = useContext(MessageContext);
  const { value, onValueChange, placeholder } = props;

  const onChange = useCallback(
    (e?: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      if (e == null) {
        return;
      }
      const value = e.currentTarget.value;
      onValueChange?.(value);
    },
    [onValueChange]
  );

  const onClear = useCallback(() => {
    onValueChange?.("");
  }, [onValueChange]);

  return (
    <SearchBox
      placeholder={placeholder ?? renderToString("search")}
      underlined={true}
      value={value}
      onChange={onChange}
      onClear={onClear}
    />
  );
}

const EMPTY_CALLOUT_PROPS: IDropdownProps["calloutProps"] = {};

export const SearchableDropdown: React.VFC<SearchableDropdownProps> =
  function SearchableDropdown(props) {
    const {
      isLoadingOptions,
      onSearchValueChange,
      searchValue,
      searchPlaceholder,
      calloutProps = EMPTY_CALLOUT_PROPS,
      ...restProps
    } = props;

    const onRenderList = useCallback<
      NonNullable<IDropdownProps["onRenderList"]>
    >(
      (props?, defaultRenderer?) => {
        if (defaultRenderer == null) {
          return null;
        }

        return (
          <>
            <div className={styles.searchBoxRow}>
              <SearchableDropdownSearchBox
                onValueChange={onSearchValueChange}
                value={searchValue}
                placeholder={searchPlaceholder}
              />
            </div>
            {isLoadingOptions ? (
              <div className={styles.optionsLoadingRow}>
                <Spinner size={SpinnerSize.xSmall} />
              </div>
            ) : (
              defaultRenderer(props)
            )}
          </>
        );
      },
      [isLoadingOptions, onSearchValueChange, searchPlaceholder, searchValue]
    );

    return (
      <Dropdown
        onRenderList={onRenderList}
        {...restProps}
        calloutProps={{
          calloutMaxHeight: 264,
          alignTargetEdge: true,
          ...calloutProps,
        }}
      />
    );
  };
