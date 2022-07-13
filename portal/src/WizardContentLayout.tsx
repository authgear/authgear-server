import React, { useCallback } from "react";
import { useNavigate } from "react-router-dom";
import {
  DefaultEffects,
  Text,
  DefaultButton,
  Link as FluentLink,
} from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ReactRouterLink from "./ReactRouterLink";
import styles from "./WizardContentLayout.module.css";

export interface WizardContentLayoutProps {
  title?: React.ReactNode;
  primaryButton?: React.ReactNode;
  backButtonDisabled?: boolean;
  children?: React.ReactNode;
  appID?: string;
}

export default function WizardContentLayout(
  props: WizardContentLayoutProps
): React.ReactElement {
  const navigate = useNavigate();
  const { title, children, primaryButton, backButtonDisabled, appID } = props;
  const onClickBackButton = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      navigate(-1);
    },
    [navigate]
  );
  return (
    <div className={styles.root}>
      <div
        className={styles.content}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        <Text className={styles.title} variant="large" block={true}>
          {title}
        </Text>
        {children}
        <div className={styles.buttons}>
          {primaryButton}
          {backButtonDisabled !== true && (
            <DefaultButton
              className={styles.backButton}
              onClick={onClickBackButton}
            >
              <FormattedMessage id="back" />
            </DefaultButton>
          )}
        </div>
      </div>
      {appID != null && (
        <ReactRouterLink
          className={styles.skip}
          to={`/project/${appID}`}
          component={FluentLink}
        >
          <FormattedMessage id="WizardContentLayout.skip.label" />
        </ReactRouterLink>
      )}
    </div>
  );
}
