/* global JSX */
import React, { useMemo, useCallback, useContext, useState } from "react";
import { FormattedMessage, Context } from "./intl";
import {
  DetailsList,
  DetailsHeader,
  DetailsRow,
  DirectionalHint,
  Dropdown,
  Dialog,
  DialogFooter,
  IconButton,
  SelectionMode,
  IColumn,
  IDropdownOption,
  IDialogContentProps,
  IDetailsHeaderProps,
  IDetailsRowProps,
  IDetailsColumnRenderTooltipProps,
  IRenderFunction,
  IIconProps,
  IDragDropEvents,
  Icon,
  Text,
} from "@fluentui/react";
import PrimaryButton from "./PrimaryButton";
import DefaultButton from "./DefaultButton";
import LabelWithTooltip from "./LabelWithTooltip";
import {
  UserProfileAttributesAccessControl,
  AccessControlLevelString,
} from "./types";
import { parseJSONPointer } from "./util/jsonpointer";
import styles from "./UserProfileAttributesList.module.css";

export type UserProfileAttributesListAccessControlAdjustment = [
  keyof UserProfileAttributesAccessControl,
  AccessControlLevelString
];

export interface UserProfileAttributesListItem {
  pointer: string;
  access_control: UserProfileAttributesAccessControl;
}

export interface ItemComponentProps<T> {
  className: string;
  item: T;
}

export interface UserProfileAttributesListProps<
  T extends UserProfileAttributesListItem
> {
  items: T[];
  ItemComponent: React.ComponentType<ItemComponentProps<T>>;
  onChangeItems: (items: T[]) => void;
  onEditButtonClick?: (index: number) => void;
  onReorderItems?: (items: T[]) => void;
}

export interface UserProfileAttributesListPendingUpdate {
  index: number;
  key: keyof UserProfileAttributesAccessControl;
  mainAdjustment: UserProfileAttributesListAccessControlAdjustment;
  otherAdjustments: UserProfileAttributesListAccessControlAdjustment[];
}

const EDIT_BUTTON_ICON_PROPS: IIconProps = {
  iconName: "Edit",
};

function intOfAccessControlLevelString(
  level: AccessControlLevelString
): number {
  switch (level) {
    case "hidden":
      return 1;
    case "readonly":
      return 2;
    case "readwrite":
      return 3;
    default:
      throw new Error("unknown value: " + String(level));
  }
}

type AccessControlAdjuster = (
  accessControl: UserProfileAttributesAccessControl,
  target: keyof UserProfileAttributesAccessControl,
  level: AccessControlLevelString
) => UserProfileAttributesListAccessControlAdjustment | undefined;

function atLeast(
  accessControl: UserProfileAttributesAccessControl,
  target: keyof UserProfileAttributesAccessControl,
  level: AccessControlLevelString
): UserProfileAttributesListAccessControlAdjustment | undefined {
  const targetLevelInt = intOfAccessControlLevelString(accessControl[target]);
  const levelInt = intOfAccessControlLevelString(level);
  if (targetLevelInt < levelInt) {
    return [target, level];
  }
  return undefined;
}

function atMost(
  accessControl: UserProfileAttributesAccessControl,
  target: keyof UserProfileAttributesAccessControl,
  level: AccessControlLevelString
): UserProfileAttributesListAccessControlAdjustment | undefined {
  const targetLevelInt = intOfAccessControlLevelString(accessControl[target]);
  const levelInt = intOfAccessControlLevelString(level);
  if (targetLevelInt > levelInt) {
    return [target, level];
  }
  return undefined;
}

