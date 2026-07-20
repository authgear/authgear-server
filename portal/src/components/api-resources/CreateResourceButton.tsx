import React from "react";
import { PlusIcon } from "@radix-ui/react-icons";
import { Context } from "../../intl";
import { PrimaryButton } from "../v2/Button/PrimaryButton/PrimaryButton";

interface CreateResourceButtonProps {
  className?: string;
  onClick: () => void;
}

export const CreateResourceButton: React.VFC<CreateResourceButtonProps> = ({
  className,
  onClick,
}) => {
  const { renderToString } = React.useContext(Context);
  return (
    <span className={className}>
      <PrimaryButton
        size="2"
        onClick={onClick}
        text={
          <span className="inline-flex items-center gap-1">
            <PlusIcon width="1rem" height="1rem" />
            {renderToString("CreateResourceButton.label")}
          </span>
        }
      />
    </span>
  );
};
