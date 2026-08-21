import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { Context } from "../../../intl";
import ChoiceButton from "../../../ChoiceButton";
import styles from "./FrameworkCard.module.css";

export interface M2MCardProps {
  selected: boolean;
  onSelect: () => void;
}

/**
 * A selectable card shown in the "Integrations & other" section of the
 * Create Application grid. Unlike a FrameworkCard it does not resolve a
 * framework or application type in place; selecting it marks the wizard as
 * machine-to-machine, and the wizard's Next button routes to the dedicated
 * machine-to-machine create screen.
 */
export const M2MCard: React.FC<M2MCardProps> = ({ selected, onSelect }) => {
  const { renderToString } = useContext(Context);

  const IconComponent = useMemo(() => {
    return function M2MIcon() {
      return (
        <i className={cn("ti", "ti-server", styles.icon)} aria-hidden={true} />
      );
    };
  }, []);

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
      text={renderToString("CreateOAuthClientScreen.framework.m2m.title")}
      secondaryText={renderToString(
        "CreateOAuthClientScreen.framework.m2m.description"
      )}
      IconComponent={IconComponent}
      onClick={onClick}
    />
  );
};
