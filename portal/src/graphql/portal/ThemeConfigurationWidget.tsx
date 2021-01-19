import React from "react";
import { DefaultEffects, Text, Label } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import PortalColorPicker from "../../PortalColorPicker";
import styles from "./ThemeConfigurationWidget.module.scss";

export interface ThemeConfigurationWidgetProps {
  primaryColor: string;
  onChangePrimaryColor: (color: string) => void;
  textColor: string;
  onChangeTextColor: (color: string) => void;
  backgroundColor: string;
  onChangeBackgroundColor: (color: string) => void;
}

export const LIGHT_MODE_PRIMARY_COLOR = "#176df3";
export const LIGHT_MODE_TEXT_COLOR = "#000000";
export const LIGHT_MODE_BACKGROUND_COLOR = "#ffffff";

export const DARK_MODE_PRIMARY_COLOR = "#317BF4";
export const DARK_MODE_TEXT_COLOR = "#ffffff";
export const DARK_MODE_BACKGROUND_COLOR = "#000000";

const ThemeConfigurationWidget: React.FC<ThemeConfigurationWidgetProps> = function ThemeConfigurationWidget(
  props: ThemeConfigurationWidgetProps
) {
  const {
    primaryColor,
    onChangePrimaryColor,
    textColor,
    onChangeTextColor,
    backgroundColor,
    onChangeBackgroundColor,
  } = props;

  return (
    <div
      className={styles.root}
      style={{ boxShadow: DefaultEffects.elevation4 }}
    >
      <Text as="h1" className={styles.title}>
        <FormattedMessage id="ThemeConfigurationWidget.light-mode" />
      </Text>
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
          />
        </div>
        <div className={styles.colorControl}>
          <Label className={styles.colorControlLabel}>
            <FormattedMessage id="ThemeConfigurationWidget.text-color" />
          </Label>
          <PortalColorPicker color={textColor} onChange={onChangeTextColor} />
        </div>
        <div className={styles.colorControl}>
          <Label className={styles.colorControlLabel}>
            <FormattedMessage id="ThemeConfigurationWidget.background-color" />
          </Label>
          <PortalColorPicker
            color={backgroundColor}
            onChange={onChangeBackgroundColor}
          />
        </div>
      </div>
    </div>
  );
};

export default ThemeConfigurationWidget;
