import React, { useMemo } from "react";
import { getShades } from "./util/theme";
import styles from "./ThemePreviewWidget.module.scss";

export interface ThemePreviewWidgetProps {
  isDarkMode: boolean;
  primaryColor: string;
  textColor: string;
  backgroundColor: string;
}

interface GetStyleOptions {
  primaryColor: string;
  textColor: string;
  backgroundColor: string;
}

function getThemeStyle(name: string, shades: string[]) {
  const style: Record<string, string> = {};
  for (let i = 0; i < shades.length; i++) {
    if (i === 0) {
      style[`--color-${name}-unshaded`] = shades[i];
    } else {
      style[`--color-${name}-shaded-${i}`] = shades[i];
    }
  }
  return style;
}

function getLightModeStyle(options: GetStyleOptions) {
  const { primaryColor, textColor, backgroundColor } = options;
  const primaryStyle = getThemeStyle("primary", getShades(primaryColor));
  const textStyle = getThemeStyle("text", getShades(textColor));
  const backgroundStyle = getThemeStyle(
    "background",
    getShades(backgroundColor)
  );
  const lightModeStyle = {
    "--color-error-unshaded": "#e23d3d",
    "--color-error-shaded-1": "#fef6f6",
    "--color-error-shaded-2": "#fbdddd",
    "--color-error-shaded-3": "#f7c1c1",
    "--color-error-shaded-4": "#ee8686",
    "--color-error-shaded-5": "#e65252",
    "--color-error-shaded-6": "#cc3737",
    "--color-error-shaded-7": "#ac2f2f",
    "--color-error-shaded-8": "#7f2222",

    "--color-white-unshaded": "#ffffff",
    "--color-white-shaded-1": "#767676",
    "--color-white-shaded-2": "#a6a6a6",
    "--color-white-shaded-3": "#c8c8c8",
    "--color-white-shaded-4": "#d0d0d0",
    "--color-white-shaded-5": "#dadada",
    "--color-white-shaded-6": "#eaeaea",
    "--color-white-shaded-7": "#f4f4f4",
    "--color-white-shaded-8": "#f8f8f8",

    "--color-black-unshaded": "#000000",
    "--color-black-shaded-1": "#898989",
    "--color-black-shaded-2": "#737373",
    "--color-black-shaded-3": "#595959",
    "--color-black-shaded-4": "#373737",
    "--color-black-shaded-5": "#2f2f2f",
    "--color-black-shaded-6": "#252525",
    "--color-black-shaded-7": "#151515",
    "--color-black-shaded-8": "#0b0b0b",

    "--color-apple-unshaded": "#000000",
    "--color-apple-shaded-1": "#898989",
    "--color-apple-shaded-2": "#737373",
    "--color-apple-shaded-3": "#595959",
    "--color-apple-shaded-4": "#373737",
    "--color-apple-shaded-5": "#2f2f2f",
    "--color-apple-shaded-6": "#252525",
    "--color-apple-shaded-7": "#151515",
    "--color-apple-shaded-8": "#0b0b0b",

    "--color-google-unshaded": "#ffffff",
    "--color-google-shaded-1": "#767676",
    "--color-google-shaded-2": "#a6a6a6",
    "--color-google-shaded-3": "#c8c8c8",
    "--color-google-shaded-4": "#d0d0d0",
    "--color-google-shaded-5": "#dadada",
    "--color-google-shaded-6": "#eaeaea",
    "--color-google-shaded-7": "#f4f4f4",
    "--color-google-shaded-8": "#f8f8f8",

    "--color-facebook-unshaded": "#3b5998",
    "--color-facebook-shaded-1": "#f5f7fb",
    "--color-facebook-shaded-2": "#d7dfef",
    "--color-facebook-shaded-3": "#b7c4e0",
    "--color-facebook-shaded-4": "#7b91c2",
    "--color-facebook-shaded-5": "#4d69a5",
    "--color-facebook-shaded-6": "#36508a",
    "--color-facebook-shaded-7": "#2d4474",
    "--color-facebook-shaded-8": "#213256",

    "--color-linkedin-unshaded": "#187fb8",
    "--color-linkedin-shaded-1": "#f3f9fc",
    "--color-linkedin-shaded-2": "#d2e8f4",
    "--color-linkedin-shaded-3": "#add4ea",
    "--color-linkedin-shaded-4": "#65add4",
    "--color-linkedin-shaded-5": "#2d8dc0",
    "--color-linkedin-shaded-6": "#1573a5",
    "--color-linkedin-shaded-7": "#12618c",
    "--color-linkedin-shaded-8": "#0d4867",

    "--color-azureadv2-unshaded": "#00a2ed",
    "--color-azureadv2-shaded-1": "#f4fbfe",
    "--color-azureadv2-shaded-2": "#d4effc",
    "--color-azureadv2-shaded-3": "#afe2fa",
    "--color-azureadv2-shaded-4": "#62c6f4",
    "--color-azureadv2-shaded-5": "#1dadef",
    "--color-azureadv2-shaded-6": "#0092d5",
    "--color-azureadv2-shaded-7": "#007bb4",
    "--color-azureadv2-shaded-8": "#005b85",

    "--color-wechat-unshaded": "#07c160",
    "--color-wechat-shaded-1": "#f3fdf8",
    "--color-wechat-shaded-2": "#d0f5e2",
    "--color-wechat-shaded-3": "#a8edc9",
    "--color-wechat-shaded-4": "#5dda99",
    "--color-wechat-shaded-5": "#1fc971",
    "--color-wechat-shaded-6": "#07ae58",
    "--color-wechat-shaded-7": "#06934a",
    "--color-wechat-shaded-8": "#046d37",

    "--color-warn": "#fbca4e",
    "--color-good": "#58ca9a",

    "--color-pane-background": "#ffffff",
    "--color-pane-shadow": "rgba(0, 0, 0, 0.25)",

    "--color-separator": "#e5e5e5",

    "--color-password-strength-meter-0": "#edebe9",
    "--color-password-strength-meter-1": "var(--color-error-unshaded)",
    "--color-password-strength-meter-2": "#ffa133",
    "--color-password-strength-meter-3": "var(--color-warn)",
    "--color-password-strength-meter-4": "#baca58",
    "--color-password-strength-meter-5": "var(--color-good)",
  };

  return {
    ...primaryStyle,
    ...textStyle,
    ...backgroundStyle,
    ...lightModeStyle,
  };
}

