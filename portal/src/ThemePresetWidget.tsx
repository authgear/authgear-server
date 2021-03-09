import React, { useCallback } from "react";
import cn from "classnames";
import {
  LightTheme,
  DarkTheme,
  isDarkThemeEqual,
  isLightThemeEqual,
} from "./util/theme";
import { Text, DefaultEffects, Icon } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import styles from "./ThemePresetWidget.module.scss";

export interface ThemePresetWidgetProps {
  className?: string;
  disabled?: boolean;
  isDarkMode: boolean;
  highlightedLightTheme?: LightTheme | null;
  highlightedDarkTheme?: DarkTheme | null;
  darkThemeIsCustom: boolean;
  lightThemeIsCustom: boolean;
  onClickLightTheme?: (lightTheme: LightTheme) => void;
  onClickDarkTheme?: (darkTheme: DarkTheme) => void;
  onClickCustom: () => void;
}

export const LIGHT_THEME_PRESETS: LightTheme[] = [
  {
    isDarkTheme: false,
    primaryColor: "#176df3",
    textColor: "#000000",
    backgroundColor: "#ffffff",
  },
  {
    isDarkTheme: false,
    primaryColor: "#00ce90",
    textColor: "#000000",
    backgroundColor: "#ffffff",
  },
  {
    isDarkTheme: false,
    primaryColor: "#f9597a",
    textColor: "#000000",
    backgroundColor: "#ffffff",
  },
  {
    isDarkTheme: false,
    primaryColor: "#000000",
    textColor: "#000000",
    backgroundColor: "#ffffff",
  },
  {
    isDarkTheme: false,
    primaryColor: "#46b1f9",
    textColor: "#000000",
    backgroundColor: "#ffffff",
  },
  {
    isDarkTheme: false,
    primaryColor: "#874bff",
    textColor: "#000000",
    backgroundColor: "#ffffff",
  },
  {
    isDarkTheme: false,
    primaryColor: "#ff8a00",
    textColor: "#000000",
    backgroundColor: "#ffffff",
  },
  {
    isDarkTheme: false,
    primaryColor: "#dbc26b",
    textColor: "#000000",
    backgroundColor: "#ffffff",
  },
];

export const DEFAULT_LIGHT_THEME = LIGHT_THEME_PRESETS[0];

export const DARK_THEME_PRESETS: DarkTheme[] = [
  {
    isDarkTheme: true,
    primaryColor: "#317BF4",
    textColor: "#ffffff",
    backgroundColor: "#000000",
  },
  {
    isDarkTheme: true,
    primaryColor: "#00ce90",
    textColor: "#ffffff",
    backgroundColor: "#000000",
  },
  {
    isDarkTheme: true,
    primaryColor: "#f9597a",
    textColor: "#ffffff",
    backgroundColor: "#000000",
  },
  {
    isDarkTheme: true,
    primaryColor: "#b6b6b6",
    textColor: "#ffffff",
    backgroundColor: "#000000",
  },
  {
    isDarkTheme: true,
    primaryColor: "#46b1f9",
    textColor: "#ffffff",
    backgroundColor: "#000000",
  },
  {
    isDarkTheme: true,
    primaryColor: "#874bff",
    textColor: "#ffffff",
    backgroundColor: "#000000",
  },
  {
    isDarkTheme: true,
    primaryColor: "#ff8a00",
    textColor: "#ffffff",
    backgroundColor: "#000000",
  },
  {
    isDarkTheme: true,
    primaryColor: "#dbc26b",
    textColor: "#ffffff",
    backgroundColor: "#000000",
  },
];

export const DEFAULT_DARK_THEME = DARK_THEME_PRESETS[0];

interface ThemePresetProps {
  disabled?: boolean;
  isDarkMode: boolean;
  isSelected: boolean;
  presetNameID: string;
  presetTheme: LightTheme | DarkTheme;
  onClickLightTheme?: (lightTheme: LightTheme) => void;
  onClickDarkTheme?: (darkTheme: DarkTheme) => void;
}

