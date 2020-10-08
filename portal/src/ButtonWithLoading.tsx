import React, { useContext, useMemo } from "react";
import {
  IButtonProps,
  ISpinnerProps,
  PrimaryButton,
  Spinner,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { theme } from "./theme";

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

  const textColor = useMemo(() => {
    const buttonTheme = props.theme ?? theme;
    const normalTextColor = buttonTheme.palette.white;
    const disableTextColor = buttonTheme.palette.neutralTertiary;
    return props.disabled ? disableTextColor : normalTextColor;
  }, [props.disabled, props.theme]);

  return (
    <PrimaryButton
      style={{ pointerEvents: loading ? "none" : undefined }}
      {...rest}
    >
      {loading ? (
        <Spinner
          label={renderToString(loadingLabelId ?? labelId)}
          styles={spinnerStyles ?? { label: { color: textColor } }}
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
