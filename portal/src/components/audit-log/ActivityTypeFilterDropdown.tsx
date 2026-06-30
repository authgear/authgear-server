import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { ChevronDownIcon, Cross2Icon } from "@radix-ui/react-icons";
import { Popover, Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";
import {
  TextField,
  TextFieldIcon,
} from "../v2/TextField/TextField";
import {
  groupActivityTypesByCategory,
  ActivityTypeCategoryId,
} from "./activityTypeCategories";
import styles from "./ActivityTypeFilterDropdown.module.css";

export type AuditLogActivityTypeAll = "ALL";
export const ACTIVITY_TYPE_ALL: AuditLogActivityTypeAll = "ALL";

export type ActivityTypeFilterDropdownOptionKey =
  | AuditLogActivityType
  | AuditLogActivityTypeAll;

interface ActivityTypeFilterDropdownProps {
  className?: string;
  value: ActivityTypeFilterDropdownOptionKey;
  onChange: (newValue: ActivityTypeFilterDropdownOptionKey) => void;
  availableActivityTypes: AuditLogActivityType[];
}

interface ActivityTypeOption {
  key: AuditLogActivityType;
  label: string;
}

interface ActivityTypeCategorySection {
  id: ActivityTypeCategoryId;
  label: string;
  options: ActivityTypeOption[];
}

export const ActivityTypeFilterDropdown: React.VFC<ActivityTypeFilterDropdownProps> =
  function ActivityTypeFilterDropdown({
    className,
    value,
    onChange,
    availableActivityTypes,
  }) {
    const { renderToString } = useContext(MessageContext);
    const [open, setOpen] = useState(false);
    const [searchValue, setSearchValue] = useState("");

    const placeholder = renderToString("AuditLogActivityType.ALL");

    const selectedLabel = useMemo(() => {
      if (value === ACTIVITY_TYPE_ALL) {
        return null;
      }
      return renderToString("AuditLogActivityType." + value);
    }, [renderToString, value]);

    const categorySections = useMemo<ActivityTypeCategorySection[]>(() => {
      const normalizedSearch = searchValue.trim().toLowerCase();
      const groups = groupActivityTypesByCategory(availableActivityTypes);

      return groups
        .map((group) => {
          const options = group.activityTypes
            .map((activityType) => ({
              key: activityType,
              label: renderToString("AuditLogActivityType." + activityType),
            }))
            .filter((option) =>
              normalizedSearch === ""
                ? true
                : option.label.toLowerCase().includes(normalizedSearch)
            )
            .sort((a, b) => a.label.localeCompare(b.label));

          if (options.length === 0) {
            return null;
          }

          return {
            id: group.id,
            label: renderToString(
              "AuditLogActivityType.category." + group.id
            ),
            options,
          };
        })
        .filter((section): section is ActivityTypeCategorySection => {
          return section != null;
        });
    }, [availableActivityTypes, renderToString, searchValue]);

    const onSelectOption = useCallback(
      (activityType: AuditLogActivityType) => {
        onChange(activityType);
        setOpen(false);
        setSearchValue("");
      },
      [onChange]
    );

    const onClearFilter = useCallback(
      (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setSearchValue("");
        onChange(ACTIVITY_TYPE_ALL);
        setOpen(false);
      },
      [onChange]
    );

    const onOpenChange = useCallback((nextOpen: boolean) => {
      setOpen(nextOpen);
      if (!nextOpen) {
        setSearchValue("");
      }
    }, []);

    const hasSelection = value !== ACTIVITY_TYPE_ALL;

    return (
      <div className={cn(styles.root, className)}>
        <Popover.Root open={open} onOpenChange={onOpenChange}>
          <Popover.Trigger>
            <button
              type="button"
              className={styles.trigger}
              aria-label={placeholder}
            >
              <span
                className={cn(
                  styles.triggerLabel,
                  !hasSelection && styles.triggerLabelPlaceholder
                )}
              >
                {selectedLabel ?? placeholder}
              </span>
              {hasSelection ? (
                <span
                  role="button"
                  tabIndex={0}
                  className={styles.clearButton}
                  aria-label={renderToString(
                    "AuditLogScreen.clear-all-filters"
                  )}
                  onPointerDown={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                  }}
                  onClick={onClearFilter}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      onClearFilter(
                        e as unknown as React.MouseEvent<HTMLSpanElement>
                      );
                    }
                  }}
                >
                  <Cross2Icon className={styles.triggerIcon} />
                </span>
              ) : (
                <ChevronDownIcon className={styles.triggerIcon} />
              )}
            </button>
          </Popover.Trigger>
          <Popover.Content
            className={styles.content}
            sideOffset={4}
            align="start"
          >
            <div className={styles.searchRow}>
              <TextField
                size="2"
                type="search"
                value={searchValue}
                placeholder={renderToString("search")}
                iconStart={TextFieldIcon.MagnifyingGlass}
                onChange={(e) => {
                  setSearchValue(e.currentTarget.value);
                }}
              />
            </div>
            <div className={styles.list}>
              {categorySections.length === 0 ? (
                <Text className={styles.emptyView}>
                  <FormattedMessage id="SearchableDropdown.empty" />
                </Text>
              ) : (
                categorySections.map((section) => (
                  <div key={section.id} className={styles.category}>
                    <Text as="div" className={styles.categoryLabel}>
                      {section.label}
                    </Text>
                    {section.options.map((option) => (
                      <button
                        key={option.key}
                        type="button"
                        className={cn(
                          styles.item,
                          value === option.key && styles.itemSelected
                        )}
                        onClick={() => {
                          onSelectOption(option.key);
                        }}
                      >
                        {option.label}
                      </button>
                    ))}
                  </div>
                ))
              )}
            </div>
          </Popover.Content>
        </Popover.Root>
      </div>
    );
  };
