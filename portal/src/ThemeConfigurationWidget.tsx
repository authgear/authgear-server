import React, {
  useCallback,
  useRef,
  useState,
  useContext,
  useMemo,
} from "react";
import { Text, Label, Dropdown, IDropdownOption } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
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
  BannerConfiguration,
  DEFAULT_BANNER_CONFIGURATION,
} from "./util/theme";
import styles from "./ThemeConfigurationWidget.module.css";
import TextField from "./TextField";
import Toggle from "./Toggle";

export interface ThemeConfigurationWidgetProps {
  className?: string;

  lightTheme?: LightTheme | null;
  darkTheme?: DarkTheme | null;
  isDarkMode: boolean;
  darkModeEnabled: boolean;
  watermarkEnabled: boolean;
  onChangeLightTheme: (lightTheme: LightTheme) => void;
  onChangeDarkTheme: (darkTheme: DarkTheme) => void;
  onChangeDarkModeEnabled: (enabled: boolean) => void;
  onChangePrimaryColor: (color: string) => void;
  onChangeTextColor: (color: string) => void;
  onChangeBackgroundColor: (color: string) => void;

  appLogoValue: string | undefined;
  onChangeAppLogo: ImageFilePickerProps["onChange"];

  bannerConfiguration?: BannerConfiguration | null;
  onChangeBannerConfiguration?: (c: BannerConfiguration) => void;
}

type DropdownKey = "fixed-height" | "fixed-width";

