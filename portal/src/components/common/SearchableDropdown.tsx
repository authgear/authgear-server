import {
  Dropdown,
  IDropdownOption,
  IDropdownProps,
  SearchBox,
  Spinner,
  SpinnerSize,
} from "@fluentui/react";
import React, { useCallback, useContext, useMemo, useState } from "react";
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
      options,
      isLoadingOptions,
      onSearchValueChange,
      onChange: propsOnChange,
      searchValue,
      searchPlaceholder,
      calloutProps = EMPTY_CALLOUT_PROPS,
      ...restProps
    } = props;

    const [selectedOptionsCache, setSelectedOptionsCache] = useState<
      Map<IDropdownOption["key"], IDropdownOption>
    >(new Map());

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

    const onChange = useCallback(
      (
        e: React.FormEvent<HTMLDivElement>,
        option?: IDropdownOption,
        idx?: number
      ) => {
        if (option == null) {
          propsOnChange?.(e, option, idx);
          return;
        }
        // In single select mode, option.selected is always undefined
        if (option.selected || option.selected == null) {
          setSelectedOptionsCache((prev) => {
            prev.set(option.key, option);
            return new Map(prev);
          });
        } else {
          setSelectedOptionsCache((prev) => {
            prev.delete(option.key);
            return new Map(prev);
          });
        }
        propsOnChange?.(e, option, idx);
      },
      [propsOnChange]
    );

    const combinedOptions = useMemo(() => {
      const providedOptionKeys = new Set(options.map((o) => o.key));
      const hiddenOptions = Array.from(selectedOptionsCache.entries()).flatMap(
        ([key, option]) => {
          if (providedOptionKeys.has(key)) {
            // If the provided `options` props already has this item,
            // we don't need to inject it to the options list
            return [];
          }
          // Else, include it in the option list as a hidden option
          // This is required for the dropdown to display the selected items correctly
          return [{ ...option, hidden: true }];
        }
      );
      return options.concat(hiddenOptions);
    }, [options, selectedOptionsCache]);

    return (
      <Dropdown
        options={combinedOptions}
        onRenderList={onRenderList}
        onChange={onChange}
        {...restProps}
        calloutProps={{
          calloutMaxHeight: 264,
          calloutMinWidth: 200,
          alignTargetEdge: true,
          ...calloutProps,
        }}
      />
    );
  };
