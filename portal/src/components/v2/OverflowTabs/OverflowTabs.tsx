import React, {
  useCallback,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { DropdownMenu, Tabs } from "@radix-ui/themes";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import cn from "classnames";
import { Context } from "../../../intl";
import styles from "./OverflowTabs.module.css";

const TAB_GAP_PX = 8;
const OVERFLOW_BUTTON_WIDTH_PX = 40;

export interface OverflowTabOption {
  value: string;
  label: React.ReactNode;
}

export interface OverflowTabsProps {
  className?: string;
  listClassName?: string;
  value: string;
  onValueChange: (value: string) => void;
  tabs: OverflowTabOption[];
}

// The selected tab is always kept visible: when it would fall into the
// overflow menu, it is shown as the last visible tab instead and the
// displaced tabs collapse into the menu (matching the old Fluent Pivot).
function calculateVisibleTabValues(
  containerWidth: number,
  tabs: readonly OverflowTabOption[],
  tabWidths: readonly number[],
  selectedValue: string,
  overflowButtonWidth: number
): string[] {
  if (tabs.length === 0) {
    return [];
  }

  const totalWidth = tabWidths.reduce(
    (sum, width, index) => sum + width + (index > 0 ? TAB_GAP_PX : 0),
    0
  );
  if (totalWidth <= containerWidth) {
    return tabs.map((tab) => tab.value);
  }

  const budget = containerWidth - TAB_GAP_PX - overflowButtonWidth;

  let usedWidth = 0;
  let visibleCount = 0;
  for (let index = 0; index < tabWidths.length; index++) {
    const tabWidth = tabWidths[index] + (visibleCount > 0 ? TAB_GAP_PX : 0);
    if (usedWidth + tabWidth > budget && visibleCount > 0) {
      break;
    }
    usedWidth += tabWidth;
    visibleCount++;
  }
  visibleCount = Math.max(1, Math.min(visibleCount, tabWidths.length));

  const selectedIndex = tabs.findIndex((tab) => tab.value === selectedValue);
  if (selectedIndex < visibleCount) {
    return tabs.slice(0, visibleCount).map((tab) => tab.value);
  }

  // The selected tab overflows: show it as the last visible tab, keeping
  // as many preceding tabs as still fit.
  usedWidth = tabWidths[selectedIndex];
  const prefixValues: string[] = [];
  for (let index = 0; index < selectedIndex; index++) {
    const tabWidth = tabWidths[index] + TAB_GAP_PX;
    if (usedWidth + tabWidth > budget) {
      break;
    }
    usedWidth += tabWidth;
    prefixValues.push(tabs[index].value);
  }
  return [...prefixValues, selectedValue];
}

export function OverflowTabs({
  className,
  listClassName,
  value,
  onValueChange,
  tabs,
}: OverflowTabsProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const containerRef = useRef<HTMLDivElement>(null);
  const measureListRef = useRef<HTMLDivElement>(null);
  const [visibleValues, setVisibleValues] = useState<string[]>(() =>
    tabs.map((tab) => tab.value)
  );

  const tabKeys = useMemo(
    () => tabs.map((tab) => tab.value).join("\0"),
    [tabs]
  );

  const updateVisibleCount = useCallback(() => {
    const container = containerRef.current;
    const measureList = measureListRef.current;
    if (container == null || measureList == null) {
      return;
    }

    const tabWidths = Array.from(measureList.children).map(
      (child) => (child as HTMLElement).offsetWidth
    );
    setVisibleValues(
      calculateVisibleTabValues(
        container.clientWidth,
        tabs,
        tabWidths,
        value,
        OVERFLOW_BUTTON_WIDTH_PX
      )
    );
  }, [tabs, value]);

  useLayoutEffect(() => {
    updateVisibleCount();
  }, [tabKeys, updateVisibleCount]);

  useLayoutEffect(() => {
    const container = containerRef.current;
    if (container == null) {
      return;
    }

    const observer = new ResizeObserver(() => {
      updateVisibleCount();
    });
    observer.observe(container);
    return () => {
      observer.disconnect();
    };
  }, [updateVisibleCount]);

  const visibleTabs = tabs.filter((tab) => visibleValues.includes(tab.value));
  const overflowTabs = tabs.filter((tab) => !visibleValues.includes(tab.value));
  const hasOverflow = overflowTabs.length > 0;

  return (
    <Tabs.Root
      className={cn(styles.root, className)}
      value={value}
      onValueChange={onValueChange}
    >
      <div ref={containerRef}>
        <Tabs.List className={cn(styles.list, listClassName)}>
          {visibleTabs.map((tab) => (
            <Tabs.Trigger key={tab.value} value={tab.value}>
              {tab.label}
            </Tabs.Trigger>
          ))}
          {hasOverflow ? (
            <DropdownMenu.Root>
              <DropdownMenu.Trigger>
                <button
                  type="button"
                  className={styles.overflowTrigger}
                  aria-label={renderToString("OverflowTabs.more-tabs")}
                >
                  <DotsHorizontalIcon className={styles.overflowIcon} />
                </button>
              </DropdownMenu.Trigger>
              <DropdownMenu.Content align="start">
                {overflowTabs.map((tab) => (
                  <DropdownMenu.Item
                    key={tab.value}
                    onSelect={() => {
                      onValueChange(tab.value);
                    }}
                  >
                    {tab.label}
                  </DropdownMenu.Item>
                ))}
              </DropdownMenu.Content>
            </DropdownMenu.Root>
          ) : null}
        </Tabs.List>
        <Tabs.List
          ref={measureListRef}
          className={styles.measureList}
          aria-hidden={true}
        >
          {tabs.map((tab) => (
            <Tabs.Trigger key={tab.value} value={tab.value} tabIndex={-1}>
              {tab.label}
            </Tabs.Trigger>
          ))}
        </Tabs.List>
      </div>
    </Tabs.Root>
  );
}
