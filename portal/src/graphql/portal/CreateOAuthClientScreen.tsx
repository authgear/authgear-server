import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import { Heading } from "@radix-ui/themes";
import { ChevronLeftIcon } from "@radix-ui/react-icons";
import { useNavigate, useParams } from "react-router-dom";
import { produce, createDraft } from "immer";
import { Context, FormattedMessage } from "../../intl";

import ScreenContent from "../../ScreenContent";
import Link from "../../Link";
import {
  OAuthClientConfig,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  Framework,
} from "../../types";
import { clearEmptyObject, ensureNonEmptyString } from "../../util/misc";
import { genRandomHexadecimalString } from "../../util/random";
import { makeValidationErrorCustomMessageIDRule } from "../../error/parse";
import styles from "./CreateOAuthClientScreen.module.css";
import { FormProvider } from "../../form";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import { useCapture } from "../../gtm_v2";
import { useLoadableView } from "../../hook/useLoadableView";
import { updateClientConfig } from "./EditOAuthClientForm";

import { FrameworkGrid } from "./CreateOAuthClientScreen/FrameworkGrid";
import { AuthMethodChoiceComponent } from "./CreateOAuthClientScreen/AuthMethodChoice";
import {
  findFramework,
  type AuthMethodChoice as Stage2Choice,
} from "./CreateOAuthClientScreen/frameworks";

const NGINX_DOCS_HREF =
  "https://docs.authgear.com/get-started/backend-api/nginx";

interface FormState {
  clients: OAuthClientConfig[];
  newClient: OAuthClientConfig;
  frameworkId: Framework | null;
  stage2: Stage2Choice | null;
  m2mSelected: boolean;
}

