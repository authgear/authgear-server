import { useCallback, useEffect, useMemo, useState } from "react";
import { SimpleFormModel, useSimpleForm } from "../../../hook/useSimpleForm";
import { produce } from "immer";
import { parse as parseCSS } from "postcss";
import { useLocation, useNavigate } from "react-router-dom";
import {
  processCompanyName,
  projectIDFromCompanyName,
  randomProjectID,
} from "../../../util/projectname";
import { useCreateAppMutation } from "../../../graphql/portal/mutations/createAppMutation";
import { useOptionalAppContext } from "../../../context/AppContext";
import { useSaveProjectWizardDataMutation } from "../../../graphql/portal/mutations/saveProjectWizardDataMutation";
import {
  LoginIDKeyConfig,
  OAuthSSOProviderConfig,
  PortalAPIAppConfig,
  PortalAPISecretConfigUpdateInstruction,
  PrimaryAuthenticatorType,
} from "../../../types";
import { useAppAndSecretConfigQuery } from "../../../graphql/portal/query/appAndSecretConfigQuery";
import { useUpdateAppAndSecretConfigMutation } from "../../../graphql/portal/mutations/updateAppAndSecretMutation";
import {
  Resource,
  ResourceDefinition,
  ResourceSpecifier,
  expandDef,
  expandSpecifier,
  resolveResource,
  specifierId,
} from "../../../util/resource";
import {
  ResourceFormModel,
  ResourcesFormState,
  useResourceForm,
} from "../../../hook/useResourceForm";
import {
  RESOURCE_APP_LOGO,
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
  RESOURCE_TRANSLATION_JSON,
} from "../../../resources";
import {
  CssAstVisitor,
  CustomisableThemeStyleGroup,
  DEFAULT_LIGHT_THEME,
  PartialCustomisableTheme,
  StyleCssVisitor,
  Theme,
  getThemeTargetSelector,
} from "../../../model/themeAuthFlowV2";
import { TranslationKey } from "../../../model/translations";
import { deriveColors } from "../../../util/theme";
import { ImageValue } from "../../../components/v2/ImageInput/ImageInput";
import { usePortalClient } from "../../../graphql/portal/apollo";
import { useQuery } from "@apollo/client";
import {
  ScreenNavQueryDocument,
  ScreenNavQueryQuery,
  ScreenNavQueryQueryVariables,
} from "../../../graphql/portal/query/screenNavQuery.generated";
import { useCapture } from "../../../gtm_v2";

export enum ProjectWizardStep {
  "step1" = "step1",
  "step2" = "step2",
  "step3" = "step3",
}

export enum LoginMethod {
  Email = "Email",
  Phone = "Phone",
  Username = "Username",
  Google = "Google",
  Apple = "Apple",
  Facebook = "Facebook",
  Github = "Github",
  LinkedIn = "LinkedIn",
  MicrosoftEntraID = "MicrosoftEntraID",
  MicrosoftADFS = "MicrosoftADFS",
  MicrosoftAzureADB2C = "MicrosoftAzureADB2C",
  WechatWeb = "WechatWeb",
  WechatMobile = "WechatMobile",
}

export enum AuthMethod {
  Passwordless = "Passwordless",
  Password = "Password",
}

export interface FormState {
  completed: boolean;
  step: ProjectWizardStep;

  // step 1
  projectName: string;
  projectID: string;

  // step 2
  loginMethods: LoginMethod[];
  authMethods: AuthMethod[];

  // step 3
  logo: ImageValue | null;
  buttonAndLinkColor: string;
  buttonLabelColor: string;
}

