import React, { useContext } from "react";
import { Context } from "@oursky/react-messageformat";
import { Label, Text, Link } from "@fluentui/react";

export interface TextLinkProps {
  className?: string;
  label: React.ReactNode;
  value?: string | null;
}

// TextLink looks like TextField.
const TextLink: React.FC<TextLinkProps> = function TextLink(props) {
  const { className, label, value } = props;
  const { renderToString } = useContext(Context);
  return (
    <div className={className}>
      <Label>{label}</Label>
      {value != null && value !== "" ? (
        <Link className="mobile:break-words" href={value} target="_blank">
          {value}
        </Link>
      ) : (
        <Text>{renderToString("not-set")}</Text>
      )}
    </div>
  );
};

export default TextLink;
