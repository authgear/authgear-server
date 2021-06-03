import React, {
  useMemo,
  useContext,
  forwardRef,
  ForwardedRef,
  Ref,
} from "react";
import cn from "classnames";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { getShades, BannerConfiguration } from "./util/theme";
import { base64EncodedDataToDataURI } from "./util/uri";
import styles from "./ThemePreviewWidget.module.scss";
import appLogo from "./images/app_logo.png";
import appLogoDark from "./images/app_logo_dark.png";

export interface ThemePreviewWidgetProps {
  ref?: Ref<HTMLElement>;
  appLogoValue: string | undefined;
  bannerConfiguration: BannerConfiguration;
  className?: string;
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

    "--color-adfs-unshaded": "#00a2ed",
    "--color-adfs-shaded-1": "#f4fbfe",
    "--color-adfs-shaded-2": "#d4effc",
    "--color-adfs-shaded-3": "#afe2fa",
    "--color-adfs-shaded-4": "#62c6f4",
    "--color-adfs-shaded-5": "#1dadef",
    "--color-adfs-shaded-6": "#0092d5",
    "--color-adfs-shaded-7": "#007bb4",
    "--color-adfs-shaded-8": "#005b85",

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

interface BannerProps {
  isDarkMode: boolean;
  appLogoValue: string | undefined;
  bannerConfiguration: BannerConfiguration;
}

function Banner(props: BannerProps) {
  const { appLogoValue, isDarkMode, bannerConfiguration } = props;
  const src = useMemo(() => {
    if (appLogoValue != null) {
      return base64EncodedDataToDataURI(appLogoValue);
    }
    return undefined;
  }, [appLogoValue]);
  return (
    <div className={cn(styles.marginV20)}>
      <div
        className={styles.bannerFrame}
        style={{
          backgroundColor: bannerConfiguration.backgroundColor,
          paddingTop: bannerConfiguration.paddingTop,
          paddingRight: bannerConfiguration.paddingRight,
          paddingBottom: bannerConfiguration.paddingBottom,
          paddingLeft: bannerConfiguration.paddingLeft,
        }}
      >
        <img
          className={styles.banner}
          src={src ?? (isDarkMode ? appLogoDark : appLogo)}
          style={{
            width: bannerConfiguration.width,
            height: bannerConfiguration.height,
          }}
        />
      </div>
    </div>
  );
}

function PageSwitch() {
  return (
    <div
      className={cn(
        styles.signinSignupSwitch,
        styles.flex,
        styles.flexDirectionRow
      )}
    >
      <a
        className={cn(
          styles.notA,
          styles.signinSignupLink,
          styles.primaryTxt,
          styles.current
        )}
      >
        <FormattedMessage id="ThemePreviewWidget.page-switch.login" />
      </a>
      <a
        className={cn(styles.notA, styles.signinSignupLink, styles.primaryTxt)}
      >
        <FormattedMessage id="ThemePreviewWidget.page-switch.signup" />
      </a>
    </div>
  );
}

function LoginToContinueLabel() {
  return (
    <p
      className={cn(
        styles.primaryTxt,
        styles.frontInherit,
        styles.fontSemibold,
        styles.textCenter,
        styles.marginT40,
        styles.marginB20,
        styles.widthFull
      )}
    >
      <FormattedMessage id="ThemePreviewWidget.login-to-continue" />
    </p>
  );
}

function LoginIDForm() {
  const { renderToString } = useContext(Context);
  return (
    <div
      className={cn(
        styles.flex,
        styles.flexDirectionColumn,
        styles.marginB20,
        styles.widthFull
      )}
    >
      <input
        className={cn(
          styles.marginB20,
          styles.input,
          styles.textInput,
          styles.primaryTxt
        )}
        placeholder={renderToString(
          "ThemePreviewWidget.input-placeholder.email"
        )}
      />

      <a
        className={cn(
          styles.link,
          styles.fontSmaller,
          styles.alignSelfFlexStart,
          styles.block,
          styles.marginB20
        )}
      >
        <FormattedMessage id="ThemePreviewWidget.use-phone-instead" />
      </a>
      <button type="button" className={cn(styles.btn, styles.primaryBtn)}>
        <FormattedMessage id="ThemePreviewWidget.next-button-label" />
      </button>
    </div>
  );
}

function Separator() {
  return (
    <div
      className={cn(
        styles.ssoLoginIDSeparator,
        styles.flex,
        styles.flexDirectionRow,
        styles.alignItemsCenter,
        styles.marginB20,
        styles.widthFull
      )}
    >
      <span className={cn(styles.primaryTxt)}>
        <FormattedMessage id="ThemePreviewWidget.separator-label" />
      </span>
    </div>
  );
}

interface SimpleOAuthButtonProps {
  providerType: string;
  iconClassName: string;
  title: React.ReactNode;
}

function SimpleOAuthButton(props: SimpleOAuthButtonProps) {
  const { providerType, iconClassName, title } = props;
  return (
    <button
      type="button"
      className={cn(
        styles.btn,
        styles.ssoBtn,
        styles.marginB20,
        styles[providerType]
      )}
    >
      <span className={styles.ssoBtnContent}>
        <div className={styles.ssoBtnIcon}>
          <i className={cn("fab", iconClassName)} aria-hidden="true" />
        </div>
        <span className={styles.title}>{title}</span>
      </span>
    </button>
  );
}

function GoogleButton() {
  return (
    <button
      type="button"
      className={cn(styles.btn, styles.ssoBtn, styles.marginB20, styles.google)}
    >
      <span className={styles.ssoBtnContent}>
        <div className={cn(styles.ssoBtnIcon, styles.googleIcon)} />
        <span className={styles.title}>
          <FormattedMessage id="ThemePreviewWidget.google-button-label" />
        </span>
      </span>
    </button>
  );
}

function OAuthForm() {
  return (
    <div
      className={cn(styles.flex, styles.flexDirectionColumn, styles.widthFull)}
    >
      <SimpleOAuthButton
        providerType="apple"
        iconClassName="fa-apple"
        title={<FormattedMessage id="ThemePreviewWidget.apple-button-label" />}
      />
      <GoogleButton />
      <SimpleOAuthButton
        providerType="facebook"
        iconClassName="fa-facebook-f"
        title={
          <FormattedMessage id="ThemePreviewWidget.facebook-button-label" />
        }
      />
      <SimpleOAuthButton
        providerType="linkedin"
        iconClassName="fa-linkedin-in"
        title={
          <FormattedMessage id="ThemePreviewWidget.linkedin-button-label" />
        }
      />
      <SimpleOAuthButton
        providerType="azureadv2"
        iconClassName="fa-microsoft"
        title={
          <FormattedMessage id="ThemePreviewWidget.azureadv2-button-label" />
        }
      />
      <SimpleOAuthButton
        providerType="adfs"
        iconClassName="fa-microsoft"
        title={<FormattedMessage id="ThemePreviewWidget.adfs-button-label" />}
      />
      <SimpleOAuthButton
        providerType="wechat"
        iconClassName="fa-weixin"
        title={<FormattedMessage id="ThemePreviewWidget.wechat-button-label" />}
      />
    </div>
  );
}

function Disclaimer() {
  return (
    <p
      className={cn(
        styles.fontSmaller,
        styles.primaryTxt,
        styles.marginB20,
        styles.widthFull
      )}
    >
      <FormattedMessage
        id="ThemePreviewWidget.disclaimer"
        values={{
          className: styles.link,
        }}
      />
    </p>
  );
}

function Footer() {
  return (
    <div
      className={cn(styles.footerWatermark, styles.marginV20, styles.widthFull)}
    />
  );
}

type Props = Omit<ThemePreviewWidgetProps, "ref">;

const ThemePreviewWidget: React.FC<Props> = forwardRef(
  function ThemePreviewWidget(props: Props, ref: ForwardedRef<HTMLElement>) {
    const {
      className,
      isDarkMode,
      primaryColor,
      textColor,
      backgroundColor,
      appLogoValue,
      bannerConfiguration,
    } = props;
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
      <div
        /* @ts-expect-error */
        ref={ref}
        className={cn(className, styles.root, isDarkMode && styles.dark)}
        /* @ts-expect-error */
        style={rootStyle}
      >
        <div className={styles.page}>
          <div className={styles.content}>
            <Banner
              appLogoValue={appLogoValue}
              isDarkMode={isDarkMode}
              bannerConfiguration={bannerConfiguration}
            />
            <div
              className={cn(
                styles.pane,
                styles.flex,
                styles.flexDirectionColumn
              )}
            >
              <div className={cn(styles.flex, styles.flexDirectionColumn)}>
                <PageSwitch />
                <div className={cn(styles.paddingH20)}>
                  <LoginToContinueLabel />
                  <LoginIDForm />
                  <Separator />
                  <OAuthForm />
                  <Disclaimer />
                </div>
                <Footer />
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }
);

export default ThemePreviewWidget;
