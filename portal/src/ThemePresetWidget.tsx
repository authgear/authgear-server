import React, { useCallback } from "react";
import cn from "classnames";
import {
  LightTheme,
  DarkTheme,
  isLightThemeEqual,
  isDarkThemeEqual,
} from "./util/theme";
import { Text, DefaultEffects } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import styles from "./ThemePresetWidget.module.scss";

export interface ThemePresetWidgetProps {
  className?: string;
  isDarkMode: boolean;
  lightTheme?: LightTheme | null;
  darkTheme?: DarkTheme | null;
  onClickLightTheme?: (lightTheme: LightTheme) => void;
  onClickDarkTheme?: (darkTheme: DarkTheme) => void;
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
];

export const DEFAULT_DARK_THEME = DARK_THEME_PRESETS[0];

interface ThemePresetProps {
  isDarkMode: boolean;
  index: number;
  lightTheme?: LightTheme | null;
  darkTheme?: DarkTheme | null;
  onClickLightTheme?: (lightTheme: LightTheme) => void;
  onClickDarkTheme?: (darkTheme: DarkTheme) => void;
}

function ThemePreset(props: ThemePresetProps) {
  const {
    isDarkMode,
    index,
    lightTheme,
    darkTheme,
    onClickLightTheme,
    onClickDarkTheme,
  } = props;
  const theme = isDarkMode
    ? DARK_THEME_PRESETS[index]
    : LIGHT_THEME_PRESETS[index];
  const isSelected = isDarkMode
    ? isDarkThemeEqual(theme as any, darkTheme ?? DEFAULT_DARK_THEME)
    : isLightThemeEqual(theme as any, lightTheme ?? DEFAULT_LIGHT_THEME);
  const onClick = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (isDarkMode) {
        onClickDarkTheme?.(theme as any);
      } else {
        onClickLightTheme?.(theme as any);
      }
    },
    [theme, isDarkMode, onClickLightTheme, onClickDarkTheme]
  );
  return (
    <div
      className={cn(styles.preset, isSelected && styles.selected)}
      onClick={onClick}
    >
      <div
        className={styles.background}
        style={{
          boxShadow: DefaultEffects.elevation4,
          backgroundColor: theme.backgroundColor,
        }}
      >
        <div
          className={styles.foreground}
          style={{ backgroundColor: theme.primaryColor }}
        >
          <Text
            style={{
              color: isDarkMode ? theme.textColor : theme.backgroundColor,
            }}
          >
            <FormattedMessage id="ThemeConfigurationWidget.sample-text" />
          </Text>
        </div>
      </div>
      <Text className={styles.presetName}>
        <FormattedMessage
          id={"ThemeConfigurationWidget.preset." + String(index)}
        />
      </Text>
    </div>
  );
}

const ThemePresetWidget: React.FC<ThemePresetWidgetProps> = function ThemePresetWidget(
  props: ThemePresetWidgetProps
) {
  const {
    className,
    isDarkMode,
    lightTheme,
    darkTheme,
    onClickLightTheme,
    onClickDarkTheme,
  } = props;
  const children = [];
  for (let i = 0; i < LIGHT_THEME_PRESETS.length; i++) {
    children.push(
      <ThemePreset
        key={String(i)}
        index={i}
        isDarkMode={isDarkMode}
        lightTheme={lightTheme}
        darkTheme={darkTheme}
        onClickLightTheme={onClickLightTheme}
        onClickDarkTheme={onClickDarkTheme}
      />
    );
  }
  return <div className={cn(styles.root, className)}>{children}</div>;
};

export default ThemePresetWidget;
