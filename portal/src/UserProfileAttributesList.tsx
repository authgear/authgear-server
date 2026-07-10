/* global JSX */
import React, { useMemo, useCallback, useContext, useState } from "react";
import cn from "classnames";
import { FormattedMessage, Context } from "./intl";
import {
  Select,
  Dialog,
  Text,
  Button,
  Flex,
  IconButton,
  DropdownMenu,
} from "@radix-ui/themes";
import {
  QuestionMarkCircledIcon,
  Pencil1Icon,
  DragHandleDots2Icon,
  DotsVerticalIcon,
} from "@radix-ui/react-icons";
import { Tooltip } from "./components/v2/Tooltip/Tooltip";
import { SecondaryButton } from "./components/v2/Button/SecondaryButton/SecondaryButton";
import ExternalLink from "./ExternalLink";
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

interface AccessControlSelectProps {
  accessControlKey: keyof UserProfileAttributesAccessControl;
  item: UserProfileAttributesListItem;
  onValueChange: (value: string) => void;
}

function AccessControlSelect({
  accessControlKey,
  item,
  onValueChange,
}: AccessControlSelectProps) {
  const { renderToString } = useContext(Context);
  const selectedValue = item.access_control[accessControlKey];
  const includeReadwrite = accessControlKey !== "bearer";

  return (
    <div className={styles.selectRoot}>
      <Select.Root value={selectedValue} onValueChange={onValueChange}>
        <Select.Trigger
          variant="surface"
          className={styles.selectTrigger}
        />
        <Select.Content>
          <Select.Item value="hidden">
            {renderToString("user-profile.access-control-level.hidden")}
          </Select.Item>
          <Select.Item value="readonly">
            {renderToString("user-profile.access-control-level.readonly")}
          </Select.Item>
          {includeReadwrite && (
            <Select.Item value="readwrite">
              {renderToString("user-profile.access-control-level.readwrite")}
            </Select.Item>
          )}
        </Select.Content>
      </Select.Root>
    </div>
  );
}

interface ItemActionsMenuProps {
  index: number;
  onEditButtonClick: (index: number) => void;
  menuLabel: string;
}

