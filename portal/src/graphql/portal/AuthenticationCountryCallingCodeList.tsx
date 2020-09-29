import React, { useCallback, useContext, useMemo } from "react";
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
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { OrderColumnButtons, swap } from "../../DetailsListWithOrdering";
import { useGetTelecomCountryName } from "../../util/translations";
import countryCallingCodeMap from "../../data/countryCodeMap.json";

import styles from "./AuthenticationCountryCallingCodeList.module.scss";

interface CountryCallingCodeListProps {
  allCountryCallingCodes: string[];
  selectedCountryCallingCodes: string[];
  onSelectedCountryCallingCodesChange: (newSelectedCodes: string[]) => void;
}

interface CountryCallingCodeListItem extends IObjectWithKey {
  key: string;
  selected: boolean;
  countryName: string;
  callingCode: string;
}

interface CountryCallingCodeListItemCheckboxProps extends ICheckboxProps {
  index?: number;
  onCheckboxClicked: (index: number, checked: boolean) => void;
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
    opacity: "1 !important",
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
      name: renderToString("AuthenticationWidget.phone.list-header.active"),
      minWidth: 90,
      maxWidth: 90,
      className: styles.callingCodeListColumn,
    },
    {
      key: "countryName",
      fieldName: "countryName",
      name: renderToString(
        "AuthenticationWidget.phone.list-header.country-or-area"
      ),
      minWidth: 240,
      maxWidth: 240,
      isMultiline: true,
      className: cn(styles.countryNameCell, styles.callingCodeListColumn),
    },
    {
      key: "callingCode",
      fieldName: "callingCode",
      name: renderToString("AuthenticationWidget.phone.list-header.code"),
      minWidth: 65,
      maxWidth: 65,
      className: styles.callingCodeListColumn,
    },
    {
      key: "order",
      name: renderToString("AuthenticationWidget.phone.list-header.order"),
      fieldName: "order",
      minWidth: 175,
      maxWidth: 175,
      className: styles.callingCodeListColumn,
    },
  ];
}

// asusume country calling code data has no duplicated code
function constructCallingCodeListItem(
  allCountryCallingCodes: string[],
  selectedCountryCallingCodes: string[],
  getTelecomCountryName: (code: string) => string
): CountryCallingCodeListItem[] {
  const selectedCountryCallingCodeSet = new Set(selectedCountryCallingCodes);

  const unselectedCountryCallingCodes: string[] = [];
  const callingCodeListItemMap: Record<
    string,
    { key: string; selected: boolean; countryName: string; callingCode: string }
  > = {};
  for (const callingCode of allCountryCallingCodes) {
    const countryCodes =
      countryCallingCodeMap[callingCode as keyof typeof countryCallingCodeMap];
    const countryName = countryCodes
      .map((code) => getTelecomCountryName(code))
      .join(", ");

    const selected = selectedCountryCallingCodeSet.has(callingCode);

    callingCodeListItemMap[callingCode] = {
      key: callingCode,
      selected,
      countryName,
      callingCode,
    };

    if (!selected) {
      unselectedCountryCallingCodes.push(callingCode);
    }
  }

  const selectedItems = selectedCountryCallingCodes.map(
    (callingCode) => callingCodeListItemMap[callingCode]
  );
  const unselectedItems = unselectedCountryCallingCodes
    .sort((code1, code2) => (Number(code1) > Number(code2) ? 1 : -1))
    .map((callingCode) => callingCodeListItemMap[callingCode]);

  return selectedItems.concat(unselectedItems);
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
    selectedCountryCallingCodes,
    onSelectedCountryCallingCodesChange,
  } = props;
  const { renderToString } = useContext(Context);
  const { getTelecomCountryName } = useGetTelecomCountryName();

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
    return constructCallingCodeListItem(
      allCountryCallingCodes,
      selectedCountryCallingCodes,
      getTelecomCountryName
    );
  }, [
    allCountryCallingCodes,
    selectedCountryCallingCodes,
    getTelecomCountryName,
  ]);

  const onCallingCodeSwap = useCallback(
    (index1: number, index2: number) => {
      onSelectedCountryCallingCodesChange(
        swap(selectedCountryCallingCodes, index1, index2)
      );
    },
    [onSelectedCountryCallingCodesChange, selectedCountryCallingCodes]
  );

  const onCallingCodeSelected = useCallback(
    (index: number, selected: boolean) => {
      const newSelectedCallingCodes = produce(
        selectedCountryCallingCodes,
        (draftSelectedCallingCodes) => {
          if (!(index >= 0 && index < countryCallingCodeList.length)) {
            return;
          }
          const modifiedItem = countryCallingCodeList[index];
          const targetIndex = draftSelectedCallingCodes.findIndex(
            (callingCode) => callingCode === modifiedItem.callingCode
          );
          if (selected && targetIndex === -1) {
            draftSelectedCallingCodes.push(modifiedItem.callingCode);
          }
          if (!selected && targetIndex > -1) {
            draftSelectedCallingCodes.splice(targetIndex, 1);
          }
        }
      );

      onSelectedCountryCallingCodesChange(newSelectedCallingCodes);
    },
    [
      countryCallingCodeList,
      onSelectedCountryCallingCodesChange,
      selectedCountryCallingCodes,
    ]
  );

  const selectAllCallingCode = useCallback(() => {
    onSelectedCountryCallingCodesChange(
      countryCallingCodeList.map((item) => item.callingCode)
    );
  }, [countryCallingCodeList, onSelectedCountryCallingCodesChange]);

  const unselectAllCallingCode = useCallback(() => {
    onSelectedCountryCallingCodesChange([]);
  }, [onSelectedCountryCallingCodesChange]);

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
          if (item?.selected) {
            return (
              <OrderColumnButtons
                index={index}
                itemCount={selectedCountryCallingCodes.length}
                onSwapClicked={onCallingCodeSwap}
                renderAriaLabel={() => item.countryName}
              />
            );
          }
          return (
            <span>
              <FormattedMessage id="AuthenticationWidget.phone.default-order" />
            </span>
          );
        case "countryName":
          return <span>{item?.countryName}</span>;
        case "callingCode":
          return <span>{item?.callingCode}</span>;
        default:
          return <span>{item?.callingCode}</span>;
      }
    },
    [onCallingCodeSwap, selectedCountryCallingCodes, onCallingCodeSelected]
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
        <DetailsHeader
          {...props}
          columns={modifiedColumns}
          onRenderDetailsCheckbox={renderCheckbox}
          selectAllVisibility={SelectAllVisibility.visible}
          styles={HEADER_STYLE}
        />
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
      const isLastSelectedRow =
        itemIndex === selectedCountryCallingCodes.length - 1;
      return (
        <DetailsRow
          {...props}
          className={cn(styles.callingCodeListRow, {
            [styles.lastSelectedCallingCode]: isLastSelectedRow,
          })}
        />
      );
    },
    [selectedCountryCallingCodes]
  );

  return (
    <section className={styles.root}>
      <DetailsList
        columns={countryCodeListColumns}
        items={countryCallingCodeList}
        selectionMode={SelectionMode.none}
        onRenderItemColumn={onRenderCallingCodeItemColumn}
        onRenderDetailsHeader={onRenderCallingCodeListHeader}
        onRenderRow={onRenderCallingCodeListRow}
      />
    </section>
  );
};

export default CountryCallingCodeList;
