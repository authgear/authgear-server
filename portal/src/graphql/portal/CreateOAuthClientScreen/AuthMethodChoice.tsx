import React from "react";
import { ChoiceGroup, type IChoiceGroupOption } from "@fluentui/react";
import { useIntl, FormattedMessage } from "react-intl";
import ExternalLink from "../../../ExternalLink";
import WidgetSubtitle from "../../../WidgetSubtitle";
import type { AuthMethodChoice as Stage2Choice } from "./frameworks";
import styles from "./AuthMethodChoice.module.css";

export interface AuthMethodChoiceProps {
  value: Stage2Choice | null;
  onChange: (value: Stage2Choice) => void;
  nginxDocsHref: string;
}

export const AuthMethodChoiceComponent: React.FC<AuthMethodChoiceProps> = ({
  value,
  onChange,
  nginxDocsHref,
}) => {
  const { formatMessage } = useIntl();
  const options: IChoiceGroupOption[] = [
    {
      key: "token",
      text: formatMessage({ id: "CreateOAuthClientScreen.stage2.option.token" }),
    },
    {
      key: "cookie",
      text: formatMessage({ id: "CreateOAuthClientScreen.stage2.option.cookie" }),
      // eslint-disable-next-line react/no-unstable-nested-components
      onRenderField: (props, render) => (
        <div>
          {render!(props)}
          <div
            className={styles.cookieHelp}
            onClick={(e) => e.stopPropagation()}
          >
            <ExternalLink href={nginxDocsHref}>
              <FormattedMessage id="CreateOAuthClientScreen.stage2.option.cookie.docs-link" />
            </ExternalLink>
          </div>
        </div>
      ),
    },
  ];
  return (
    <div className={styles.root}>
      <WidgetSubtitle>
        <FormattedMessage id="CreateOAuthClientScreen.stage2.question" />
      </WidgetSubtitle>
      <ChoiceGroup
        selectedKey={value ?? undefined}
        options={options}
        onChange={(_, option) =>
          option && onChange(option.key as Stage2Choice)
        }
      />
    </div>
  );
};
