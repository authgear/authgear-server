import React, { useCallback, useMemo } from "react";
import ChoiceButton from "../../../ChoiceButton";
import type { FrameworkEntry } from "./frameworks";
import styles from "./FrameworkCard.module.css";

export interface FrameworkCardProps {
  framework: FrameworkEntry;
  selected: boolean;
  onSelect: () => void;
}

export const FrameworkCard: React.FC<FrameworkCardProps> = ({
  framework,
  selected,
  onSelect,
}) => {
  const IconComponent = useMemo(() => {
    return function FrameworkLogo() {
      return (
        <img
          className={styles.logo}
          src={framework.logo}
          alt=""
        />
      );
    };
  }, [framework.logo]);

  const onClick = useCallback(
    (e: React.MouseEvent<HTMLElement>) => {
      e.preventDefault();
      e.stopPropagation();
      onSelect();
    },
    [onSelect]
  );

  return (
    <ChoiceButton
      className={styles.card}
      checked={selected}
      text={framework.displayName}
      secondaryText={framework.helperText}
      IconComponent={IconComponent}
      onClick={onClick}
    />
  );
};
