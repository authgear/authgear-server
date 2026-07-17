import React, { useCallback, useMemo } from "react";
import cn from "classnames";
import { IconButton, Text } from "@fluentui/react";
import { FormattedMessage } from "../../intl";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ExternalLink from "../../ExternalLink";
import PortalLink from "../../Link";
import LinkButton from "../../LinkButton";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import type { StarterKit } from "./CreateOAuthClientScreen/frameworks";
import { buildConfigContent } from "./CreateOAuthClientScreen/starterKit";
import styles from "./EditOAuthClientFormFrameworkQuickStart.module.css";
import { QuickStartStep } from "./QuickStartStep";

export interface StarterKitSectionProps {
  starterKit: StarterKit;
  frameworkDisplayName: string;
  clientID: string;
  publicOrigin: string;
  usersPath: string;
  redirectURIIsSet: boolean;
  saving: boolean;
  onSetRedirectURI: (value: string) => void;
  onGoToSettings: () => void;
}

const inlineCode = (chunks: React.ReactNode) => (
  <code className={styles.inlineCode}>{chunks}</code>
);

const renderGitHubIcon = () => (
  <i
    className={`ti ti-brand-github ${styles.buttonBrandIcon}`}
    aria-hidden={true}
  />
);

export function StarterKitSection(
  props: StarterKitSectionProps
): React.ReactElement {
  const {
    starterKit,
    frameworkDisplayName,
    clientID,
    publicOrigin,
    usersPath,
    redirectURIIsSet,
    saving,
    onSetRedirectURI,
    onGoToSettings,
  } = props;
  const configContent = useMemo(
    () =>
      buildConfigContent(starterKit, {
        clientID,
        endpoint: publicOrigin,
      }),
    [starterKit, clientID, publicOrigin]
  );

  const { copyButtonProps, Feedback } = useCopyFeedback({
    textToCopy: configContent,
  });

  const onClickAuthorize = useCallback(() => {
    onSetRedirectURI(starterKit.redirectURI);
  }, [onSetRedirectURI, starterKit.redirectURI]);

  return (
    <>
      <Text variant="xLarge" block={true} className={styles.starterKitTitle}>
        <FormattedMessage id="StarterKit.section-title" />
      </Text>

      <QuickStartStep
        className="mt-4"
        stepNumber="1"
        title={<FormattedMessage id="StarterKit.step1.title" />}
      >
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage
            id="StarterKit.step1.body"
            values={{ displayName: frameworkDisplayName }}
          />
        </Text>
        <div className={styles.stepButtonRow}>
          <PrimaryButton
            iconProps={{ iconName: "Download" }}
            href={starterKit.downloadUrl}
            target="_blank"
            rel="noreferrer"
            text={<FormattedMessage id="StarterKit.step1.download" />}
          />
          <DefaultButton
            onRenderIcon={renderGitHubIcon}
            href={starterKit.repoUrl}
            target="_blank"
            rel="noreferrer"
            text={<FormattedMessage id="StarterKit.step1.view-github" />}
          />
        </div>
      </QuickStartStep>

      <QuickStartStep
        className="mt-4"
        stepNumber="2"
        title={<FormattedMessage id="StarterKit.step2.title" />}
      >
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage id="StarterKit.step2.body" />
        </Text>
        <TextFieldWithCopyButton
          value={starterKit.redirectURI}
          readOnly={true}
        />
        <div className={styles.redirectStatusRow}>
          {redirectURIIsSet ? (
            <>
              <span
                className={cn(styles.statusChip, styles.statusChipAuthorized)}
              >
                <i className="ti ti-circle-check" aria-hidden={true} />
                <FormattedMessage id="StarterKit.step2.status.authorized" />
              </span>
              <LinkButton
                className={styles.redirectManageLink}
                onClick={onGoToSettings}
              >
                <FormattedMessage id="StarterKit.step2.manage" />
              </LinkButton>
            </>
          ) : (
            <>
              <span
                className={cn(styles.statusChip, styles.statusChipUnauthorized)}
              >
                <i className="ti ti-alert-triangle" aria-hidden={true} />
                <FormattedMessage id="StarterKit.step2.status.unauthorized" />
              </span>
              <PrimaryButton
                onClick={onClickAuthorize}
                disabled={saving}
                text={<FormattedMessage id="StarterKit.step2.authorize" />}
              />
            </>
          )}
        </div>
      </QuickStartStep>

      <QuickStartStep
        className="mt-4"
        stepNumber="3"
        title={
          <FormattedMessage
            id={`StarterKit.step3.title.${starterKit.config.format}`}
          />
        }
      >
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage
            id={`StarterKit.step3.body.${starterKit.config.format}`}
            values={{ code: inlineCode }}
          />
        </Text>
        <div className={styles.envCard}>
          <div className={styles.envCopyWrap}>
            <IconButton {...copyButtonProps} />
            <Feedback />
          </div>
          <pre className={styles.snippetCode}>
            <code>{configContent}</code>
          </pre>
        </div>
      </QuickStartStep>

      <QuickStartStep
        className="mt-4"
        stepNumber="4"
        title={<FormattedMessage id="StarterKit.step4.title" />}
      >
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage
            id="StarterKit.step4.body"
            values={{ installCmd: starterKit.installCmd, code: inlineCode }}
          />
        </Text>
      </QuickStartStep>

      <QuickStartStep
        className="mt-4"
        stepNumber="5"
        title={<FormattedMessage id="StarterKit.step5.title" />}
      >
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage
            id="StarterKit.step5.body"
            values={{
              startCmd: starterKit.startCmd,
              homepageUrl: starterKit.homepageUrl,
              code: inlineCode,
              // eslint-disable-next-line react/no-unstable-nested-components
              link: (chunks: React.ReactNode) => (
                <ExternalLink href={starterKit.homepageUrl}>
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        </Text>
      </QuickStartStep>

      <QuickStartStep
        className="mt-4"
        stepNumber="6"
        title={<FormattedMessage id="StarterKit.step6.title" />}
      >
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage
            id="StarterKit.step6.body"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              usersLink: (chunks: React.ReactNode) => (
                <PortalLink to={usersPath}>{chunks}</PortalLink>
              ),
            }}
          />
        </Text>
      </QuickStartStep>

      <QuickStartStep
        className="mt-4"
        stepNumber="7"
        title={<FormattedMessage id="StarterKit.step7.title" />}
      >
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage id="StarterKit.step7.body" />
        </Text>
        <div className={styles.stepButtonRow}>
          <DefaultButton
            href={starterKit.guideUrl}
            target="_blank"
            rel="noreferrer"
            text={
              <FormattedMessage
                id="StarterKit.step7.guide"
                values={{ displayName: frameworkDisplayName }}
              />
            }
          />
        </div>
      </QuickStartStep>

      <QuickStartStep
        className="mt-4 mb-16"
        stepNumber="8"
        title={<FormattedMessage id="StarterKit.step8.title" />}
      >
        <div className={styles.stepButtonRow}>
          <DefaultButton
            onClick={onGoToSettings}
            text={<FormattedMessage id="StarterKit.step8.settings" />}
          />
        </div>
      </QuickStartStep>
    </>
  );
}
