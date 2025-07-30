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
import { useNavigate } from "react-router-dom";
import { useEndpoints } from "../../hook/useEndpoints";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../types";
import HorizontalDivider from "../../HorizontalDivider";
import { CodeField } from "../../components/common/CodeField";
import { startReauthentication } from "../../graphql/portal/Authenticated";
import { LocationState } from "./APIResourceDetailsScreen";
import { useSearchParamsState } from "../../hook/useSearchParamsState";

enum ExampleCodeTabKey {
  CURL = "CURL",
  Python = "Python",
  Go = "Go",
  NodeJS = "NodeJS",
}

export function APIResourceDetailsScreenTestTab({
  resource,
  effectiveAppConfig,
  secretConfig,
}: {
  resource: Resource;
  effectiveAppConfig: PortalAPIAppConfig;
  secretConfig: PortalAPISecretConfig | null;
}): React.ReactElement | null {
  return (
    <APIResourceDetailsScreenTestTabContent
      resource={resource}
      effectiveAppConfig={effectiveAppConfig}
      secretConfig={secretConfig}
    />
  );
}

function APIResourceDetailsScreenTestTabContent({
  resource,
  effectiveAppConfig,
  secretConfig,
}: {
  resource: Resource;
  effectiveAppConfig: PortalAPIAppConfig;
  secretConfig: PortalAPISecretConfig | null;
}) {
  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();

  const [selectedClientId, setSelectedClientId] = useSearchParamsState<string>(
    "client",
    ""
  );
  const [accessToken, _setAccessToken] = useState<string | null>(null);
  const [selectedPivotKey, setSelectedPivotKey] = useState<ExampleCodeTabKey>(
    ExampleCodeTabKey.CURL
  );

  const selectedClient = useMemo(() => {
    return effectiveAppConfig.oauth?.clients?.find(
      (client) => client.client_id === selectedClientId
    );
  }, [effectiveAppConfig, selectedClientId]);

  const selectedClientSecret = useMemo((): string | null => {
    if (!secretConfig || !selectedClient?.client_id) {
      return null;
    }
    const secret = secretConfig.oauthClientSecrets?.find(
      (secret) => secret.clientID === selectedClient.client_id
    );
    if (secret?.keys != null && secret.keys.length > 0 && secret.keys[0].key) {
      return secret.keys[0].key;
    }
    return null;
  }, [secretConfig, selectedClient]);

  const { token: _tokenEndpoint } = useEndpoints(
    effectiveAppConfig.http?.public_origin ?? ""
  );

  const authorizedApplicationsOptions = useMemo(() => {
    const authorizedClientIDs = new Set(resource.clientIDs);
    return (
      effectiveAppConfig.oauth?.clients
        ?.filter((clientConfig) => {
          return authorizedClientIDs.has(clientConfig.client_id);
        })
        .map((clientConfig) => ({
          key: clientConfig.client_id,
          text: clientConfig.name ?? clientConfig.client_name ?? "",
        })) ?? []
    );
  }, [effectiveAppConfig.oauth?.clients, resource.clientIDs]);

  const handleDropdownChange = useCallback(
    (_: unknown, option?: IDropdownOption) => {
      setSelectedClientId(String(option?.key ?? ""));
    },
    [setSelectedClientId]
  );

  const handlePivotClick = useCallback((item?: PivotItem) => {
    if (item?.props.itemKey) {
      setSelectedPivotKey(item.props.itemKey as ExampleCodeTabKey);
    }
  }, []);

  const revealSecrets = useCallback(() => {
    startReauthentication<LocationState>(navigate, {
      isClientSecretRevealed: true,
    }).catch((e) => {
      // Normally there should not be any error.
      console.error(e);
    });
  }, [navigate]);

  const onGenerate = useCallback(() => {
    if (selectedClientSecret == null) {
      revealSecrets();
    } else {
      // TODO
    }
  }, [selectedClientSecret, revealSecrets]);

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
          selectedKey={selectedClient?.client_id}
          onChange={handleDropdownChange}
        />
        {selectedClient != null ? (
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
                  onClick={onGenerate}
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
              <div className="mt-4 flex space-x-4">
                <PrimaryButton
                  text={<FormattedMessage id="reveal" />}
                  onClick={revealSecrets}
                  disabled={selectedClientSecret != null}
                />
                <DefaultButton text={<FormattedMessage id="copy" />} />
              </div>
            </section>
          </>
        ) : null}
      </div>
    </div>
  );
}
