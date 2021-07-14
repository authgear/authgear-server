import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import produce from "immer";
import {
  Checkbox,
  DetailsList,
  IColumn,
  SelectionMode,
  ICheckboxProps,
  IDetailsHeaderProps,
  IRenderFunction,
  DetailsHeader,
  DetailsRow,
  IDetailsRowProps,
  SelectAllVisibility,
  SearchBox,
  CheckboxVisibility,
  ScrollablePane,
  StickyPositionType,
  Sticky,
  IconButton,
  IIconProps,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useTextField } from "../../hook/useInput";
import OrderButtons, { swap } from "../../OrderButtons";
import { useGetTelecomCountryName } from "../../util/translations";
import ALL_COUNTRIES from "../../data/country.json";
import { useExactKeywordSearch } from "../../util/search";

import styles from "./AuthenticationCountryCallingCodeList.module.scss";

export interface CountryCallingCodeListProps {
  className?: string;
  pinnedAlpha2: string[];
  allowedAlpha2: string[];
  onChange: (newPinnedCodes: string[], newSelectedCodes: string[]) => void;
}

interface ListItem {
  key: string;
  selected: boolean;
  pinned: boolean;
  alpha2: string;
  countryCallingCode: string;
  displayName: string;
}

type Country = typeof ALL_COUNTRIES[number];
type CountryMap = Record<string, Country>;

const COUNTRY_MAP: CountryMap = ALL_COUNTRIES.reduce<CountryMap>(
  (acc: CountryMap, currValue: Country) => {
    acc[currValue.Alpha2] = currValue;
    return acc;
  },
  {}
);

interface CountryCallingCodeListItemCheckboxProps extends ICheckboxProps {
  index?: number;
  onCheckboxClicked: (index: number, checked: boolean) => void;
}

interface CountryCallingCodeListPinButtonProps {
  className?: string;
  index?: number;
  pinned?: boolean;
  onPinClick: (index: number, checked: boolean) => void;
}

interface CountryCallingCodeListSelectAllProps extends ICheckboxProps {
  isPartiallySelected: boolean;
  isAllSelected: boolean;
  selectAll: () => void;
  unselectAll: () => void;
}

const HEADER_STYLE = {
  check: {
    width: "35px !important",
    paddingLeft: "15px !important",
  },
};

function makeCountryCodeListColumns(
  renderToString: (messageId: string) => string
): IColumn[] {
  return [
    {
      key: "selected",
      fieldName: "selected",
      name: renderToString("LoginIDConfigurationScreen.phone.columns.active"),
      minWidth: 90,
      maxWidth: 90,
      className: styles.callingCodeListColumn,
    },
    {
      key: "countryName",
      fieldName: "countryName",
      name: renderToString(
        "LoginIDConfigurationScreen.phone.columns.country-or-area"
      ),
      minWidth: 180,
      maxWidth: 180,
      isMultiline: true,
      className: cn(styles.countryNameCell, styles.callingCodeListColumn),
    },
    {
      key: "callingCode",
      fieldName: "callingCode",
      name: renderToString("LoginIDConfigurationScreen.phone.columns.code"),
      minWidth: 65,
      maxWidth: 65,
      className: styles.callingCodeListColumn,
    },
    {
      key: "order",
      name: renderToString("LoginIDConfigurationScreen.phone.columns.order"),
      fieldName: "order",
      minWidth: 140,
      maxWidth: 140,
      className: styles.callingCodeListColumn,
    },
    {
      key: "pinned",
      name: renderToString("LoginIDConfigurationScreen.phone.columns.pinned"),
      fieldName: "pinned",
      minWidth: 140,
      maxWidth: 140,
      className: styles.callingCodeListColumn,
    },
  ];
}

function indexArrayOrNull<T>(list: T[], index: number): T | null {
  if (index >= 0 && index < list.length) {
    return list[index];
  }
  return null;
}

function edit(values: string[], target: string, checked: boolean): string[] {
  return produce(values, (values) => {
    const index = values.findIndex((a) => a === target);
    if (checked && index < 0) {
      values.push(target);
    }
    if (!checked && index >= 0) {
      values.splice(index, 1);
    }
  });
}

const CountryCallingCodeListItemCheckbox: React.FC<CountryCallingCodeListItemCheckboxProps> =
  function CountryCallingCodeListItemCheckbox(
    props: CountryCallingCodeListItemCheckboxProps
  ) {
    const { onCheckboxClicked, index, ...rest } = props;

    const onChange = useCallback(
      (_event, checked?: boolean) => {
        if (index == null || checked == null) {
          return;
        }
        onCheckboxClicked(index, checked);
      },
      [onCheckboxClicked, index]
    );

    return <Checkbox {...rest} onChange={onChange} />;
  };

