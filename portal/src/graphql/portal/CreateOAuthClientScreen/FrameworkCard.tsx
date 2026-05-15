import React from "react";
import cn from "classnames";
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
  return (
    <button
      type="button"
      role="radio"
      aria-checked={selected}
      className={cn(styles.card, { [styles.selected]: selected })}
      onClick={onSelect}
    >
      <img className={styles.logo} src={framework.logo} alt="" />
      <span className={styles.labels}>
        <span className={styles.name}>{framework.displayName}</span>
        <span className={styles.helper}>{framework.helperText}</span>
      </span>
    </button>
  );
};
