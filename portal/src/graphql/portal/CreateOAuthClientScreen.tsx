import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupOptionProps,
  Text,
} from "@fluentui/react";
import { useNavigate, useParams, Link } from "react-router-dom";
import { produce, createDraft } from "immer";
import { Context, FormattedMessage } from "../../intl";
import { SearchBox } from "@fluentui/react/lib/SearchBox";
import { useResourcesQueryQuery } from "../adminapi/query/resourcesQuery.generated";
import {
  ApplicationResourcesList,
  ApplicationResourceListItem,
} from "../../components/api-resources/ApplicationResourcesList";
import { encodeOffsetToCursor } from "../../util/pagination";
import { PaginationProps } from "../../PaginationWidget";
import { useDebounced } from "../../hook/useDebounced";
import { useAddResourceToClientIdMutation } from "../adminapi/mutations/addResourceToClientID.generated";

import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import { updateClientConfig } from "./EditOAuthClientForm";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import {
  OAuthClientConfig,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../../types";
import { clearEmptyObject, ensureNonEmptyString } from "../../util/misc";
import { genRandomHexadecimalString } from "../../util/random";
import { makeValidationErrorCustomMessageIDRule } from "../../error/parse";
import styles from "./CreateOAuthClientScreen.module.css";
import { FormProvider } from "../../form";
import FormTextField from "../../FormTextField";
import { useTextField } from "../../hook/useInput";
import Widget from "../../Widget";
import ButtonWithLoading from "../../ButtonWithLoading";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import LinkButton from "../../LinkButton";
import { useAppContext } from "../../context/AppContext";
import { useLoadableView } from "../../hook/useLoadableView";

interface FormState {
  clients: OAuthClientConfig[];
  newClient: OAuthClientConfig;
  step: FormStep;
  authorizeResourceURIs: string[];
}

enum FormStep {
  SelectType = "select_type",
  AuthorizeResource = "authorize_resource",
}

function constructFormState(
  config: PortalAPIAppConfig,
  _secretConfig: PortalAPISecretConfig
): FormState {
  return {
    clients: config.oauth?.clients ?? [],
    newClient: {
      x_application_type: "spa",
      client_id: genRandomHexadecimalString(),
    },
    step: FormStep.SelectType,
    authorizeResourceURIs: [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secretConfig: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const [newConfig, _] = produce(
    [config, currentState],
    ([config, currentState]) => {
      config.oauth ??= {};
      config.oauth.clients = currentState.clients;
      const draft = createDraft(currentState.newClient);
      if (draft.x_application_type == null) {
        throw new Error("unexpected null x_application_type");
      }
      switch (draft.x_application_type) {
        case "spa":
        case "traditional_webapp":
          draft.redirect_uris = ["http://localhost/after-authentication"];
          draft.post_logout_redirect_uris = ["http://localhost/after-logout"];
          draft.grant_types = ["authorization_code", "refresh_token"];
          draft.response_types = ["code", "none"];
          draft.issue_jwt_access_token = true;
          break;
        case "native":
          draft.redirect_uris = ["com.example.myapp://host/path"];
          draft.grant_types = ["authorization_code", "refresh_token"];
          draft.response_types = ["code", "none"];
          draft.issue_jwt_access_token = true;
          break;
        case "confidential":
        case "third_party_app":
          draft.client_name = draft.name;
          draft.redirect_uris = ["http://localhost/after-authentication"];
          draft.grant_types = ["authorization_code", "refresh_token"];
          draft.response_types = ["code", "none"];
          draft.issue_jwt_access_token = true;
          break;
        case "m2m":
          draft.issue_jwt_access_token = true;
          break;
      }
      config.oauth.clients.push(draft);
      clearEmptyObject(config);
    }
  );
  return [newConfig, secretConfig];
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  _secrets: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  const clientTypesWithSecret: OAuthClientConfig["x_application_type"][] = [
    "confidential",
    "third_party_app",
    "m2m",
  ];
  if (
    clientTypesWithSecret.includes(currentState.newClient.x_application_type)
  ) {
    return {
      oauthClientSecrets: {
        action: "generate",
        generateData: {
          clientID: currentState.newClient.client_id,
        },
      },
    };
  }
  return undefined;
}

function constructInitialCurrentState(state: FormState): FormState {
  return produce(state, (state) => {
    state.newClient.name = "My App";
  });
}

function getNextStep(state: FormState): FormStep | null {
  if (state.newClient.x_application_type === "m2m") {
    if (state.step === FormStep.SelectType) {
      return FormStep.AuthorizeResource;
    }
  }
  return null;
}

interface CreateOAuthClientContentProps {
  form: AppSecretConfigFormModel<FormState>;
  hasNoAPIResources: boolean;
}

interface StepSelectApplicationTypeProps {
  client: OAuthClientConfig;
  form: AppSecretConfigFormModel<FormState>;
  onClickSave: () => void;
  hasNoAPIResources: boolean;
}

const StepSelectApplicationType: React.VFC<StepSelectApplicationTypeProps> =
  function StepSelectApplicationType(props) {
    const { client, form, onClickSave, hasNoAPIResources } = props;
    const { appNodeID } = useAppContext();
    const { state, setState, isDirty, isUpdating } = form;
    const { renderToString } = useContext(Context);

    const onClientConfigChange = useCallback(
      (newClient: OAuthClientConfig) => {
        setState((s) => ({ ...s, newClient }));
      },
      [setState]
    );

    const { onChange: onClientNameChange } = useTextField((value) => {
      onClientConfigChange(
        updateClientConfig(client, "name", ensureNonEmptyString(value))
      );
    });

    const onRenderLabel = useCallback((description: string) => {
      return (option?: IChoiceGroupOption | IChoiceGroupOptionProps) => {
        return (
          <div className={styles.optionLabel}>
            <Text className={styles.optionLabelText} block={true}>
              {option?.text}
            </Text>
            <Text className={styles.optionLabelDescription} block={true}>
              {description}
            </Text>
          </div>
        );
      };
    }, []);

    const options: IChoiceGroupOption[] = useMemo(() => {
      return [
        {
          key: "spa",
          text: renderToString("oauth-client.application-type.spa"),
          onRenderLabel: onRenderLabel(
            renderToString(
              "CreateOAuthClientScreen.application-type.description.spa"
            )
          ),
        },
        {
          key: "traditional_webapp",
          text: renderToString(
            "oauth-client.application-type.traditional-webapp"
          ),
          onRenderLabel: onRenderLabel(
            renderToString(
              "CreateOAuthClientScreen.application-type.description.traditional-webapp"
            )
          ),
        },
        {
          key: "native",
          text: renderToString("oauth-client.application-type.native"),
          onRenderLabel: onRenderLabel(
            renderToString(
              "CreateOAuthClientScreen.application-type.description.native"
            )
          ),
        },
        {
          key: "confidential",
          text: renderToString("oauth-client.application-type.confidential"),
          onRenderLabel: onRenderLabel(
            renderToString(
              "CreateOAuthClientScreen.application-type.description.confidential"
            )
          ),
        },
        // Do not show this option.
        //{
        //  key: "third_party_app",
        //  text: renderToString("oauth-client.application-type.third-party-app"),
        //  onRenderLabel: onRenderLabel(
        //    renderToString(
        //      "CreateOAuthClientScreen.application-type.description.third-party-app"
        //    )
        //  ),
        //  disabled: true,
        //},
        {
          key: "m2m",
          text: renderToString("oauth-client.application-type.m2m"),
          onRenderLabel: onRenderLabel(
            (
              <FormattedMessage
                id={
                  hasNoAPIResources
                    ? "CreateOAuthClientScreen.application-type.description.m2m.disabled"
                    : "CreateOAuthClientScreen.application-type.description.m2m"
                }
                values={{
                  reactRouterLink: (chunks: React.ReactNode) => (
                    <Link
                      to={`/project/${encodeURIComponent(
                        appNodeID
                      )}/api-resources/create`}
                    >
                      {chunks}
                    </Link>
                  ),
                }}
              />
            ) as any as string
          ),
          disabled: hasNoAPIResources,
        },
      ];
    }, [renderToString, onRenderLabel, appNodeID, hasNoAPIResources]);

    const onApplicationChange = useCallback(
      (_e: unknown, option?: IChoiceGroupOption) => {
        if (option != null) {
          let issueJwtAccessToken: boolean | undefined;
          switch (option.key) {
            case "spa":
            case "native":
              issueJwtAccessToken = true;
              break;
            case "m2m":
              issueJwtAccessToken = true;
              break;
            default:
              issueJwtAccessToken = undefined;
              break;
          }
          onClientConfigChange(
            updateClientConfig(
              updateClientConfig(
                client,
                "x_application_type",
                option.key as OAuthClientConfig["x_application_type"]
              ),
              "issue_jwt_access_token",
              issueJwtAccessToken
            )
          );
        }
      },
      [onClientConfigChange, client]
    );

    return (
      <Widget className={cn(styles.widget)}>
        <FormTextField
          parentJSONPointer={/\/oauth\/clients\/\d+/}
          fieldName="name"
          label={renderToString("CreateOAuthClientScreen.name.label")}
          description={renderToString(
            "CreateOAuthClientScreen.name.description"
          )}
          value={client.name ?? ""}
          onChange={onClientNameChange}
          required={true}
        />
        <ChoiceGroup
          label={renderToString(
            "CreateOAuthClientScreen.application-type.label"
          )}
          options={options}
          selectedKey={client.x_application_type}
          onChange={onApplicationChange}
        />
        <div className={styles.buttons}>
          <ButtonWithLoading
            onClick={onClickSave}
            loading={isUpdating}
            disabled={!isDirty}
            labelId={getNextStep(state) != null ? "next" : "save"}
          />
        </div>
      </Widget>
    );
  };

interface StepAuthorizeResourceProps {
  client: OAuthClientConfig;
  form: AppSecretConfigFormModel<FormState>;
  onClickSave: () => void;
}

const StepAuthorizeResource: React.VFC<StepAuthorizeResourceProps> =
  function StepAuthorizeResource(props) {
    const { form, onClickSave } = props;
    const { isDirty, isUpdating } = form;
    const { renderToString } = useContext(Context);
    const [searchKeyword, setSearchKeyword] = useState("");
    const [offset, setOffset] = useState(0);

    const [debouncedSearchKeyword] = useDebounced(searchKeyword, 300);

    const PAGE_SIZE = 10;

    const { data, loading, error, refetch } = useResourcesQueryQuery({
      variables: {
        first: PAGE_SIZE,
        after: encodeOffsetToCursor(offset),
        searchKeyword:
          debouncedSearchKeyword === "" ? undefined : debouncedSearchKeyword,
      },
      fetchPolicy: "cache-and-network",
    });

    const resourceListData: ApplicationResourceListItem[] = useMemo(() => {
      const resources =
        data?.resources?.edges
          ?.map((edge) => edge?.node)
          .filter((node) => !!node) ?? [];
      return resources.map((resource) => {
        const isAuthorized = form.state.authorizeResourceURIs.includes(
          resource.resourceURI
        );
        return {
          id: resource.id,
          name: resource.name,
          resourceURI: resource.resourceURI,
          isAuthorized: isAuthorized,
        };
      });
    }, [data?.resources?.edges, form.state.authorizeResourceURIs]);

    const handleToggleAuthorization = useCallback(
      (item: ApplicationResourceListItem, isAuthorized: boolean) => {
        form.setState((s) => {
          const uris = new Set(s.authorizeResourceURIs);
          if (isAuthorized) {
            uris.add(item.resourceURI);
          } else {
            uris.delete(item.resourceURI);
          }
          return { ...s, authorizeResourceURIs: Array.from(uris) };
        });
      },
      [form]
    );

    const onChangeOffset = useCallback((newOffset: number) => {
      setOffset(newOffset);
    }, []);

    const pagination: PaginationProps = {
      offset,
      pageSize: PAGE_SIZE,
      totalCount: data?.resources?.totalCount ?? undefined,
      onChangeOffset,
    };

    if (error != null) {
      return <ShowError error={error} onRetry={refetch} />;
    }

    return (
      <div className={cn(styles.widget, "flex flex-col gap-y-4")}>
        <Text block={true}>
          <FormattedMessage id="CreateOAuthClientScreen.authorize-resource.description" />
        </Text>
        <SearchBox
          placeholder={renderToString("search")}
          styles={{ root: { width: 300 } }}
          onChange={(_e, newValue) => {
            setSearchKeyword(newValue ?? "");
            setOffset(0);
          }}
        />
        <ApplicationResourcesList
          className="flex-1"
          resources={resourceListData}
          loading={loading}
          pagination={pagination}
          onToggleAuthorization={handleToggleAuthorization}
        />
        <div className={styles.buttons}>
          <ButtonWithLoading
            onClick={onClickSave}
            loading={isUpdating}
            disabled={!isDirty}
            labelId="save"
          />
          <LinkButton
            onClick={() => {
              form.setState((s) => ({ ...s, step: FormStep.SelectType }));
            }}
          >
            <FormattedMessage id="back" />
          </LinkButton>
        </div>
      </div>
    );
  };

const CreateOAuthClientContent: React.VFC<CreateOAuthClientContentProps> =
  function CreateOAuthClientContent(props) {
    const { form, hasNoAPIResources } = props;
    const { state, setState, save } = form;
    const { appID } = useParams() as { appID: string };
    const navigate = useNavigate();

    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: "~/configuration/apps",
          label: (
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          ),
        },
        {
          to: ".",
          label: <FormattedMessage id="CreateOAuthClientScreen.title" />,
        },
      ];
    }, []);

    const [clientId] = useState(state.newClient.client_id);
    const client =
      state.clients.find((c) => c.client_id === clientId) ?? state.newClient;

    const onClickSave = useCallback(() => {
      const nextStep = getNextStep(state);
      if (nextStep != null) {
        setState((s) => ({ ...s, step: nextStep }));
        return;
      }
      save()
        .then(
          () => {
            const applicationTypesWithQuickStart: OAuthClientConfig["x_application_type"][] =
              [
                "confidential",
                "native",
                "spa",
                "third_party_app",
                "traditional_webapp",
              ];
            const nextPath = `/project/${appID}/configuration/apps/${encodeURIComponent(
              clientId
            )}/edit`;
            const searchParams = new URLSearchParams();
            if (
              applicationTypesWithQuickStart.includes(client.x_application_type)
            ) {
              searchParams.set("quickstart", "true");
            }
            navigate(
              {
                pathname: nextPath,
                search: searchParams.toString(),
              },
              {
                replace: true,
              }
            );
          },
          () => { }
        )
        .catch(() => { });
    }, [
      save,
      appID,
      clientId,
      client.x_application_type,
      navigate,
      setState,
      state,
    ]);

    return (
      <ScreenContent className="flex-1-0-auto" layout={"list"}>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        {state.step === FormStep.SelectType ? (
          <StepSelectApplicationType
            client={client}
            form={form}
            onClickSave={onClickSave}
            hasNoAPIResources={hasNoAPIResources}
          />
        ) : null}
        {state.step === FormStep.AuthorizeResource ? (
          <StepAuthorizeResource
            client={client}
            form={form}
            onClickSave={onClickSave}
          />
        ) : null}
      </ScreenContent>
    );
  };

