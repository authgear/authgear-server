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
  const overrideStyles = useMemo(() => {
    return {
      root: {
        color: theme.semanticColors.bodyText,
      },
    };
  }, [theme]);
  return (
    <Text
      styles={overrideStyles}
      className={styles.SurveyTitle}
      variant="xxLarge"
      block={true}
    >
      {props.children}
    </Text>
  );
}

export interface SurveySubtitleProps {
  children?: React.ReactNode;
}

export function SurveySubtitle(props: SurveySubtitleProps): React.ReactElement {
  const theme = useTheme();
  const overrideStyles = useMemo(() => {
    return {
      root: {
        color: theme.semanticColors.bodySubtext,
      },
    };
  }, [theme]);
  return (
    <Text
      styles={overrideStyles}
      className={styles.SurveySubtitle}
      variant="large"
      block={true}
    >
      {props.children}
    </Text>
  );
}

export interface SurveyLayoutProps {
  title: string;
  subtitle: string;
  nextButton: React.ReactNode;
  backButton?: React.ReactNode;
  children?: React.ReactNode;
}

export default function SurveyLayout(
  props: SurveyLayoutProps
): React.ReactElement {
  const { title, subtitle, nextButton, backButton, children } = props;
  const theme = useTheme();
  const overrideBodyStyles = useMemo(() => {
    return {
      backgroundColor: theme.semanticColors.bodyStandoutBackground,
    };
  }, [theme]);
  return (
    <div style={overrideBodyStyles} className={styles.body}>
      <Logo />
      <div className={styles.centerDiv}>
        <div className={styles.titlesDiv}>
          <SurveyTitle>{title}</SurveyTitle>
          <SurveySubtitle>{subtitle}</SurveySubtitle>
        </div>
        <div className={styles.contentDiv}>{children}</div>
        <div className={styles.navigationDiv}>
          {backButton}
          {nextButton}
        </div>
      </div>
    </div>
  );
}