function sanitizeFormState(state: FormState): FormState {
  return produce(state, (newState) => {
    newState.projectID = state.projectID.trim();
    newState.projectName = state.projectName.trim();

    const isEmailSelected = state.loginMethods.includes(LoginMethod.Email);
    const isPhoneSelected = state.loginMethods.includes(LoginMethod.Phone);
    const isUsernameSelected = state.loginMethods.includes(
      LoginMethod.Username
    );

    const selectedAuthMethods = new Set(state.authMethods);

    if (!isEmailSelected && !isPhoneSelected) {
      // Passwordless is not allowed if no email or phone
      selectedAuthMethods.delete(AuthMethod.Passwordless);
    }

    if (isUsernameSelected) {
      // Password is required if username is enabled
      selectedAuthMethods.add(AuthMethod.Password);
    }

    newState.authMethods = Array.from(selectedAuthMethods);
    return newState;
  });
}

export interface ProjectWizardFormModel extends SimpleFormModel<FormState> {
  toPreviousStep: () => void;
  canSave: boolean;
  isInitializing: boolean;
  initializeError: unknown;

  effectiveAuthMethods: AuthMethod[];
  isProjectIDEditable: boolean;
}

function makeDefaultState({
  projectID,
  companyName,
}: {
  projectID?: string;
  companyName?: string;
}): FormState {
  const processedCompanyName = companyName
    ? processCompanyName(companyName)
    : null;
  return {
    completed: false,
    step: ProjectWizardStep.step1,

    projectName: companyName ? companyName : "",
    projectID: projectID
      ? projectID
      : processedCompanyName
      ? projectIDFromCompanyName(processedCompanyName)
      : randomProjectID(),

    loginMethods: [LoginMethod.Email],
    authMethods: [AuthMethod.Passwordless],

    logo: null,
    buttonAndLinkColor: "#176DF3",
    buttonLabelColor: "#FFFFFF",
  };
}

interface LocationState {
  company_name: string;
}

function deriveLoginIDKeysFromFormState(
  formState: FormState
): LoginIDKeyConfig[] {
  const keys: LoginIDKeyConfig[] = [];
  for (const method of formState.loginMethods) {
    switch (method) {
      case LoginMethod.Email:
        keys.push({ type: "email" });
        break;
      case LoginMethod.Phone:
        keys.push({ type: "phone" });
        break;
      case LoginMethod.Username:
        keys.push({ type: "username" });
        break;
      case LoginMethod.Google:
        break;
      case LoginMethod.Apple:
        break;
      case LoginMethod.Facebook:
        break;
      case LoginMethod.Github:
        break;
      case LoginMethod.LinkedIn:
        break;
      case LoginMethod.MicrosoftEntraID:
        break;
      case LoginMethod.MicrosoftADFS:
        break;
      case LoginMethod.MicrosoftAzureADB2C:
        break;
      case LoginMethod.WechatWeb:
        break;
      case LoginMethod.WechatMobile:
        break;
    }
  }
  return keys;
}

