import React, { useContext } from "react";
import { Context } from "@oursky/react-messageformat";
import PrimaryButton from "../../PrimaryButton";

interface CreateResourceButtonProps {
  onClick: () => void;
  className?: string;
}

export const CreateResourceButton: React.VFC<CreateResourceButtonProps> = ({
  onClick,
  className,
}) => {
  const { renderToString } = useContext(Context);
  return (
    <PrimaryButton
      text={renderToString("CreateResourceButton.label")}
      iconProps={{ iconName: "Add" }}
      onClick={onClick}
      className={className}
    />
  );
};
