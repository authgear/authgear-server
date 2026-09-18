import React, { useContext, useMemo } from "react";
import cn from "classnames";
import styles from "./Logo.module.css";
import { Context } from "../../intl";
import { useAppearance } from "../../hook/useAppearance";

export function Logo({
  inverted,
  containerClassName,
}: {
  inverted?: boolean;
  containerClassName?: string;
}): React.ReactElement {
  const { renderToString } = useContext(Context);
  const { resolved } = useAppearance();
  // `inverted` asks for the colored logo, which only reads well on a light
  // surface; in dark mode the plain (white) logo is shown instead.
  const colored = inverted === true && resolved === "light";
  const src = useMemo(() => {
    if ((import.meta as any).env.DEV) {
      // In local, system.logo-inverted-uri does not exist, use the image in production for development
      return colored
        ? "https://portal.authgear.com/img/logo-inverted.png"
        : "https://portal.authgear.com/img/logo.png";
    }
    return renderToString(
      colored ? "system.logo-inverted-uri" : "system.logo-uri"
    );
  }, [colored, renderToString]);

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
