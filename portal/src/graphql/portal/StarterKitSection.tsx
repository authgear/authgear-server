import React, { useMemo } from "react";
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
  onAuthorize: () => void;
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
    onAuthorize,
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

  // Mobile/hybrid kits insert "Run on iOS" and "Run on Android" steps after
  // the web-test step, shifting the trailing step numbers.
  const { mobileRun } = starterKit;
  const mobileStepCount = mobileRun != null ? 2 : 0;

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
          <FormattedMessage
            id="StarterKit.step2.body"
            values={{ count: starterKit.redirectURIs.length }}
          />
        </Text>
        <div className={styles.redirectURIList}>
          {starterKit.redirectURIs.map((uri) => (
            <TextFieldWithCopyButton key={uri} value={uri} readOnly={true} />
          ))}
        </div>
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
                onClick={onAuthorize}
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
            values={{ code: inlineCode, fileName: starterKit.config.fileName }}
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
          {starterKit.homepageUrl != null ? (
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
          ) : (
            <FormattedMessage
              id="StarterKit.step5.body.device"
              values={{ startCmd: starterKit.startCmd, code: inlineCode }}
            />
          )}
        </Text>
      </QuickStartStep>

      {mobileRun != null ? (
        <>
          <QuickStartStep
            className="mt-4"
            stepNumber="6"
            title={<FormattedMessage id="StarterKit.mobile.ios.title" />}
          >
            <Text block={true} className={styles.stepBody}>
              <FormattedMessage
                id="StarterKit.mobile.ios.body"
                values={{
                  code: inlineCode,
                  buildCmd: mobileRun.buildCmd,
                  syncCmd: mobileRun.syncCmd,
                  openCmd: mobileRun.iosCmd,
                }}
              />
            </Text>
          </QuickStartStep>

          <QuickStartStep
            className="mt-4"
            stepNumber="7"
            title={<FormattedMessage id="StarterKit.mobile.android.title" />}
          >
            <Text block={true} className={styles.stepBody}>
              <FormattedMessage
                id="StarterKit.mobile.android.body"
                values={{
                  code: inlineCode,
                  buildCmd: mobileRun.buildCmd,
                  syncCmd: mobileRun.syncCmd,
                  openCmd: mobileRun.androidCmd,
                }}
              />
            </Text>
          </QuickStartStep>
        </>
      ) : null}

      <QuickStartStep
        className="mt-4"
        stepNumber={String(6 + mobileStepCount)}
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
        stepNumber={String(7 + mobileStepCount)}
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
        stepNumber={String(8 + mobileStepCount)}
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
