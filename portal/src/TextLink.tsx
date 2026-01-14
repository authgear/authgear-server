import React, { useContext } from "react";
import { Context } from "./intl";
import { Label, Text } from "@fluentui/react";
import ExternalLink from "./ExternalLink";
import styles from "./TextLink.module.css";

export interface TextLinkProps {
  className?: string;
  label: React.ReactNode;
  value?: string | null;
}

// TextLink looks like TextField.
const TextLink: React.VFC<TextLinkProps> = function TextLink(props) {
  const { className, label, value } = props;
  const { renderToString } = useContext(Context);
  return (
    <div className={className}>
      <Label>{label}</Label>
      {value != null && value !== "" ? (
        <ExternalLink className={styles.link} href={value}>
          {value}
        </ExternalLink>
      ) : (
        <Text>{renderToString("not-set")}</Text>
      )}
    </div>
  );
};

export default TextLink;