const CountryCallingCodeListPinButton: React.FC<CountryCallingCodeListPinButtonProps> =
  function CountryCallingCodeListPinButton(
    props: CountryCallingCodeListPinButtonProps
  ) {
    const { className, index, pinned, onPinClick } = props;

    const iconProps: IIconProps = useMemo(() => {
      const iconName = pinned ? "PinnedSolid" : "Pinned";
      return { iconName };
    }, [pinned]);

    const onButtonClick = useCallback(() => {
      if (index == null || pinned == null) {
        return;
      }
      onPinClick(index, !pinned);
    }, [index, pinned, onPinClick]);

    return (
      <IconButton
        className={className}
        iconProps={iconProps}
        onClick={onButtonClick}
      />
    );
  };

const CountryCallingCodeListSelectAll: React.FC<CountryCallingCodeListSelectAllProps> =
  function CountryCallingCodeListSelectAll(
    props: CountryCallingCodeListSelectAllProps
  ) {
    const {
      isPartiallySelected,
      isAllSelected,
      selectAll,
      unselectAll,
      ...rest
    } = props;

    const onChange = useCallback(
      (_event, checked?: boolean) => {
        if (checked == null) {
          return;
        }
        if (checked) {
          selectAll();
        } else {
          unselectAll();
        }
      },
      [selectAll, unselectAll]
    );

    return (
      <Checkbox
        {...rest}
        indeterminate={isPartiallySelected}
        checked={isAllSelected}
        onChange={onChange}
      />
    );
  };

