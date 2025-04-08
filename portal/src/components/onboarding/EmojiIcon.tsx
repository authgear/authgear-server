import React from "react";
import styles from "./EmojiIcon.module.css";

export function EmojiIcon({
  children,
}: {
  children?: React.ReactNode;
}): React.ReactElement {
  return <span className={styles.emojiIcon}>{children}</span>;
}
