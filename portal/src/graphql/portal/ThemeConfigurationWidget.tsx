import React, { useCallback, useRef } from "react";
import { DefaultEffects, Text, Label, Toggle } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import PortalColorPicker from "../../PortalColorPicker";
import ScaleContainer from "../../ScaleContainer";
import ThemePreviewWidget from "../../ThemePreviewWidget";
import styles from "./ThemeConfigurationWidget.module.scss";

export interface ThemeConfigurationWidgetProps {
  isDarkMode: boolean;
  darkModeEnabled: boolean;
  onChangeDarkModeEnabled: (enabled: boolean) => void;
  primaryColor: string;
  onChangePrimaryColor: (color: string) => void;
  textColor: string;
  onChangeTextColor: (color: string) => void;
  backgroundColor: string;
  onChangeBackgroundColor: (color: string) => void;
}

const ThemeConfigurationWidget: React.FC<ThemeConfigurationWidgetProps> = function ThemeConfigurationWidget(
  props: ThemeConfigurationWidgetProps
) {
  const previewWidgetRef = useRef<HTMLElement | null>(null);
  const {
    isDarkMode,
    darkModeEnabled,
    onChangeDarkModeEnabled,
    primaryColor,
    onChangePrimaryColor,
    textColor,
    onChangeTextColor,
    backgroundColor,
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

  return (
    <div
      className={styles.root}
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
        <div className={styles.colorControlSection}>
          <Text as="h2" className={styles.colorControlTitle}>
            <FormattedMessage id="ThemeConfigurationWidget.custom-color" />
          </Text>
          <div className={styles.colorControl}>
            <Label className={styles.colorControlLabel}>
              <FormattedMessage id="ThemeConfigurationWidget.primary-color" />
            </Label>
            <PortalColorPicker
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
              color={backgroundColor}
              onChange={onChangeBackgroundColor}
              disabled={isDarkMode && !darkModeEnabled}
            />
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
