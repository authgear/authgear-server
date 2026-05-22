import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import {
  Dialog,
  DialogFooter,
  IDialogContentProps,
  Text,
} from "@fluentui/react";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import { useNavigate, useParams } from "react-router-dom";
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
import { useLoadableView } from "../../hook/useLoadableView";

interface FormState {
  clients: OAuthClientConfig[];
  newClient: OAuthClientConfig;
  authorizeResourceURIs: string[];
}

function constructFormState(
  config: PortalAPIAppConfig,
  _secretConfig: PortalAPISecretConfig
): FormState {
  return {
    clients: config.oauth?.clients ?? [],
    newClient: {
      x_application_type: "m2m",
      client_id: genRandomHexadecimalString(),
    },
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
      draft.x_application_type = "m2m";
      draft.issue_jwt_access_token = true;
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
  return {
    oauthClientSecrets: {
      action: "generate",
      generateData: {
        clientID: currentState.newClient.client_id,
      },
    },
  };
}

function constructInitialCurrentState(state: FormState): FormState {
  return produce(state, (state) => {
    state.newClient.name = "My App";
  });
}

interface StepAuthorizeResourceProps {
  client: OAuthClientConfig;
  form: AppSecretConfigFormModel<FormState>;
  onClickSave: () => void;
}

const StepAuthorizeResource: React.VFC<StepAuthorizeResourceProps> =
  function StepAuthorizeResource(props) {
    const { client, form, onClickSave } = props;
    const { isDirty, isUpdating, setState } = form;
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
      // eslint-disable-next-line @typescript-eslint/strict-void-return
      return <ShowError error={error} onRetry={refetch} />;
    }

    return (
      <Widget className={cn(styles.widget, "flex flex-col gap-y-4")}>
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
        <div
          style={{ minHeight: 500, display: "flex", flexDirection: "column" }}
        >
          <ApplicationResourcesList
            className="flex-1"
            resources={resourceListData}
            loading={loading}
            pagination={pagination}
            onToggleAuthorization={handleToggleAuthorization}
          />
        </div>
        <div className={styles.buttons}>
          <ButtonWithLoading
            onClick={onClickSave}
            loading={isUpdating}
            disabled={!isDirty}
            labelId="save"
          />
        </div>
      </Widget>
    );
  };

interface CreateM2MClientContentProps {
  form: AppSecretConfigFormModel<FormState>;
}

const CreateM2MClientContent: React.VFC<CreateM2MClientContentProps> =
  function CreateM2MClientContent(props) {
    const { form } = props;
    const { state, save } = form;
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
      save()
        .then(
          () => {
            const nextPath = `/project/${appID}/configuration/apps/${encodeURIComponent(
              clientId
            )}/edit`;
            navigate(
              {
                pathname: nextPath,
              },
              {
                replace: true,
              }
            );
          },
          () => {}
        )
        .catch(() => {});
    }, [save, appID, clientId, navigate]);

    return (
      <ScreenContent className="flex-1-0-auto" layout={"list"}>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        <StepAuthorizeResource
          client={client}
          form={form}
          onClickSave={onClickSave}
        />
      </ScreenContent>
    );
  };

const CreateM2MClientScreen: React.VFC = function CreateM2MClientScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);
  const [addResource] = useAddResourceToClientIdMutation();

  const resourcesCountQuery = useResourcesQueryQuery({
    variables: { first: 1 },
    fetchPolicy: "cache-and-network",
  });
  const noAPIResources =
    !resourcesCountQuery.loading &&
    resourcesCountQuery.error == null &&
    (resourcesCountQuery.data?.resources?.totalCount ?? 0) === 0;

  const goToAPIResources = useCallback(() => {
    navigate(`/project/${appID}/api-resources/create`);
  }, [appID, navigate]);

  const goToAppsList = useCallback(() => {
    navigate(`/project/${appID}/configuration/apps`);
  }, [appID, navigate]);

  const noResourcesDialogContent: IDialogContentProps = useMemo(
    () => ({
      title: renderToString("CreateM2MClientScreen.no-resources-dialog.title"),
      subText: renderToString("CreateM2MClientScreen.no-resources-dialog.body"),
    }),
    [renderToString]
  );

  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: null,
    constructFormState,
    constructConfig,
    constructInitialCurrentState,
    constructSecretUpdateInstruction,
    postSave: useCallback(
      async (state: FormState) => {
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
    loadables: [form] as const,
    render: ([form]) => (
      <FormProvider
        loading={form.isUpdating}
        error={form.updateError}
        rules={errorRules}
      >
        <FormErrorMessageBar />
        <div className="flex-1 overflow-y-auto flex flex-col">
          <CreateM2MClientContent form={form} />
        </div>
        <Dialog
          hidden={!noAPIResources}
          dialogContentProps={noResourcesDialogContent}
          modalProps={{ isBlocking: true }}
          onDismiss={goToAppsList}
        >
          <DialogFooter>
            <PrimaryButton
              onClick={goToAPIResources}
              text={
                <FormattedMessage id="CreateM2MClientScreen.no-resources-dialog.cta" />
              }
            />
            <DefaultButton
              onClick={goToAppsList}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
      </FormProvider>
    ),
  });
};

export default CreateM2MClientScreen;
