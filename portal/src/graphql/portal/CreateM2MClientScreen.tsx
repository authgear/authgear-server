import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { produce, createDraft } from "immer";
import { Context, FormattedMessage } from "../../intl";
import { useResourcesQueryQuery } from "../adminapi/query/resourcesQuery.generated";
import {
  ApplicationResourcesList,
  ApplicationResourceListItem,
} from "../../components/api-resources/ApplicationResourcesList";
import { encodeOffsetToCursor } from "../../util/pagination";
import { PaginationProps } from "../../PaginationWidget";
import { useDebounced } from "../../hook/useDebounced";
import { useCapture } from "../../gtm_v2";
import { useAddResourceToClientIdMutation } from "../adminapi/mutations/addResourceToClientID.generated";

import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import Link from "../../Link";
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
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import { useLoadableView } from "../../hook/useLoadableView";
import { FROM_CREATE_APPLICATION_FLOW_STATE } from "./ApplicationsConfigurationScreen";

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

interface StepAuthorizeResourceProps {
  client: OAuthClientConfig;
  form: AppSecretConfigFormModel<FormState>;
  onClickSave: () => void;
  onClickBack: () => void;
}

const StepAuthorizeResource: React.VFC<StepAuthorizeResourceProps> =
  function StepAuthorizeResource(props) {
    const { client, form, onClickSave, onClickBack } = props;
    const { getIsDirty, isUpdating } = form;
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const { renderToString } = useContext(Context);
    const [searchKeyword, setSearchKeyword] = useState("");
    const [offset, setOffset] = useState(0);

    const [debouncedSearchKeyword] = useDebounced(searchKeyword, 300);

    const PAGE_SIZE = 10;

    const onSearchChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setSearchKeyword(e.currentTarget.value);
        setOffset(0);
      },
      []
    );

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
      // eslint-disable-next-line @typescript-eslint/strict-void-return
      return <ShowError error={error} onRetry={refetch} />;
    }

    return (
      <div className={cn(styles.widget, styles.wizardPanel)}>
        <div className={styles.m2mNameBlock}>
          <Text as="p" size="2" weight="medium" className={styles.m2mNameLabel}>
            {renderToString("CreateOAuthClientScreen.name.label")}
          </Text>
          <Text as="p" size="2" className={styles.m2mNameValue}>
            {client.name}
          </Text>
        </div>
        <Text as="p" size="2" className={styles.m2mDescription}>
          <FormattedMessage id="CreateOAuthClientScreen.authorize-resource.description" />
        </Text>
        <div className={styles.searchField}>
          <TextField
            size="2"
            type="search"
            placeholder={renderToString("search")}
            value={searchKeyword}
            iconStart={TextFieldIcon.MagnifyingGlass}
            onChange={onSearchChange}
          />
        </div>
        <div className={styles.resourceListContainer}>
          <ApplicationResourcesList
            className="flex-1"
            resources={resourceListData}
            loading={loading}
            pagination={pagination}
            onToggleAuthorization={handleToggleAuthorization}
            isSearchActive={debouncedSearchKeyword !== ""}
          />
        </div>
        <div className={styles.footer}>
          <SecondaryButton
            size="2"
            text={renderToString("back")}
            onClick={onClickBack}
          />
          <PrimaryButton
            size="2"
            onClick={onClickSave}
            loading={isUpdating}
            disabled={!isDirty}
            text={<FormattedMessage id="CreateOAuthClientScreen.submit" />}
          />
        </div>
      </div>
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
    const capture = useCapture();

    const [clientId] = useState(state.newClient.client_id);
    const client =
      state.clients.find((c) => c.client_id === clientId) ?? state.newClient;

    const onClickSave = useCallback(() => {
      save()
        .then(
          () => {
            capture("createApplication.created", {
              client_id: clientId,
              application_type: "m2m",
              wizard_version: "framework_first",
            });
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
    }, [save, appID, clientId, navigate, capture]);

    const onClickBack = useCallback(() => {
      navigate(`/project/${appID}/configuration/apps/add`);
    }, [appID, navigate]);

    return (
      <ScreenContent className="flex-1-0-auto" layout={"list"}>
        <div className={cn(styles.widget, styles.pageHeader)}>
          <Link
            to={`/project/${appID}/configuration/apps`}
            state={FROM_CREATE_APPLICATION_FLOW_STATE}
            className={styles.backLink}
          >
            <ChevronLeftIcon className={styles.backLinkIcon} />
            <span>
              <FormattedMessage id="ApplicationsConfigurationScreen.title" />
            </span>
          </Link>
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="CreateOAuthClientScreen.title" />
          </Text>
        </div>
        <StepAuthorizeResource
          client={client}
          form={form}
          onClickSave={onClickSave}
          onClickBack={onClickBack}
        />
      </ScreenContent>
    );
  };

const CreateM2MClientScreen: React.VFC = function CreateM2MClientScreen() {
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const location = useLocation();
  const [addResource] = useAddResourceToClientIdMutation();

  // The application name is entered on the New Application screen (/add) and
  // passed here via router state; fall back to a default on direct navigation.
  const initialName = (location.state as { name?: string } | null)?.name ?? "";
  const constructInitialCurrentStateWithName = useCallback(
    (state: FormState): FormState =>
      produce(state, (state) => {
        state.newClient.name = ensureNonEmptyString(initialName) ?? "My App";
      }),
    [initialName]
  );

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
    navigate(`/project/${appID}/configuration/apps`, {
      state: FROM_CREATE_APPLICATION_FLOW_STATE,
    });
  }, [appID, navigate]);

  const onNoResourcesDialogOpenChange = useCallback(
    (open: boolean) => {
      if (!open) {
        goToAppsList();
      }
    },
    [goToAppsList]
  );

  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: null,
    constructFormState,
    constructConfig,
    constructInitialCurrentState: constructInitialCurrentStateWithName,
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
        <div className={styles.scrollArea}>
          <CreateM2MClientContent form={form} />
        </div>
        <ConfirmationDialog
          open={noAPIResources}
          onOpenChange={onNoResourcesDialogOpenChange}
          title={
            <FormattedMessage id="CreateM2MClientScreen.no-resources-dialog.title" />
          }
          description={
            <FormattedMessage id="CreateM2MClientScreen.no-resources-dialog.body" />
          }
          confirmText={
            <FormattedMessage id="CreateM2MClientScreen.no-resources-dialog.cta" />
          }
          cancelText={<FormattedMessage id="cancel" />}
          onConfirm={goToAPIResources}
          onCancel={goToAppsList}
          confirmColor="indigo"
        />
      </FormProvider>
    ),
  });
};

export default CreateM2MClientScreen;
