import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupOptionProps,
  Text,
} from "@fluentui/react";
import { useNavigate, useParams } from "react-router-dom";
import produce, { createDraft } from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { updateClientConfig } from "./EditOAuthClientForm";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject, ensureNonEmptyString } from "../../util/misc";
import { genRandomHexadecimalString } from "../../util/random";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
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
  AuthgearGTMEvent,
  AuthgearGTMEventType,
  useAuthgearGTMEventBase,
  useGTMDispatch,
} from "../../GTMProvider";

interface FormState {
  clients: OAuthClientConfig[];
  newClient: OAuthClientConfig;
}

const errorRules = [
  makeValidationErrorMatchUnknownKindParseRule(
    "general",
    /^\/oauth\/clients$/,
    "error.client-quota-exceeded",
    {
      to: "./../../../billing",
    }
  ),
];

function constructFormState(config: PortalAPIAppConfig): FormState {
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
      issue_jwt_access_token: undefined,
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
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
    }
    config.oauth.clients.push(draft);
    clearEmptyObject(config);
  });
}

function constructInitialCurrentState(state: FormState): FormState {
  return produce(state, (state) => {
    state.newClient.name = "My App";
  });
}

interface CreateOAuthClientContentProps {
  form: AppConfigFormModel<FormState>;
}

const CreateOAuthClientContent: React.VFC<CreateOAuthClientContentProps> =
  function CreateOAuthClientContent(props) {
    const { state, setState, save, isDirty, isUpdating } = props.form;
    const navigate = useNavigate();
    const { renderToString } = useContext(Context);

    const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
      return [
        {
          to: "./..",
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
      return (option?: IChoiceGroupOptionProps) => {
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
      ];
    }, [renderToString, onRenderLabel]);

    const onApplicationChange = useCallback(
      (_e, option) => {
        if (option != null) {
          onClientConfigChange(
            updateClientConfig(client, "x_application_type", option.key)
          );
        }
      },
      [onClientConfigChange, client]
    );

    const gtmEventBase = useAuthgearGTMEventBase();
    const sendDataToGTM = useGTMDispatch();
    const onClickSave = useCallback(() => {
      save()
        .then(
          () => {
            const event: AuthgearGTMEvent = {
              ...gtmEventBase,
              event: AuthgearGTMEventType.CreatedApplication,
              event_data: {
                application_type: client.x_application_type,
              },
            };
            sendDataToGTM(event);
            navigate(
              `./../${encodeURIComponent(clientId)}/edit?quickstart=true`,
              {
                replace: true,
              }
            );
          },
          () => {}
        )
        .catch(() => {});
    }, [
      save,
      navigate,
      clientId,
      sendDataToGTM,
      gtmEventBase,
      client.x_application_type,
    ]);

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
  const form = useAppConfigForm({
    appID,
    constructFormState,
    constructConfig,
    constructInitialCurrentState,
  });

  const { isLoading, loadError, reload, updateError, isUpdating } = form;

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
