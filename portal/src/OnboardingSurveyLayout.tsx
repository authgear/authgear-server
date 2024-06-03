import React, { useContext, useMemo } from "react";
import { useTheme, Text } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import styles from "./OnboardingSurveyLayout.module.css";
import authgearLogoURL from "./images/authgear_logo_color.svg";

interface LogoProps {}

const Logo: React.VFC<LogoProps> = (_props: LogoProps) => {
  const { renderToString } = useContext(Context);
  const theme = useTheme();
  const logoStyles = useMemo(() => {
    return {
      fill: theme.semanticColors.bodyText,
    };
  }, [theme]);
  return (
    <img
      style={logoStyles}
      className={styles.logo}
      alt={renderToString("system.name")}
      src={authgearLogoURL}
    />
  );
};

export interface SurveyTitleProps {
  children?: React.ReactNode;
}

export function SurveyTitle(props: SurveyTitleProps): React.ReactElement {
  const theme = useTheme();
  const styles = useMemo(() => {
    return {
      root: {
        "font-size": "x-large",
        "font-weight": 600,
        "white-space": "pre-line",
        "text-align": "center",
        color: theme.semanticColors.bodyText,
        "margin-bottom": "20px",
      },
    };
  }, [theme]);
  return (
    <Text styles={styles} variant="large" block={true}>
      {props.children}
    </Text>
  );
}

export interface SurveySubtitleProps {
  children?: React.ReactNode;
}

export function SurveySubtitle(props: SurveySubtitleProps): React.ReactElement {
  const theme = useTheme();
  const styles = useMemo(() => {
    return {
      root: {
        "font-size": "medium",
        "font-weight": 400,
        "white-space": "pre-line",
        "text-align": "center",
        color: theme.semanticColors.bodySubtext,
      },
    };
  }, [theme]);
  return (
    <Text styles={styles} variant="small" block={true}>
      {props.children}
    </Text>
  );
}

export interface SurveyLayoutProps {
  title: string;
  subtitle: string;
  backButtonDisabled: boolean;
  primaryButton: React.ReactNode;
  secondaryButton: React.ReactNode;
  children?: React.ReactNode;
}

export default function SurveyLayout(
  props: SurveyLayoutProps
): React.ReactElement {
  const {
    title,
    subtitle,
    backButtonDisabled,
    primaryButton,
    secondaryButton,
    children,
  } = props;
  const theme = useTheme();
  const bodyStyles = useMemo(() => {
    return {
      backgroundColor: theme.semanticColors.bodyStandoutBackground,
      height: "100vh",
    };
  }, [theme]);
  return (
    <div style={bodyStyles}>
      <Logo />
      <div className={styles.center}>
        <SurveyTitle>{title}</SurveyTitle>
        <SurveySubtitle>{subtitle}</SurveySubtitle>
        <div className={styles.content}>{children}</div>
        <div className={styles.navigation}>
          {backButtonDisabled ? null : secondaryButton}
          {primaryButton}
        </div>
      </div>
    </div>
  );
}
