import React from "react";
import { Text } from "@radix-ui/themes";
import { Toggle } from "../v2/Toggle/Toggle";
import styles from "./MechanismSectionHeader.module.css";

export interface MechanismSectionHeaderProps {
  title: React.ReactNode;
  description: React.ReactNode;
  toggleLabel: React.ReactNode;
  checked: boolean;
  disabled?: boolean;
  onCheckedChange: (checked: boolean) => void;
}

/**
 * The header of one client-onboarding mechanism's group of settings cards:
 * its name, what it does, and the switch that turns it on.
 *
 * The switch sits in the header rather than in a card of its own so the
 * cards below it read as belonging to it. CIMD and DCR are configured
 * independently and their cards are otherwise identical in appearance, so
 * without a header band an admin cannot tell where one mechanism's settings
 * end and the other's begin.
 */
export const MechanismSectionHeader: React.VFC<MechanismSectionHeaderProps> =
  function MechanismSectionHeader({
    title,
    description,
    toggleLabel,
    checked,
    disabled,
    onCheckedChange,
  }) {
    return (
      <div className={styles.header}>
        <div className={styles.headingGroup}>
          <Text as="p" size="4" weight="bold" className={styles.title}>
            {title}
          </Text>
          <Text as="p" size="2" color="gray" className={styles.description}>
            {description}
          </Text>
        </div>
        <div className={styles.toggle}>
          <Toggle
            checked={checked}
            disabled={disabled}
            onCheckedChange={onCheckedChange}
            text={toggleLabel}
          />
        </div>
      </div>
    );
  };
