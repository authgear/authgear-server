import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { produce } from "immer";
import { Checkbox, IconButton, Text } from "@radix-ui/themes";
import {
  Cross2Icon,
  DrawingPinFilledIcon,
  DrawingPinIcon,
} from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import OrderButtons, { swap } from "../../OrderButtons";
import { useGetTelecomCountryName } from "../../util/translations";
import ALL_COUNTRIES from "../../data/country.json";
import { useExactKeywordSearch } from "../../util/search";
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";

import styles from "./AuthenticationCountryCallingCodeList.module.css";

export interface CountryCallingCodeListProps {
  className?: string;
  title?: React.ReactNode;
  pinnedAlpha2: string[];
  allowedAlpha2: string[];
  featureAllowlist?: string[];
  onChange: (newAllowedCodes: string[], newPinnedCodes: string[]) => void;
  disabled: boolean;
}

interface ListItem {
  key: string;
  selected: boolean;
  pinned: boolean;
  alpha2: string;
  countryCallingCode: string;
  displayName: string;
  disabled: boolean;
}

type Country = (typeof ALL_COUNTRIES)[number];
type CountryMap = Record<string, Country>;

const COUNTRY_MAP: CountryMap = ALL_COUNTRIES.reduce<CountryMap>(
  (acc: CountryMap, currValue: Country) => {
    acc[currValue.Alpha2] = currValue;
    return acc;
  },
  {}
);

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

function PinButton(props: {
  pinned: boolean;
  disabled: boolean;
  onClick: () => void;
}): React.ReactElement {
  const { pinned, disabled, onClick } = props;
  const { renderToString } = useContext(Context);
  const PinIcon = pinned ? DrawingPinFilledIcon : DrawingPinIcon;

  return (
    <IconButton
      type="button"
      className={styles.pinButton}
      variant="ghost"
      color="gray"
      size="1"
      disabled={disabled}
      onClick={onClick}
      aria-label={renderToString(
        "LoginIDConfigurationScreen.phone.columns.pinned"
      )}
      aria-pressed={pinned}
    >
      <PinIcon
        className={cn(styles.pinIcon, pinned && styles.pinIconPinned)}
        width="1rem"
        height="1rem"
      />
    </IconButton>
  );
}

