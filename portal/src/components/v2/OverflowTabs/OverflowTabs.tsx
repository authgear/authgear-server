import React, {
  useCallback,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { DropdownMenu, Tabs } from "@radix-ui/themes";
import { DotsHorizontalIcon } from "@radix-ui/react-icons";
import cn from "classnames";
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

function calculateVisibleTabCount(
  containerWidth: number,
  tabWidths: readonly number[],
  overflowButtonWidth: number
): number {
  if (tabWidths.length === 0) {
    return 0;
  }

  const totalWidth = tabWidths.reduce(
    (sum, width, index) => sum + width + (index > 0 ? TAB_GAP_PX : 0),
    0
  );
  if (totalWidth <= containerWidth) {
    return tabWidths.length;
  }

  let usedWidth = 0;
  let visibleCount = 0;
  for (let index = 0; index < tabWidths.length; index++) {
    const tabWidth = tabWidths[index] + (visibleCount > 0 ? TAB_GAP_PX : 0);
    const remainingTabs = tabWidths.length - (index + 1);
    const reserveOverflow =
      remainingTabs > 0 ? TAB_GAP_PX + overflowButtonWidth : 0;

    if (
      usedWidth + tabWidth + reserveOverflow > containerWidth &&
      visibleCount > 0
    ) {
      break;
    }

    usedWidth += tabWidth;
    visibleCount++;
  }

  return Math.max(1, Math.min(visibleCount, tabWidths.length));
}

export function OverflowTabs({
  className,
  listClassName,
  value,
  onValueChange,
  tabs,
}: OverflowTabsProps): React.ReactElement {
  const containerRef = useRef<HTMLDivElement>(null);
  const measureListRef = useRef<HTMLDivElement>(null);
  const [visibleCount, setVisibleCount] = useState(tabs.length);

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
    setVisibleCount(
      calculateVisibleTabCount(
        container.clientWidth,
        tabWidths,
        OVERFLOW_BUTTON_WIDTH_PX
      )
    );
  }, []);

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

  const visibleTabs = tabs.slice(0, visibleCount);
  const overflowTabs = tabs.slice(visibleCount);
  const hasOverflow = overflowTabs.length > 0;
  const hasSelectedOverflow = overflowTabs.some((tab) => tab.value === value);

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
                  className={cn(
                    styles.overflowTrigger,
                    hasSelectedOverflow && styles.overflowTriggerActive
                  )}
                  aria-label="More tabs"
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
