import React, { useCallback, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import {
  DefaultEffects,
  Text,
  DefaultButton,
  Link as FluentLink,
  useTheme,
} from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import ReactRouterLink from "./ReactRouterLink";
import styles from "./WizardContentLayout.module.css";
import {
  AuthgearGTMEventType,
  EventDataAttributes,
  useMakeAuthgearGTMEventDataAttributes,
} from "./GTMProvider";

export function WizardDivider(): React.ReactElement {
  const theme = useTheme();
  return (
    <hr
      style={{
        border: "0",
        height: "0",
        borderTopWidth: "1px",
        borderTopStyle: "solid",
        borderTopColor: theme.palette.neutralTertiaryAlt,
        backgroundColor: theme.palette.neutralTertiaryAlt,
      }}
    />
  );
}

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
import cn from "classnames";

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
    <div className={cn(styles.root, "mobile:grid-cols-6")}>
      <div
        className={styles.content}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        {children}
        <div className={styles.buttons}>
          {primaryButton}
          {backButtonDisabled !== true && (
            <DefaultButton onClick={onClickBackButton}>
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
          {...(trackSkipButtonClick ? gtmEventDataAttributes : {})}
        >
          <FormattedMessage id="WizardContentLayout.skip.label" />
        </ReactRouterLink>
      )}
    </div>
  );
}
