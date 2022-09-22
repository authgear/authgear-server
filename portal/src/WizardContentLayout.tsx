import React, { useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { DefaultEffects, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import Link from "./Link";
import DefaultButton from "./DefaultButton";
import styles from "./WizardContentLayout.module.css";
import {
  AuthgearGTMEventType,
  EventDataAttributes,
  useMakeAuthgearGTMEventDataAttributes,
} from "./GTMProvider";

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
  trackSkipButtonClick?: boolean;
  trackSkipButtonEventData?: EventDataAttributes;
}

export default function WizardContentLayout(
  props: WizardContentLayoutProps
): React.ReactElement {
  const navigate = useNavigate();
  const {
    children,
    primaryButton,
    backButtonDisabled,
    appID,
    trackSkipButtonClick,
    trackSkipButtonEventData,
  } = props;
  const onClickBackButton = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate(-1);
    },
    [navigate]
  );

  const makeGTMEventDataAttributes = useMakeAuthgearGTMEventDataAttributes();
  const gtmEventDataAttributes = useMemo(() => {
    return makeGTMEventDataAttributes({
      event: AuthgearGTMEventType.ClickedSkipInProjectWizard,
      eventDataAttributes: trackSkipButtonEventData,
    });
  }, [makeGTMEventDataAttributes, trackSkipButtonEventData]);

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
          {...(trackSkipButtonClick ? gtmEventDataAttributes : {})}
        >
          <FormattedMessage id="WizardContentLayout.skip.label" />
        </Link>
      ) : null}
    </div>
  );
}
