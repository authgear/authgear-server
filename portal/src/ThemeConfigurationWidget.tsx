import React, { useCallback, useRef, useState } from "react";
import cn from "classnames";
import { Text, Label, Toggle } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ScaleContainer from "./ScaleContainer";
import Widget from "./Widget";
import WidgetTitle from "./WidgetTitle";
import ThemePreviewWidget from "./ThemePreviewWidget";
import PortalColorPicker from "./PortalColorPicker";
import ImageFilePicker, { ImageFilePickerProps } from "./ImageFilePicker";
import ThemePresetWidget, {
  DEFAULT_LIGHT_THEME,
  DEFAULT_DARK_THEME,
  LIGHT_THEME_PRESETS,
  DARK_THEME_PRESETS,
} from "./ThemePresetWidget";
import {
  LightTheme,
  DarkTheme,
  isLightThemeEqual,
  isDarkThemeEqual,
} from "./util/theme";
import styles from "./ThemeConfigurationWidget.module.scss";

export interface ThemeConfigurationWidgetProps {
  className?: string;
  lightTheme?: LightTheme | null;
  darkTheme?: DarkTheme | null;
  isDarkMode: boolean;
  darkModeEnabled: boolean;
  appLogoValue: string | undefined;
  onChangeAppLogo: ImageFilePickerProps["onChange"];
  onChangeLightTheme: (lightTheme: LightTheme) => void;
  onChangeDarkTheme: (darkTheme: DarkTheme) => void;
  onChangeDarkModeEnabled: (enabled: boolean) => void;
  onChangePrimaryColor: (color: string) => void;
  onChangeTextColor: (color: string) => void;
  onChangeBackgroundColor: (color: string) => void;
}

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
    appLogoValue,
    onChangeAppLogo,
    onChangeLightTheme: onChangeLightThemeProp,
    onChangeDarkTheme: onChangeDarkThemeProp,
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

  const [darkThemeIsCustom, setDarkThemeIsCustom] = useState(() => {
    let equal = false;
    for (const theme of DARK_THEME_PRESETS) {
      if (isDarkThemeEqual(theme, darkTheme ?? DEFAULT_DARK_THEME)) {
        equal = true;
      }
    }
    return !equal;
  });

  const [lightThemeIsCustom, setLightThemeIsCustom] = useState(() => {
    let equal = false;
    for (const theme of LIGHT_THEME_PRESETS) {
      if (isLightThemeEqual(theme, lightTheme ?? DEFAULT_LIGHT_THEME)) {
        equal = true;
      }
    }
    return !equal;
  });

  const onChangeLightTheme = useCallback(
    (lightTheme: LightTheme) => {
      setLightThemeIsCustom(false);
      onChangeLightThemeProp(lightTheme);
    },
    [onChangeLightThemeProp]
  );

  const onChangeDarkTheme = useCallback(
    (darkTheme: DarkTheme) => {
      setDarkThemeIsCustom(false);
      onChangeDarkThemeProp(darkTheme);
    },
    [onChangeDarkThemeProp]
  );

  const onClickCustom = useCallback(() => {
    setLightThemeIsCustom(true);
    setDarkThemeIsCustom(true);
  }, []);

  const disabled =
    (isDarkMode && !darkModeEnabled) ||
    (isDarkMode ? !darkThemeIsCustom : !lightThemeIsCustom);

  const highlightedLightTheme = lightThemeIsCustom
    ? null
    : lightTheme ?? DEFAULT_LIGHT_THEME;

  const highlightedDarkTheme = darkThemeIsCustom
    ? null
    : darkTheme ?? DEFAULT_DARK_THEME;

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
    <Widget className={cn(styles.root, className)}>
      <div className={styles.titleSection}>
        {isDarkMode && (
          <Toggle
            className={styles.toggle}
            checked={darkModeEnabled}
            onChange={onChangeChecked}
          />
        )}
        <WidgetTitle>
          <FormattedMessage
            id={
              isDarkMode
                ? "ThemeConfigurationWidget.dark-mode"
                : "ThemeConfigurationWidget.light-mode"
            }
          />
        </WidgetTitle>
      </div>
      <div className={styles.rootSection}>
        <div className={styles.leftSection}>
          <div className={styles.themeColorSection}>
            <Text as="h3" className={styles.sectionTitle}>
              <FormattedMessage id="ThemeConfigurationWidget.theme-color-title" />
            </Text>
            <ThemePresetWidget
              className={styles.presetWidget}
              isDarkMode={isDarkMode}
              highlightedLightTheme={highlightedLightTheme}
              highlightedDarkTheme={highlightedDarkTheme}
              darkThemeIsCustom={darkThemeIsCustom}
              lightThemeIsCustom={lightThemeIsCustom}
              onClickLightTheme={onChangeLightTheme}
              onClickDarkTheme={onChangeDarkTheme}
              onClickCustom={onClickCustom}
            />
            <div className={styles.colorControl}>
              <Label className={styles.colorControlLabel}>
                <FormattedMessage id="ThemeConfigurationWidget.primary-color" />
              </Label>
              <PortalColorPicker
                className={styles.colorPicker}
                color={primaryColor}
                onChange={onChangePrimaryColor}
                disabled={disabled}
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
                disabled={disabled}
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
                disabled={disabled}
              />
            </div>
          </div>
          <div className={styles.appLogoSection}>
            <Text as="h3" className={styles.sectionTitle}>
              <FormattedMessage id="ThemeConfigurationWidget.app-logo-title" />
            </Text>
            <Text variant="small" className={styles.themeColorTitle}>
              <FormattedMessage id="ThemeConfigurationWidget.app-logo-description" />
            </Text>
            <ImageFilePicker
              base64EncodedData={appLogoValue}
              onChange={onChangeAppLogo}
            />
          </div>
        </div>
        <div className={styles.previewSection}>
          <Text as="h3" className={styles.sectionTitle}>
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
              appLogoValue={appLogoValue}
              primaryColor={primaryColor}
              textColor={textColor}
              backgroundColor={backgroundColor}
            />
          </ScaleContainer>
        </div>
      </div>
    </Widget>
  );
};

export default ThemeConfigurationWidget;
