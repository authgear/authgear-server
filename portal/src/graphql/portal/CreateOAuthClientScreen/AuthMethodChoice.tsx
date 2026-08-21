import React from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "react-intl";
import ExternalLink from "../../../ExternalLink";
import {
  RadioCards,
  type RadioCardOption,
} from "../../../components/v2/RadioCards/RadioCards";
import type { AuthMethodChoice as Stage2Choice } from "./frameworks";
import styles from "./AuthMethodChoice.module.css";

export interface AuthMethodChoiceProps {
  value: Stage2Choice | null;
  onChange: (value: Stage2Choice) => void;
  nginxDocsHref: string;
}

const options: RadioCardOption<Stage2Choice>[] = [
  {
    value: "token",
    title: (
      <FormattedMessage id="CreateOAuthClientScreen.stage2.option.token" />
    ),
    subtitle: (
      <FormattedMessage id="CreateOAuthClientScreen.stage2.option.token.description" />
    ),
  },
  {
    value: "cookie",
    title: (
      <FormattedMessage id="CreateOAuthClientScreen.stage2.option.cookie" />
    ),
    subtitle: (
      <FormattedMessage id="CreateOAuthClientScreen.stage2.option.cookie.description" />
    ),
  },
];

export const AuthMethodChoiceComponent: React.FC<AuthMethodChoiceProps> = ({
  value,
  onChange,
  nginxDocsHref,
}) => {
  return (
    <div className={styles.root}>
      <Text size="2" weight="medium" className={styles.question}>
        <FormattedMessage id="CreateOAuthClientScreen.stage2.question" />
      </Text>
      <RadioCards
        size="2"
        options={options}
        value={value}
        onValueChange={onChange}
        itemMinWidth={240}
        itemFillSpaces={true}
      />
      <div className={styles.cookieHelp}>
        <ExternalLink href={nginxDocsHref}>
          <FormattedMessage id="CreateOAuthClientScreen.stage2.option.cookie.docs-link" />
        </ExternalLink>
      </div>
    </div>
  );
};
