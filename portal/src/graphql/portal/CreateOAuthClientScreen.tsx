import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupOptionProps,
  Text,
} from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import { produce, createDraft } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
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
import { makeValidationErrorMatchUnknownKindParseRule } from "../../error/parse";
import styles from "./CreateOAuthClientScreen.module.css";
import { FormProvider } from "../../form";
import FormTextField from "../../FormTextField";
import { useTextField } from "../../hook/useInput";
import Widget from "../../Widget";
import ButtonWithLoading from "../../ButtonWithLoading";
import { FormErrorMessageBar } from "../../FormErrorMessageBar";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";

interface FormState {
  clients: OAuthClientConfig[];
  newClient: OAuthClientConfig;
}

function constructFormState(
  config: PortalAPIAppConfig,
  _secretConfig: PortalAPISecretConfig
): FormState {
  return {
    clients: config.oauth?.clients ?? [],
    newClient: {
      name: undefined,
      x_application_type: "spa",
      client_id: genRandomHexadecimalString(),
      redirect_uris: [],
      grant_types: ["authorization_code", "refresh_token"],
      response_types: ["code", "none"],
      access_token_lifetime_seconds: undefined,
      refresh_token_lifetime_seconds: undefined,
      post_logout_redirect_uris: undefined,
      issue_jwt_access_token: true,
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secretConfig: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  return produce([config, secretConfig], ([config, _secretConfig]) => {
    config.oauth ??= {};
    config.oauth.clients = currentState.clients.slice();
    const draft = createDraft(currentState.newClient);
    if (
      draft.x_application_type === "spa" ||
      draft.x_application_type === "traditional_webapp"
    ) {
      draft.redirect_uris = ["http://localhost/after-authentication"];
      draft.post_logout_redirect_uris = ["http://localhost/after-logout"];
    } else if (draft.x_application_type === "native") {
      draft.redirect_uris = ["com.example.myapp://host/path"];
      draft.post_logout_redirect_uris = undefined;
    } else if (
      draft.x_application_type === "confidential" ||
      draft.x_application_type === "third_party_app"
    ) {
      draft.client_name = draft.name;
      draft.redirect_uris = ["http://localhost/after-authentication"];
      draft.post_logout_redirect_uris = undefined;
    }
    config.oauth.clients.push(draft);
    clearEmptyObject(config);
  });
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  _secrets: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  if (
    currentState.newClient.x_application_type === "confidential" ||
    currentState.newClient.x_application_type === "third_party_app"
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

interface CreateOAuthClientContentProps {
  form: AppSecretConfigFormModel<FormState>;
}

const CreateOAuthClientContent: React.VFC<CreateOAuthClientContentProps> =
  function CreateOAuthClientContent(props) {
    const { state, setState, save, isDirty, isUpdating } = props.form;
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
    const client =
      state.clients.find((c) => c.client_id === clientId) ?? state.newClient;

    const onClientConfigChange = useCallback(
      (newClient: OAuthClientConfig) => {
        setState((state) => ({ ...state, newClient }));
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
      ];
    }, [renderToString, onRenderLabel]);

    const onApplicationChange = useCallback(
      (_e, option) => {
        if (option != null) {
          let issueJwtAccessToken: boolean | undefined;
          switch (option.key) {
            case "spa" || "native":
              issueJwtAccessToken = true;
              break;
            default:
              issueJwtAccessToken = undefined;
              break;
          }
          onClientConfigChange(
            updateClientConfig(
              updateClientConfig(client, "x_application_type", option.key),
              "issue_jwt_access_token",
              issueJwtAccessToken
            )
          );
        }
      },
      [onClientConfigChange, client]
    );

    const onClickSave = useCallback(() => {
      save()
        .then(
          () => {
            navigate(
              `/project/${appID}/configuration/apps/${encodeURIComponent(
                clientId
              )}/edit?quickstart=true`,
              {
                replace: true,
              }
            );
          },
          () => {}
        )
        .catch(() => {});
    }, [navigate, appID, clientId, save]);

    const parentJSONPointer = /\/oauth\/clients\/\d+/;

    return (
      <ScreenContent>
        <NavBreadcrumb className={styles.widget} items={navBreadcrumbItems} />
        <Widget className={styles.widget}>
          <FormTextField
            parentJSONPointer={parentJSONPointer}
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
              labelId="save"
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

  const { isLoading, loadError, reload, updateError, isUpdating } = form;

  const errorRules = useMemo(
    () => [
      makeValidationErrorMatchUnknownKindParseRule(
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

  if (isLoading) {
    return <ShowLoading />;
  }

  if (loadError) {
    return <ShowError error={loadError} onRetry={reload} />;
  }

  return (
    <FormProvider loading={isUpdating} error={updateError} rules={errorRules}>
      <FormErrorMessageBar />
      <ScreenLayoutScrollView>
        <CreateOAuthClientContent form={form} />
      </ScreenLayoutScrollView>
    </FormProvider>
  );
};

export default CreateOAuthClientScreen;
