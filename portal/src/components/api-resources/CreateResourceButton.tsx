import React, { useContext } from "react";
import { Context } from "../../intl";
import PrimaryButton from "../../PrimaryButton";
import { useNavigate, useParams } from "react-router-dom";

interface CreateResourceButtonProps {
  className?: string;
}

export const CreateResourceButton: React.VFC<CreateResourceButtonProps> = ({
  className,
}) => {
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const { appID } = useParams<{ appID: string }>();
  const handleClick = React.useCallback(() => {
    navigate(
      `/project/${encodeURIComponent(appID ?? "")}/api-resources/create`
    );
  }, [navigate, appID]);
  return (
    <PrimaryButton
      text={renderToString("CreateResourceButton.label")}
      iconProps={{ iconName: "Add" }}
      onClick={handleClick}
      className={className}
    />
  );
};
