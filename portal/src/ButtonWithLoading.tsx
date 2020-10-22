import React, { useContext, useMemo } from "react";
import {
  IButtonProps,
  ISpinnerProps,
  PrimaryButton,
  Spinner,
  SpinnerSize,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useSystemConfig } from "./context/SystemConfigContext";

interface ButtonWithLoadingProps extends IButtonProps {
  loading: boolean;
  labelId: string;
  loadingLabelId?: string;
  spinnerStyles?: ISpinnerProps["styles"];
}

const ButtonWithLoading: React.FC<ButtonWithLoadingProps> = function ButtonWithLoading(
  props: ButtonWithLoadingProps
) {
  const {
    loading,
    labelId,
    loadingLabelId,
    spinnerStyles,
    disabled: disabledProps,
    ...rest
  } = props;
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();

  const disabled = loading || disabledProps;
  const textColor = useMemo(() => {
    const buttonTheme = props.theme ?? themes.main;
    const normalTextColor = buttonTheme.palette.white;
    const disableTextColor = buttonTheme.palette.neutralTertiary;
    return disabled ? disableTextColor : normalTextColor;
  }, [disabled, props.theme, themes.main]);

  return (
    <PrimaryButton disabled={disabled} {...rest}>
      {loading ? (
        <Spinner
          label={renderToString(loadingLabelId ?? labelId)}
          size={SpinnerSize.xSmall}
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
