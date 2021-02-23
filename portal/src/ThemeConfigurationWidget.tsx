import React, { useCallback, useRef } from "react";
import cn from "classnames";
import { DefaultEffects, Text, Label, Toggle } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScaleContainer from "./ScaleContainer";
import ThemePreviewWidget from "./ThemePreviewWidget";
import PortalColorPicker from "./PortalColorPicker";
import ThemePresetWidget, {
  DEFAULT_LIGHT_THEME,
  DEFAULT_DARK_THEME,
} from "./ThemePresetWidget";
import { LightTheme, DarkTheme } from "./util/theme";
import styles from "./ThemeConfigurationWidget.module.scss";

export interface ThemeConfigurationWidgetProps {
  className?: string;
  lightTheme?: LightTheme | null;
  darkTheme?: DarkTheme | null;
  isDarkMode: boolean;
  darkModeEnabled: boolean;
  onChangeLightTheme: (lightTheme: LightTheme) => void;
  onChangeDarkTheme: (darkTheme: DarkTheme) => void;
  onChangeDarkModeEnabled: (enabled: boolean) => void;
  onChangePrimaryColor: (color: string) => void;
  onChangeTextColor: (color: string) => void;
  onChangeBackgroundColor: (color: string) => void;
}

/*
 */

// eslint-disable-next-line complexity
const ThemeConfigurationWidget: React.FC<ThemeConfigurationWidgetProps> = function ThemeConfigurationWidget(
  props: ThemeConfigurationWidgetProps
) {
  const previewWidgetRef = useRef<HTMLElement | null>(null);
  const {
    className,
    lightTheme,
    darkTheme,
    isDarkMode,
    darkModeEnabled,
    onChangeLightTheme,
    onChangeDarkTheme,
    onChangeDarkModeEnabled,
    onChangePrimaryColor,
    onChangeTextColor,
    onChangeBackgroundColor,
  } = props;

  const onChangeChecked = useCallback(
    (_e, checked) => {
      if (checked != null) {
        onChangeDarkModeEnabled(checked);
      }
    },
    [onChangeDarkModeEnabled]
  );

  const primaryColor = isDarkMode
    ? (darkTheme ?? DEFAULT_DARK_THEME).primaryColor
    : (lightTheme ?? DEFAULT_LIGHT_THEME).primaryColor;

  const textColor = isDarkMode
    ? (darkTheme ?? DEFAULT_DARK_THEME).textColor
    : (lightTheme ?? DEFAULT_LIGHT_THEME).textColor;

  const backgroundColor = isDarkMode
    ? (darkTheme ?? DEFAULT_DARK_THEME).backgroundColor
    : (lightTheme ?? DEFAULT_LIGHT_THEME).backgroundColor;

  return (
    <div
      className={cn(styles.root, className)}
      style={{ boxShadow: DefaultEffects.elevation4 }}
    >
      <div className={styles.titleSection}>
        {isDarkMode && (
          <Toggle
            className={styles.toggle}
            checked={darkModeEnabled}
            onChange={onChangeChecked}
          />
        )}
        <Text as="h1" className={styles.title}>
          <FormattedMessage
            id={
              isDarkMode
                ? "ThemeConfigurationWidget.dark-mode"
                : "ThemeConfigurationWidget.light-mode"
            }
          />
        </Text>
      </div>
      <div className={styles.rootSection}>
        <div className={styles.leftSection}>
          <div className={styles.presetSection}>
            <Text as="h2" className={styles.presetTitle}>
              <FormattedMessage id="ThemeConfigurationWidget.preset-title" />
            </Text>
            <ThemePresetWidget
              className={styles.presetWidget}
              isDarkMode={isDarkMode}
              lightTheme={lightTheme}
              darkTheme={darkTheme}
              onClickLightTheme={onChangeLightTheme}
              onClickDarkTheme={onChangeDarkTheme}
            />
          </div>
          <div className={styles.colorControlSection}>
            <Text as="h2" className={styles.colorControlTitle}>
              <FormattedMessage id="ThemeConfigurationWidget.custom-color" />
            </Text>
            <div className={styles.colorControl}>
              <Label className={styles.colorControlLabel}>
                <FormattedMessage id="ThemeConfigurationWidget.primary-color" />
              </Label>
              <PortalColorPicker
                className={styles.colorPicker}
                color={primaryColor}
                onChange={onChangePrimaryColor}
                disabled={isDarkMode && !darkModeEnabled}
              />
            </div>
            <div className={styles.colorControl}>
              <Label className={styles.colorControlLabel}>
                <FormattedMessage id="ThemeConfigurationWidget.text-color" />
              </Label>
              <PortalColorPicker
                className={styles.colorPicker}
                color={textColor}
                onChange={onChangeTextColor}
                disabled={isDarkMode && !darkModeEnabled}
              />
            </div>
            <div className={styles.colorControl}>
              <Label className={styles.colorControlLabel}>
                <FormattedMessage id="ThemeConfigurationWidget.background-color" />
              </Label>
              <PortalColorPicker
                className={styles.colorPicker}
                color={backgroundColor}
                onChange={onChangeBackgroundColor}
                disabled={isDarkMode && !darkModeEnabled}
              />
            </div>
          </div>
        </div>
        <div className={styles.previewSection}>
          <Text as="h2" className={styles.colorControlTitle}>
            <FormattedMessage id="ThemeConfigurationWidget.preview-mode" />
          </Text>
          <ScaleContainer
            className={styles.previewContainer}
            childrenRef={previewWidgetRef}
            mode="fixed-width"
          >
            <ThemePreviewWidget
              /* @ts-expect-error */
              ref={previewWidgetRef}
              className={styles.previewWidget}
              isDarkMode={isDarkMode}
              primaryColor={primaryColor}
              textColor={textColor}
              backgroundColor={backgroundColor}
            />
          </ScaleContainer>
        </div>
      </div>
    </div>
  );
};

export default ThemeConfigurationWidget;
