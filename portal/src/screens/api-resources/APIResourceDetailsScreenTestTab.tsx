import React, { useContext, useMemo, useState, useCallback } from "react";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import {
  Context as MessageContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
import WidgetTitle from "../../WidgetTitle";
import {
  Text,
  Dropdown,
  IDropdownOption,
  Pivot,
  PivotItem,
} from "@fluentui/react";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useParams } from "react-router-dom";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useLoadableView } from "../../hook/useLoadableView";
import { PortalAPIAppConfig } from "../../types";
import HorizontalDivider from "../../HorizontalDivider";
import { CodeField } from "../../components/common/CodeField";

enum ExampleCodeTabKey {
  CURL = "CURL",
  Python = "Python",
  Go = "Go",
  NodeJS = "NodeJS",
}

export function APIResourceDetailsScreenTestTab({
  resource,
}: {
  resource: Resource;
}): React.ReactElement | null {
  const { appID } = useParams() as { appID: string };
  const appConfigQuery = useAppAndSecretConfigQuery(appID);

  return useLoadableView({
    loadables: [appConfigQuery] as const,
    render: ([{ effectiveAppConfig }]) => {
      return (
        <APIResourceDetailsScreenTestTabContent
          resource={resource}
          appConfig={effectiveAppConfig!}
        />
      );
    },
  });
}

function APIResourceDetailsScreenTestTabContent({
  resource,
  appConfig,
}: {
  resource: Resource;
  appConfig: PortalAPIAppConfig;
}) {
  const { renderToString } = useContext(MessageContext);

  const [selectedClientId, setSelectedClientId] = useState<string | null>(null);
  const [accessToken, _setAccessToken] = useState<string | null>(null);
  const [selectedPivotKey, setSelectedPivotKey] = useState<ExampleCodeTabKey>(
    ExampleCodeTabKey.CURL
  );

  const authorizedApplicationsOptions = useMemo(() => {
    const authorizedClientIDs = new Set(resource.clientIDs);
    return (
      appConfig.oauth?.clients
        ?.filter((clientConfig) => {
          return authorizedClientIDs.has(clientConfig.client_id);
        })
        .map((clientConfig) => ({
          key: clientConfig.client_id,
          text: clientConfig.name ?? clientConfig.client_name ?? "",
        })) ?? []
    );
  }, [appConfig.oauth?.clients, resource.clientIDs]);

  const handleDropdownChange = useCallback(
    (_: unknown, option?: IDropdownOption) => {
      setSelectedClientId(String(option?.key ?? ""));
    },
    []
  );

  const handlePivotClick = useCallback((item?: PivotItem) => {
    if (item?.props.itemKey) {
      setSelectedPivotKey(item.props.itemKey as ExampleCodeTabKey);
    }
  }, []);

  return (
    <div className="pt-5 flex-1 flex flex-col space-y-4">
      <header className="space-y-2">
        <WidgetTitle>
          <FormattedMessage id="APIResourceDetailsScreen.tab.test" />
        </WidgetTitle>
        <Text block={true}>
          <FormattedMessage id="APIResourceDetailsScreen.test.description" />
        </Text>
      </header>
      <div className="flex-1 flex flex-col max-w-180 space-y-5">
        <Dropdown
          label={renderToString(
            "APIResourceDetailsScreen.test.authorizedApplications"
          )}
          placeholder={
            authorizedApplicationsOptions.length === 0
              ? renderToString(
                  "APIResourceDetailsScreen.test.selectApplication.empty"
                )
              : renderToString(
                  "APIResourceDetailsScreen.test.selectApplication"
                )
          }
          options={authorizedApplicationsOptions}
          disabled={authorizedApplicationsOptions.length === 0}
          selectedKey={selectedClientId}
          onChange={handleDropdownChange}
        />
        {selectedClientId !== null ? (
          <>
            <HorizontalDivider />
            <section>
              <WidgetTitle>
                <FormattedMessage id="APIResourceDetailsScreen.test.accessToken.title" />
              </WidgetTitle>
              <Text block={true} className="mt-2">
                <FormattedMessage id="APIResourceDetailsScreen.test.accessToken.description" />
              </Text>
              <CodeField
                className="mt-3"
                codeClassName="h-25 overflow-y-auto"
                placeholder={
                  <FormattedMessage id="APIResourceDetailsScreen.test.accessToken.placeholder" />
                }
              >
                {accessToken}
              </CodeField>
              <div className="mt-4 flex space-x-4">
                <PrimaryButton
                  text={
                    <FormattedMessage id="APIResourceDetailsScreen.test.generateButton.text" />
                  }
                />
                <DefaultButton
                  text={<FormattedMessage id="copy" />}
                  disabled={accessToken == null}
                />
              </div>
            </section>
            <HorizontalDivider />
            <section>
              <WidgetTitle>
                <FormattedMessage id="APIResourceDetailsScreen.test.requestToken.title" />
              </WidgetTitle>
              <Pivot
                className="mt-2"
                selectedKey={selectedPivotKey}
                onLinkClick={handlePivotClick}
              >
                <PivotItem
                  headerText={renderToString(
                    "APIResourceDetailsScreen.test.pivot.curl.headerText"
                  )}
                  itemKey={ExampleCodeTabKey.CURL}
                />
                <PivotItem
                  headerText={renderToString(
                    "APIResourceDetailsScreen.test.pivot.python.headerText"
                  )}
                  itemKey={ExampleCodeTabKey.Python}
                />
                <PivotItem
                  headerText={renderToString(
                    "APIResourceDetailsScreen.test.pivot.go.headerText"
                  )}
                  itemKey={ExampleCodeTabKey.Go}
                />
                <PivotItem
                  headerText={renderToString(
                    "APIResourceDetailsScreen.test.pivot.nodejs.headerText"
                  )}
                  itemKey={ExampleCodeTabKey.NodeJS}
                />
              </Pivot>
              <CodeField className="mt-4">{"TODO: Example code"}</CodeField>
              <div className="mt-4">
                <PrimaryButton text={<FormattedMessage id="reveal" />} />
              </div>
            </section>
          </>
        ) : null}
      </div>
    </div>
  );
}
