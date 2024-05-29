import React, { useContext } from "react";
import { Text } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import styles from "./OnboardingSurveyLayout.module.css";

interface LogoProps {}

const Logo: React.VFC<LogoProps> = (_props: LogoProps) => {
  const { renderToString } = useContext(Context);

  return (
    <img
      className={styles.logo}
      alt={renderToString("system.name")}
      src={renderToString("system.logo-inverted-uri")}
    />
  );
};

export interface SurveyTitleProps {
  children?: React.ReactNode;
}

export function SurveyTitle(props: SurveyTitleProps): React.ReactElement {
  return (
    <Text className={styles.title} variant="large" block={true}>
      {props.children}
    </Text>
  );
}

export interface SurveySubtitleProps {
  children?: React.ReactNode;
}

export function SurveySubtitle(props: SurveySubtitleProps): React.ReactElement {
  return (
    <Text className={styles.subtitle} variant="small" block={true}>
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
  return (
    <div className={styles.root}>
      <Logo />
      <div className={styles.center}>
        <SurveyTitle>{title}</SurveyTitle>
        <SurveySubtitle>{subtitle}</SurveySubtitle>
        {children}
        <div className={styles.navigation}>
          {backButtonDisabled ? null : secondaryButton}
          {primaryButton}
        </div>
      </div>
    </div>
  );
}