const CountryCallingCodeList: React.FC<CountryCallingCodeListProps> =
  function CountryCallingCodeList(props: CountryCallingCodeListProps) {
    const { className, pinnedAlpha2, allowedAlpha2, onChange } = props;
    const { renderToString } = useContext(Context);
    const { getTelecomCountryName } = useGetTelecomCountryName();

    const [searchString, setSearchString] = useState("");
    const { onChange: onSearchBoxChange } = useTextField((value) => {
      setSearchString(value);
    });

    const allItems: ListItem[] = useMemo(() => {
      const pinned = new Set(pinnedAlpha2);
      const allowed = new Set(allowedAlpha2);

      const lst = [];

      for (const alpha2 of pinnedAlpha2) {
        const country = COUNTRY_MAP[alpha2];
        lst.push({
          key: country.Alpha2,
          selected: allowed.has(country.Alpha2),
          pinned: pinned.has(country.Alpha2),
          alpha2: country.Alpha2,
          countryCallingCode: country.CountryCallingCode,
          displayName: getTelecomCountryName(country.Alpha2),
        });
      }

      for (const country of ALL_COUNTRIES) {
        if (pinned.has(country.Alpha2)) {
          continue;
        }

        lst.push({
          key: country.Alpha2,
          selected: allowed.has(country.Alpha2),
          pinned: pinned.has(country.Alpha2),
          alpha2: country.Alpha2,
          countryCallingCode: country.CountryCallingCode,
          displayName: getTelecomCountryName(country.Alpha2),
        });
      }

      return lst;
    }, [allowedAlpha2, pinnedAlpha2, getTelecomCountryName]);

    const { search } = useExactKeywordSearch(allItems, [
      "alpha2",
      "countryCallingCode",
      "displayName",
    ]);

    const countryCodeListColumns = useMemo(
      () => makeCountryCodeListColumns(renderToString),
      [renderToString]
    );

    const isPartiallySelected = useMemo(() => {
      return (
        allowedAlpha2.length > 0 && allowedAlpha2.length < ALL_COUNTRIES.length
      );
    }, [allowedAlpha2]);

    const isAllSelected = useMemo(() => {
      return allowedAlpha2.length === ALL_COUNTRIES.length;
    }, [allowedAlpha2]);

    const filteredItems: ListItem[] = useMemo(() => {
      return search(searchString);
    }, [search, searchString]);

    const onSwap = useCallback(
      (index1: number, index2: number) => {
        onChange(allowedAlpha2, swap(pinnedAlpha2, index1, index2));
      },
      [onChange, allowedAlpha2, pinnedAlpha2]
    );

    // NOTE: pinned code must be selected
    // if unselected code is pinned, select the code
    const onPinClick = useCallback(
      (index: number, pinned: boolean) => {
        const modifiedItem = indexArrayOrNull(filteredItems, index);
        if (modifiedItem == null) {
          return;
        }

        const newPinned = edit(pinnedAlpha2, modifiedItem.alpha2, pinned);

        let newAllowed = allowedAlpha2;
        if (pinned && !modifiedItem.selected) {
          newAllowed = edit(allowedAlpha2, modifiedItem.alpha2, true);
        }

        onChange(newAllowed, newPinned);
      },
      [onChange, filteredItems, pinnedAlpha2, allowedAlpha2]
    );

    // NOTE: pinned code must be selected
    // if pinned code is deselected, unpin the code
    const onSelect = useCallback(
      (index: number, selected: boolean) => {
        const modifiedItem = indexArrayOrNull(filteredItems, index);
        if (modifiedItem == null) {
          return;
        }

        const newAllowed = edit(allowedAlpha2, modifiedItem.alpha2, selected);

        let newPinned = pinnedAlpha2;
        if (!selected && modifiedItem.pinned) {
          newPinned = edit(pinnedAlpha2, modifiedItem.alpha2, false);
        }

        onChange(newAllowed, newPinned);
      },
      [onChange, filteredItems, pinnedAlpha2, allowedAlpha2]
    );

    const selectAll = useCallback(() => {
      onChange(
        filteredItems.map((a) => a.alpha2),
        pinnedAlpha2
      );
    }, [onChange, filteredItems, pinnedAlpha2]);

    const unselectAll = useCallback(() => {
      onChange([], []);
    }, [onChange]);

    const onRenderCallingCodeItemColumn = React.useCallback(
      (item?: ListItem, index?: number, column?: IColumn) => {
        switch (column?.key) {
          case "selected":
            return (
              <CountryCallingCodeListItemCheckbox
                index={index}
                checked={item?.selected}
                onCheckboxClicked={onSelect}
              />
            );
          case "order":
            if (item?.pinned) {
              return (
                <OrderButtons
                  index={index}
                  itemCount={pinnedAlpha2.length}
                  onSwapClicked={onSwap}
                  renderAriaLabel={() => item.displayName}
                />
              );
            }
            return (
              <span>
                <FormattedMessage id="LoginIDConfigurationScreen.phone.default-order" />
              </span>
            );
          case "pinned":
            return (
              <CountryCallingCodeListPinButton
                index={index}
                className={styles.pin}
                pinned={item?.pinned ?? false}
                onPinClick={onPinClick}
              />
            );
          case "countryName":
            return <span>{item?.displayName}</span>;
          case "callingCode":
            return <span>{item?.countryCallingCode}</span>;
          default:
            return null;
        }
      },
      [onSwap, pinnedAlpha2, onPinClick, onSelect]
    );

    const onRenderCallingCodeListHeader = useCallback<
      IRenderFunction<IDetailsHeaderProps>
    >(
      (props) => {
        if (props == null) {
          return null;
        }
        const renderCheckbox = () => {
          return (
            <CountryCallingCodeListSelectAll
              selectAll={selectAll}
              unselectAll={unselectAll}
              isPartiallySelected={isPartiallySelected}
              isAllSelected={isAllSelected}
            />
          );
        };

        // modify column width for select all checkbox
        const modifiedColumns = produce(props.columns, (draftColumn) => {
          const activeColumnWidth = draftColumn[0].calculatedWidth!;
          draftColumn[0].calculatedWidth = activeColumnWidth - 35;
        });

        return (
          <Sticky stickyPosition={StickyPositionType.Header}>
            <DetailsHeader
              {...props}
              columns={modifiedColumns}
              onRenderDetailsCheckbox={renderCheckbox}
              selectAllVisibility={SelectAllVisibility.visible}
              styles={HEADER_STYLE}
            />
          </Sticky>
        );
      },
      [selectAll, unselectAll, isPartiallySelected, isAllSelected]
    );

    const onRenderCallingCodeListRow = useCallback<
      IRenderFunction<IDetailsRowProps>
    >(
      (props) => {
        if (props == null) {
          return null;
        }
        const { itemIndex } = props;
        const isLastPinnedRow = itemIndex === pinnedAlpha2.length - 1;
        return (
          <DetailsRow
            {...props}
            className={cn(styles.callingCodeListRow, {
              [styles.lastPinnedCallingCode]: isLastPinnedRow,
            })}
          />
        );
      },
      [pinnedAlpha2]
    );

    return (
      <div className={className}>
        <SearchBox
          className={styles.searchBox}
          placeholder={renderToString("search")}
          onChange={onSearchBoxChange}
        />
        <div className={styles.listWrapper}>
          <ScrollablePane>
            <DetailsList
              className={styles.detailsList}
              columns={countryCodeListColumns}
              items={filteredItems}
              selectionMode={SelectionMode.none}
              onRenderItemColumn={onRenderCallingCodeItemColumn}
              onRenderDetailsHeader={onRenderCallingCodeListHeader}
              onRenderRow={onRenderCallingCodeListRow}
              checkboxVisibility={CheckboxVisibility.always}
            />
          </ScrollablePane>
        </div>
      </div>
    );
  };

export default CountryCallingCodeList;
