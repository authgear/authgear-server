import React, { useContext, useMemo } from "react";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  DefaultEffects,
  Icon,
  IIconProps,
  PrimaryButton,
  Text,
} from "@fluentui/react";
import { useParams } from "react-router-dom";
import ScreenHeader from "../../ScreenHeader";
import styles from "./ProjectWizardDoneScreen.module.scss";
import SignupScreenImg from "../../images/onboarding_signup_screen.png";
import SettingScreenImg from "../../images/onboarding_settings_screen.png";
import SSOLogoImg from "../../images/onboarding_sso_logo.png";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import ShowLoading from "../../ShowLoading";
import ReactRouterLink from "../../ReactRouterLink";
import { PortalAPIAppConfig } from "../../types";

export interface ProjectWizardDoneStepContentProps {
  image: string;
  titleId: string;
  messageId: string;
  stepCount?: number;
  imageAlignRight?: boolean;
  actionLabelId?: string;
  actionHref?: string;
}

const ProjectWizardDoneStepContent: React.FC<ProjectWizardDoneStepContentProps> =
  function ProjectWizardDoneStepContent(props) {
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

function makeActionButton(
  iconProps: IIconProps
): React.FC<React.AnchorHTMLAttributes<HTMLAnchorElement>> {
  return function ActionButton(props) {
    const { children, ...rest } = props;
    return (
      <a
        className={styles.actionButton}
        style={{ boxShadow: DefaultEffects.elevation4 }}
        {...rest}
      >
        <Icon {...iconProps} className={styles.actionIcon} />
        <Text className={styles.labelText}>{children}</Text>
        <Icon className={styles.arrowIcon} iconName="ChromeBackMirrored" />
      </a>
    );
  };
}
const ActionButtonPortal = makeActionButton({ iconName: "PlugConnected" });
const ActionButtonDocs = makeActionButton({ iconName: "ReadingMode" });

interface ProjectWizardDoneContentProps {
  config: PortalAPIAppConfig;
}

const ProjectWizardDoneContent: React.FC<ProjectWizardDoneContentProps> =
  function ProjectWizardDoneContent(props) {
    const { appID } = useParams() as { appID: string };

    const { config } = props;

    const endpoint = `${config.http?.public_origin}?x_tutorial=true`;

    const portalAppEndpoint = `/project/${encodeURIComponent(appID)}`;
    const portalSSOEndpoint = `/project/${encodeURIComponent(
      appID
    )}/configuration/single-sign-on`;

    // when login id is disabled, the only identities are sso or anonymous
    // show configure sso step
    // the condition may change when more identities are supported
    const loginIDDisabled = useMemo(() => {
      return (
        (config.authentication?.identities ?? []).indexOf("login_id") === -1
      );
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
                <FormattedMessage id="ProjectWizardDoneScreen.title" />
              </Text>
              <Text className={styles.pageDesc} block={true} variant="small">
                <FormattedMessage id="ProjectWizardDoneScreen.desc" />
              </Text>
              <Text className={styles.completionMessage} block={true}>
                <FormattedMessage id="ProjectWizardDoneScreen.completion-message" />
              </Text>
              {!loginIDDisabled && (
                <ProjectWizardDoneStepContent
                  image={SignupScreenImg}
                  titleId="ProjectWizardDoneScreen.signup-login.title"
                  messageId="ProjectWizardDoneScreen.signup-login.desc"
                  stepCount={1}
                  actionLabelId="ProjectWizardDoneScreen.signup-login.action"
                  actionHref={endpoint}
                />
              )}
              {!loginIDDisabled && (
                <ProjectWizardDoneStepContent
                  image={SettingScreenImg}
                  titleId="ProjectWizardDoneScreen.settings.title"
                  messageId="ProjectWizardDoneScreen.settings.desc"
                  stepCount={2}
                  imageAlignRight={true}
                />
              )}
              {loginIDDisabled && (
                <ProjectWizardDoneStepContent
                  image={SSOLogoImg}
                  titleId="ProjectWizardDoneScreen.sso.title"
                  messageId="ProjectWizardDoneScreen.sso.desc"
                  actionLabelId="ProjectWizardDoneScreen.sso.action"
                  actionHref={portalSSOEndpoint}
                />
              )}
            </div>
            <div className={styles.nowYouMaySection}>
              <Text className={styles.title}>
                <FormattedMessage id="ProjectWizardDoneScreen.now-you-may.title" />
              </Text>
              <ReactRouterLink
                to={portalAppEndpoint}
                component={ActionButtonPortal}
              >
                <FormattedMessage id="ProjectWizardDoneScreen.now-you-may.portal.label" />
              </ReactRouterLink>
              <ActionButtonDocs
                href="https://docs.authgear.com/"
                target="_blank"
                rel="noreferrer"
              >
                <FormattedMessage id="ProjectWizardDoneScreen.now-you-may.doc.label" />
              </ActionButtonDocs>
            </div>
          </div>
        </div>
      </div>
    );
  };

const ProjectWizardDoneScreen: React.FC = function ProjectWizardDoneScreen() {
  const { appID } = useParams() as { appID: string };

  const { effectiveAppConfig, loading } = useAppAndSecretConfigQuery(appID);

  if (loading || !effectiveAppConfig) {
    return <ShowLoading />;
  }

  return (
    <div className={styles.root}>
      <ScreenHeader />
      <ProjectWizardDoneContent config={effectiveAppConfig} />
    </div>
  );
};

export default ProjectWizardDoneScreen;