const CreateOAuthClientScreen: React.VFC = function CreateOAuthClientScreen() {
  const { appID } = useParams() as { appID: string };
  const [addResource] = useAddResourceToClientIdMutation();

  const resourceCountQuery = useResourcesQueryQuery({
    variables: {
      first: 1,
    },
    fetchPolicy: "cache-and-network",
  });

  const hasNoAPIResources =
    (resourceCountQuery.data?.resources?.totalCount ?? 0) === 0;

  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: null,
    constructFormState,
    constructConfig,
    constructInitialCurrentState,
    constructSecretUpdateInstruction,
    postSave: useCallback(
      async (state: FormState) => {
        if (state.newClient.x_application_type !== "m2m") {
          return;
        }
        const clientID = state.newClient.client_id;
        const uris = state.authorizeResourceURIs;
        for (const resourceURI of uris) {
          await addResource({
            variables: {
              clientID,
              resourceURI,
            },
          });
        }
      },
      [addResource]
    ),
  });

  const errorRules = useMemo(
    () => [
      makeValidationErrorCustomMessageIDRule(
        "general",
        /^\/oauth\/clients$/,
        "error.client-quota-exceeded",
        {
          to: `/project/${appID}/billing`,
        }
      ),
    ],
    [appID]
  );

  return useLoadableView({
    loadables: [
      form,
      {
        isLoading: resourceCountQuery.loading,
        loadError: resourceCountQuery.error,
        reload: resourceCountQuery.refetch,
      },
    ] as const,
    render: ([form]) => (
      <FormProvider
        loading={form.isUpdating}
        error={form.updateError}
        rules={errorRules}
      >
        <FormErrorMessageBar />
        <div className="flex-1 overflow-y-auto flex flex-col">
          <CreateOAuthClientContent
            form={form}
            hasNoAPIResources={hasNoAPIResources}
          />
        </div>
      </FormProvider>
    ),
  });
};

export default CreateOAuthClientScreen;
