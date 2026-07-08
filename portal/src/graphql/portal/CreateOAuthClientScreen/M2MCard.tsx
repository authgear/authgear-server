import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { Context } from "../../../intl";
import ChoiceButton from "../../../ChoiceButton";
import styles from "./FrameworkCard.module.css";

export interface M2MCardProps {
  onSelect: () => void;
}

/**
 * A navigational card shown in the "Integrations & other" section of the
 * Create Application grid. Unlike a FrameworkCard it does not select a
 * framework or resolve an application type; selecting it routes to the
 * dedicated machine-to-machine create screen.
 */
export const M2MCard: React.FC<M2MCardProps> = ({ onSelect }) => {
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
      checked={false}
      text={renderToString("CreateOAuthClientScreen.framework.m2m.title")}
      secondaryText={renderToString(
        "CreateOAuthClientScreen.framework.m2m.description"
      )}
      IconComponent={IconComponent}
      onClick={onClick}
    />
  );
};
