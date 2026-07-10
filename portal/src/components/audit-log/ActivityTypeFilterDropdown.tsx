import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import cn from "classnames";
import {
  ChevronDownIcon,
  ChevronRightIcon,
  Cross2Icon,
} from "@radix-ui/react-icons";
import { Checkbox, Popover, Text } from "@radix-ui/themes";
import { Context as MessageContext, FormattedMessage } from "../../intl";
import { AuditLogActivityType } from "../../graphql/adminapi/globalTypes.generated";
import {
  TextField,
  TextFieldIcon,
} from "../v2/TextField/TextField";
import {
  ActivityTypeCategoryGroupId,
  ActivityTypeSubcategoryId,
  getActivityTypeCategoryGroup,
  getActivityTypeSubcategory,
  groupActivityTypesByHierarchy,
  hasMultipleSubcategories,
} from "./activityTypeCategories";
import styles from "./ActivityTypeFilterDropdown.module.css";

export type AuditLogActivityTypeAll = "ALL";
export const ACTIVITY_TYPE_ALL: AuditLogActivityTypeAll = "ALL";

export function parseActivityTypesFromQuery(
  query: string | null,
  availableActivityTypes: AuditLogActivityType[]
): AuditLogActivityType[] {
  if (query == null || query === "" || query === ACTIVITY_TYPE_ALL) {
    return [];
  }
  return query
    .split(",")
    .filter((type): type is AuditLogActivityType =>
      availableActivityTypes.includes(type as AuditLogActivityType)
    );
}

export function serializeActivityTypesToQuery(
  activityTypes: AuditLogActivityType[]
): string {
  if (activityTypes.length === 0) {
    return ACTIVITY_TYPE_ALL;
  }
  return activityTypes.join(",");
}

interface ActivityTypeFilterDropdownProps {
  className?: string;
  value: AuditLogActivityType[];
  onChange: (newValue: AuditLogActivityType[]) => void;
  availableActivityTypes: AuditLogActivityType[];
  wideContent?: boolean;
}

interface ActivityTypeOption {
  key: AuditLogActivityType;
  label: string;
}

interface ActivityTypeSubcategorySection {
  id: ActivityTypeSubcategoryId;
  label: string;
  options: ActivityTypeOption[];
}

interface ActivityTypeGroupSection {
  id: ActivityTypeCategoryGroupId;
  label: string;
  hasMultipleSubcategories: boolean;
  subcategories: ActivityTypeSubcategorySection[];
}

