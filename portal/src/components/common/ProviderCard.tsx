import React from "react";
import cn from "classnames";
import {
  DefaultEffects,
  FontIcon,
  IButtonProps,
  IIconProps,
  Image,
  Label,
  Text,
} from "@fluentui/react";
import { useSystemConfig } from "../../context/SystemConfigContext";
import styles from "./ProviderCard.module.css";

interface ProviderCardProps {
  className?: string;
  iconProps?: IIconProps;
  logoSrc?: any;
  children?: React.ReactNode;
  onClick?: IButtonProps["onClick"];
  isSelected?: boolean;
  disabled?: boolean;
}

const PROVIDER_CARD_ICON_STYLE = {
  width: "32px",
  height: "32px",
  fontSize: "32px",
};

export function ProviderCard(props: ProviderCardProps): React.ReactElement {
  const {
    className,
    disabled,
    isSelected,
    children,
    onClick,
    iconProps,
    logoSrc,
  } = props;
  const {
    themes: {
      main: {
        palette: { themePrimary },
        semanticColors: { disabledBackground: backgroundColor },
      },
    },
  } = useSystemConfig();
  return (
    <div
      style={{
        boxShadow: disabled ? undefined : DefaultEffects.elevation4,
        borderColor: isSelected ? themePrimary : "transparent",
        backgroundColor: disabled ? backgroundColor : undefined,
        cursor: disabled ? "not-allowed" : undefined,
      }}
      className={cn(className, styles.providerCard)}
      onClick={disabled ? undefined : onClick}
    >
      {iconProps != null ? (
        <FontIcon {...iconProps} style={PROVIDER_CARD_ICON_STYLE} />
      ) : null}
      {logoSrc != null ? <Image src={logoSrc} width={32} height={32} /> : null}
      <Label>{children}</Label>
    </div>
  );
}

interface ProviderDescriptionProps {
  children?: React.ReactNode;
}

export function ProviderCardDescription(
  props: ProviderDescriptionProps
): React.ReactElement {
  const { children } = props;
  const {
    themes: {
      main: {
        semanticColors: { bodySubtext: color },
      },
    },
  } = useSystemConfig();

  return (
    <Text
      variant="small"
      block={true}
      style={{
        color,
      }}
      className={styles.columnFull}
    >
      {children}
    </Text>
  );
}