function ItemActionsMenu({
  index,
  onEditButtonClick,
  menuLabel,
}: ItemActionsMenuProps) {
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger>
        <IconButton
          variant="ghost"
          size="2"
          aria-label={menuLabel}
          onClick={(e) => {
            e.stopPropagation();
          }}
        >
          <DotsVerticalIcon />
        </IconButton>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content align="end">
        <DropdownMenu.Item
          onSelect={() => {
            onEditButtonClick(index);
          }}
        >
          <Pencil1Icon />
          <FormattedMessage id="edit" />
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu.Root>
  );
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

  const canReorder = onReorderItems != null;
  const hasEdit = onEditButtonClick != null;

  const gridClassName = hasEdit ? styles.gridWithActions : undefined;

  const reorder = useCallback(
    (index: number, targetItem: T) => {
      const insertIndex = items.indexOf(targetItem);
      if (insertIndex >= 0) {
        const itemsWithoutIndex = [
          ...items.slice(0, index),
          ...items.slice(index + 1),
        ];
        itemsWithoutIndex.splice(insertIndex, 0, items[index]);
        onReorderItems?.(itemsWithoutIndex);
      }
    },
    [items, onReorderItems]
  );

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

  const makeSelectOnChange = useCallback(
    (index: number, key: keyof UserProfileAttributesAccessControl) => {
      return (value: string) => {
        const update = makeUpdate(
          items,
          index,
          key,
          value as AccessControlLevelString
        );

        if (update.otherAdjustments.length !== 0) {
          setPendingUpdate(update);
          return;
        }

        const newItems = applyUpdate(items, update);
        onChangeItems(newItems);
      };
    },
    [items, onChangeItems]
  );

  const pendingDialogTitle = useMemo(() => {
    if (pendingUpdate == null) return "";
    const pointer = items[pendingUpdate.index].pointer;
    const fieldName = parseJSONPointer(pointer)[0];
    return (
      <FormattedMessage
        id="UserProfileAttributesList.dialog.title.pending-update"
        values={{
          fieldName,
          party: pendingUpdate.mainAdjustment[0],
        }}
      />
    );
  }, [pendingUpdate, items]);

  const pendingDialogDescription = useMemo(() => {
    if (pendingUpdate == null) return null;
    const pointer = items[pendingUpdate.index].pointer;
    const fieldName = parseJSONPointer(pointer)[0];
    return (
      <>
        <Text as="p" size="2">
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
        {pendingUpdate.otherAdjustments.map((a, i) => (
          <Text key={i} as="p" size="2" className={styles.consequence}>
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
        ))}
      </>
    );
  }, [renderToString, pendingUpdate, items]);

  const endUserTooltipContent = useMemo(
    () => (
      <FormattedMessage
        id="UserProfileAttributesList.header.tooltip.end_user"
        values={{
          DocLink: (chunks: React.ReactNode) => (
            <ExternalLink href="https://docs.authgear.com/customization/built-in-ui/user-settings">
              {chunks}
            </ExternalLink>
          ),
        }}
      />
    ),
    []
  );

  const bearerTooltipContent = useMemo(
    () => (
      <FormattedMessage
        id="UserProfileAttributesList.header.tooltip.bearer"
        values={{
          DocLink: (chunks: React.ReactNode) => (
            <ExternalLink href="https://docs.authgear.com/integration/user-profiles/user-profile">
              {chunks}
            </ExternalLink>
          ),
        }}
      />
    ),
    []
  );

  const portalUiTooltipContent = useMemo(
    () => (
      <FormattedMessage id="UserProfileAttributesList.header.tooltip.portal_ui" />
    ),
    []
  );

  return (
    <>
      <div className={styles.table}>
        <div className={cn(styles.headerRow, gridClassName)}>
          <div className={styles.headerCellMain}>
            {canReorder ? (
              <span className={styles.headerReorderSpacer} aria-hidden={true} />
            ) : null}
            <Text size="2" weight="medium">
              <FormattedMessage id="UserProfileAttributesList.header.label.attribute-name" />
            </Text>
          </div>
          <div className={styles.headerCell}>
            <Text size="2" weight="medium">
              <FormattedMessage id="UserProfileAttributesList.header.label.portal_ui" />
            </Text>
            <Tooltip content={portalUiTooltipContent}>
              <QuestionMarkCircledIcon className={styles.tooltipIcon} />
            </Tooltip>
          </div>
          <div className={styles.headerCell}>
            <Text size="2" weight="medium">
              <FormattedMessage id="UserProfileAttributesList.header.label.bearer" />
            </Text>
            <Tooltip content={bearerTooltipContent}>
              <QuestionMarkCircledIcon className={styles.tooltipIcon} />
            </Tooltip>
          </div>
          <div className={styles.headerCell}>
            <Text size="2" weight="medium">
              <FormattedMessage id="UserProfileAttributesList.header.label.end_user" />
            </Text>
            <Tooltip content={endUserTooltipContent}>
              <QuestionMarkCircledIcon className={styles.tooltipIcon} />
            </Tooltip>
          </div>
          {hasEdit ? <div className={styles.headerCellSmall} /> : null}
        </div>

        {items.map((item, index) => (
          <div
            key={item.pointer}
            className={cn(styles.row, gridClassName)}
            draggable={canReorder}
            onDragStart={() => setDNDIndex(index)}
            onDragEnd={() => setDNDIndex(undefined)}
            onDragOver={(e) => e.preventDefault()}
            onDrop={() => {
              if (dndIndex != null && dndIndex !== index) {
                reorder(dndIndex, item);
              }
              setDNDIndex(undefined);
            }}
            data-dnd-before={
              dndIndex != null && index < dndIndex ? true : undefined
            }
            data-dnd-after={
              dndIndex != null && index > dndIndex ? true : undefined
            }
          >
            <div className={styles.cellMain}>
              {canReorder ? (
                <div className={styles.reorderHandle}>
                  <DragHandleDots2Icon />
                </div>
              ) : null}
              <div className={styles.cellMainContent}>
                <ItemComponent className="" item={item} />
              </div>
            </div>
            <div className={styles.cell}>
              <AccessControlSelect
                accessControlKey="portal_ui"
                item={item}
                onValueChange={makeSelectOnChange(index, "portal_ui")}
              />
            </div>
            <div className={styles.cell}>
              <AccessControlSelect
                accessControlKey="bearer"
                item={item}
                onValueChange={makeSelectOnChange(index, "bearer")}
              />
            </div>
            <div className={styles.cell}>
              <AccessControlSelect
                accessControlKey="end_user"
                item={item}
                onValueChange={makeSelectOnChange(index, "end_user")}
              />
            </div>
            {hasEdit ? (
              <div className={styles.cellSmall}>
                <ItemActionsMenu
                  index={index}
                  onEditButtonClick={onEditButtonClick!}
                  menuLabel={renderToString(
                    "UserProfileAttributesList.row-actions"
                  )}
                />
              </div>
            ) : null}
          </div>
        ))}
      </div>

      <Dialog.Root
        open={pendingUpdate != null}
        onOpenChange={(open) => {
          if (!open) onDismissPendingUpdateDialog();
        }}
      >
        <Dialog.Content maxWidth="400px" size="3">
          <Dialog.Title>{pendingDialogTitle}</Dialog.Title>
          <Dialog.Description size="2">
            <div>{pendingDialogDescription}</div>
          </Dialog.Description>
          <Flex gap="3" justify="end" mt="4">
            <SecondaryButton
              size="2"
              text={<FormattedMessage id="cancel" />}
              onClick={onDismissPendingUpdateDialog}
            />
            <Button
              size="2"
              variant="solid"
              color="indigo"
              onClick={onClickConfirmPendingUpdate}
            >
              <FormattedMessage id="confirm" />
            </Button>
          </Flex>
        </Dialog.Content>
      </Dialog.Root>
    </>
  );
}

export default UserProfileAttributesList;
