import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import produce from "immer";
import {
  Checkbox,
  DetailsList,
  IColumn,
  IObjectWithKey,
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
import countryCallingCodeMap from "../../data/countryCodeMap.json";
import { useExactKeywordSearch } from "../../util/search";

import styles from "./AuthenticationCountryCallingCodeList.module.scss";

interface CountryCallingCodeListProps {
  allCountryCallingCodes: string[];
  pinnedCountryCallingCodes: string[];
  selectedCountryCallingCodes: string[];
  onChange: (newPinnedCodes: string[], newSelectedCodes: string[]) => void;
}

interface CountryCallingCodeListItem extends IObjectWithKey {
  key: string;
  selected: boolean;
  pinned: boolean;
  countryName: string;
  callingCode: string;
}

type CountryCallingCodeListData = Record<
  string,
  {
    key: string;
    countryName: string;
    callingCode: string;
  }
>;

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
      name: renderToString(
        "AuthenticationLoginIDSettingsScreen.phone.columns.active"
      ),
      minWidth: 90,
      maxWidth: 90,
      className: styles.callingCodeListColumn,
    },
    {
      key: "countryName",
      fieldName: "countryName",
      name: renderToString(
        "AuthenticationLoginIDSettingsScreen.phone.columns.country-or-area"
      ),
      minWidth: 180,
      maxWidth: 180,
      isMultiline: true,
      className: cn(styles.countryNameCell, styles.callingCodeListColumn),
    },
    {
      key: "callingCode",
      fieldName: "callingCode",
      name: renderToString(
        "AuthenticationLoginIDSettingsScreen.phone.columns.code"
      ),
      minWidth: 65,
      maxWidth: 65,
      className: styles.callingCodeListColumn,
    },
    {
      key: "order",
      name: renderToString(
        "AuthenticationLoginIDSettingsScreen.phone.columns.order"
      ),
      fieldName: "order",
      minWidth: 140,
      maxWidth: 140,
      className: styles.callingCodeListColumn,
    },
    {
      key: "pinned",
      name: renderToString(
        "AuthenticationLoginIDSettingsScreen.phone.columns.pinned"
      ),
      fieldName: "pinned",
      minWidth: 140,
      maxWidth: 140,
      className: styles.callingCodeListColumn,
    },
  ];
}

// asusume country calling code data has no duplicated code
function constructCallingCodeListData(
  allCountryCallingCodes: string[],
  getTelecomCountryName: (code: string) => string
): CountryCallingCodeListData {
  const callingCodeListData: CountryCallingCodeListData = {};
  for (const callingCode of allCountryCallingCodes) {
    const countryCodes =
      countryCallingCodeMap[callingCode as keyof typeof countryCallingCodeMap];
    const countryName =
      countryCodes.length > 0 ? getTelecomCountryName(countryCodes[0]) : "";

    callingCodeListData[callingCode] = {
      key: callingCode,
      countryName,
      callingCode,
    };
  }
  return callingCodeListData;
}

function constructCallingCodeListItem(
  selectedCountryCallingCodes: string[],
  pinnedCountryCallingCodes: string[],
  callingCodeListData: CountryCallingCodeListData,
  matchedCallingCodes: {
    key: string;
    countryName: string;
    callingCode: string;
  }[]
): CountryCallingCodeListItem[] {
  const pinnedCountryCallingCodeSet = new Set(pinnedCountryCallingCodes);
  const selectedCountryCallingCodesSet = new Set(selectedCountryCallingCodes);

  const rawUnpinnedCodeList: string[] = matchedCallingCodes
    .filter((item) => !pinnedCountryCallingCodeSet.has(item.callingCode))
    .map((item) => item.callingCode);

  const unpinnedCodeList = rawUnpinnedCodeList.sort(
    (code1, code2) => Number(code1) - Number(code2)
  );

  const codeList = pinnedCountryCallingCodes.concat(unpinnedCodeList);

  return codeList.map((callingCode) => ({
    ...callingCodeListData[callingCode],
    pinned: pinnedCountryCallingCodeSet.has(callingCode),
    selected: selectedCountryCallingCodesSet.has(callingCode),
  }));
}

function getModifiedItem(
  countryCallingCodeList: CountryCallingCodeListItem[],
  index: number
): CountryCallingCodeListItem | null {
  if (!(index >= 0 && index < countryCallingCodeList.length)) {
    return null;
  }
  return countryCallingCodeList[index];
}