function deriveOAuthProvidersFromFormState(
  formState: FormState
): OAuthSSOProviderConfig[] {
  const configs: OAuthSSOProviderConfig[] = [];
  for (const method of formState.loginMethods) {
    switch (method) {
      case LoginMethod.Apple:
        configs.push({
          type: "apple",
          alias: "apple",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.Google:
        configs.push({
          type: "google",
          alias: "google",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.Facebook:
        configs.push({
          type: "facebook",
          alias: "facebook",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.Github:
        configs.push({
          type: "github",
          alias: "github",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.LinkedIn:
        configs.push({
          type: "linkedin",
          alias: "linkedin",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.MicrosoftEntraID:
        configs.push({
          type: "azureadv2",
          alias: "azureadv2",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.MicrosoftADFS:
        configs.push({
          type: "adfs",
          alias: "adfs",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.MicrosoftAzureADB2C:
        configs.push({
          type: "azureadb2c",
          alias: "azureadb2c",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.WechatWeb:
        configs.push({
          type: "wechat",
          app_type: "web",
          alias: "wechat_web",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.WechatMobile:
        configs.push({
          type: "wechat",
          app_type: "mobile",
          alias: "wechat_mobile",
          credentials_behavior: "use_demo_credentials",
        });
        break;
      case LoginMethod.Email:
        break;
      case LoginMethod.Phone:
        break;
      case LoginMethod.Username:
        break;
    }
  }
  return configs;
}

function derivePrimaryAuthenticatorsFromFormState(
  formState: FormState
): PrimaryAuthenticatorType[] {
  const authenticators: PrimaryAuthenticatorType[] = [];
  for (const method of formState.authMethods) {
    switch (method) {
      case AuthMethod.Password:
        authenticators.push("password");
        break;
      case AuthMethod.Passwordless:
        if (formState.loginMethods.includes(LoginMethod.Email)) {
          authenticators.push("oob_otp_email");
        }
        if (formState.loginMethods.includes(LoginMethod.Phone)) {
          authenticators.push("oob_otp_sms");
        }
        break;
    }
  }
  return authenticators;
}

function constructConfig(
  config: PortalAPIAppConfig,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.identity ??= {};
    config.identity.login_id ??= {};
    config.identity.login_id.keys =
      deriveLoginIDKeysFromFormState(currentState);
    config.identity.oauth = {
      providers: deriveOAuthProvidersFromFormState(currentState),
    };

    config.authentication ??= {};
    config.authentication.identities = ["oauth", "login_id"];

    config.authentication.primary_authenticators =
      derivePrimaryAuthenticatorsFromFormState(currentState);

    config.ui ??= {};
    config.ui.dark_theme_disabled = true;
    config.ui.light_theme_disabled = undefined;
  });
}

function constructSecretUpdateInstruction(
  newConfig: PortalAPIAppConfig
): PortalAPISecretConfigUpdateInstruction | undefined {
  if (newConfig.identity?.oauth?.providers == null) {
    return undefined;
  }
  return {
    oauthSSOProviderClientSecrets: {
      action: "set",
      data: newConfig.identity.oauth.providers.map((provider) => ({
        newAlias: provider.alias,
        newClientSecret: "",
      })),
    },
  };
}

export function useProjectWizardForm(
  initialState: FormState | null
): ProjectWizardFormModel {
  const capture = useCapture();
  const appContext = useOptionalAppContext();
  const existingAppNodeID = appContext?.appNodeID;
  const navigate = useNavigate();
  const { state } = useLocation();

  const portalClient = usePortalClient();
  const { createApp } = useCreateAppMutation();
  const { mutate: saveProjectWizardDataMutation } =
    useSaveProjectWizardDataMutation();

  const skip = existingAppNodeID == null;
  const {
    loadError: appConfigLoadError,
    rawAppConfig,
    rawAppConfigChecksum,
    refetch: reloadAppConfig,
  } = useAppAndSecretConfigQuery(existingAppNodeID!, null, skip);
  const { updateAppAndSecretConfig: updateAppConfig } =
    useUpdateAppAndSecretConfigMutation(existingAppNodeID!);
  const { refetch: reloadScreenNavQuery } = useQuery<
    ScreenNavQueryQuery,
    ScreenNavQueryQueryVariables
  >(ScreenNavQueryDocument, {
    client: portalClient,
    variables: {
      id: existingAppNodeID!,
    },
    skip: true,
  });

  const [defaultState, setDefaultState] = useState(() => {
    if (initialState != null) {
      return initialState;
    }
    const typedState: LocationState | null = state as LocationState | null;
    const defaultState = makeDefaultState({
      projectID: appContext?.appID,
      companyName: typedState?.company_name,
    });
    return defaultState;
  });

  const [
    resourceForm,
    { setAppName, setAppLogo, setButtonAndLinkColor, setLabelColor },
  ] = useProjectWizardResourceForm(
    existingAppNodeID,
    rawAppConfig?.localization?.fallback_language ?? "en"
  );

  const submit = useCallback(
    async (formState: FormState): Promise<string | null> => {
      const sanitizedFormState = sanitizeFormState(formState);
      if (!computeCanSave(sanitizedFormState)) {
        throw new Error(
          "Cannot navigate to next step, check canNavigateToNextStep"
        );
      }
      const updatedState = produce(sanitizedFormState, (draft) => {
        switch (draft.step) {
          case ProjectWizardStep.step1: {
            draft.step = ProjectWizardStep.step2;
            break;
          }
          case ProjectWizardStep.step2:
            draft.step = ProjectWizardStep.step3;
            break;
          case ProjectWizardStep.step3:
            draft.completed = true;
            break;
        }
        return draft;
      });
      switch (formState.step) {
        case ProjectWizardStep.step1: {
          capture("projectWizard.set-name");
          if (!existingAppNodeID) {
            const appID = await createApp(
              sanitizedFormState.projectID,
              updatedState
            );
            setDefaultState(updatedState);
            return `/project/${encodeURIComponent(appID!)}/wizard`;
            // eslint-disable-next-line no-else-return
          } else {
            await saveProjectWizardDataMutation(
              existingAppNodeID,
              updatedState
            );
            setDefaultState(updatedState);
            return null;
          }
        }
        case ProjectWizardStep.step2:
          capture("projectWizard.set-auth");
          await saveProjectWizardDataMutation(existingAppNodeID!, updatedState);
          setDefaultState(updatedState);
          return null;
        case ProjectWizardStep.step3: {
          capture("projectWizard.set-branding");
          if (rawAppConfig == null) {
            throw new Error("unexpected error: rawAppConfig is null");
          }
          const newConfig = constructConfig(rawAppConfig, updatedState);
          const secretUpdateInstruction =
            constructSecretUpdateInstruction(newConfig);
          await updateAppConfig({
            appConfig: newConfig,
            appConfigChecksum: rawAppConfigChecksum,
            secretConfigUpdateInstructions: secretUpdateInstruction,
            ignoreConflict: true,
          });
          await resourceForm.save();
          // Set it to null to indicate the flow is finished
          await saveProjectWizardDataMutation(existingAppNodeID!, updatedState);
          await reloadAppConfig();
          // Reload the tutorial data, so portal will not redirect user back to wizard again
          await reloadScreenNavQuery();
          setDefaultState(updatedState);

          return `/project/${existingAppNodeID}`;
        }
      }
    },
    [
      capture,
      createApp,
      existingAppNodeID,
      rawAppConfig,
      rawAppConfigChecksum,
      reloadAppConfig,
      reloadScreenNavQuery,
      resourceForm,
      saveProjectWizardDataMutation,
      updateAppConfig,
    ]
  );

  const form = useSimpleForm<FormState>({
    stateMode: "UpdateInitialStateWithUseEffect",
    defaultState: defaultState,
    submit,
  });

  const formState = form.state;
  const nextPath = form.submissionResult;

  useEffect(() => {
    if (nextPath) {
      navigate(nextPath, { replace: true });
    }
  }, [navigate, nextPath]);

  const formStateSanitized = useMemo(
    () => sanitizeFormState(formState),
    [formState]
  );

  // Sync resource form
  useEffect(() => {
    setAppName(formState.projectName);
  }, [formState.projectName, setAppName]);
  useEffect(() => {
    setAppLogo(formState.logo);
  }, [formState.logo, setAppLogo]);

  useEffect(() => {
    setButtonAndLinkColor(formState.buttonAndLinkColor);
  }, [formState.buttonAndLinkColor, setButtonAndLinkColor]);
  useEffect(() => {
    setLabelColor(formState.buttonLabelColor);
  }, [formState.buttonLabelColor, setLabelColor]);

  const toPreviousStep = useCallback(() => {
    capture("projectWizard.clicked-back");
    form.setState((prev) => {
      return produce(prev, (draft) => {
        switch (formState.step) {
          case ProjectWizardStep.step2:
            draft.step = ProjectWizardStep.step1;
            break;
          case ProjectWizardStep.step3:
            draft.step = ProjectWizardStep.step2;
            break;

          default:
            throw new Error("no previous step is available");
        }
        return draft;
      });
    });
  }, [capture, form, formState.step]);

  const canSave = useMemo(
    () => computeCanSave(formStateSanitized),
    [formStateSanitized]
  );

  return useMemo(
    () => ({
      ...form,
      toPreviousStep: toPreviousStep,
      canSave,
      isInitializing: rawAppConfig == null && existingAppNodeID != null,
      initializeError: appConfigLoadError,
      effectiveAuthMethods: formStateSanitized.authMethods,
      isProjectIDEditable: existingAppNodeID == null,
    }),
    [
      form,
      toPreviousStep,
      canSave,
      rawAppConfig,
      appConfigLoadError,
      formStateSanitized.authMethods,
      existingAppNodeID,
    ]
  );
}

const LOCALE_BASED_RESOUCE_DEFINITIONS = [
  RESOURCE_TRANSLATION_JSON,
  RESOURCE_APP_LOGO,
];

const THEME_RESOURCE_DEFINITIONS = [
  RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
];

const LightThemeResourceSpecifier = {
  def: RESOURCE_AUTHGEAR_AUTHFLOW_V2_LIGHT_THEME_CSS,
  locale: null,
  extension: null,
};

function getLightThemeFromResourceFormState(
  state: ResourcesFormState
): PartialCustomisableTheme {
  const themeResource =
    state.resources[specifierId(LightThemeResourceSpecifier)];
  if (themeResource?.nullableValue == null) {
    return DEFAULT_LIGHT_THEME;
  }
  const root = parseCSS(themeResource.nullableValue);
  const styleCSSVisitor = new StyleCssVisitor(
    getThemeTargetSelector(Theme.Light),
    new CustomisableThemeStyleGroup()
  );
  return styleCSSVisitor.getStyle(root);
}

function useProjectWizardResourceForm(
  appNodeID: string | undefined,
  locale: string
): [
  ResourceFormModel<ResourcesFormState>,
  {
    setAppName: (appName: string) => void;
    setAppLogo: (logo: ImageValue | null) => void;
    setButtonAndLinkColor: (color: string) => void;
    setLabelColor: (color: string) => void;
  }
] {
  const specifiers = useMemo<ResourceSpecifier[]>(() => {
    const specifiers: ResourceSpecifier[] = [];
    for (const def of THEME_RESOURCE_DEFINITIONS) {
      specifiers.push({
        def,
        locale: null,
        extension: null,
      });
    }
    for (const def of LOCALE_BASED_RESOUCE_DEFINITIONS) {
      specifiers.push(...expandDef(def, locale));
    }
    return specifiers;
  }, [locale]);

  const resourceForm = useResourceForm(appNodeID, specifiers);
  const resourceFormSetState = resourceForm.setState;

  const resourceMutator = useMemo(() => {
    return {
      setTranslationValue: (key: string, value: string) => {
        resourceFormSetState((s) => {
          return produce(s, (draft) => {
            const specifier: ResourceSpecifier = {
              def: RESOURCE_TRANSLATION_JSON,
              locale: locale,
              extension: null,
            };
            const translationResource = resolveResource(s.resources, [
              specifier,
            ]);
            if (!translationResource?.nullableValue) {
              return;
            }
            const jsonValue = JSON.parse(translationResource.nullableValue);
            if (value === "") {
              delete jsonValue[key];
            } else {
              jsonValue[key] = value;
            }
            draft.resources[specifierId(specifier)] = {
              specifier: specifier,
              path: expandSpecifier(specifier),
              nullableValue: JSON.stringify(jsonValue, null, 2),
            };
          });
        });
      },
      setImage: (
        def: ResourceDefinition,
        image: {
          base64EncodedData: string;
          extension: string;
        } | null
      ) => {
        resourceFormSetState((prev) => {
          return produce(prev, (draft) => {
            const specifiers = expandDef(def, locale);
            for (const specifier of specifiers) {
              const resource = draft.resources[specifierId(specifier)];
              if (resource != null) {
                resource.nullableValue = "";
              }
            }
            if (image == null) {
              return;
            }
            const specifier = {
              def,
              extension: image.extension,
              locale: locale,
            };
            const resource: Resource = {
              specifier,
              path: expandSpecifier(specifier),
              nullableValue: image.base64EncodedData,
            };
            draft.resources[specifierId(specifier)] = resource;
          });
        });
      },
      updateCustomisableTheme: (
        updater: (prev: PartialCustomisableTheme) => PartialCustomisableTheme
      ) => {
        resourceFormSetState((s) => {
          const newState = updater(getLightThemeFromResourceFormState(s));
          return produce(s, (draft) => {
            const resourceSpecifier = LightThemeResourceSpecifier;
            const themeResource = draft.resources[
              specifierId(resourceSpecifier)
            ] ?? {
              specifier: resourceSpecifier,
              path: expandSpecifier(resourceSpecifier),
            };

            themeResource.nullableValue = (() => {
              const cssAstVisitor = new CssAstVisitor(
                getThemeTargetSelector(Theme.Light)
              );
              const styleGroup = new CustomisableThemeStyleGroup(newState);
              styleGroup.acceptCssAstVisitor(cssAstVisitor);
              if (cssAstVisitor.getDeclarations().length <= 0) {
                return "";
              }
              return cssAstVisitor.getCSS().toResult().css;
            })();

            draft.resources[specifierId(resourceSpecifier)] = themeResource;
          });
        });
      },
    };
  }, [resourceFormSetState, locale]);

  const setAppName = useCallback(
    (appName: string) => {
      resourceMutator.setTranslationValue(TranslationKey.AppName, appName);
    },
    [resourceMutator]
  );

  const setAppLogo = useCallback(
    (imageValue: ImageValue | null) => {
      resourceMutator.setImage(RESOURCE_APP_LOGO, imageValue);
    },
    [resourceMutator]
  );

  const setButtonAndLinkColor = useCallback(
    (color: string) => {
      const derivedColors = deriveColors(color);
      resourceMutator.updateCustomisableTheme((prev) => {
        return produce(prev, (draft) => {
          if (derivedColors != null) {
            draft.primaryButton.backgroundColor = color;
            draft.primaryButton.backgroundColorActive = derivedColors.variant;
            draft.primaryButton.backgroundColorHover = derivedColors.variant;

            draft.icon.color = color;

            draft.link.color = color;
            draft.link.colorActive = derivedColors.variant;
            draft.link.colorHover = derivedColors.variant;
          }
          return draft;
        });
      });
    },
    [resourceMutator]
  );

  const setLabelColor = useCallback(
    (color: string) => {
      resourceMutator.updateCustomisableTheme((prev) => {
        return produce(prev, (draft) => {
          draft.primaryButton.labelColor = color;
          return draft;
        });
      });
    },
    [resourceMutator]
  );

  return [
    resourceForm,
    { setAppName, setAppLogo, setButtonAndLinkColor, setLabelColor },
  ];
}

function computeCanSave(formStateSanitized: FormState): boolean {
  switch (formStateSanitized.step) {
    case ProjectWizardStep.step1:
      return (
        formStateSanitized.projectID.trim() !== "" &&
        formStateSanitized.projectName.trim() !== ""
      );
    case ProjectWizardStep.step2: {
      if (formStateSanitized.loginMethods.length === 0) {
        return false;
      }
      const loginMethods = new Set(formStateSanitized.loginMethods);
      if (
        formStateSanitized.authMethods.length === 0 &&
        [LoginMethod.Email, LoginMethod.Phone, LoginMethod.Username].some(
          (method) => loginMethods.has(method)
        )
      ) {
        return false;
      }
      return true;
    }
    case ProjectWizardStep.step3:
      return true;
  }
}
