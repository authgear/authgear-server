import React, { useCallback, useContext, useMemo } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import {
  CubeIcon,
  DesktopIcon,
  MobileIcon,
  ReaderIcon,
  StackIcon,
} from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import Link from "../../Link";
import { OAuthClientConfig } from "../../types";
import {
  IconRadioCards,
  IconRadioCardOption,
} from "../v2/IconRadioCards/IconRadioCards";
import { SquareIcon, SquareIconProps } from "../v2/SquareIcon/SquareIcon";
import styles from "./ApplicationTypeRadioCards.module.css";

type ApplicationType = NonNullable<OAuthClientConfig["x_application_type"]>;

interface ApplicationTypeRadioCardsProps {
  className?: string;
  value: ApplicationType | null | undefined;
  hasNoAPIResources: boolean;
  appNodeID: string;
  onValueChange: (value: ApplicationType) => void;
}

function applicationTypeIcon(Icon: SquareIconProps["Icon"]) {
  return (
    <SquareIcon
      className="text-[var(--accent-9)]"
      Icon={Icon}
      size="7"
      radius="4"
      iconSize="1.125rem"
    />
  );
}

function ApplicationTypeCardSubtitle(props: { description: React.ReactNode }) {
  const { description } = props;
  return (
    <Text as="p" size="1" className={styles.subtitleLine}>
      {description}
    </Text>
  );
}

export const ApplicationTypeRadioCards: React.VFC<ApplicationTypeRadioCardsProps> =
  function ApplicationTypeRadioCards(props) {
    const { className, value, hasNoAPIResources, appNodeID, onValueChange } =
      props;
    const { renderToString } = useContext(Context);

    const options = useMemo((): IconRadioCardOption<ApplicationType>[] => {
      return [
        {
          value: "spa",
          icon: applicationTypeIcon(ReaderIcon as SquareIconProps["Icon"]),
          title: renderToString("oauth-client.application-type.spa"),
          subtitle: (
            <ApplicationTypeCardSubtitle
              description={renderToString(
                "CreateOAuthClientScreen.application-type.description.spa"
              )}
            />
          ),
        },
        {
          value: "native",
          icon: applicationTypeIcon(MobileIcon as SquareIconProps["Icon"]),
          title: renderToString("oauth-client.application-type.native"),
          subtitle: (
            <ApplicationTypeCardSubtitle
              description={renderToString(
                "CreateOAuthClientScreen.application-type.description.native"
              )}
            />
          ),
        },
        {
          value: "traditional_webapp",
          icon: applicationTypeIcon(DesktopIcon as SquareIconProps["Icon"]),
          title: renderToString(
            "oauth-client.application-type.traditional-webapp"
          ),
          subtitle: (
            <ApplicationTypeCardSubtitle
              description={renderToString(
                "CreateOAuthClientScreen.application-type.description.traditional-webapp"
              )}
            />
          ),
        },
        {
          value: "confidential",
          icon: applicationTypeIcon(CubeIcon as SquareIconProps["Icon"]),
          title: renderToString("oauth-client.application-type.confidential"),
          subtitle: (
            <ApplicationTypeCardSubtitle
              description={renderToString(
                "CreateOAuthClientScreen.application-type.description.confidential"
              )}
            />
          ),
        },
        {
          value: "m2m",
          icon: applicationTypeIcon(StackIcon as SquareIconProps["Icon"]),
          title: renderToString("oauth-client.application-type.m2m"),
          subtitle: (
            <ApplicationTypeCardSubtitle
              description={
                hasNoAPIResources ? (
                  <FormattedMessage
                    id="CreateOAuthClientScreen.application-type.description.m2m.disabled"
                    values={{
                      // eslint-disable-next-line react/no-unstable-nested-components
                      reactRouterLink: (chunks: React.ReactNode) => (
                        <Link
                          to={`/project/${encodeURIComponent(
                            appNodeID
                          )}/api-resources?create=1`}
                        >
                          {chunks}
                        </Link>
                      ),
                    }}
                  />
                ) : (
                  renderToString(
                    "CreateOAuthClientScreen.application-type.description.m2m"
                  )
                )
              }
            />
          ),
          disabled: hasNoAPIResources,
        },
      ];
    }, [appNodeID, hasNoAPIResources, renderToString]);

    const handleValueChange = useCallback(
      (newValue: ApplicationType) => {
        onValueChange(newValue);
      },
      [onValueChange]
    );

    return (
      <div className={cn(styles.applicationTypeCards, className)}>
        <IconRadioCards
          size="2"
          value={value ?? null}
          onValueChange={handleValueChange}
          options={options}
          numberOfColumns={2}
          itemFillSpaces={true}
          itemMinWidth={280}
        />
      </div>
    );
  };