function updateCountryCallingCodeList(
  codes: string[],
  targetCode: string,
  checked: boolean
) {
  return produce(codes, (draftCodes) => {
    const targetIndex = codes.findIndex(
      (callingCode) => callingCode === targetCode
    );
    if (checked && targetIndex === -1) {
      draftCodes.push(targetCode);
    }
    if (!checked && targetIndex > -1) {
      draftCodes.splice(targetIndex, 1);
    }
  });
}

const CountryCallingCodeListItemCheckbox: React.FC<CountryCallingCodeListItemCheckboxProps> = function CountryCallingCodeListItemCheckbox(
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

const CountryCallingCodeListPinButton: React.FC<CountryCallingCodeListPinButtonProps> = function CountryCallingCodeListPinButton(
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

const CountryCallingCodeListSelectAll: React.FC<CountryCallingCodeListSelectAllProps> = function CountryCallingCodeListSelectAll(
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

const CountryCallingCodeList: React.FC<CountryCallingCodeListProps> = function CountryCallingCodeList(
  props: CountryCallingCodeListProps
) {
  const {
    allCountryCallingCodes,
    pinnedCountryCallingCodes,
    selectedCountryCallingCodes,
    onChange,
  } = props;
  const { renderToString } = useContext(Context);
  const { getTelecomCountryName } = useGetTelecomCountryName();

  const [searchString, setSearchString] = useState("");
  const { onChange: onSearchBoxChange } = useTextField((value) => {
    setSearchString(value);
  });

  const countryCallingCodeListData = useMemo(() => {
    return constructCallingCodeListData(
      allCountryCallingCodes,
      getTelecomCountryName
    );
  }, [allCountryCallingCodes, getTelecomCountryName]);

  const dataList = useMemo(() => {
    return allCountryCallingCodes.map(
      (callingCode) => countryCallingCodeListData[callingCode]
    );
  }, [countryCallingCodeListData, allCountryCallingCodes]);

  const { search } = useExactKeywordSearch(dataList, [
    "callingCode",
    "countryName",
  ]);

  const countryCodeListColumns = useMemo(
    () => makeCountryCodeListColumns(renderToString),
    [renderToString]
  );

  const isCallingCodePartiallySelected = useMemo(() => {
    return (
      selectedCountryCallingCodes.length > 0 &&
      selectedCountryCallingCodes.length < allCountryCallingCodes.length
    );
  }, [selectedCountryCallingCodes, allCountryCallingCodes]);

  const isCallingCodeAllSelected = useMemo(() => {
    return selectedCountryCallingCodes.length === allCountryCallingCodes.length;
  }, [selectedCountryCallingCodes, allCountryCallingCodes]);

  const countryCallingCodeList: CountryCallingCodeListItem[] = useMemo(() => {
    const matchedCallingCodes = search(searchString);
    return constructCallingCodeListItem(
      selectedCountryCallingCodes,
      pinnedCountryCallingCodes,
      countryCallingCodeListData,
      matchedCallingCodes
    );
  }, [
    pinnedCountryCallingCodes,
    selectedCountryCallingCodes,
    countryCallingCodeListData,
    searchString,
    search,
  ]);

  const onCallingCodeSwap = useCallback(
    (index1: number, index2: number) => {
      onChange(
        selectedCountryCallingCodes,
        swap(pinnedCountryCallingCodes, index1, index2)
      );
    },
    [onChange, selectedCountryCallingCodes, pinnedCountryCallingCodes]
  );

  // NOTE: pinned code must be selected
  // if unselected code is pinned, select the code
  const onCallingCodePinned = useCallback(
    (index: number, pinned: boolean) => {
      const modifiedItem = getModifiedItem(countryCallingCodeList, index);
      if (modifiedItem == null) {
        return;
      }

      const pinnedCodes = updateCountryCallingCodeList(
        pinnedCountryCallingCodes,
        modifiedItem.callingCode,
        pinned
      );
      let selectedCodes = selectedCountryCallingCodes;
      if (pinned && !modifiedItem.selected) {
        selectedCodes = updateCountryCallingCodeList(
          selectedCountryCallingCodes,
          modifiedItem.callingCode,
          true
        );
      }

      onChange(selectedCodes, pinnedCodes);
    },
    [
      countryCallingCodeList,
      onChange,
      pinnedCountryCallingCodes,
      selectedCountryCallingCodes,
    ]
  );

  // NOTE: pinned code must be selected
  // if pinned code is deselected, unpin the code
  const onCallingCodeSelected = useCallback(
    (index: number, selected: boolean) => {
      const modifiedItem = getModifiedItem(countryCallingCodeList, index);
      if (modifiedItem == null) {
        return;
      }

      const selectedCodes = updateCountryCallingCodeList(
        selectedCountryCallingCodes,
        modifiedItem.callingCode,
        selected
      );
      let pinnedCodes = pinnedCountryCallingCodes;
      if (!selected && modifiedItem.pinned) {
        pinnedCodes = updateCountryCallingCodeList(
          pinnedCountryCallingCodes,
          modifiedItem.callingCode,
          false
        );
      }
      onChange(selectedCodes, pinnedCodes);
    },
    [
      countryCallingCodeList,
      onChange,
      selectedCountryCallingCodes,
      pinnedCountryCallingCodes,
    ]
  );

  const selectAllCallingCode = useCallback(() => {
    onChange(
      countryCallingCodeList.map((item) => item.callingCode),
      pinnedCountryCallingCodes
    );
  }, [onChange, countryCallingCodeList, pinnedCountryCallingCodes]);

  const unselectAllCallingCode = useCallback(() => {
    onChange([], []);
  }, [onChange]);

  const onRenderCallingCodeItemColumn = React.useCallback(
    (item?: CountryCallingCodeListItem, index?: number, column?: IColumn) => {
      switch (column?.key) {
        case "selected":
          return (
            <CountryCallingCodeListItemCheckbox
              index={index}
              checked={item?.selected}
              onCheckboxClicked={onCallingCodeSelected}
            />
          );
        case "order":
          if (item?.pinned) {
            return (
              <OrderButtons
                index={index}
                itemCount={pinnedCountryCallingCodes.length}
                onSwapClicked={onCallingCodeSwap}
                renderAriaLabel={() => item.countryName}
              />
            );
          }
          return (
            <span>
              <FormattedMessage id="AuthenticationLoginIDSettingsScreen.phone.default-order" />
            </span>
          );
        case "pinned":
          return (
            <CountryCallingCodeListPinButton
              index={index}
              className={styles.pin}
              pinned={item?.pinned ?? false}
              onPinClick={onCallingCodePinned}
            />
          );
        case "countryName":
          return <span>{item?.countryName}</span>;
        case "callingCode":
          return <span>{item?.callingCode}</span>;
        default:
          return null;
      }
    },
    [
      onCallingCodeSwap,
      pinnedCountryCallingCodes,
      onCallingCodePinned,
      onCallingCodeSelected,
    ]
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
            selectAll={selectAllCallingCode}
            unselectAll={unselectAllCallingCode}
            isPartiallySelected={isCallingCodePartiallySelected}
            isAllSelected={isCallingCodeAllSelected}
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
    [
      selectAllCallingCode,
      unselectAllCallingCode,
      isCallingCodePartiallySelected,
      isCallingCodeAllSelected,
    ]
  );

  const onRenderCallingCodeListRow = useCallback<
    IRenderFunction<IDetailsRowProps>
  >(
    (props) => {
      if (props == null) {
        return null;
      }
      const { itemIndex } = props;
      const isLastPinnedRow =
        itemIndex === pinnedCountryCallingCodes.length - 1;
      return (
        <DetailsRow
          {...props}
          className={cn(styles.callingCodeListRow, {
            [styles.lastPinnedCallingCode]: isLastPinnedRow,
          })}
        />
      );
    },
    [pinnedCountryCallingCodes]
  );

  return (
    <section className={styles.root}>
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
            items={countryCallingCodeList}
            selectionMode={SelectionMode.none}
            onRenderItemColumn={onRenderCallingCodeItemColumn}
            onRenderDetailsHeader={onRenderCallingCodeListHeader}
            onRenderRow={onRenderCallingCodeListRow}
            checkboxVisibility={CheckboxVisibility.always}
          />
        </ScrollablePane>
      </div>
    </section>
  );
};

export default CountryCallingCodeList;