function getDarkModeStyle(options: GetStyleOptions) {
  const lightModeStyle = getLightModeStyle(options);
  const darkModeStyle = {
    "--color-google-unshaded": "#4285f4",
    "--color-google-shaded-1": "#f7faff",
    "--color-google-shaded-2": "#e0ebfd",
    "--color-google-shaded-3": "#c5dafc",
    "--color-google-shaded-4": "#8cb6f9",
    "--color-google-shaded-5": "#5895f6",
    "--color-google-shaded-6": "#3b79dc",
    "--color-google-shaded-7": "#3266ba",
    "--color-google-shaded-8": "#254b89",
    "--color-pane-background": "#252525",
    "--color-separator": "#3e3e3e",
    "--color-password-strength-meter-0": "#8a8886",
  };

  return {
    ...lightModeStyle,
    ...darkModeStyle,
  };
}

const ThemePreviewWidget: React.FC<ThemePreviewWidgetProps> = function ThemePreviewWidget(
  props: ThemePreviewWidgetProps
) {
  const { isDarkMode, primaryColor, textColor, backgroundColor } = props;
  const rootStyle = useMemo(() => {
    return isDarkMode
      ? getDarkModeStyle({
          primaryColor,
          textColor,
          backgroundColor,
        })
      : getLightModeStyle({
          primaryColor,
          textColor,
          backgroundColor,
        });
  }, [isDarkMode, primaryColor, textColor, backgroundColor]);
  return (
    <div className={styles.root} style={rootStyle as any}>
      <div className={styles.page}>
        <div className={styles.content}></div>
      </div>
    </div>
  );
};

export default ThemePreviewWidget;
