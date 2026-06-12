import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { useNavigate, useParams } from "react-router-dom";
import { produce, createDraft } from "immer";
import { Context, FormattedMessage } from "../../intl";

import ScreenContent from "../../ScreenContent";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
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
import FormTextField from "../../FormTextField";
import { useTextField } from "../../hook/useInput";
import Widget from "../../Widget";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
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
        case "m2m":
          // M2M is handled by CreateM2MClientScreen; not reachable here.
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
    const client = state.newClient;

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

    const framework = state.frameworkId
      ? findFramework(state.frameworkId)
      : undefined;
    const needsStage2 = framework?.stage2 === "token-or-cookie";

    const onSelectFramework = useCallback(
      (id: Framework) => {
        const picked = findFramework(id);
        const defaultStage2: Stage2Choice | null =
          picked?.stage2 === "token-or-cookie" ? "token" : null;
        setState((s) => ({ ...s, frameworkId: id, stage2: defaultStage2 }));
      },
      [setState]
    );

    const onChangeStage2 = useCallback(
      (value: Stage2Choice) => {
        setState((s) => ({ ...s, stage2: value }));
      },
      [setState]
    );

    const canSubmit = useMemo(() => {
      if (!ensureNonEmptyString(client.name ?? "")) return false;
      if (!framework) return false;
      if (needsStage2 && state.stage2 == null) return false;
      return true;
    }, [client.name, framework, needsStage2, state.stage2]);

    const onClickCancel = useCallback(() => {
      navigate(`/project/${appID}/configuration/apps`);
    }, [appID, navigate]);

    const onClickSave = useCallback(() => {
      if (!canSubmit) return;
      save()
        .then(
          () => {
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
    }, [canSubmit, save, appID, clientId, navigate]);

    return (
      <ScreenContent className="flex-1-0-auto" layout={"list"}>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        <Widget className={cn(styles.widget, styles.wizardWidget)}>
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
            autoFocus={true}
          />
          <FrameworkGrid
            selectedId={state.frameworkId}
            onSelect={onSelectFramework}
          />
          {needsStage2 ? (
            <AuthMethodChoiceComponent
              value={state.stage2}
              onChange={onChangeStage2}
              nginxDocsHref={NGINX_DOCS_HREF}
            />
          ) : null}
          <div className={styles.footer}>
            <DefaultButton
              text={renderToString("CreateOAuthClientScreen.cancel")}
              onClick={onClickCancel}
            />
            <ButtonWithLoading
              onClick={onClickSave}
              loading={isUpdating}
              disabled={!canSubmit}
              labelId="CreateOAuthClientScreen.submit"
            />
          </div>
        </Widget>
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
        <div className="flex-1 overflow-y-auto flex flex-col">
          <CreateOAuthClientContent form={form} />
        </div>
      </FormProvider>
    ),
  });
};

export default CreateOAuthClientScreen;