function ThemePreset(props: ThemePresetProps) {
  const {
    disabled,
    isDarkMode,
    isSelected,
    presetNameID,
    presetTheme,
    onClickLightTheme,
    onClickDarkTheme,
  } = props;
  const onClick = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (isDarkMode) {
        onClickDarkTheme?.(presetTheme as any);
      } else {
        onClickLightTheme?.(presetTheme as any);
      }
    },
    [presetTheme, isDarkMode, onClickLightTheme, onClickDarkTheme]
  );
  return (
    <div
      className={cn(
        styles.preset,
        isSelected && styles.selected,
        disabled && styles.disabled
      )}
      onClick={disabled ? undefined : onClick}
    >
      <div
        className={styles.background}
        style={{
          boxShadow: DefaultEffects.elevation4,
          backgroundColor: presetTheme.backgroundColor,
        }}
      >
        <div
          className={styles.foreground}
          style={{ backgroundColor: presetTheme.primaryColor }}
        >
          <Text
            style={{
              color: isDarkMode
                ? presetTheme.textColor
                : presetTheme.backgroundColor,
            }}
          >
            <FormattedMessage id="ThemeConfigurationWidget.sample-text" />
          </Text>
        </div>
      </div>
      <Text className={styles.presetName}>
        <FormattedMessage id={presetNameID} />
      </Text>
    </div>
  );
}

interface CustomProps {
  disabled?: boolean;
  isSelected: boolean;
  onClickCustom: () => void;
}

function Custom(props: CustomProps) {
  const { disabled, onClickCustom, isSelected } = props;
  return (
    <div
      className={cn(
        styles.preset,
        isSelected && styles.selected,
        disabled && styles.disabled
      )}
      onClick={disabled ? undefined : onClickCustom}
    >
      <Icon iconName="Add" className={styles.customIcon} />
      <Text className={styles.presetName}>
        <FormattedMessage id="ThemeConfigurationWidget.custom" />
      </Text>
    </div>
  );
}

const ThemePresetWidget: React.FC<ThemePresetWidgetProps> = function ThemePresetWidget(
  props: ThemePresetWidgetProps
) {
  const {
    disabled,
    className,
    isDarkMode,
    highlightedLightTheme,
    highlightedDarkTheme,
    darkThemeIsCustom,
    lightThemeIsCustom,
    onClickLightTheme,
    onClickDarkTheme,
    onClickCustom,
  } = props;
  const children = [];

  if (isDarkMode) {
    for (let i = 0; i < DARK_THEME_PRESETS.length; i++) {
      const presetTheme = DARK_THEME_PRESETS[i];
      const isSelected =
        highlightedDarkTheme != null &&
        isDarkThemeEqual(presetTheme, highlightedDarkTheme);
      children.push(
        <ThemePreset
          key={String(i)}
          disabled={disabled}
          isDarkMode={isDarkMode}
          isSelected={isSelected}
          presetNameID={"ThemeConfigurationWidget.preset." + String(i)}
          presetTheme={presetTheme}
          onClickLightTheme={onClickLightTheme}
          onClickDarkTheme={onClickDarkTheme}
        />
      );
    }
  } else {
    for (let i = 0; i < LIGHT_THEME_PRESETS.length; i++) {
      const presetTheme = LIGHT_THEME_PRESETS[i];
      const isSelected =
        highlightedLightTheme != null &&
        isLightThemeEqual(presetTheme, highlightedLightTheme);
      children.push(
        <ThemePreset
          key={String(i)}
          isDarkMode={isDarkMode}
          isSelected={isSelected}
          presetNameID={"ThemeConfigurationWidget.preset." + String(i)}
          presetTheme={presetTheme}
          onClickLightTheme={onClickLightTheme}
          onClickDarkTheme={onClickDarkTheme}
        />
      );
    }
  }

  children.push(
    <Custom
      key="custom"
      disabled={disabled}
      onClickCustom={onClickCustom}
      isSelected={isDarkMode ? darkThemeIsCustom : lightThemeIsCustom}
    />
  );

  return <div className={cn(styles.root, className)}>{children}</div>;
};

export default ThemePresetWidget;
