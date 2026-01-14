import React, { useContext, useMemo } from "react";
import styles from "./Logo.module.css";
import { Context } from "../../intl";

export function Logo({ inverted }: { inverted?: boolean }): React.ReactElement {
  const { renderToString } = useContext(Context);
  const src = useMemo(() => {
    if ((import.meta as any).env.DEV) {
      // In local, system.logo-inverted-uri does not exist, use the image in production for development
      return inverted
        ? "https://portal.authgear.com/img/logo-inverted.png"
        : "https://portal.authgear.com/img/logo.png";
    }
    return renderToString(
      inverted ? "system.logo-inverted-uri" : "system.logo-uri"
    );
  }, [inverted, renderToString]);

  return (
    <div className={styles.logo__container}>
      <img
        className={styles.logo__img}
        alt={renderToString("system.name")}
        src={src}
      />
    </div>
  );
}
