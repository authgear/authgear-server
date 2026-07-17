import React, { useCallback, useContext, useMemo, useState } from "react";
import { IconButton, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ExternalLink from "../../ExternalLink";
import PortalLink from "../../Link";
import LinkButton from "../../LinkButton";
import TextField from "../../TextField";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import type { StarterKit } from "./CreateOAuthClientScreen/frameworks";
import { buildEnvFileContent } from "./CreateOAuthClientScreen/starterKit";
import styles from "./EditOAuthClientFormFrameworkQuickStart.module.css";

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

interface StepProps {
  index: number;
  title: React.ReactNode;
  children?: React.ReactNode;
}

function Step({ index, title, children }: StepProps) {
  return (
    <div className={styles.step}>
      <div className={styles.stepNumber}>{index}</div>
      <div className={styles.stepContent}>
        <Text block={true} className={styles.stepTitle}>
          {title}
        </Text>
        {children}
      </div>
    </div>
  );
}

const inlineCode = (chunks: React.ReactNode) => (
  <code className={styles.inlineCode}>{chunks}</code>
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
  const { renderToString } = useContext(Context);

  const [redirectInput, setRedirectInput] = useState(starterKit.redirectURI);
  const [editing, setEditing] = useState(false);

  // Adjust `editing` when `redirectURIIsSet` changes, following React's
  // "adjusting state when a prop changes" pattern instead of an effect,
  // to avoid a redundant extra render from setState-in-effect.
  const [prevRedirectURIIsSet, setPrevRedirectURIIsSet] =
    useState(redirectURIIsSet);
  if (redirectURIIsSet !== prevRedirectURIIsSet) {
    setPrevRedirectURIIsSet(redirectURIIsSet);
    if (redirectURIIsSet) {
      setEditing(false);
    }
  }

  const envContent = useMemo(
    () =>
      buildEnvFileContent(starterKit, {
        clientID,
        endpoint: publicOrigin,
      }),
    [starterKit, clientID, publicOrigin]
  );

  const { copyButtonProps, Feedback } = useCopyFeedback({
    textToCopy: envContent,
  });

  const onChangeRedirect = useCallback((_e: unknown, newValue?: string) => {
    setRedirectInput(newValue ?? "");
  }, []);

  const onClickSet = useCallback(() => {
    onSetRedirectURI(redirectInput);
  }, [onSetRedirectURI, redirectInput]);

  const onClickEdit = useCallback(() => {
    setEditing(true);
  }, []);

  const showEditor = editing || !redirectURIIsSet;

  return (
    <>
      <Text variant="xLarge" block={true} className={styles.starterKitTitle}>
        <FormattedMessage id="StarterKit.section-title" />
      </Text>

      <Step index={1} title={<FormattedMessage id="StarterKit.step1.title" />}>
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
            iconProps={{ iconName: "OpenInNewTab" }}
            href={starterKit.repoUrl}
            target="_blank"
            rel="noreferrer"
            text={<FormattedMessage id="StarterKit.step1.view-github" />}
          />
        </div>
      </Step>

      <Step index={2} title={<FormattedMessage id="StarterKit.step2.title" />}>
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage id="StarterKit.step2.body" />
        </Text>
        <TextFieldWithCopyButton value={starterKit.redirectURI} readOnly={true} />
        {showEditor ? (
          <div className={styles.redirectSetRow}>
            <TextField
              label={renderToString("StarterKit.step2.redirect-label")}
              placeholder={renderToString("StarterKit.step2.placeholder")}
              value={redirectInput}
              onChange={onChangeRedirect}
              styles={{ root: { flex: 1 } }}
            />
            <PrimaryButton
              onClick={onClickSet}
              disabled={saving || redirectInput === ""}
              text={<FormattedMessage id="StarterKit.step2.set" />}
            />
          </div>
        ) : (
          <div className={styles.redirectSetRow}>
            <TextFieldWithCopyButton
              value={starterKit.redirectURI}
              readOnly={true}
              disabled={true}
            />
            <LinkButton
              className={styles.redirectEditLink}
              onClick={onClickEdit}
            >
              <FormattedMessage id="StarterKit.step2.edit" />
            </LinkButton>
          </div>
        )}
      </Step>

      <Step index={3} title={<FormattedMessage id="StarterKit.step3.title" />}>
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage
            id="StarterKit.step3.body"
            values={{ code: inlineCode }}
          />
        </Text>
        <div className={styles.envCard}>
          <div className={styles.envCopyWrap}>
            <IconButton {...copyButtonProps} />
            <Feedback />
          </div>
          <pre className={styles.snippetCode}>
            <code>{envContent}</code>
          </pre>
        </div>
      </Step>

      <Step index={4} title={<FormattedMessage id="StarterKit.step4.title" />}>
        <Text block={true} className={styles.stepBody}>
          <FormattedMessage
            id="StarterKit.step4.body"
            values={{ installCmd: starterKit.installCmd, code: inlineCode }}
          />
        </Text>
      </Step>

      <Step index={5} title={<FormattedMessage id="StarterKit.step5.title" />}>
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
      </Step>

      <Step index={6} title={<FormattedMessage id="StarterKit.step6.title" />}>
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
      </Step>

      <Step index={7} title={<FormattedMessage id="StarterKit.step7.title" />}>
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
      </Step>

      <Step index={8} title={<FormattedMessage id="StarterKit.step8.title" />}>
        <div className={styles.stepButtonRow}>
          <DefaultButton
            onClick={onGoToSettings}
            text={<FormattedMessage id="StarterKit.step8.settings" />}
          />
        </div>
      </Step>
    </>
  );
}