function makeUpdate(
  prevItems: UserProfileAttributesListItem[],
  index: number,
  key: keyof UserProfileAttributesAccessControl,
  newValue: AccessControlLevelString
): UserProfileAttributesListPendingUpdate {
  const accessControl = prevItems[index].access_control;

  const mainAdjustment: UserProfileAttributesListAccessControlAdjustment = [
    key,
    newValue,
  ];

  const adjustments: ReturnType<AccessControlAdjuster>[] = [];
  switch (key) {
    case "end_user": {
      switch (newValue) {
        case "hidden": {
          adjustments.push(atLeast(accessControl, "bearer", "hidden"));
          adjustments.push(atLeast(accessControl, "portal_ui", "hidden"));
          break;
        }
        case "readonly": {
          adjustments.push(atLeast(accessControl, "bearer", "readonly"));
          adjustments.push(atLeast(accessControl, "portal_ui", "readonly"));
          break;
        }
        case "readwrite": {
          adjustments.push(atLeast(accessControl, "bearer", "readonly"));
          adjustments.push(atLeast(accessControl, "portal_ui", "readwrite"));
          break;
        }
      }
      break;
    }
    case "bearer": {
      switch (newValue) {
        case "hidden": {
          adjustments.push(atMost(accessControl, "end_user", "hidden"));
          break;
        }
        case "readonly": {
          adjustments.push(atLeast(accessControl, "portal_ui", "readonly"));
          break;
        }
        case "readwrite": {
          // Unreachable because readwrite is not a valid value for bearer.
          break;
        }
      }
      break;
    }
    case "portal_ui": {
      switch (newValue) {
        case "hidden": {
          adjustments.push(atMost(accessControl, "end_user", "hidden"));
          adjustments.push(atMost(accessControl, "bearer", "hidden"));
          break;
        }
        case "readonly": {
          adjustments.push(atMost(accessControl, "end_user", "readonly"));
          break;
        }
        case "readwrite": {
          // Nothing to adjust.
          break;
        }
      }
      break;
    }
  }

  const otherAdjustments: UserProfileAttributesListAccessControlAdjustment[] =
    adjustments.filter(
      (a): a is UserProfileAttributesListAccessControlAdjustment => a != null
    );

  return {
    index,
    key,
    mainAdjustment,
    otherAdjustments,
  };
}

function applyUpdate<T extends UserProfileAttributesListItem>(
  prevItems: T[],
  update: UserProfileAttributesListPendingUpdate
): T[] {
  const { index, mainAdjustment, otherAdjustments } = update;
  let accessControl = prevItems[index].access_control;
  const adjustments = [mainAdjustment, ...otherAdjustments];

  for (const adjustment of adjustments) {
    accessControl = {
      ...accessControl,
      [adjustment[0]]: adjustment[1],
    };
  }

  const newItems = [...prevItems];
  newItems[index] = {
    ...newItems[index],
    access_control: accessControl,
  };

  return newItems;
}