function constructFormState(
  config: PortalAPIAppConfig,
  _secretConfig: PortalAPISecretConfig
): FormState {
  return {
    clients: config.oauth?.clients ?? [],
    newClient: {
      client_id: genRandomHexadecimalString(),
    },
    frameworkId: null,
    stage2: null,
    m2mSelected: false,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secretConfig: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const framework = currentState.frameworkId
    ? findFramework(currentState.frameworkId)
    : undefined;
  if (framework == null) {
    // Before the user picks a framework, the form is not yet dirty.
    // Return the input config unchanged.
    return [config, secretConfig];
  }
  if (framework.stage2 === "token-or-cookie" && currentState.stage2 == null) {
    return [config, secretConfig];
  }
  const xType = framework.resolveType(currentState.stage2 ?? undefined);

  const [newConfig, _] = produce(
    [config, currentState],
    ([config, currentState]) => {
      config.oauth ??= {};
      config.oauth.clients = currentState.clients;
      const draft = createDraft(currentState.newClient);
      draft.x_application_type = xType;
      draft.x_framework = framework.id;
      switch (xType) {
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
  const framework = currentState.frameworkId
    ? findFramework(currentState.frameworkId)
    : undefined;
  if (framework == null) {
    return undefined;
  }
  if (framework.stage2 === "token-or-cookie" && currentState.stage2 == null) {
    return undefined;
  }
  const xType = framework.resolveType(currentState.stage2 ?? undefined);
  const clientTypesWithSecret: OAuthClientConfig["x_application_type"][] = [
    "confidential",
    "third_party_app",
  ];
  if (clientTypesWithSecret.includes(xType)) {
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
  return state;
}

interface CreateOAuthClientContentProps {
  form: AppSecretConfigFormModel<FormState>;
}

const CreateOAuthClientContent: React.VFC<CreateOAuthClientContentProps> =
  function CreateOAuthClientContent(props) {
    const { form } = props;
    const { state, setState, save, isUpdating } = form;
    const { appID } = useParams() as { appID: string };
    const navigate = useNavigate();
    const { renderToString } = useContext(Context);
    const capture = useCapture();
    const viewedRef = useRef(false);
    useEffect(() => {
      if (viewedRef.current) {
        return;
      }
      viewedRef.current = true;
      capture("createApplication.viewed", {
        wizard_version: "framework_first",
      });
    }, [capture]);

    const [clientId] = useState(state.newClient.client_id);
    const client = state.newClient;

    const onClientConfigChange = useCallback(
      (newClient: OAuthClientConfig) => {
        setState((s) => ({ ...s, newClient }));
      },
      [setState]
    );

    const onClientNameChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        onClientConfigChange(
          updateClientConfig(
            client,
            "name",
            ensureNonEmptyString(e.currentTarget.value)
          )
        );
      },
      [onClientConfigChange, client]
    );

    const framework = state.frameworkId
      ? findFramework(state.frameworkId)
      : undefined;
    const needsStage2 = framework?.stage2 === "token-or-cookie";

    const onSelectFramework = useCallback(
      (id: Framework) => {
        const picked = findFramework(id);
        const defaultStage2: Stage2Choice | null =
          picked?.stage2 === "token-or-cookie" ? "token" : null;
        if (picked != null) {
          capture("createApplication.selected-type", {
            application_type: picked.resolveType(defaultStage2 ?? undefined),
            framework_id: picked.id,
            wizard_version: "framework_first",
          });
        }
        setState((s) => ({
          ...s,
          frameworkId: id,
          stage2: defaultStage2,
          m2mSelected: false,
        }));
      },
      [setState, capture]
    );

    const onSelectM2M = useCallback(() => {
      capture("createApplication.selected-type", {
        application_type: "m2m",
        wizard_version: "framework_first",
      });
      setState((s) => ({
        ...s,
        frameworkId: null,
        stage2: null,
        m2mSelected: true,
      }));
    }, [setState, capture]);

    const onChangeStage2 = useCallback(
      (value: Stage2Choice) => {
        setState((s) => ({ ...s, stage2: value }));
      },
      [setState]
    );

    const nameValid = useMemo(
      () => Boolean(ensureNonEmptyString(client.name ?? "")),
      [client.name]
    );

    const canSubmit = useMemo(() => {
      if (!nameValid) return false;
      if (!framework) return false;
      if (needsStage2 && state.stage2 == null) return false;
      return true;
    }, [nameValid, framework, needsStage2, state.stage2]);

    const onClickNext = useCallback(() => {
      navigate(`/project/${appID}/configuration/apps/add-m2m`, {
        state: { name: client.name ?? "" },
      });
    }, [appID, navigate, client.name]);

    const onClickCancel = useCallback(() => {
      navigate(`/project/${appID}/configuration/apps`);
    }, [appID, navigate]);

    const onClickSave = useCallback(() => {
      if (!canSubmit) return;
      save()
        .then(
          () => {
            if (framework != null) {
              capture("createApplication.created", {
                client_id: clientId,
                application_type: framework.resolveType(
                  state.stage2 ?? undefined
                ),
                framework_id: framework.id,
                wizard_version: "framework_first",
              });
            }
            const nextPath = `/project/${appID}/configuration/apps/${encodeURIComponent(
              clientId
            )}/edit`;
            const searchParams = new URLSearchParams();
            searchParams.set("tab", "quick-start");
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
          () => {}
        )
        .catch(() => {});
    }, [
      canSubmit,
      save,
      appID,
      clientId,
      navigate,
      framework,
      state.stage2,
      capture,
    ]);

    const onFormSubmit = useCallback(
      (e: React.FormEvent) => {
        e.preventDefault();
        if (state.m2mSelected) {
          if (nameValid) {
            onClickNext();
          }
        } else {
          // onClickSave no-ops unless canSubmit.
          onClickSave();
        }
      },
      [state.m2mSelected, nameValid, onClickNext, onClickSave]
    );

    return (
      <ScreenContent className="flex-1-0-auto" layout={"list"}>
        <div className={cn(styles.widget, styles.pageHeader)}>
          <Link
            to={`/project/${appID}/configuration/apps`}
            className={styles.backLink}
          >
            <ChevronLeftIcon className={styles.backLinkIcon} />
            <span>
              <FormattedMessage id="ApplicationsConfigurationScreen.title" />
            </span>
          </Link>
          <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="CreateOAuthClientScreen.title" />
          </Heading>
        </div>
        <form
          className={cn(styles.widget, styles.wizardPanel)}
          onSubmit={onFormSubmit}
        >
          <TextField
            size="2"
            parentJSONPointer={/\/oauth\/clients\/\d+/}
            fieldName="name"
            label={renderToString("CreateOAuthClientScreen.name.label")}
            hint={renderToString("CreateOAuthClientScreen.name.description")}
            value={client.name ?? ""}
            onChange={onClientNameChange}
            required={true}
          />
          <FrameworkGrid
            selectedId={state.frameworkId}
            m2mSelected={state.m2mSelected}
            onSelect={onSelectFramework}
            onSelectM2M={onSelectM2M}
          />
          {needsStage2 ? (
            <AuthMethodChoiceComponent
              value={state.stage2}
              onChange={onChangeStage2}
              nginxDocsHref={NGINX_DOCS_HREF}
            />
          ) : null}
          <div className={styles.footer}>
            <SecondaryButton
              size="2"
              text={renderToString("CreateOAuthClientScreen.cancel")}
              onClick={onClickCancel}
            />
            <PrimaryButton
              size="2"
              type="submit"
              loading={isUpdating}
              disabled={state.m2mSelected ? !nameValid : !canSubmit}
              text={
                <FormattedMessage
                  id={
                    state.m2mSelected
                      ? "CreateOAuthClientScreen.next"
                      : "CreateOAuthClientScreen.submit"
                  }
                />
              }
            />
          </div>
        </form>
      </ScreenContent>
    );
  };

const CreateOAuthClientScreen: React.VFC = function CreateOAuthClientScreen() {
  const { appID } = useParams() as { appID: string };

  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: null,
    constructFormState,
    constructConfig,
    constructInitialCurrentState,
    constructSecretUpdateInstruction,
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
          <CreateOAuthClientContent form={form} />
        </div>
      </FormProvider>
    ),
  });
};

export default CreateOAuthClientScreen;