const CountryCallingCodeList: React.VFC<CountryCallingCodeListProps> =
  function CountryCallingCodeList(props: CountryCallingCodeListProps) {
    const {
      disabled,
      className,
      title,
      pinnedAlpha2,
      allowedAlpha2,
      featureAllowlist,
      onChange,
    } = props;
    const { renderToString } = useContext(Context);
    const { getTelecomCountryName } = useGetTelecomCountryName();

    const [searchString, setSearchString] = useState("");

    const onSearchChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchString(e.target.value);
      },
      []
    );

    const onSearchClear = useCallback(() => {
      setSearchString("");
    }, []);

    const allItems: ListItem[] = useMemo(() => {
      const pinned = new Set(pinnedAlpha2);
      const allowed = new Set(allowedAlpha2);
      const featureSet =
        featureAllowlist && featureAllowlist.length > 0
          ? new Set(featureAllowlist)
          : null;

      const lst: ListItem[] = [];
      const enabledItems: ListItem[] = [];
      const disabledItems: ListItem[] = [];

      const makeItem = (alpha2: string): ListItem => {
        const country = COUNTRY_MAP[alpha2];
        const isFeatureAllowed = featureSet ? featureSet.has(alpha2) : true;
        return {
          key: country.Alpha2,
          selected: allowed.has(country.Alpha2),
          pinned: pinned.has(country.Alpha2),
          alpha2: country.Alpha2,
          countryCallingCode: country.CountryCallingCode,
          displayName: getTelecomCountryName(country.Alpha2),
          disabled: disabled || !isFeatureAllowed,
        };
      };

      for (const alpha2 of pinnedAlpha2) {
        lst.push(makeItem(alpha2));
      }

      for (const country of ALL_COUNTRIES) {
        if (pinned.has(country.Alpha2)) {
          continue;
        }

        const item = makeItem(country.Alpha2);
        if (item.disabled) {
          disabledItems.push(item);
        } else {
          enabledItems.push(item);
        }
      }

      lst.push(...enabledItems);
      lst.push(...disabledItems);

      return lst;
    }, [
      disabled,
      allowedAlpha2,
      pinnedAlpha2,
      getTelecomCountryName,
      featureAllowlist,
    ]);

    const { search } = useExactKeywordSearch(allItems, [
      "alpha2",
      "countryCallingCode",
      "displayName",
    ]);

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
        filteredItems.filter((item) => !item.disabled).map((a) => a.alpha2),
        pinnedAlpha2
      );
    }, [onChange, filteredItems, pinnedAlpha2]);

    const unselectAll = useCallback(() => {
      onChange([], []);
    }, [onChange]);

    const onSelectAllCheckedChange = useCallback(
      (checked: boolean | "indeterminate") => {
        if (checked === "indeterminate") {
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
      <div className={cn(styles.root, className)}>
        <div className={styles.toolbar}>
          {title != null ? (
            <Text as="p" size="2" weight="medium" className={styles.toolbarTitle}>
              {title}
            </Text>
          ) : (
            <span />
          )}
          <div className={styles.searchField}>
            <TextField
              size="2"
              type="search"
              placeholder={renderToString("search")}
              value={searchString}
              iconStart={TextFieldIcon.MagnifyingGlass}
              onChange={onSearchChange}
              disabled={disabled}
              suffixPlain={true}
              suffix={
                searchString !== "" ? (
                  <button
                    type="button"
                    className={styles.searchClearButton}
                    aria-label={renderToString(
                      "APIResourcesScreen.clear-search"
                    )}
                    onClick={onSearchClear}
                  >
                    <Cross2Icon className={styles.searchClearIcon} />
                  </button>
                ) : undefined
              }
            />
          </div>
        </div>
        <div className={styles.listWrapper}>
          <div className={styles.table}>
            <div className={styles.tableHeader}>
              <div className={styles.cellSelected}>
                <Checkbox
                  checked={
                    isAllSelected
                      ? true
                      : isPartiallySelected
                        ? "indeterminate"
                        : false
                  }
                  onCheckedChange={onSelectAllCheckedChange}
                  disabled={disabled}
                  aria-label={renderToString("activate")}
                />
              </div>
              <div className={styles.headerCellCountry}>
                <FormattedMessage id="LoginIDConfigurationScreen.phone.columns.country-or-area" />
              </div>
              <div className={styles.headerCellCode}>
                <FormattedMessage id="LoginIDConfigurationScreen.phone.columns.code" />
              </div>
              <div className={styles.headerCellOrder}>
                <FormattedMessage id="LoginIDConfigurationScreen.phone.columns.order" />
              </div>
              <div className={styles.headerCellPinned} aria-hidden={true} />
            </div>
            {filteredItems.map((item, index) => {
              const isLastPinnedRow = index === pinnedAlpha2.length - 1;
              return (
                <div
                  key={item.key}
                  className={cn(
                    styles.tableRow,
                    isLastPinnedRow && styles.lastPinnedCallingCode
                  )}
                >
                  <div className={styles.cellSelected}>
                    <Checkbox
                      checked={item.selected}
                      disabled={item.disabled}
                      onCheckedChange={(checked) => {
                        if (checked === "indeterminate") {
                          return;
                        }
                        onSelect(index, checked);
                      }}
                    />
                  </div>
                  <div className={styles.cellCountry}>
                    <Text size="2">{item.displayName}</Text>
                  </div>
                  <div className={styles.cellCode}>
                    <Text size="2">{item.countryCallingCode}</Text>
                  </div>
                  <div className={styles.cellOrder}>
                    {item.pinned ? (
                      <OrderButtons
                        disabled={item.disabled}
                        index={index}
                        itemCount={pinnedAlpha2.length}
                        onSwapClicked={onSwap}
                      />
                    ) : null}
                  </div>
                  <div className={styles.cellPinned}>
                    <PinButton
                      pinned={item.pinned}
                      disabled={item.disabled}
                      onClick={() => {
                        onPinClick(index, !item.pinned);
                      }}
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      </div>
    );
  };

export default CountryCallingCodeList;
