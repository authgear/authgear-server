import React, { useContext, useMemo } from "react";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  DefaultEffects,
  Icon,
  IIconProps,
  PrimaryButton,
  Text,
} from "office-ui-fabric-react";
import { useParams } from "react-router-dom";
import ScreenHeader from "../../ScreenHeader";
import styles from "./OnboardingCompletionScreen.module.scss";
import SignupScreenImg from "../../images/onboarding_signup_screen.png";
import SettingScreenImg from "../../images/onboarding_settings_screen.png";
import SSOLogoImg from "../../images/onboarding_sso_logo.png";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ShowLoading from "../../ShowLoading";
import { PortalAPIAppConfig } from "../../types";

export interface OnboardingCompletionStepContentProps {
  image: string;
  titleId: string;
  messageId: string;
  stepCount?: number;
  imageAlignRight?: boolean;
  actionLabelId?: string;
  actionHref?: string;
}

const OnboardingCompletionStepContent: React.FC<OnboardingCompletionStepContentProps> = function OnboardingCompletionStepContent(
  props
) {
  const { renderToString } = useContext(Context);
  const {
    image,
    titleId,
    messageId,
    stepCount,
    imageAlignRight,
    actionLabelId,
    actionHref,
  } = props;
  return (
    <div
      className={cn(styles.stepSection, {
        [styles.rightScreenshot]: imageAlignRight,
      })}
    >
      <div className={styles.screenshot}>
        <img className={styles.logo} src={image} />
      </div>
      <div className={styles.info}>
        <Text className={styles.title} block={true}>
          {stepCount && <span className={styles.stepCount}>{stepCount}</span>}
          <FormattedMessage id={titleId} />
        </Text>
        <Text className={styles.desc} block={true} variant="small">
          <FormattedMessage id={messageId} />
        </Text>
        {!!actionLabelId && !!actionHref && (
          <PrimaryButton
            text={renderToString(actionLabelId)}
            href={actionHref}
            target="_blank"
            rel="noreferrer"
          />
        )}
      </div>
    </div>
  );
};

interface ActionButtonProps {
  iconProps: IIconProps;
  labelId: string;
  href: string;
}

const ActionButton: React.FC<ActionButtonProps> = function ActionButton(props) {
  const { iconProps, labelId, href } = props;
  return (
    <a
      className={styles.actionButton}
      href={href}
      target="_blank"
      rel="noreferrer"
      style={{ boxShadow: DefaultEffects.elevation4 }}
    >
      <Icon {...iconProps} className={styles.actionIcon} />
      <Text className={styles.labelText}>
        <FormattedMessage id={labelId} />
      </Text>
      <Icon className={styles.arrowIcon} iconName="ChromeBackMirrored" />
    </a>
  );
};

interface OnboardingCompletionContentProps {
  config: PortalAPIAppConfig;
}

const OnboardingCompletionContent: React.FC<OnboardingCompletionContentProps> = function OnboardingCompletionContent(
  props
) {
  const { appID } = useParams();
  const { appHostSuffix } = useSystemConfig();

  const { config } = props;

  const rawAppID = config.id;
  const endpoint = rawAppID ? "https://" + rawAppID + appHostSuffix : undefined;

  const portalAppEndpoint = `/app/${encodeURIComponent(appID)}`;
  const portalSSOEndpoint = `/app/${encodeURIComponent(
    appID
  )}/configuration/single-sign-on`;

  // when login id is disabled, the only identities are sso or anonymous
  // show configure sso step
  // the condition may change when more identities are supported
  const loginIDDisabled = useMemo(() => {
    return (config.authentication?.identities ?? []).indexOf("login_id") === -1;
  }, [config.authentication?.identities]);

  return (
    <div className={styles.pageWrapper}>
      <div className={styles.page}>
        <div
          className={styles.pageContent}
          style={{ boxShadow: DefaultEffects.elevation4 }}
        >
          <div className={styles.mainSection}>
            <Text className={styles.pageTitle} block={true} variant="xLarge">
              <FormattedMessage id="OnboardingCompletion.title" />
            </Text>
            <Text className={styles.pageDesc} block={true} variant="small">
              <FormattedMessage id="OnboardingCompletion.desc" />
            </Text>
            <Text className={styles.completionMessage} block={true}>
              <FormattedMessage id="OnboardingCompletion.completion-message" />
            </Text>
            {!loginIDDisabled && (
              <OnboardingCompletionStepContent
                image={SignupScreenImg}
                titleId="OnboardingCompletion.signup-login.title"
                messageId="OnboardingCompletion.signup-login.desc"
                stepCount={1}
                actionLabelId="OnboardingCompletion.signup-login.action"
                actionHref={endpoint}
              />
            )}
            {!loginIDDisabled && (
              <OnboardingCompletionStepContent
                image={SettingScreenImg}
                titleId="OnboardingCompletion.settings.title"
                messageId="OnboardingCompletion.settings.desc"
                stepCount={2}
                imageAlignRight={true}
              />
            )}
            {loginIDDisabled && (
              <OnboardingCompletionStepContent
                image={SSOLogoImg}
                titleId="OnboardingCompletion.sso.title"
                messageId="OnboardingCompletion.sso.desc"
                actionLabelId="OnboardingCompletion.sso.action"
                actionHref={portalSSOEndpoint}
              />
            )}
          </div>
          <div className={styles.nowYouMaySection}>
            <Text className={styles.title}>
              <FormattedMessage id="OnboardingCompletion.now-you-may.title" />
            </Text>

            <ActionButton
              labelId="OnboardingCompletion.now-you-may.portal.label"
              iconProps={{ iconName: "PlugConnected" }}
              href={portalAppEndpoint}
            />
            <ActionButton
              labelId="OnboardingCompletion.now-you-may.doc.label"
              iconProps={{ iconName: "ReadingMode" }}
              href="https://docs.authgear.com/"
            />
          </div>
        </div>
      </div>
    </div>
  );
};

const OnboardingCompletionScreen: React.FC = function OnboardingCompletionScreen() {
  const { appID } = useParams();

  const { effectiveAppConfig, loading } = useAppConfigQuery(appID);

  if (loading || !effectiveAppConfig) {
    return <ShowLoading />;
  }

  return (
    <div className={styles.root}>
      <ScreenHeader />
      <OnboardingCompletionContent config={effectiveAppConfig} />
    </div>
  );
};

export default OnboardingCompletionScreen;
