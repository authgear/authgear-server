import React, { useContext, useMemo } from "react";
import cn from "classnames";
import styles from "./Logo.module.css";
import { Context } from "../../intl";
import { useAppearance } from "../../hook/useAppearance";

/**
 * Which mark to draw. "auto" picks from the appearance, which is what a logo on
 * the page background wants. Name one explicitly when the logo sits on a
 * surface that does not follow the appearance, such as the onboarding gradient.
 */
export type LogoVariant = "auto" | "color" | "white";

export function Logo({
  variant = "auto",
  containerClassName,
}: {
  variant?: LogoVariant;
  containerClassName?: string;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { resolved } = useAppearance();
  const colored =
    variant === "auto" ? resolved === "light" : variant === "color";
  const src = useMemo(
    () =>
      renderToString(colored ? "system.logo-inverted-uri" : "system.logo-uri"),
    [colored, renderToString]
  );

  return (
    <div className={cn(styles.logo__container, containerClassName)}>
      <img
        className={styles.logo__img}
        alt={renderToString("system.name")}
        src={src}
      />
    </div>
  );
}
