import {
  Checkbox as RadixCheckbox,
  DropdownMenu,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import {
  ChevronDownIcon,
  DotsVerticalIcon,
  InfoCircledIcon,
  ListBulletIcon,
  Pencil1Icon,
  PlusIcon,
  StarFilledIcon,
  TrashIcon,
} from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import cn from "classnames";
import { produce } from "immer";
import React, { useCallback, useContext, useMemo, useState } from "react";
import { TextField } from "../../components/v2/TextField/TextField";
import { TextArea } from "../../components/v2/TextArea/TextArea";
import {
  IconRadioCards,
  IconRadioCardOption,
} from "../../components/v2/IconRadioCards/IconRadioCards";
import { SquareIcon } from "../../components/v2/SquareIcon/SquareIcon";
import { FormField } from "../../components/v2/FormField/FormField";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import {
  createOAuthSSOProviderItemKey,
  isOAuthSSOProvider,
  OAuthSSOFeatureConfig,
  OAuthSSOProviderConfig,
  OAuthSSOProviderFeatureConfig,
  OAuthSSOProviderItemKey,
  OAuthSSOProviderType,
  OAuthSSOWeChatAppType,
  parseOAuthSSOProviderItemKey,
  SSOProviderFormSecretViewModel,
} from "../../types";
import ExternalLink from "../../ExternalLink";

import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import styles from "./SingleSignOnConfigurationWidget.module.css";
import {
  OAuthProviderFormModel,
  SSOProviderFormState,
} from "../../hook/useOAuthProviderForm";
import { Badge } from "../../components/v2/Badge/Badge";
import { Callout } from "../../components/v2/Callout/Callout";
import { isOAuthProviderMissingCredential } from "../../model/oauthProviders";
import { EffectiveSecretConfig } from "./globalTypes.generated";

const MASKED_SECRET = "***************";

type CredentialStatus = "active" | "missing_credential" | "demo";

interface SingleSignOnConfigurationWidgetProps {
  className?: string;

  jsonPointer: string;
  clientSecretParentJsonPointer: RegExp;

  isDemoCredentialAvailable: boolean;
  config: OAuthSSOProviderConfig;
  secret: SSOProviderFormSecretViewModel;
  onChange: (
    config: OAuthSSOProviderConfig,
    secret: SSOProviderFormSecretViewModel
  ) => void;

  disabled: boolean;
  publicOrigin: string;
}

type WidgetTextFieldKey =
  | keyof Omit<OAuthSSOProviderConfig, "type" | "claims">
  | "client_secret"
  | "email_required";

interface OAuthProviderInfo {
  providerType: OAuthSSOProviderType;
  iconClassName: string;
  fields: Set<WidgetTextFieldKey>;
  isSecretFieldTextArea: boolean;
  appType?: OAuthSSOWeChatAppType;
  titleId: string;
  subtitleId?: string;
  descriptionId: string;
  inactiveMessageId: string;
  docUrl: string;
  redirectUrlLabelId: string | null;
}

const oauthProviders: Record<OAuthSSOProviderItemKey, OAuthProviderInfo> = {
  apple: {
    providerType: "apple",
    iconClassName: "fa-apple",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "key_id",
      "team_id",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: true,
    titleId: "AddSingleSignOnConfigurationScreen.card.apple.title",
    descriptionId: "AddSingleSignOnConfigurationScreen.card.apple.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.apple.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/social-login-providers/apple",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.apple",
  },
  google: {
    providerType: "google",
    iconClassName: "fa-google",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.google.title",
    descriptionId: "AddSingleSignOnConfigurationScreen.card.google.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.google.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/social-login-providers/google",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.google",
  },
  facebook: {
    providerType: "facebook",
    iconClassName: "fa-facebook",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.facebook.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.facebook.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.facebook.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/social-login-providers/facebook",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.facebook",
  },
  github: {
    providerType: "github",
    iconClassName: "fa-github",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.github.title",
    descriptionId: "AddSingleSignOnConfigurationScreen.card.github.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.github.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/social-login-providers/github",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.github",
  },
  linkedin: {
    providerType: "linkedin",
    iconClassName: "fa-linkedin",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.linkedin.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.linkedin.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.linkedin.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/social-login-providers/linkedin",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.linkedin",
  },
  azureadv2: {
    providerType: "azureadv2",
    iconClassName: "fa-microsoft",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "tenant",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.azureadv2.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.azureadv2.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.azureadv2.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/enterprise-login-providers/azureadv2",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.azure",
  },
  azureadb2c: {
    providerType: "azureadb2c",
    iconClassName: "fa-microsoft",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "tenant",
      "policy",
      "domain_hint",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.azureadb2c.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.azureadb2c.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.azureadb2c.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/enterprise-login-providers/azureadb2c",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.azure",
  },
  adfs: {
    providerType: "adfs",
    iconClassName: "fa-microsoft",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "discovery_document_endpoint",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.adfs.title",
    descriptionId: "AddSingleSignOnConfigurationScreen.card.adfs.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.adfs.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/enterprise-login-providers/adfs",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.azure",
  },
  "wechat.web": {
    providerType: "wechat",
    appType: "web",
    iconClassName: "fa-weixin",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "account_id",
      "is_sandbox_account",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.wechat.web.title",
    subtitleId: "AddSingleSignOnConfigurationScreen.card.wechat.web.subtitle",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.wechat.web.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.wechat.web.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/social-login-providers/wechat-web",
    redirectUrlLabelId:
      "SingleSignOnConfigurationWidget.redirectUrl.label.wechat",
  },
  "wechat.mobile": {
    providerType: "wechat",
    appType: "mobile",
    iconClassName: "fa-weixin",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "account_id",
      "wechat_redirect_uris",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.wechat.mobile.title",
    subtitleId:
      "AddSingleSignOnConfigurationScreen.card.wechat.mobile.subtitle",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.wechat.mobile.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.wechat.mobile.inactiveMessage",
    docUrl:
      "https://docs.authgear.com/authentication-and-access/social-enterprise-login-providers/social-login-providers/wechat-mobile",
    redirectUrlLabelId: null,
  },
};

interface OAuthClientIconProps {
  className?: string;
  providerItemKey: OAuthSSOProviderItemKey;
}

const OAuthClientIcon: React.VFC<OAuthClientIconProps> =
  function OAuthClientIcon(props) {
    const { providerItemKey } = props;
    const { iconClassName } = oauthProviders[providerItemKey];
    return <i className={cn("fab", iconClassName, styles.widgetLabelIcon)} />;
  };

function ProviderStatus({
  providerConfig,
  providersWithDemoCredentials,
}: {
  providerConfig: OAuthSSOProviderConfig;
  providersWithDemoCredentials: Set<string>;
}) {
  if (providerConfig.credentials_behavior === "use_demo_credentials") {
    if (providersWithDemoCredentials.has(providerConfig.type)) {
      return (
        <Badge
          size="1"
          variant="warning"
          text={
            <FormattedMessage id="SingleSignOnConfigurationScreen.providerStatus.demo" />
          }
        />
      );
    }
    return (
      <Badge
        size="1"
        variant="error"
        text={
          <FormattedMessage id="SingleSignOnConfigurationScreen.providerStatus.inactive" />
        }
      />
    );
  }
  return (
    <Badge
      size="1"
      variant="success"
      text={
        <FormattedMessage id="SingleSignOnConfigurationScreen.providerStatus.active" />
      }
    />
  );
}

export function useSingleSignOnConfigurationWidget(
  initialAlias: string,
  providerItemKey: OAuthSSOProviderItemKey,
  form: OAuthProviderFormModel,
  effectiveSecretConfig: EffectiveSecretConfig | undefined,
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig
): Omit<SingleSignOnConfigurationWidgetProps, "className" | "publicOrigin"> {
  const {
    state: { providers },
    setState,
  } = form;

  const [providerType, appType] = parseOAuthSSOProviderItemKey(providerItemKey);

  const [providerIndex] = useState<number>(() => {
    const existingIndex = providers.findIndex((p) =>
      isOAuthSSOProvider(p.config, providerType, initialAlias, appType)
    );
    if (existingIndex !== -1) {
      return existingIndex;
    }
    // Insert at the end if it does not exist
    return providers.length;
  });

  const disabled = useMemo(() => {
    const providersConfig = oauthSSOFeatureConfig?.providers ?? {};
    const providerConfig = providersConfig[
      providerType
    ] as OAuthSSOProviderFeatureConfig | null;
    return providerConfig?.disabled ?? false;
  }, [oauthSSOFeatureConfig?.providers, providerType]);

  const provider = useMemo<SSOProviderFormState>(() => {
    const newConfig = {
      config: {
        type: providerType,
        alias: initialAlias,
        ...(appType && { app_type: appType }),
      },
      secret: {
        originalAlias: null,
        newAlias: initialAlias,
        newClientSecret: "",
      },
    } satisfies SSOProviderFormState;
    return providers.length > providerIndex
      ? providers[providerIndex]
      : newConfig;
  }, [providerType, initialAlias, appType, providers, providerIndex]);

  const jsonPointer = useMemo(() => {
    return `/identity/oauth/providers/${providerIndex}`;
  }, [providerIndex]);
  const clientSecretParentJsonPointer = new RegExp(
    `/secrets/\\d+/data/items/${providerIndex}`
  );

  const providersWithDemoCredentials = useMemo<Set<string>>(() => {
    return new Set(
      effectiveSecretConfig?.oauthSSOProviderDemoSecrets?.map((it) => it.type)
    );
  }, [effectiveSecretConfig]);
  const isDemoCredentialAvailable = providersWithDemoCredentials.has(
    provider.config.type
  );

  const onChange = useCallback(
    (
      newConfig: OAuthSSOProviderConfig,
      secret: SSOProviderFormSecretViewModel
    ) =>
      setState((state) =>
        produce(state, (state) => {
          const config = produce(newConfig, (config) => {
            if (isDemoCredentialAvailable) {
              // If demo credential is avaiable, the user have to choose between demo credential and custom credential
              return;
            }
            // Else, set it automatically
            config.credentials_behavior = isOAuthProviderMissingCredential(
              config,
              secret
            )
              ? "use_demo_credentials"
              : "use_project_credentials";
          });

          if (providerIndex === -1) {
            state.providers.push({
              config,
              secret: {
                originalAlias: null,
                newAlias: secret.newAlias,
                newClientSecret: secret.newClientSecret,
              },
            });
          } else {
            state.providers[providerIndex] = {
              config,
              secret: {
                originalAlias: secret.originalAlias,
                newAlias: secret.newAlias,
                newClientSecret: secret.newClientSecret,
              },
            };
          }
        })
      ),
    [setState, providerIndex, isDemoCredentialAvailable]
  );

  return {
    jsonPointer: jsonPointer,
    clientSecretParentJsonPointer: clientSecretParentJsonPointer,
    isDemoCredentialAvailable: isDemoCredentialAvailable,
    config: provider.config,
    secret: provider.secret,
    onChange: onChange,
    disabled: disabled,
  };
}

// If we do not do this, then some optional config, like domain_hint, when being clear,
// is domain_hint="".
// The JSON schema rejects empty string.
// So when it is an empty string, it should be set to undefined instead.
function emptyStringToUndefined(value: string | undefined): string | undefined {
  if (value == null || value === "") {
    return undefined;
  }
  return value;
}

function FieldLabelWithTooltip(props: {
  labelId: string;
  tooltipId: string;
}): React.ReactElement {
  const { labelId, tooltipId } = props;
  return (
    <span className={styles.tooltipLabel}>
      <FormattedMessage id={labelId} />
      <Tooltip content={<FormattedMessage id={tooltipId} />}>
        <InfoCircledIcon className={styles.infoIcon} />
      </Tooltip>
    </span>
  );
}

function WidgetCheckbox(props: {
  label: React.ReactNode;
  checked: boolean;
  onCheckedChange: (checked: boolean | "indeterminate") => void;
  disabled?: boolean;
}): React.ReactElement {
  const { label, checked, onCheckedChange, disabled } = props;
  return (
    <label className={styles.checkboxRow}>
      <RadixCheckbox
        checked={checked}
        onCheckedChange={onCheckedChange}
        disabled={disabled}
      />
      <Text size="2">{label}</Text>
    </label>
  );
}

type DemoCredentialOptionValue = "custom" | "demo";

const SingleSignOnConfigurationWidget: React.VFC<SingleSignOnConfigurationWidgetProps> =
  function SingleSignOnConfigurationWidget(
    props: SingleSignOnConfigurationWidgetProps
  ) {
    const {
      className,
      jsonPointer,
      clientSecretParentJsonPointer,
      isDemoCredentialAvailable,
      config,
      secret,
      onChange,
      disabled: featureDisabled,
      publicOrigin,
    } = props;
    const isMissingCredential =
      config.credentials_behavior === "use_demo_credentials";

    const { renderToString } = useContext(Context);

    const providerItemKey = createOAuthSSOProviderItemKey(
      config.type,
      config.app_type
    );

    const {
      isSecretFieldTextArea,
      fields: visibleFields,
      inactiveMessageId,
      docUrl,
      redirectUrlLabelId,
    } = oauthProviders[providerItemKey];
    const [advancedFolded, setAdvancedFolded] = useState(true);
    const redirectURL = useMemo(() => {
      if (!publicOrigin || !config.alias) {
        return "";
      }
      const normalizedOrigin = publicOrigin.replace(/\/+$/, "");
      return `${normalizedOrigin}/sso/oauth2/callback/${encodeURIComponent(
        config.alias
      )}`;
    }, [publicOrigin, config.alias]);

    const messageID = "OAuthBranding." + providerItemKey;

    const onAliasChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        onChange({ ...config, alias: value }, { ...secret, newAlias: value });
      },
      [onChange, config, secret]
    );

    const onClientIDChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange(
          { ...config, client_id: emptyStringToUndefined(e.target.value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onTenantChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange(
          { ...config, tenant: emptyStringToUndefined(e.target.value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onPolicyChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange(
          { ...config, policy: emptyStringToUndefined(e.target.value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onDomainHintChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange(
          { ...config, domain_hint: emptyStringToUndefined(e.target.value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onDiscoveryDocumentEndpointChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange(
          {
            ...config,
            discovery_document_endpoint: emptyStringToUndefined(e.target.value),
          },
          secret
        ),
      [onChange, config, secret]
    );
    const onKeyIDChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange(
          { ...config, key_id: emptyStringToUndefined(e.target.value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onTeamIDChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange(
          { ...config, team_id: emptyStringToUndefined(e.target.value) },
          secret
        ),
      [onChange, config, secret]
    );

    const onClientSecretChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) =>
        onChange(config, { ...secret, newClientSecret: e.target.value }),
      [onChange, config, secret]
    );
    const onAccountIDChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) =>
        onChange(
          { ...config, account_id: emptyStringToUndefined(e.target.value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onIsSandBoxAccountChange = useCallback(
      (checked: boolean | "indeterminate") =>
        onChange({ ...config, is_sandbox_account: checked === true }, secret),
      [onChange, config, secret]
    );
    const onWeChatRedirectUrisChange = useCallback(
      (list: string[]) =>
        onChange(
          { ...config, wechat_redirect_uris: list.length > 0 ? list : [] },
          secret
        ),
      [onChange, config, secret]
    );
    const onCreateDisabledChange = useCallback(
      (checked: boolean | "indeterminate") =>
        onChange({ ...config, create_disabled: checked === true }, secret),
      [onChange, config, secret]
    );
    const onDeleteDisabledChange = useCallback(
      (checked: boolean | "indeterminate") =>
        onChange({ ...config, delete_disabled: checked === true }, secret),
      [onChange, config, secret]
    );
    const onEmailRequiredChange = useCallback(
      (checked: boolean | "indeterminate") => {
        const value = checked === true;
        const newConfig = produce(config, (config) => {
          config.claims ??= {};
          config.claims.email ??= {};
          if (!value) {
            config.claims.email.required = false;
          } else {
            delete config.claims.email.required;
          }
        });
        onChange(newConfig, secret);
      },
      [onChange, config, secret]
    );

    const wechatRedirectUris = useMemo(
      () => config.wechat_redirect_uris ?? [],
      [config.wechat_redirect_uris]
    );
    const onWechatUriItemChange = useCallback(
      (index: number, value: string) => {
        const newList = wechatRedirectUris.slice();
        newList[index] = value;
        onWeChatRedirectUrisChange(newList);
      },
      [wechatRedirectUris, onWeChatRedirectUrisChange]
    );
    const onWechatUriItemAdd = useCallback(() => {
      const newList = wechatRedirectUris.slice();
      newList.push("");
      onWeChatRedirectUrisChange(newList);
    }, [wechatRedirectUris, onWeChatRedirectUrisChange]);
    const onWechatUriItemDelete = useCallback(
      (index: number) => {
        const newList = wechatRedirectUris.slice();
        newList.splice(index, 1);
        onWeChatRedirectUrisChange(newList);
      },
      [wechatRedirectUris, onWeChatRedirectUrisChange]
    );

    const noneditable = featureDisabled;

    const credentialStatus = useMemo<CredentialStatus>(() => {
      if (isMissingCredential && !isDemoCredentialAvailable) {
        return "missing_credential";
      }
      if (isMissingCredential && isDemoCredentialAvailable) {
        return "demo";
      }
      return "active";
    }, [isDemoCredentialAvailable, isMissingCredential]);

    const isDemoCredentialSelected = credentialStatus === "demo";

    const handleDemoCredentialSelectedChange = useCallback(
      (value: boolean) => {
        const newConfig = produce(config, (config) => {
          config.credentials_behavior = value
            ? "use_demo_credentials"
            : "use_project_credentials";
        });
        onChange(newConfig, secret);
      },
      [config, onChange, secret]
    );

    const onDemoCredentialValueChange = useCallback(
      (value: DemoCredentialOptionValue) => {
        handleDemoCredentialSelectedChange(value === "demo");
      },
      [handleDemoCredentialSelectedChange]
    );

    const demoCredentialOptions = useMemo<
      IconRadioCardOption<DemoCredentialOptionValue>[]
    >(
      () => [
        {
          value: "custom",
          icon: (
            <SquareIcon
              className="text-[var(--accent-9)]"
              Icon={ListBulletIcon}
              size="7"
              radius="4"
              iconSize="1.375rem"
            />
          ),
          title: renderToString(
            "SingleSignOnConfigurationWidget.credentialStatusButton.custom.text"
          ),
          subtitle: renderToString(
            "SingleSignOnConfigurationWidget.credentialStatusButton.custom.secondaryText"
          ),
        },
        {
          value: "demo",
          icon: (
            <SquareIcon
              className="text-[var(--accent-9)]"
              Icon={StarFilledIcon}
              size="7"
              radius="4"
              iconSize="1.375rem"
            />
          ),
          title: renderToString(
            "SingleSignOnConfigurationWidget.credentialStatusButton.demo.text"
          ),
          subtitle: renderToString(
            "SingleSignOnConfigurationWidget.credentialStatusButton.demo.secondaryText"
          ),
        },
      ],
      [renderToString]
    );

    const onToggleAdvancedFolded = useCallback(() => {
      setAdvancedFolded((folded) => !folded);
    }, []);

    return (
      <SettingsSectionCard
        className={className}
        contentClassName={styles.cardContent}
        title={
          <FormattedMessage id="SingleSignOnConfigurationWidget.settings.label" />
        }
      >
        <div className={styles.contentHeader}>
          <div className={styles.widgetHeader}>
            <span className={styles.widgetHeaderIcon}>
              <OAuthClientIcon providerItemKey={providerItemKey} />
            </span>
            <Text
              as="p"
              size="3"
              weight="medium"
              className={styles.contentTitle}
            >
              {renderToString(messageID)}
            </Text>
          </div>
          <Text
            as="p"
            size="2"
            color="gray"
            className={styles.contentDescription}
          >
            <FormattedMessage
              id="SingleSignOnConfigurationWidget.setupGuide"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                docLink: (chunks: React.ReactNode) => (
                  <ExternalLink href={docUrl}>{chunks}</ExternalLink>
                ),
              }}
            />
          </Text>
        </div>
        <Text as="p" size="2" weight="medium" className={styles.sectionTitle}>
          <FormattedMessage id="SingleSignOnConfigurationWidget.credentials.label" />
        </Text>
        {featureDisabled ? (
          <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
        ) : null}
        {isDemoCredentialAvailable ? (
          <IconRadioCards
            size="2"
            value={
              isDemoCredentialSelected ? ("demo" as const) : ("custom" as const)
            }
            onValueChange={onDemoCredentialValueChange}
            options={demoCredentialOptions}
            numberOfColumns={2}
            itemFillSpaces={true}
          />
        ) : null}
        {credentialStatus === "missing_credential" ? (
          <Callout
            className="w-full"
            type="error"
            text={<FormattedMessage id={inactiveMessageId} />}
            showCloseButton={false}
          />
        ) : credentialStatus === "demo" ? (
          <Callout
            className="w-full"
            type="warning"
            text={
              <FormattedMessage id="SingleSignOnConfigurationWidget.hint.usingDemoCredential" />
            }
            showCloseButton={false}
          />
        ) : null}
        {credentialStatus !== "demo" ? (
          <>
            {redirectURL && redirectUrlLabelId ? (
              <TextField
                size="2"
                labelSize="2"
                label={renderToString(redirectUrlLabelId)}
                value={redirectURL}
                readOnly={true}
                suffix={<CopyIconButton textToCopy={redirectURL} />}
                suffixPlain={true}
              />
            ) : null}
            {visibleFields.has("client_id") ? (
              <TextField
                size="2"
                labelSize="2"
                parentJSONPointer={jsonPointer}
                fieldName="client_id"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.client-id"
                )}
                value={config.client_id ?? ""}
                onChange={onClientIDChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("client_secret") ? (
              isSecretFieldTextArea ? (
                <TextArea
                  size="2"
                  labelSize="2"
                  className={styles.clientSecretTextArea}
                  parentJSONPointer={clientSecretParentJsonPointer}
                  fieldName="client_secret"
                  label={renderToString(
                    "SingleSignOnConfigurationScreen.widget.client-secret"
                  )}
                  value={
                    noneditable || secret.newClientSecret == null
                      ? MASKED_SECRET
                      : secret.newClientSecret
                  }
                  onChange={onClientSecretChange}
                  disabled={noneditable || secret.newClientSecret == null}
                />
              ) : (
                <TextField
                  size="2"
                  labelSize="2"
                  parentJSONPointer={clientSecretParentJsonPointer}
                  fieldName="client_secret"
                  label={renderToString(
                    "SingleSignOnConfigurationScreen.widget.client-secret"
                  )}
                  value={
                    noneditable || secret.newClientSecret == null
                      ? MASKED_SECRET
                      : secret.newClientSecret
                  }
                  onChange={onClientSecretChange}
                  disabled={noneditable || secret.newClientSecret == null}
                />
              )
            ) : null}
            {visibleFields.has("tenant") ? (
              <TextField
                size="2"
                labelSize="2"
                parentJSONPointer={jsonPointer}
                fieldName="tenant"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.tenant"
                )}
                value={config.tenant ?? ""}
                onChange={onTenantChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("policy") ? (
              <TextField
                size="2"
                labelSize="2"
                parentJSONPointer={jsonPointer}
                fieldName="policy"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.policy"
                )}
                value={config.policy ?? ""}
                placeholder={renderToString(
                  "SingleSignOnConfigurationScreen.widget.policy.placeholder"
                )}
                onChange={onPolicyChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("domain_hint") ? (
              <TextField
                size="2"
                labelSize="2"
                parentJSONPointer={jsonPointer}
                fieldName="domain_hint"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.domain-hint"
                )}
                placeholder={renderToString(
                  "SingleSignOnConfigurationScreen.widget.domain-hint.placeholder"
                )}
                hint={
                  <FormattedMessage
                    id="SingleSignOnConfigurationScreen.widget.domain-hint.description"
                    values={{
                      // eslint-disable-next-line react/no-unstable-nested-components
                      externalLink: (chunks: React.ReactNode) => (
                        <ExternalLink
                          href="https://docs.microsoft.com/en-us/azure/active-directory-b2c/direct-signin?pivots=b2c-user-flow#redirect-sign-in-to-a-social-provider"
                          target="_blank"
                          rel="noreferrer"
                        >
                          {chunks}
                        </ExternalLink>
                      ),
                    }}
                  />
                }
                value={config.domain_hint ?? ""}
                onChange={onDomainHintChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("discovery_document_endpoint") ? (
              <TextField
                size="2"
                labelSize="2"
                parentJSONPointer={jsonPointer}
                fieldName="discovery_document_endpoint"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.discovery-document-endpoint"
                )}
                value={config.discovery_document_endpoint ?? ""}
                onChange={onDiscoveryDocumentEndpointChange}
                placeholder="http://example.com/.well-known/openid-configuration"
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("key_id") ? (
              <TextField
                size="2"
                labelSize="2"
                parentJSONPointer={jsonPointer}
                fieldName="key_id"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.key-id"
                )}
                value={config.key_id ?? ""}
                onChange={onKeyIDChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("team_id") ? (
              <TextField
                size="2"
                labelSize="2"
                parentJSONPointer={jsonPointer}
                fieldName="team_id"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.team-id"
                )}
                value={config.team_id ?? ""}
                onChange={onTeamIDChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("account_id") ? (
              <TextField
                size="2"
                labelSize="2"
                parentJSONPointer={jsonPointer}
                fieldName="account_id"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.account-id"
                )}
                value={config.account_id ?? ""}
                onChange={onAccountIDChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("is_sandbox_account") ? (
              <WidgetCheckbox
                label={
                  <FormattedMessage id="SingleSignOnConfigurationScreen.widget.is-sandbox-account" />
                }
                checked={config.is_sandbox_account ?? false}
                onCheckedChange={onIsSandBoxAccountChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("wechat_redirect_uris") ? (
              <FormField
                size="2"
                labelSize="2"
                labelSpace="1"
                label={
                  <FieldLabelWithTooltip
                    labelId="SingleSignOnConfigurationScreen.widget.wechat-redirect-uris-label"
                    tooltipId="SingleSignOnConfigurationScreen.widget.wechat-redirect-uris-tooltip-message"
                  />
                }
              >
                <div className={styles.uriList}>
                  {wechatRedirectUris.map((uri, index) => (
                    <div key={index} className={styles.uriListItem}>
                      <div className={styles.uriListItemField}>
                        <TextField
                          size="2"
                          value={uri}
                          onChange={(e) => {
                            onWechatUriItemChange(index, e.target.value);
                          }}
                          disabled={noneditable}
                        />
                      </div>
                      <RadixIconButton
                        type="button"
                        variant="ghost"
                        color="red"
                        size="2"
                        aria-label={renderToString("delete")}
                        onClick={() => {
                          onWechatUriItemDelete(index);
                        }}
                        disabled={noneditable}
                      >
                        <TrashIcon width="1rem" height="1rem" />
                      </RadixIconButton>
                    </div>
                  ))}
                </div>
                <div className={styles.uriListAddButton}>
                  <SecondaryButton
                    size="2"
                    onClick={onWechatUriItemAdd}
                    disabled={noneditable}
                    text={
                      <FormattedMessage id="SingleSignOnConfigurationScreen.widget.add-uri" />
                    }
                  />
                </div>
              </FormField>
            ) : null}
            {visibleFields.has("email_required") ||
            visibleFields.has("create_disabled") ||
            visibleFields.has("delete_disabled") ? (
              <div className={styles.policySection}>
                <Text
                  as="p"
                  size="2"
                  weight="medium"
                  className={styles.sectionTitle}
                >
                  <FormattedMessage id="SingleSignOnConfigurationWidget.signInPolicy.label" />
                </Text>
                <div className={styles.checkboxGroup}>
                  {visibleFields.has("email_required") ? (
                    <WidgetCheckbox
                      label={
                        <FormattedMessage id="SingleSignOnConfigurationScreen.widget.email-required" />
                      }
                      checked={config.claims?.email?.required ?? true}
                      onCheckedChange={onEmailRequiredChange}
                      disabled={noneditable}
                    />
                  ) : null}
                  {visibleFields.has("create_disabled") ? (
                    <WidgetCheckbox
                      label={
                        <FormattedMessage id="SingleSignOnConfigurationScreen.widget.create-disabled" />
                      }
                      checked={config.create_disabled ?? false}
                      onCheckedChange={onCreateDisabledChange}
                      disabled={noneditable}
                    />
                  ) : null}
                  {visibleFields.has("delete_disabled") ? (
                    <WidgetCheckbox
                      label={
                        <FormattedMessage id="SingleSignOnConfigurationScreen.widget.delete-disabled" />
                      }
                      checked={config.delete_disabled ?? false}
                      onCheckedChange={onDeleteDisabledChange}
                      disabled={noneditable}
                    />
                  ) : null}
                </div>
              </div>
            ) : null}
            {visibleFields.has("alias") ? (
              <div className={styles.advancedSection}>
                <button
                  type="button"
                  className={styles.advancedToggle}
                  onClick={onToggleAdvancedFolded}
                >
                  <Text size="2" weight="medium">
                    <FormattedMessage id="SingleSignOnConfigurationWidget.advancedOptions" />
                  </Text>
                  <ChevronDownIcon
                    className={cn(
                      styles.advancedToggleIcon,
                      !advancedFolded ? styles.advancedToggleIconOpen : null
                    )}
                    aria-hidden={true}
                  />
                </button>
                {advancedFolded ? null : (
                  <TextField
                    size="2"
                    labelSize="2"
                    parentJSONPointer={jsonPointer}
                    fieldName="alias"
                    label={renderToString(
                      "SingleSignOnConfigurationScreen.widget.alias"
                    )}
                    hint={renderToString(
                      "SingleSignOnConfigurationScreen.widget.alias.description"
                    )}
                    value={config.alias}
                    onChange={onAliasChange}
                    disabled={noneditable}
                  />
                )}
              </div>
            ) : null}
          </>
        ) : null}
      </SettingsSectionCard>
    );
  };

interface OAuthClientCardProps {
  className?: string;
  providerItemKey: OAuthSSOProviderItemKey;
  isAdded?: boolean;
  onAddClick?: (k: OAuthSSOProviderItemKey) => void;
}

function canAddMultiple(provider: OAuthSSOProviderItemKey): boolean {
  switch (provider) {
    case "azureadb2c":
    case "azureadv2":
    case "adfs":
      return true;
    default:
      return false;
  }
}

export const OAuthClientCard: React.VFC<OAuthClientCardProps> =
  function OAuthClientCard(props) {
    const { className, providerItemKey, isAdded, onAddClick } = props;
    const { renderToString } = useContext(Context);

    const {
      titleId: cardTitleId,
      subtitleId: cardSubtitleId,
      descriptionId: cardDescriptionId,
    } = oauthProviders[providerItemKey];

    const handleAddClick = useCallback(() => {
      onAddClick?.(providerItemKey);
    }, [onAddClick, providerItemKey]);

    return (
      <div className={cn(styles.cardContainer, className)}>
        <div className={styles.cardHeader}>
          <div className={styles.cardTitleRow}>
            <div className={styles.cardIcon}>
              <OAuthClientIcon providerItemKey={providerItemKey} />
            </div>
            <div className={styles.cardName}>
              <Text
                as="p"
                size="2"
                weight="medium"
                className={styles.cardTitle}
              >
                <FormattedMessage id={cardTitleId} />
              </Text>
              {cardSubtitleId != null ? (
                <Text
                  as="p"
                  size="1"
                  color="gray"
                  className={styles.cardSubtitle}
                >
                  <FormattedMessage id={cardSubtitleId} />
                </Text>
              ) : null}
            </div>
          </div>
          {isAdded && !canAddMultiple(providerItemKey) ? (
            <div className={styles.cardAddedBadge}>
              <Text as="p" size="1" color="gray">
                <FormattedMessage id="AddSingleSignOnConfigurationScreen.card.button.added" />
              </Text>
            </div>
          ) : (
            <RadixIconButton
              type="button"
              variant="soft"
              size="2"
              aria-label={renderToString("add")}
              onClick={handleAddClick}
            >
              <PlusIcon width="1rem" height="1rem" />
            </RadixIconButton>
          )}
        </div>
        <div className={styles.cardBody}>
          <Text as="p" size="1" color="gray" className={styles.cardDescription}>
            <FormattedMessage id={cardDescriptionId} />
          </Text>
        </div>
      </div>
    );
  };

interface OAuthClientRowProps {
  className?: string;
  providerConfig: OAuthSSOProviderConfig;
  providersWithDemoCredentials: Set<string>;
  onEditClick?: (provider: OAuthSSOProviderConfig) => void;
  onDeleteClick?: (provider: OAuthSSOProviderConfig) => void;
}

export const OAuthClientRow: React.VFC<OAuthClientRowProps> =
  function OAuthClientRow(props) {
    const {
      className,
      providerConfig,
      providersWithDemoCredentials,
      onEditClick,
      onDeleteClick,
    } = props;
    const { renderToString } = useContext(Context);

    const providerItemKey = useMemo(
      () =>
        createOAuthSSOProviderItemKey(
          providerConfig.type,
          providerConfig.app_type
        ),
      [providerConfig]
    );

    const { titleId, subtitleId, descriptionId } =
      oauthProviders[providerItemKey];

    const handleEditClick = useCallback(() => {
      onEditClick?.(providerConfig);
    }, [onEditClick, providerConfig]);

    const handleDeleteClick = useCallback(() => {
      onDeleteClick?.(providerConfig);
    }, [onDeleteClick, providerConfig]);

    const onRowKeyDown = useCallback(
      (e: React.KeyboardEvent<HTMLDivElement>) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          handleEditClick();
        }
      },
      [handleEditClick]
    );

    return (
      <div
        className={cn(styles.tableRow, className)}
        role="button"
        tabIndex={0}
        onClick={handleEditClick}
        onKeyDown={onRowKeyDown}
      >
        <div className={styles.tableCellProvider}>
          <div className={styles.rowIcon}>
            <OAuthClientIcon providerItemKey={providerItemKey} />
          </div>
          <div className={styles.rowContent}>
            <Text as="p" size="2" className={styles.rowTitle}>
              {`${renderToString(titleId)}${
                subtitleId != null ? ` (${renderToString(subtitleId)})` : ""
              }`}
            </Text>
            <Text
              as="p"
              size="1"
              color="gray"
              className={styles.rowDescription}
            >
              <FormattedMessage id={descriptionId} />
            </Text>
          </div>
        </div>
        <div className={styles.tableCellAlias}>
          <Text as="p" size="2" className={styles.rowAlias}>
            {providerConfig.alias}
          </Text>
        </div>
        <div className={styles.tableCellConfiguration}>
          <ProviderStatus
            providerConfig={providerConfig}
            providersWithDemoCredentials={providersWithDemoCredentials}
          />
        </div>
        <div
          className={styles.tableCellActions}
          onClick={(e) => e.stopPropagation()}
          onKeyDown={(e) => e.stopPropagation()}
        >
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <RadixIconButton
                className={styles.rowActionsButton}
                variant="soft"
                color="gray"
                size="2"
                aria-label={renderToString(
                  "SingleSignOnConfigurationScreen.row-actions"
                )}
              >
                <DotsVerticalIcon width="1rem" height="1rem" />
              </RadixIconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end">
              <DropdownMenu.Item onSelect={handleEditClick}>
                <Pencil1Icon />
                <FormattedMessage id="SingleSignOnConfigurationScreen.edit" />
              </DropdownMenu.Item>
              <DropdownMenu.Item color="red" onSelect={handleDeleteClick}>
                <TrashIcon />
                <FormattedMessage id="SingleSignOnConfigurationScreen.delete" />
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </div>
      </div>
    );
  };

export const OAuthClientRowHeader: React.VFC<{ className?: string }> = ({
  className,
}) => {
  return (
    <div className={cn(styles.tableHeader, className)}>
      <div className={styles.tableHeaderCellProvider}>
        <FormattedMessage id="SingleSignOnConfigurationScreen.header.provider" />
      </div>
      <div className={styles.tableHeaderCellAlias}>
        <FormattedMessage id="SingleSignOnConfigurationScreen.header.alias" />
      </div>
      <div className={styles.tableHeaderCellConfiguration}>
        <FormattedMessage id="SingleSignOnConfigurationScreen.header.configuration" />
      </div>
      <div className={styles.tableHeaderCellActions} />
    </div>
  );
};

export default SingleSignOnConfigurationWidget;
