import React, { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { DefaultEffects, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import Link from "./Link";
import DefaultButton from "./DefaultButton";
import styles from "./WizardContentLayout.module.css";
import { useCapture } from "./gtm_v2";

export interface WizardTitleProps {
  children?: React.ReactNode;
}

export function WizardTitle(props: WizardTitleProps): React.ReactElement {
  return (
    <Text className={styles.title} variant="large" block={true}>
      {props.children}
    </Text>
  );
}

export interface WizardDescriptionProps {
  children?: React.ReactNode;
}

export function WizardDescription(
  props: WizardDescriptionProps
): React.ReactElement {
  return <Text block={true}>{props.children}</Text>;
}

export interface WizardContentLayoutProps {
  primaryButton?: React.ReactNode;
  backButtonDisabled?: boolean;
  children?: React.ReactNode;
  appID?: string;
}

export default function WizardContentLayout(
  props: WizardContentLayoutProps
): React.ReactElement {
  const navigate = useNavigate();
  const { children, primaryButton, backButtonDisabled, appID } = props;
  const capture = useCapture();
  const onClickBackButton = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      capture("projectWizard.clicked-back");
      navigate(-1);
    },
    [navigate, capture]
  );

  const onClickSkip = useCallback(() => {
    capture("projectWizard.clicked-skip");
  }, [capture]);

  return (
    <div className={styles.root}>
      <div
        className={styles.content}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        {children}
        <div className={styles.buttons}>
          {primaryButton}
          {backButtonDisabled !== true ? (
            <DefaultButton
              onClick={onClickBackButton}
              text={<FormattedMessage id="back" />}
            />
          ) : null}
        </div>
      </div>
      {appID != null ? (
        <Link
          className={styles.skip}
          to={`/project/${appID}`}
          onClick={onClickSkip}
        >
          <FormattedMessage id="WizardContentLayout.skip.label" />
        </Link>
      ) : null}
    </div>
  );
}
