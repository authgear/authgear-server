import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { Context } from "../../../intl";
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
  const { renderToString } = useContext(Context);

  const IconComponent = useMemo(() => {
    return function FrameworkIcon() {
      return (
        <i
          className={cn("ti", `ti-${framework.iconName}`, styles.icon)}
          aria-hidden={true}
        />
      );
    };
  }, [framework.iconName]);

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
      text={renderToString(framework.displayNameMessageId)}
      secondaryText={renderToString(framework.helperTextMessageId)}
      IconComponent={IconComponent}
      onClick={onClick}
    />
  );
};