function UserProfileAttributesList<T extends UserProfileAttributesListItem>(
  props: UserProfileAttributesListProps<T>
): React.ReactElement<any, any> | null {
  const {
    items,
    onChangeItems,
    ItemComponent,
    onEditButtonClick,
    onReorderItems,
  } = props;
  const { renderToString } = useContext(Context);
  const [pendingUpdate, setPendingUpdate] = useState<
    UserProfileAttributesListPendingUpdate | undefined
  >();
  const [dndIndex, setDNDIndex] = useState<number | undefined>(undefined);

  const reorder = useCallback(
    (index: number, item: T) => {
      const itemsWithoutIndex = [
        ...items.slice(0, index),
        ...items.slice(index + 1),
      ];
      const insertIndex = items.indexOf(item);
      if (insertIndex >= 0) {
        itemsWithoutIndex.splice(insertIndex, 0, items[index]);
        onReorderItems?.(itemsWithoutIndex);
      }
    },
    [items, onReorderItems]
  );

  const dragDropEvents: IDragDropEvents = useMemo(() => {
    return {
      canDrop: () => true,
      canDrag: () => true,
      onDragEnter: () => styles.onDragEnter,
      onDragLeave: () => {},
      onDragStart: (_item?: T, index?: number) => {
        if (index != null) {
          setDNDIndex(index);
        }
      },
      onDragEnd: (_item?: T) => {
        setDNDIndex(undefined);
      },
      onDrop: (item?: T) => {
        if (dndIndex != null && item != null) {
          reorder(dndIndex, item);
        }
      },
    };
  }, [reorder, dndIndex]);

  const onClickConfirmPendingUpdate = useCallback(
    (e: React.MouseEvent<unknown>) => {
      e.preventDefault();
      e.stopPropagation();

      if (pendingUpdate != null) {
        const newItems = applyUpdate(items, pendingUpdate);
        setPendingUpdate(undefined);
        onChangeItems(newItems);
      }
    },
    [items, onChangeItems, pendingUpdate]
  );

  const onDismissPendingUpdateDialog = useCallback(() => {
    setPendingUpdate(undefined);
  }, []);

  // title and subText are typed as string but they can actually be any JSX.Element.
  // @ts-expect-error
  const pendingUpdateDialogContentProps: IDialogContentProps = useMemo(() => {
    if (pendingUpdate == null) {
      return {
        title: "",
        subText: "",
      };
    }

    const { index } = pendingUpdate;

    const pointer = items[index].pointer;
    const fieldName = parseJSONPointer(pointer)[0];

    return {
      title: (
        <FormattedMessage
          id="UserProfileAttributesList.dialog.title.pending-update"
          values={{
            fieldName,
            party: pendingUpdate.mainAdjustment[0],
          }}
        />
      ),
      subText: (
        <>
          <Text block={true}>
            <FormattedMessage
              id="UserProfileAttributesList.dialog.adjustment.condition"
              values={{
                fieldName,
                party: pendingUpdate.mainAdjustment[0],
                level: renderToString(
                  "user-profile.access-control-level." +
                    pendingUpdate.mainAdjustment[1]
                ),
              }}
            />
          </Text>
          {pendingUpdate.otherAdjustments.map((a, i) => {
            return (
              <Text key={i} block={true} className={styles.consequence}>
                <FormattedMessage
                  id="UserProfileAttributesList.dialog.adjustment.consequence"
                  values={{
                    party: a[0],
                    level: renderToString(
                      "user-profile.access-control-level." + a[1]
                    ),
                  }}
                />
              </Text>
            );
          })}
        </>
      ),
    };
  }, [renderToString, pendingUpdate, items]);

  const makeDropdownOnChange = useCallback(
    (index: number, key: keyof UserProfileAttributesAccessControl) => {
      return (
        _e: React.FormEvent<unknown>,
        option?: IDropdownOption<AccessControlLevelString>,
        _index?: number
      ) => {
        if (option == null) {
          return;
        }

        const pendingUpdate = makeUpdate(
          items,
          index,
          key,
          option.key as AccessControlLevelString
        );

        if (pendingUpdate.otherAdjustments.length !== 0) {
          setPendingUpdate(pendingUpdate);
          return;
        }

        const newItems = applyUpdate(items, pendingUpdate);
        onChangeItems(newItems);
      };
    },
    [items, onChangeItems]
  );

  const makeRenderDropdown = useCallback(
    (key: keyof UserProfileAttributesAccessControl) => {
      return (
        item?: UserProfileAttributesListItem,
        index?: number,
        _column?: IColumn
      ) => {
        if (item == null || index == null) {
          return null;
        }

        const optionHidden: IDropdownOption = {
          key: "hidden",
          text: renderToString("user-profile.access-control-level.hidden"),
        };

        const optionReadonly: IDropdownOption = {
          key: "readonly",
          text: renderToString("user-profile.access-control-level.readonly"),
        };

        const optionReadwrite: IDropdownOption = {
          key: "readwrite",
          text: renderToString("user-profile.access-control-level.readwrite"),
        };

        let options: IDropdownOption<AccessControlLevelString>[] = [];
        let selectedKey: string | undefined;
        switch (key) {
          case "portal_ui":
            options = [optionHidden, optionReadonly, optionReadwrite];
            selectedKey = item.access_control.portal_ui;
            break;
          case "bearer":
            options = [optionHidden, optionReadonly];
            selectedKey = item.access_control.bearer;
            break;
          case "end_user":
            options = [optionHidden, optionReadonly, optionReadwrite];
            selectedKey = item.access_control.end_user;
            break;
        }

        return (
          <Dropdown
            options={options}
            selectedKey={selectedKey}
            onChange={makeDropdownOnChange(index, key)}
          />
        );
      };
    },
    [renderToString, makeDropdownOnChange]
  );

  const onRenderPointer = useCallback(
    (item?: T, _index?: number, _column?: IColumn) => {
      if (item == null) {
        return null;
      }
      return <ItemComponent className="" item={item} />;
    },
    [ItemComponent]
  );

  const onRenderEditButton = useCallback(
    (
      _item?: UserProfileAttributesListItem,
      index?: number,
      _column?: IColumn
    ) => {
      if (index == null) {
        return null;
      }
      const onClick = (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();
        onEditButtonClick?.(index);
      };
      return (
        <IconButton
          iconProps={EDIT_BUTTON_ICON_PROPS}
          title={renderToString("edit")}
          ariaLabel={renderToString("edit")}
          onClick={onClick}
        />
      );
    },
    [onEditButtonClick, renderToString]
  );

  const onRenderReorderHandle = useCallback(() => {
    return (
      <div className={styles.reorderHandle}>
        <Icon iconName="GlobalNavButton" />
      </div>
    );
  }, []);

  const columns: IColumn[] = useMemo(() => {
    const columns: IColumn[] = [
      {
        key: "pointer",
        minWidth: 200,
        name: renderToString(
          "UserProfileAttributesList.header.label.attribute-name"
        ),
        onRender: onRenderPointer,
        isMultiline: true,
      },
      {
        key: "portal_ui",
        minWidth: 200,
        maxWidth: 200,
        name: "",
        onRender: makeRenderDropdown("portal_ui"),
      },
      {
        key: "bearer",
        minWidth: 200,
        maxWidth: 200,
        name: "",
        onRender: makeRenderDropdown("bearer"),
      },
      {
        key: "end_user",
        minWidth: 200,
        maxWidth: 200,
        name: "",
        onRender: makeRenderDropdown("end_user"),
      },
    ];
    if (onEditButtonClick != null) {
      columns.push({
        key: "edit",
        minWidth: 24,
        maxWidth: 24,
        name: "",
        onRender: onRenderEditButton,
      });
    }
    if (onReorderItems != null) {
      columns.push({
        key: "reorder",
        minWidth: 24,
        maxWidth: 24,
        name: "",
        onRender: onRenderReorderHandle,
      });
    }
    return columns;
  }, [
    onEditButtonClick,
    onReorderItems,
    renderToString,
    makeRenderDropdown,
    onRenderPointer,
    onRenderEditButton,
    onRenderReorderHandle,
  ]);

  const onRenderColumnHeaderTooltip: IRenderFunction<IDetailsColumnRenderTooltipProps> =
    useCallback(
      (
        props?: IDetailsColumnRenderTooltipProps,
        defaultRender?: (
          props: IDetailsColumnRenderTooltipProps
        ) => JSX.Element | null
      ) => {
        if (props == null || defaultRender == null) {
          return null;
        }
        if (props.column == null) {
          return null;
        }
        if (
          props.column.key === "portal_ui" ||
          props.column.key === "bearer" ||
          props.column.key === "end_user"
        ) {
          return (
            <LabelWithTooltip
              labelId={
                "UserProfileAttributesList.header.label." + props.column.key
              }
              tooltipMessageId={
                "UserProfileAttributesList.header.tooltip." + props.column.key
              }
              directionalHint={DirectionalHint.topCenter}
            />
          );
        }
        return defaultRender(props);
      },
      []
    );

  const onRenderDetailsHeader: IRenderFunction<IDetailsHeaderProps> =
    useCallback(
      (props?: IDetailsHeaderProps) => {
        if (props == null) {
          return null;
        }
        return (
          <DetailsHeader
            {...props}
            onRenderColumnHeaderTooltip={onRenderColumnHeaderTooltip}
          />
        );
      },
      [onRenderColumnHeaderTooltip]
    );

  const onRenderRow: IRenderFunction<IDetailsRowProps> = useCallback(
    (props?: IDetailsRowProps) => {
      if (props == null) {
        return null;
      }
      let className = "";
      const { itemIndex } = props;
      if (dndIndex != null) {
        if (itemIndex < dndIndex) {
          className = styles.before;
        } else if (itemIndex > dndIndex) {
          className = styles.after;
        }
      }
      return <DetailsRow {...props} className={className} />;
    },
    [dndIndex]
  );

  return (
    <>
      <DetailsList
        columns={columns}
        items={items}
        selectionMode={SelectionMode.none}
        onRenderDetailsHeader={onRenderDetailsHeader}
        onRenderRow={onRenderRow}
        dragDropEvents={onReorderItems != null ? dragDropEvents : undefined}
      />
      <Dialog
        hidden={pendingUpdate == null}
        onDismiss={onDismissPendingUpdateDialog}
        dialogContentProps={pendingUpdateDialogContentProps}
      >
        <DialogFooter>
          <PrimaryButton
            onClick={onClickConfirmPendingUpdate}
            text={<FormattedMessage id="confirm" />}
          />
          <DefaultButton
            onClick={onDismissPendingUpdateDialog}
            text={<FormattedMessage id="cancel" />}
          />
        </DialogFooter>
      </Dialog>
    </>
  );
}

export default UserProfileAttributesList;
