import {
  Dropdown,
  IDropdownOption,
  IDropdownProps,
  IconButton,
  SearchBox,
  Spinner,
  SpinnerSize,
  Text,
} from "@fluentui/react";
import React, { useCallback, useContext, useMemo } from "react";
import { Context as MessageContext } from "@oursky/react-messageformat";

import styles from "./SearchableDropdown.module.css";

interface SearchableDropdownProps
  extends Omit<
    IDropdownProps,
    | "defaultSelectedKey"
    | "selectedKey"
    | "defaultSelectedKeys"
    | "selectedKeys"
  > {
  isLoadingOptions?: boolean;
  onSearchValueChange?: (value: string) => void;
  searchValue?: string;
  searchPlaceholder?: string;
  selectedItem?: IDropdownOption | null;
  selectedItems?: IDropdownOption[];
  optionsEmptyMessage?: React.ReactNode;
  onClear?: () => void;
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

function EmptyView(props: { message?: React.ReactNode }) {
  const { message } = props;
  const { renderToString } = useContext(MessageContext);

  return (
    <Text block={true} className={styles.emptyView}>
      {message ?? renderToString("SearchableDropdown.empty")}
    </Text>
  );
}

const ClearButton = React.memo(function ClearButton(props: {
  onClick: React.MouseEventHandler<HTMLButtonElement>;
}) {
  const { onClick } = props;
  return (
    <IconButton
      onClick={onClick}
      styles={{
        root: {
          right: -7,
          width: 30,
          height: 30,
        },
      }}
      iconProps={{
        iconName: "Clear",
        styles: {
          root: {
            color: "#605E5C",
            fontSize: 12,
          },
        },
      }}
    />
  );
});

const EMPTY_CALLOUT_PROPS: IDropdownProps["calloutProps"] = {};

export const SearchableDropdown: React.VFC<SearchableDropdownProps> =
  function SearchableDropdown(props) {
    const {
      options,
      isLoadingOptions,
      onSearchValueChange,
      searchValue,
      searchPlaceholder,
      calloutProps = EMPTY_CALLOUT_PROPS,
      selectedItem,
      selectedItems,
      optionsEmptyMessage,
      onClear,
      ...restProps
    } = props;

    const onRenderList = useCallback<
      NonNullable<IDropdownProps["onRenderList"]>
    >(
      (props?, defaultRenderer?) => {
        if (defaultRenderer == null) {
          return null;
        }

        const isOptionsEmpty = props?.options?.length === 0;

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
            ) : isOptionsEmpty ? (
              <EmptyView message={optionsEmptyMessage} />
            ) : (
              defaultRenderer(props)
            )}
          </>
        );
      },
      [
        isLoadingOptions,
        onSearchValueChange,
        optionsEmptyMessage,
        searchPlaceholder,
        searchValue,
      ]
    );

    const onClearButtonClick = useCallback(
      (e: React.MouseEvent<HTMLButtonElement>) => {
        e.stopPropagation();
        e.preventDefault();
        onClear?.();
      },
      [onClear]
    );

    const onRenderCaretDown = useCallback<
      NonNullable<IDropdownProps["onRenderCaretDown"]>
    >(
      (props, defaultRenderer) => {
        if (
          selectedItem != null ||
          (selectedItems && selectedItems.length > 0)
        ) {
          return <ClearButton onClick={onClearButtonClick} />;
        }
        return defaultRenderer?.(props) ?? <></>;
      },
      [onClearButtonClick, selectedItem, selectedItems]
    );

    const combinedOptions = useMemo(() => {
      const providedOptionKeys = new Set(options.map((o) => o.key));

      // Include all selected items as hidden options, if they are not in `options`.
      // This is needed for the dropdown to display selected options correctly.
      const hiddenOptions: IDropdownOption[] = [];
      if (selectedItem && !providedOptionKeys.has(selectedItem.key)) {
        hiddenOptions.push({ ...selectedItem, hidden: true });
      }
      if (selectedItems) {
        for (const item of selectedItems) {
          if (!providedOptionKeys.has(item.key)) {
            hiddenOptions.push({ ...item, hidden: true });
          }
        }
      }
      return options.concat(hiddenOptions);
    }, [options, selectedItem, selectedItems]);

    const selectedKey = useMemo(() => {
      if (selectedItem === null) {
        return null;
      }
      return selectedItem?.key;
    }, [selectedItem]);

    const selectedKeys = useMemo(() => {
      return selectedItems?.map((item) => item.key);
    }, [selectedItems]);

    return (
      <Dropdown
        options={combinedOptions}
        onRenderList={onRenderList}
        onRenderCaretDown={onRenderCaretDown}
        {...restProps}
        calloutProps={{
          calloutMaxHeight: 264,
          calloutMinWidth: 200,
          alignTargetEdge: true,
          ...calloutProps,
        }}
        selectedKey={selectedKey}
        selectedKeys={selectedKeys as IDropdownProps["selectedKeys"]}
      />
    );
  };
