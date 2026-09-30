import React, { useCallback, useContext } from "react";
import cn from "classnames";
import { SegmentedControl } from "@radix-ui/themes";
import { DesktopIcon, MoonIcon, SunIcon } from "@radix-ui/react-icons";
import { Context } from "../../intl";
import { useAppearance } from "../../hook/useAppearance";
import { Appearance } from "../../util/appearance";
import styles from "./AppearanceSwitcher.module.css";

const APPEARANCE_LABEL_IDS: Record<Appearance, string> = {
  light: "ScreenHeader.appearance.light",
  system: "ScreenHeader.appearance.device",
  dark: "ScreenHeader.appearance.dark",
};

// Display order: light, device, dark.
const APPEARANCE_OPTIONS: readonly Appearance[] = ["light", "system", "dark"];

function AppearanceIcon({
  appearance,
  className,
}: {
  appearance: Appearance;
  className?: string;
}): React.ReactElement {
  switch (appearance) {
    case "light":
      return <SunIcon className={className} aria-hidden={true} />;
    case "system":
      return <DesktopIcon className={className} aria-hidden={true} />;
    case "dark":
      return <MoonIcon className={className} aria-hidden={true} />;
  }
}

export interface AppearanceSwitcherProps {
  className?: string;
}

export function AppearanceSwitcher({
  className,
}: AppearanceSwitcherProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { preference, setPreference } = useAppearance();

  const onValueChange = useCallback(
    (value: string) => {
      if (APPEARANCE_OPTIONS.includes(value as Appearance)) {
        setPreference(value as Appearance);
      }
    },
    [setPreference]
  );

  return (
    <SegmentedControl.Root
      className={cn(styles.root, className)}
      size="2"
      radius="full"
      value={preference}
      onValueChange={onValueChange}
      aria-label={renderToString("ScreenHeader.appearance")}
    >
      {APPEARANCE_OPTIONS.map((appearance) => {
        const label = renderToString(APPEARANCE_LABEL_IDS[appearance]);
        return (
          // The native title is used instead of the Radix Tooltip: the tooltip
          // trigger sets data-state on its child, which clobbers the toggle
          // item's own data-state="on" and breaks the selected styling.
          <SegmentedControl.Item
            key={appearance}
            value={appearance}
            aria-label={label}
            title={label}
          >
            <AppearanceIcon appearance={appearance} className={styles.icon} />
          </SegmentedControl.Item>
        );
      })}
    </SegmentedControl.Root>
  );
}
