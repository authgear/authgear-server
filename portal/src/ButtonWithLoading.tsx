import React, { useContext } from "react";
import {
  IButtonProps,
  ISpinnerProps,
  PrimaryButton,
  Spinner,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

interface ButtonWithLoadingProps extends IButtonProps {
  loading: boolean;
  labelId: string;
  loadingLabelId?: string;
  spinnerStyles?: ISpinnerProps["styles"];
}

const ButtonWithLoading: React.FC<ButtonWithLoadingProps> = function ButtonWithLoading(
  props: ButtonWithLoadingProps
) {
  const { loading, labelId, loadingLabelId, spinnerStyles, ...rest } = props;
  const { renderToString } = useContext(Context);

  return (
    <PrimaryButton {...rest}>
      {loading ? (
        <Spinner
          label={renderToString(loadingLabelId ?? labelId)}
          styles={spinnerStyles ?? { label: { color: "white" } }}
          ariaLive="assertive"
          labelPosition="left"
        />
      ) : (
        <FormattedMessage id={labelId} />
      )}
    </PrimaryButton>
  );
};

export default ButtonWithLoading;