export const ActivityTypeFilterDropdown: React.VFC<ActivityTypeFilterDropdownProps> =
  function ActivityTypeFilterDropdown({
    className,
    value,
    onChange,
    availableActivityTypes,
    wideContent = false,
  }) {
    const { renderToString } = useContext(MessageContext);
    const [open, setOpen] = useState(false);
    const [searchValue, setSearchValue] = useState("");
    const [expandedGroupIds, setExpandedGroupIds] = useState<
      Set<ActivityTypeCategoryGroupId>
    >(new Set());
    const [expandedSubcategoryIds, setExpandedSubcategoryIds] = useState<
      Set<ActivityTypeSubcategoryId>
    >(new Set());

    const placeholder = renderToString("AuditLogActivityType.ALL");

    const selectedLabel = useMemo(() => {
      if (value.length === 0) {
        return null;
      }
      if (value.length === 1) {
        return renderToString("AuditLogActivityType." + value[0]);
      }
      return renderToString("AuditLogScreen.activity-types-selected", {
        count: value.length,
      });
    }, [renderToString, value]);

    const groupSections = useMemo<ActivityTypeGroupSection[]>(() => {
      const normalizedSearch = searchValue.trim().toLowerCase();
      const groups = groupActivityTypesByHierarchy(availableActivityTypes);

      return groups
        .map((group) => {
          const subcategories = group.subcategories
            .map((subcategory) => {
              const options = subcategory.activityTypes
                .map((activityType) => ({
                  key: activityType,
                  label: renderToString(
                    "AuditLogActivityType." + activityType
                  ),
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
                id: subcategory.id,
                label: renderToString(
                  "AuditLogActivityType.subcategory." + subcategory.id
                ),
                options,
              };
            })
            .filter(
              (section): section is ActivityTypeSubcategorySection =>
                section != null
            );

          if (subcategories.length === 0) {
            return null;
          }

          return {
            id: group.id,
            label: renderToString(
              "AuditLogActivityType.categoryGroup." + group.id
            ),
            hasMultipleSubcategories: hasMultipleSubcategories(group.id),
            subcategories,
          };
        })
        .filter((section): section is ActivityTypeGroupSection => {
          return section != null;
        });
    }, [availableActivityTypes, renderToString, searchValue]);

    const selectedSubcategoryIds = useMemo(() => {
      if (value.length === 0) {
        return new Set<ActivityTypeSubcategoryId>();
      }
      return new Set(
        value.map((activityType) => getActivityTypeSubcategory(activityType))
      );
    }, [value]);

    const selectedGroupIds = useMemo(() => {
      if (selectedSubcategoryIds.size === 0) {
        return new Set<ActivityTypeCategoryGroupId>();
      }
      return new Set(
        [...selectedSubcategoryIds].map((subcategoryId) =>
          getActivityTypeCategoryGroup(subcategoryId)
        )
      );
    }, [selectedSubcategoryIds]);

    useEffect(() => {
      if (!open) {
        return;
      }

      const normalizedSearch = searchValue.trim();
      if (normalizedSearch !== "") {
        setExpandedGroupIds(
          new Set(groupSections.map((section) => section.id))
        );
        setExpandedSubcategoryIds(
          new Set(
            groupSections.flatMap((section) =>
              section.subcategories.map((subcategory) => subcategory.id)
            )
          )
        );
        return;
      }

      if (selectedGroupIds.size > 0 && selectedSubcategoryIds.size > 0) {
        setExpandedGroupIds(selectedGroupIds);
        setExpandedSubcategoryIds(selectedSubcategoryIds);
        return;
      }

      setExpandedGroupIds(new Set());
      setExpandedSubcategoryIds(new Set());
    }, [
      open,
      searchValue,
      selectedGroupIds,
      selectedSubcategoryIds,
      groupSections,
    ]);

    const toggleGroup = useCallback((groupId: ActivityTypeCategoryGroupId) => {
      setExpandedGroupIds((prev) => {
        const next = new Set(prev);
        if (next.has(groupId)) {
          next.delete(groupId);
        } else {
          next.add(groupId);
        }
        return next;
      });
    }, []);

    const toggleSubcategory = useCallback(
      (subcategoryId: ActivityTypeSubcategoryId) => {
        setExpandedSubcategoryIds((prev) => {
          const next = new Set(prev);
          if (next.has(subcategoryId)) {
            next.delete(subcategoryId);
          } else {
            next.add(subcategoryId);
          }
          return next;
        });
      },
      []
    );

    const onToggleOption = useCallback(
      (activityType: AuditLogActivityType, checked: boolean) => {
        if (checked) {
          onChange([...value, activityType]);
          return;
        }
        onChange(value.filter((selected) => selected !== activityType));
      },
      [onChange, value]
    );

    const onClearFilter = useCallback(
      (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        setSearchValue("");
        onChange([]);
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

    const hasSelection = value.length > 0;

    const renderOptions = useCallback(
      (options: ActivityTypeOption[], indentClassName: string) => {
        return options.map((option) => {
          const isSelected = value.includes(option.key);
          return (
            <label
              key={option.key}
              className={cn(
                styles.item,
                indentClassName,
                isSelected && styles.itemSelected
              )}
            >
              <Checkbox
                checked={isSelected}
                onCheckedChange={(checked) => {
                  onToggleOption(option.key, checked === true);
                }}
              />
              <Text as="span" className={styles.itemLabel}>
                {option.label}
              </Text>
            </label>
          );
        });
      },
      [onToggleOption, value]
    );

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
            className={cn(styles.content, wideContent && styles.contentWide)}
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
              {groupSections.length === 0 ? (
                <Text className={styles.emptyView}>
                  <FormattedMessage id="SearchableDropdown.empty" />
                </Text>
              ) : (
                groupSections.map((group) => {
                  const isGroupExpanded = expandedGroupIds.has(group.id);
                  return (
                    <div key={group.id} className={styles.group}>
                      <button
                        type="button"
                        className={styles.groupHeader}
                        aria-expanded={isGroupExpanded}
                        onClick={() => {
                          toggleGroup(group.id);
                        }}
                      >
                        {isGroupExpanded ? (
                          <ChevronDownIcon className={styles.groupChevron} />
                        ) : (
                          <ChevronRightIcon className={styles.groupChevron} />
                        )}
                        <Text as="span" className={styles.groupLabel}>
                          {group.label}
                        </Text>
                      </button>
                      {isGroupExpanded ? (
                        <div className={styles.groupContent}>
                          {group.hasMultipleSubcategories
                            ? group.subcategories.map((subcategory) => {
                                const isSubcategoryExpanded =
                                  expandedSubcategoryIds.has(subcategory.id);
                                return (
                                  <div
                                    key={subcategory.id}
                                    className={styles.subcategory}
                                  >
                                    <button
                                      type="button"
                                      className={styles.subcategoryHeader}
                                      aria-expanded={isSubcategoryExpanded}
                                      onClick={() => {
                                        toggleSubcategory(subcategory.id);
                                      }}
                                    >
                                      {isSubcategoryExpanded ? (
                                        <ChevronDownIcon
                                          className={styles.subcategoryChevron}
                                        />
                                      ) : (
                                        <ChevronRightIcon
                                          className={styles.subcategoryChevron}
                                        />
                                      )}
                                      <Text
                                        as="span"
                                        className={styles.subcategoryLabel}
                                      >
                                        {subcategory.label}
                                      </Text>
                                    </button>
                                    {isSubcategoryExpanded
                                      ? renderOptions(
                                          subcategory.options,
                                          styles.itemNested
                                        )
                                      : null}
                                  </div>
                                );
                              })
                            : renderOptions(
                                group.subcategories[0]?.options ?? [],
                                styles.itemInGroup
                              )}
                        </div>
                      ) : null}
                    </div>
                  );
                })
              )}
            </div>
          </Popover.Content>
        </Popover.Root>
      </div>
    );
  };