const ThemeConfigurationWidget: React.VFC<ThemeConfigurationWidgetProps> =
  function ThemeConfigurationWidget(props: ThemeConfigurationWidgetProps) {
    const { renderToString } = useContext(Context);
    const previewWidgetRef = useRef<HTMLElement | null>(null);
    const {
      className,
      lightTheme,
      darkTheme,
      isDarkMode,
      darkModeEnabled,
      watermarkEnabled,
      appLogoValue,
      onChangeAppLogo,
      onChangeLightTheme: onChangeLightThemeProp,
      onChangeDarkTheme: onChangeDarkThemeProp,
      onChangeDarkModeEnabled,
      onChangePrimaryColor,
      onChangeTextColor,
      onChangeBackgroundColor,
      bannerConfiguration: bannerConfigurationProp,
      onChangeBannerConfiguration,
    } = props;

    const bannerConfiguration =
      bannerConfigurationProp ?? DEFAULT_BANNER_CONFIGURATION;

    const dropdownSelectedKey: DropdownKey =
      bannerConfiguration.width === "initial" ? "fixed-height" : "fixed-width";

    const onChangeDropdown = useCallback(
      (_e: React.FormEvent<HTMLDivElement>, item?: IDropdownOption) => {
        if (item == null) {
          return;
        }
        switch (item.key) {
          case "fixed-width":
            onChangeBannerConfiguration?.({
              ...bannerConfiguration,
              width: "100%",
              height: "initial",
            });
            break;
          case "fixed-height":
            onChangeBannerConfiguration?.({
              ...bannerConfiguration,
              height: DEFAULT_BANNER_CONFIGURATION.height,
              width: "initial",
            });
            break;
          default:
            break;
        }
      },
      [bannerConfiguration, onChangeBannerConfiguration]
    );

    const appLogoDimensionValue =
      dropdownSelectedKey === "fixed-height"
        ? bannerConfiguration.height
        : bannerConfiguration.width;

    const onChangeAppLogoDimensionValue = useCallback(
      (
        _e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        value?: string
      ) => {
        if (value == null) {
          return;
        }
        if (dropdownSelectedKey === "fixed-height") {
          onChangeBannerConfiguration?.({
            ...bannerConfiguration,
            height: value,
          });
        } else {
          onChangeBannerConfiguration?.({
            ...bannerConfiguration,
            width: value,
          });
        }
      },
      [bannerConfiguration, onChangeBannerConfiguration, dropdownSelectedKey]
    );

    const appLogoHorizontalPadding = bannerConfiguration.paddingLeft;
    const appLogoVerticalPadding = bannerConfiguration.paddingTop;
    const onChangeAppLogoHorizontalPadding = useCallback(
      (
        _e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        value?: string
      ) => {
        if (value == null) {
          return;
        }
        onChangeBannerConfiguration?.({
          ...bannerConfiguration,
          paddingLeft: value,
          paddingRight: value,
        });
      },
      [bannerConfiguration, onChangeBannerConfiguration]
    );
    const onChangeAppLogoVerticalPadding = useCallback(
      (
        _e: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>,
        value?: string
      ) => {
        if (value == null) {
          return;
        }
        onChangeBannerConfiguration?.({
          ...bannerConfiguration,
          paddingTop: value,
          paddingBottom: value,
        });
      },
      [bannerConfiguration, onChangeBannerConfiguration]
    );
    const onChangeAppLogoBackgroundColor = useCallback(
      (color: string) => {
        onChangeBannerConfiguration?.({
          ...bannerConfiguration,
          backgroundColor: color,
        });
      },
      [bannerConfiguration, onChangeBannerConfiguration]
    );

    const appLogoOptions = useMemo(() => {
      return [
        {
          key: "fixed-width",
          text: renderToString("ThemeConfigurationWidget.fixed-width"),
        },
        {
          key: "fixed-height",
          text: renderToString("ThemeConfigurationWidget.fixed-height"),
        },
      ];
    }, [renderToString]);

    const onChangeChecked = useCallback(
      (_e, checked) => {
        if (checked != null) {
          onChangeDarkModeEnabled(checked);
        }
      },
      [onChangeDarkModeEnabled]
    );

    const [customTheme, setCustomTheme] = useState<
      Omit<LightTheme | DarkTheme, "isDarkTheme">
    >(() => {
      if (isDarkMode) {
        if (darkTheme) {
          const { isDarkTheme, ...colors } = darkTheme;
          return colors;
        }
        const { isDarkTheme, ...colors } = DEFAULT_DARK_THEME;
        return colors;
      }
      if (lightTheme) {
        const { isDarkTheme, ...colors } = lightTheme;
        return colors;
      }
      const { isDarkTheme, ...colors } = DEFAULT_LIGHT_THEME;
      return colors;
    });

    const onPickerChangePrimaryColor = useCallback(
      (color: string) => {
        setCustomTheme((theme) => ({ ...theme, primaryColor: color }));
        onChangePrimaryColor(color);
      },
      [onChangePrimaryColor]
    );

    const onPickerChangeTextColor = useCallback(
      (color: string) => {
        setCustomTheme((theme) => ({ ...theme, textColor: color }));
        onChangeTextColor(color);
      },
      [onChangeTextColor]
    );

    const onPickerChangeBackgroundColor = useCallback(
      (color: string) => {
        setCustomTheme((theme) => ({ ...theme, backgroundColor: color }));
        onChangeBackgroundColor(color);
      },
      [onChangeBackgroundColor]
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
      if (isDarkMode) {
        setDarkThemeIsCustom(true);
        onChangeDarkThemeProp({ isDarkTheme: true, ...customTheme });
      } else {
        setLightThemeIsCustom(true);
        onChangeLightThemeProp({ isDarkTheme: false, ...customTheme });
      }
    }, [
      customTheme,
      isDarkMode,
      onChangeDarkThemeProp,
      onChangeLightThemeProp,
    ]);

    const disabled = isDarkMode && !darkModeEnabled;

    const colorControlsDisabled = isDarkMode
      ? !darkThemeIsCustom
      : !lightThemeIsCustom;

    const highlightedLightTheme = lightThemeIsCustom
      ? null
      : lightTheme ?? DEFAULT_LIGHT_THEME;

    const highlightedDarkTheme = darkThemeIsCustom
      ? null
      : darkTheme ?? DEFAULT_DARK_THEME;

    const primaryColor = isDarkMode
      ? darkThemeIsCustom
        ? customTheme.primaryColor
        : (darkTheme ?? DEFAULT_DARK_THEME).primaryColor
      : lightThemeIsCustom
      ? customTheme.primaryColor
      : (lightTheme ?? DEFAULT_LIGHT_THEME).primaryColor;

    const textColor = isDarkMode
      ? darkThemeIsCustom
        ? customTheme.textColor
        : (darkTheme ?? DEFAULT_DARK_THEME).textColor
      : lightThemeIsCustom
      ? customTheme.textColor
      : (lightTheme ?? DEFAULT_LIGHT_THEME).textColor;

    const backgroundColor = isDarkMode
      ? darkThemeIsCustom
        ? customTheme.backgroundColor
        : (darkTheme ?? DEFAULT_DARK_THEME).backgroundColor
      : lightThemeIsCustom
      ? customTheme.backgroundColor
      : (lightTheme ?? DEFAULT_LIGHT_THEME).backgroundColor;

    return (
      <Widget className={className}>
        <div className={styles.root}>
          <div className={styles.titleSection}>
            {isDarkMode ? (
              <Toggle
                toggleClassName={styles.toggle}
                checked={darkModeEnabled}
                onChange={onChangeChecked}
              />
            ) : null}
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

          <div className={styles.themeColorSection}>
            <Label>
              <FormattedMessage id="ThemeConfigurationWidget.theme-color-title" />
            </Label>
            <ThemePresetWidget
              disabled={disabled}
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
                onChange={onPickerChangePrimaryColor}
                disabled={disabled || colorControlsDisabled}
              />
            </div>
            <div className={styles.colorControl}>
              <Label className={styles.colorControlLabel}>
                <FormattedMessage id="ThemeConfigurationWidget.text-color" />
              </Label>
              <PortalColorPicker
                className={styles.colorPicker}
                color={textColor}
                onChange={onPickerChangeTextColor}
                disabled={disabled || colorControlsDisabled}
              />
            </div>
            <div className={styles.colorControl}>
              <Label className={styles.colorControlLabel}>
                <FormattedMessage id="ThemeConfigurationWidget.background-color" />
              </Label>
              <PortalColorPicker
                className={styles.colorPicker}
                color={backgroundColor}
                onChange={onPickerChangeBackgroundColor}
                disabled={disabled || colorControlsDisabled}
              />
            </div>

            <div className={styles.appLogoSection}>
              <Label>
                <FormattedMessage id="ThemeConfigurationWidget.app-logo-title" />
              </Label>
              <Text variant="small" className={styles.themeColorTitle}>
                <FormattedMessage id="ThemeConfigurationWidget.app-logo-description" />
              </Text>
              <ImageFilePicker
                sizeLimitInBytes={100 * 1000}
                disabled={disabled}
                base64EncodedData={appLogoValue}
                onChange={onChangeAppLogo}
              />
              <div className={styles.appLogoControl}>
                <Dropdown
                  className={styles.appLogoDropdown}
                  label={renderToString(
                    "ThemeConfigurationWidget.app-logo-dropown-title"
                  )}
                  options={appLogoOptions}
                  selectedKey={dropdownSelectedKey}
                  onChange={onChangeDropdown}
                  disabled={disabled}
                />
                <TextField
                  className={styles.appLogoDimension}
                  label={renderToString("ThemeConfigurationWidget.value")}
                  value={appLogoDimensionValue}
                  onChange={onChangeAppLogoDimensionValue}
                  disabled={disabled}
                />
              </div>
              <div className={styles.appLogoControl}>
                <TextField
                  className={styles.appLogoPadding}
                  label={renderToString(
                    "ThemeConfigurationWidget.left-right-padding"
                  )}
                  value={appLogoHorizontalPadding}
                  onChange={onChangeAppLogoHorizontalPadding}
                  disabled={disabled}
                />
                <TextField
                  className={styles.appLogoPadding}
                  label={renderToString(
                    "ThemeConfigurationWidget.top-bottom-padding"
                  )}
                  value={appLogoVerticalPadding}
                  onChange={onChangeAppLogoVerticalPadding}
                  disabled={disabled}
                />
              </div>
              <div className={styles.colorControl}>
                <Label className={styles.colorControlLabel}>
                  <FormattedMessage id="ThemeConfigurationWidget.background-color" />
                </Label>
                <PortalColorPicker
                  className={styles.colorPicker}
                  color={bannerConfiguration.backgroundColor}
                  onChange={onChangeAppLogoBackgroundColor}
                  disabled={disabled}
                  alphaType="alpha"
                />
              </div>
            </div>
          </div>
          <div className={styles.previewSection}>
            <Label>
              <FormattedMessage id="ThemeConfigurationWidget.preview-mode" />
            </Label>
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
                watermarkEnabled={watermarkEnabled}
                appLogoValue={appLogoValue}
                bannerConfiguration={bannerConfiguration}
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
