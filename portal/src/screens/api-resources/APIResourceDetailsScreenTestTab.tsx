import React, {
  useContext,
  useMemo,
  useState,
  useCallback,
  useEffect,
} from "react";
import DefaultButton from "../../DefaultButton";
import {
  Context as MessageContext,
  FormattedMessage,
} from "../../intl";
import WidgetTitle from "../../WidgetTitle";
import { Text, Dropdown, IDropdownOption, PivotItem } from "@fluentui/react";
import { AGPivot } from "../../components/common/AGPivot";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { useNavigate } from "react-router-dom";
import { useEndpoints } from "../../hook/useEndpoints";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../types";
import HorizontalDivider from "../../HorizontalDivider";
import { CodeField } from "../../components/common/CodeField";
import { useStartReauthentication } from "../../graphql/portal/Authenticated";
import { LocationState } from "./APIResourceDetailsScreen";
import { useSearchParamsState } from "../../hook/useSearchParamsState";
import { useErrorMessageBarContext } from "../../ErrorMessageBar";
import { parseRawError } from "../../error/parse";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import {
  ExampleCodeVariant,
  useExampleCode,
} from "../../components/api-resources/useExampleCode";
import ButtonWithLoading from "../../ButtonWithLoading";

export function APIResourceDetailsScreenTestTab({
  resource,
  effectiveAppConfig,
  secretConfig,
}: {
  resource: Resource;
  effectiveAppConfig: PortalAPIAppConfig;
  secretConfig: PortalAPISecretConfig | null;
}): React.ReactElement | null {
  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();
  const { setErrors } = useErrorMessageBarContext();
  const [selectedClientId, setSelectedClientId] = useSearchParamsState<string>(
    "client",
    ""
  );
  const [isGenerating, setIsGenerating] = useState(false);
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [selectedCodeVariant, setSelectedCodeVariant] =
    useState<ExampleCodeVariant>(ExampleCodeVariant.curl);
  const { startReauthentication, isRevealing } =
    useStartReauthentication<LocationState>();

  const selectedClient = useMemo(() => {
    return effectiveAppConfig.oauth?.clients?.find(
      (client) => client.client_id === selectedClientId
    );
  }, [effectiveAppConfig, selectedClientId]);

  useEffect(() => {
    setAccessToken(null);
  }, [selectedClient]);

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

  const { token: tokenEndpoint } = useEndpoints(
    effectiveAppConfig.http?.public_origin ?? "",
    selectedClient?.x_application_type
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
      setSelectedCodeVariant(item.props.itemKey as ExampleCodeVariant);
    }
  }, []);

  const revealSecrets = useCallback(() => {
    startReauthentication(navigate, {
      isClientSecretRevealed: true,
    }).catch((e) => {
      // Normally there should not be any error.
      console.error(e);
    });
  }, [navigate, startReauthentication]);

  const onGenerate = useCallback(async () => {
    if (selectedClientSecret == null) {
      revealSecrets();
    } else {
      const body = new URLSearchParams();
      body.append("client_id", selectedClientId);
      body.append("grant_type", "client_credentials");
      body.append("resource", resource.resourceURI);
      body.append("client_secret", selectedClientSecret);
      setIsGenerating(true);
      try {
        const response = await fetch(tokenEndpoint, {
          method: "POST",
          headers: {
            "Content-Type": "application/x-www-form-urlencoded",
          },
          body: body.toString(),
        });

        if (!response.ok) {
          throw new Error(`invalid response status: ${response.status}`);
        }

        const data = await response.json();
        setAccessToken(data.access_token);
      } catch (error) {
        console.error("Error generating access token:", error);
        setErrors(parseRawError(error));
        setAccessToken(null);
      } finally {
        setIsGenerating(false);
      }
    }
  }, [
    selectedClientSecret,
    revealSecrets,
    selectedClientId,
    resource.resourceURI,
    tokenEndpoint,
    setErrors,
  ]);

  const exampleCode = useExampleCode({
    variant: selectedCodeVariant,
    tokenEndpoint,
    resourceURI: resource.resourceURI,
    clientSecret: selectedClientSecret,
    clientID: selectedClientId,
  });

  const { copyButtonProps: copyTokenButtonProps, Feedback: CopyTokenFeedback } =
    useCopyFeedback({
      textToCopy: accessToken ?? "",
    });

  const { copyButtonProps: copyCodeButtonProps, Feedback: CopyCodeFeedback } =
    useCopyFeedback({
      textToCopy: exampleCode,
    });

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
                <ButtonWithLoading
                  labelId="APIResourceDetailsScreen.test.generateButton.text"
                  onClick={onGenerate}
                  disabled={isGenerating}
                  loading={isRevealing || isGenerating}
                />
                <DefaultButton
                  {...copyTokenButtonProps}
                  text={<FormattedMessage id="copy" />}
                  disabled={accessToken == null}
                  iconProps={undefined}
                />
                <CopyTokenFeedback />
              </div>
            </section>
            <HorizontalDivider />
            <section>
              <WidgetTitle>
                <FormattedMessage
                  id="APIResourceDetailsScreen.test.requestToken.title"
                  values={{
                    clientName:
                      selectedClient.client_name ??
                      selectedClient.name ??
                      selectedClient.client_id,
                  }}
                />
              </WidgetTitle>
              <AGPivot
                className="mt-2"
                selectedKey={selectedCodeVariant}
                onLinkClick={handlePivotClick}
              >
                <PivotItem
                  headerText={renderToString(
                    "APIResourceDetailsScreen.test.pivot.curl.headerText"
                  )}
                  itemKey={ExampleCodeVariant.curl}
                />
                <PivotItem
                  headerText={renderToString(
                    "APIResourceDetailsScreen.test.pivot.python.headerText"
                  )}
                  itemKey={ExampleCodeVariant.Python}
                />
                <PivotItem
                  headerText={renderToString(
                    "APIResourceDetailsScreen.test.pivot.go.headerText"
                  )}
                  itemKey={ExampleCodeVariant.Go}
                />
                <PivotItem
                  headerText={renderToString(
                    "APIResourceDetailsScreen.test.pivot.nodejs.headerText"
                  )}
                  itemKey={ExampleCodeVariant.NodeJS}
                />
              </AGPivot>
              <CodeField className="mt-4">{exampleCode}</CodeField>
              <div className="mt-4 flex space-x-4">
                <ButtonWithLoading
                  labelId="reveal"
                  onClick={revealSecrets}
                  disabled={selectedClientSecret != null}
                  loading={isRevealing}
                />
                <DefaultButton
                  {...copyCodeButtonProps}
                  text={<FormattedMessage id="copy" />}
                  iconProps={undefined}
                />
                <CopyCodeFeedback />
              </div>
            </section>
          </>
        ) : null}
      </div>
    </div>
  );
}
