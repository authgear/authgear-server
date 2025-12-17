export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  /** The `AppConfig` scalar type represents an app config JSON object */
  AppConfig: { input: GQL_AppConfig; output: GQL_AppConfig; }
  /** The `Date` scalar type represents a Date. The Date is serialized in ISO 8601 format */
  Date: { input: GQL_Date; output: GQL_Date; }
  /** The `DateTime` scalar type represents a DateTime. The DateTime is serialized as an RFC 3339 quoted string */
  DateTime: { input: GQL_DateTime; output: GQL_DateTime; }
  /** The `FeatureConfig` scalar type represents an feature config JSON object */
  FeatureConfig: { input: GQL_FeatureConfig; output: GQL_FeatureConfig; }
  /** The `ProjectWizardData` scalar type represents form data of project wizard */
  ProjectWizardData: { input: any; output: any; }
  /** The `StripeError` scalar type represents Stripe error */
  StripeError: { input: GQL_StripeError; output: GQL_StripeError; }
  /** The `TutorialStatusData` scalar type represents tutorial status data */
  TutorialStatusData: { input: GQL_TutorialStatusData; output: GQL_TutorialStatusData; }
};

export type AcceptCollaboratorInvitationInput = {
  /** Invitation code. */
  code: Scalars['String']['input'];
};

export type AcceptCollaboratorInvitationPayload = {
  __typename?: 'AcceptCollaboratorInvitationPayload';
  app: App;
};

export type AdminApiAuthKeyDeleteDataInput = {
  keyID: Scalars['String']['input'];
};

export type AdminApiAuthKeyUpdateInstructionInput = {
  action: Scalars['String']['input'];
  deleteData?: InputMaybe<AdminApiAuthKeyDeleteDataInput>;
};

/** Admin API secret */
export type AdminApiSecret = {
  __typename?: 'AdminAPISecret';
  createdAt?: Maybe<Scalars['DateTime']['output']>;
  keyID: Scalars['String']['output'];
  privateKeyPEM?: Maybe<Scalars['String']['output']>;
  publicKeyPEM: Scalars['String']['output'];
};

/** Authgear app */
export type App = Node & {
  __typename?: 'App';
  collaboratorInvitations: Array<CollaboratorInvitation>;
  collaborators: Array<Collaborator>;
  domains: Array<Domain>;
  effectiveAppConfig: Scalars['AppConfig']['output'];
  effectiveFeatureConfig: Scalars['FeatureConfig']['output'];
  effectiveSecretConfig: EffectiveSecretConfig;
  /** The ID of an object */
  id: Scalars['ID']['output'];
  isProcessingSubscription: Scalars['Boolean']['output'];
  lastStripeError?: Maybe<Scalars['StripeError']['output']>;
  planName: Scalars['String']['output'];
  rawAppConfig: Scalars['AppConfig']['output'];
  rawAppConfigChecksum: Scalars['AppConfig']['output'];
  resources: Array<AppResource>;
  samlIdpEntityID: Scalars['String']['output'];
  secretConfig: SecretConfig;
  secretConfigChecksum: Scalars['String']['output'];
  subscription?: Maybe<Subscription>;
  subscriptionUsage?: Maybe<SubscriptionUsage>;
  tutorialStatus: TutorialStatus;
  usage?: Maybe<Usage>;
  viewer: Collaborator;
};


/** Authgear app */
export type AppResourcesArgs = {
  paths?: InputMaybe<Array<Scalars['String']['input']>>;
};


/** Authgear app */
export type AppSecretConfigArgs = {
  token?: InputMaybe<Scalars['String']['input']>;
};


/** Authgear app */
export type AppSubscriptionUsageArgs = {
  date: Scalars['DateTime']['input'];
};


/** Authgear app */
export type AppUsageArgs = {
  date: Scalars['DateTime']['input'];
};

export type AppListItem = {
  __typename?: 'AppListItem';
  appID: Scalars['String']['output'];
  publicOrigin: Scalars['String']['output'];
};

/** Resource file for an app */
export type AppResource = {
  __typename?: 'AppResource';
  /** The checksum of the resource file. It is an opaque string that will be used to detect conflict. */
  checksum?: Maybe<Scalars['String']['output']>;
  data?: Maybe<Scalars['String']['output']>;
  effectiveData?: Maybe<Scalars['String']['output']>;
  languageTag?: Maybe<Scalars['String']['output']>;
  path: Scalars['String']['output'];
};

/** Update to resource file. */
export type AppResourceUpdate = {
  /** The checksum of the original resource file. If provided, it will be used to detect conflict. */
  checksum?: InputMaybe<Scalars['String']['input']>;
  /** New data of the resource file. Set to null to remove it. */
  data?: InputMaybe<Scalars['String']['input']>;
  /** Path of the resource file to update. */
  path: Scalars['String']['input'];
};

export enum AppSecretKey {
  AdminApiSecrets = 'ADMIN_API_SECRETS',
  BotProtectionProviderSecret = 'BOT_PROTECTION_PROVIDER_SECRET',
  OauthClientSecrets = 'OAUTH_CLIENT_SECRETS',
  OauthSsoProviderClientSecrets = 'OAUTH_SSO_PROVIDER_CLIENT_SECRETS',
  SamlIdpSigningSecrets = 'SAML_IDP_SIGNING_SECRETS',
  SamlSpSigningSecrets = 'SAML_SP_SIGNING_SECRETS',
  SmsProviderSecrets = 'SMS_PROVIDER_SECRETS',
  SmtpSecret = 'SMTP_SECRET',
  WebhookSecret = 'WEBHOOK_SECRET'
}

/** Bot protection provider secret */
export type BotProtectionProviderSecret = {
  __typename?: 'BotProtectionProviderSecret';
  secretKey?: Maybe<Scalars['String']['output']>;
  type: Scalars['String']['output'];
};

export type BotProtectionProviderSecretInput = {
  secretKey?: InputMaybe<Scalars['String']['input']>;
  type: Scalars['String']['input'];
};

export type BotProtectionProviderSecretUpdateInstructionsInput = {
  action: Scalars['String']['input'];
  data?: InputMaybe<BotProtectionProviderSecretInput>;
};

export type CancelFailedSubscriptionPayload = {
  __typename?: 'CancelFailedSubscriptionPayload';
  app: App;
};

export type Chart = {
  __typename?: 'Chart';
  dataset: Array<Maybe<DataPoint>>;
};

export type CheckCollaboratorInvitationPayload = {
  __typename?: 'CheckCollaboratorInvitationPayload';
  appID: Scalars['String']['output'];
  isInvitee: Scalars['Boolean']['output'];
};

/** Collaborator of an app */
export type Collaborator = {
  __typename?: 'Collaborator';
  createdAt: Scalars['DateTime']['output'];
  id: Scalars['String']['output'];
  role: CollaboratorRole;
  user: User;
};

/** Collaborator invitation of an app */
export type CollaboratorInvitation = {
  __typename?: 'CollaboratorInvitation';
  createdAt: Scalars['DateTime']['output'];
  expireAt: Scalars['DateTime']['output'];
  id: Scalars['String']['output'];
  invitedBy: User;
  inviteeEmail: Scalars['String']['output'];
};

export enum CollaboratorRole {
  Editor = 'EDITOR',
  Owner = 'OWNER'
}

export type CreateAppInput = {
  /** ID of the new app. */
  id: Scalars['String']['input'];
  /** Data of project wizard */
  projectWizardData?: InputMaybe<Scalars['ProjectWizardData']['input']>;
};

export type CreateAppPayload = {
  __typename?: 'CreateAppPayload';
  app: App;
};

export type CreateCheckoutSessionInput = {
  /** App ID. */
  appID: Scalars['ID']['input'];
  /** Plan name. */
  planName: Scalars['String']['input'];
};

export type CreateCheckoutSessionPayload = {
  __typename?: 'CreateCheckoutSessionPayload';
  url: Scalars['String']['output'];
};

export type CreateCollaboratorInvitationInput = {
  /** Target app ID. */
  appID: Scalars['ID']['input'];
  /** Invitee email address. */
  inviteeEmail: Scalars['String']['input'];
};

export type CreateCollaboratorInvitationPayload = {
  __typename?: 'CreateCollaboratorInvitationPayload';
  app: App;
  collaboratorInvitation: CollaboratorInvitation;
};

export type CreateDomainInput = {
  /** Target app ID. */
  appID: Scalars['ID']['input'];
  /** Domain name. */
  domain: Scalars['String']['input'];
};

export type CreateDomainPayload = {
  __typename?: 'CreateDomainPayload';
  app: App;
  domain: Domain;
};

export type CustomSmsProviderSecretsInput = {
  timeout?: InputMaybe<Scalars['Int']['input']>;
  url: Scalars['String']['input'];
};

export type DataPoint = {
  __typename?: 'DataPoint';
  data: Scalars['Float']['output'];
  label: Scalars['String']['output'];
};

export type DeleteCollaboratorInput = {
  /** Collaborator ID. */
  collaboratorID: Scalars['String']['input'];
};

export type DeleteCollaboratorInvitationInput = {
  /** Collaborator invitation ID. */
  collaboratorInvitationID: Scalars['String']['input'];
};

export type DeleteCollaboratorInvitationPayload = {
  __typename?: 'DeleteCollaboratorInvitationPayload';
  app: App;
};

export type DeleteCollaboratorPayload = {
  __typename?: 'DeleteCollaboratorPayload';
  app: App;
};

export type DeleteDomainInput = {
  /** Target app ID. */
  appID: Scalars['ID']['input'];
  /** Domain ID. */
  domainID: Scalars['String']['input'];
};

export type DeleteDomainPayload = {
  __typename?: 'DeleteDomainPayload';
  app: App;
};

/** DNS domain of an app */
export type Domain = {
  __typename?: 'Domain';
  apexDomain: Scalars['String']['output'];
  cookieDomain: Scalars['String']['output'];
  createdAt: Scalars['DateTime']['output'];
  domain: Scalars['String']['output'];
  id: Scalars['String']['output'];
  isCustom: Scalars['Boolean']['output'];
  isVerified: Scalars['Boolean']['output'];
  verificationDNSRecord: Scalars['String']['output'];
};

/** Effective secret config */
export type EffectiveSecretConfig = {
  __typename?: 'EffectiveSecretConfig';
  oauthSSOProviderDemoSecrets?: Maybe<Array<OAuthSsoProviderDemoSecretItem>>;
};

export type GenerateAppSecretVisitTokenInput = {
  /** ID of the app. */
  id: Scalars['ID']['input'];
  /** Secrets to visit. */
  secrets: Array<AppSecretKey>;
};

export type GenerateAppSecretVisitTokenPayloadPayload = {
  __typename?: 'GenerateAppSecretVisitTokenPayloadPayload';
  /** The generated token */
  token: Scalars['String']['output'];
};

export type GenerateShortLivedAdminApiTokenPayload = {
  __typename?: 'GenerateShortLivedAdminAPITokenPayload';
  token: Scalars['String']['output'];
};

export type GenerateStripeCustomerPortalSessionInput = {
  /** Target app ID. */
  appID: Scalars['ID']['input'];
};

export type GenerateStripeCustomerPortalSessionPayload = {
  __typename?: 'GenerateStripeCustomerPortalSessionPayload';
  url: Scalars['String']['output'];
};

export type GenerateTestTokenInput = {
  /** ID of the app. */
  id: Scalars['ID']['input'];
  /** URI to return to in the tester page */
  returnUri: Scalars['String']['input'];
};

export type GenerateTestTokenPayload = {
  __typename?: 'GenerateTestTokenPayload';
  /** The generated token */
  token: Scalars['String']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  /** Accept collaborator invitation to the target app. */
  acceptCollaboratorInvitation: AcceptCollaboratorInvitationPayload;
  /** Cancel failed subscription */
  cancelFailedSubscription: CancelFailedSubscriptionPayload;
  /** Check Deno Hook */
  checkDenoHook?: Maybe<Scalars['Boolean']['output']>;
  /** Create new app */
  createApp: CreateAppPayload;
  /** Create stripe checkout session */
  createCheckoutSession: CreateCheckoutSessionPayload;
  /** Invite a collaborator to the target app. */
  createCollaboratorInvitation: CreateCollaboratorInvitationPayload;
  /** Create domain for target app */
  createDomain: CreateDomainPayload;
  /** Delete collaborator of target app. */
  deleteCollaborator: DeleteCollaboratorPayload;
  /** Delete collaborator invitation of target app. */
  deleteCollaboratorInvitation: DeleteCollaboratorInvitationPayload;
  /** Delete domain of target app */
  deleteDomain: DeleteDomainPayload;
  /** Generate a token for visiting app secrets */
  generateAppSecretVisitToken: GenerateAppSecretVisitTokenPayloadPayload;
  /** Generate short-lived admin API token */
  generateShortLivedAdminAPIToken?: Maybe<GenerateShortLivedAdminApiTokenPayload>;
  /** Generate Stripe customer portal session */
  generateStripeCustomerPortalSession: GenerateStripeCustomerPortalSessionPayload;
  /** Generate a token for tester */
  generateTesterToken: GenerateTestTokenPayload;
  /** Preview update subscription */
  previewUpdateSubscription: PreviewUpdateSubscriptionPayload;
  /** Reconcile the completed checkout session */
  reconcileCheckoutSession: ReconcileCheckoutSessionPayload;
  /** Updates the current user's custom attribute with 'survey' key */
  saveOnboardingSurvey?: Maybe<Scalars['Boolean']['output']>;
  /** Save the progress of project wizard of the app */
  saveProjectWizardData: SaveProjectWizardDataPayload;
  /** Send a SMS to test the configuration */
  sendTestSMSConfiguration?: Maybe<Scalars['Boolean']['output']>;
  /** Send test STMP configuration email */
  sendTestSMTPConfigurationEmail?: Maybe<Scalars['Boolean']['output']>;
  /** Set app subscription cancellation status */
  setSubscriptionCancelledStatus: SetSubscriptionCancelledStatusPayload;
  /** Skip the tutorial of the app */
  skipAppTutorial: SkipAppTutorialPayload;
  /** Skip a progress of the tutorial of the app */
  skipAppTutorialProgress: SkipAppTutorialProgressPayload;
  /** Update app */
  updateApp: UpdateAppPayload;
  /** Update subscription */
  updateSubscription: UpdateSubscriptionPayload;
  /** Request verification of a domain of target app */
  verifyDomain: VerifyDomainPayload;
};


export type MutationAcceptCollaboratorInvitationArgs = {
  input: AcceptCollaboratorInvitationInput;
};


export type MutationCancelFailedSubscriptionArgs = {
  input: CancelFailedSubscriptionInput;
};


export type MutationCheckDenoHookArgs = {
  input: SendDenoHookInput;
};


export type MutationCreateAppArgs = {
  input: CreateAppInput;
};


export type MutationCreateCheckoutSessionArgs = {
  input: CreateCheckoutSessionInput;
};


export type MutationCreateCollaboratorInvitationArgs = {
  input: CreateCollaboratorInvitationInput;
};


export type MutationCreateDomainArgs = {
  input: CreateDomainInput;
};


export type MutationDeleteCollaboratorArgs = {
  input: DeleteCollaboratorInput;
};


export type MutationDeleteCollaboratorInvitationArgs = {
  input: DeleteCollaboratorInvitationInput;
};


export type MutationDeleteDomainArgs = {
  input: DeleteDomainInput;
};


export type MutationGenerateAppSecretVisitTokenArgs = {
  input: GenerateAppSecretVisitTokenInput;
};


export type MutationGenerateShortLivedAdminApiTokenArgs = {
  input: GenerateShortLivedAdminApiTokenInput;
};


export type MutationGenerateStripeCustomerPortalSessionArgs = {
  input: GenerateStripeCustomerPortalSessionInput;
};


export type MutationGenerateTesterTokenArgs = {
  input: GenerateTestTokenInput;
};


export type MutationPreviewUpdateSubscriptionArgs = {
  input: PreviewUpdateSubscriptionInput;
};


export type MutationReconcileCheckoutSessionArgs = {
  input: ReconcileCheckoutSession;
};


export type MutationSaveOnboardingSurveyArgs = {
  input: SaveOnboardingSurveyInput;
};


export type MutationSaveProjectWizardDataArgs = {
  input: SaveProjectWizardDataInput;
};


export type MutationSendTestSmsConfigurationArgs = {
  input: SendTestSmsInput;
};


export type MutationSendTestSmtpConfigurationEmailArgs = {
  input: SendTestSmtpConfigurationEmailInput;
};


export type MutationSetSubscriptionCancelledStatusArgs = {
  input: SetSubscriptionCancelledStatusInput;
};


export type MutationSkipAppTutorialArgs = {
  input: SkipAppTutorialInput;
};


export type MutationSkipAppTutorialProgressArgs = {
  input: SkipAppTutorialProgressInput;
};


export type MutationUpdateAppArgs = {
  input: UpdateAppInput;
};


export type MutationUpdateSubscriptionArgs = {
  input: UpdateSubscriptionInput;
};


export type MutationVerifyDomainArgs = {
  input: VerifyDomainInput;
};

/** An object with an ID */
export type Node = {
  /** The id of the object */
  id: Scalars['ID']['output'];
};

export type OAuthClientSecretsCleanupDataInput = {
  keepClientIDs: Array<Scalars['String']['input']>;
};

export type OAuthClientSecretsDeleteDataInput = {
  clientID: Scalars['String']['input'];
  keyID: Scalars['String']['input'];
};

export type OAuthClientSecretsGenerateDataInput = {
  clientID: Scalars['String']['input'];
};

export type OAuthClientSecretsUpdateInstructionsInput = {
  action: Scalars['String']['input'];
  cleanupData?: InputMaybe<OAuthClientSecretsCleanupDataInput>;
  deleteData?: InputMaybe<OAuthClientSecretsDeleteDataInput>;
  generateData?: InputMaybe<OAuthClientSecretsGenerateDataInput>;
};

/** OAuth client secret */
export type OAuthSsoProviderClientSecret = {
  __typename?: 'OAuthSSOProviderClientSecret';
  alias: Scalars['String']['output'];
  clientSecret?: Maybe<Scalars['String']['output']>;
};

export type OAuthSsoProviderClientSecretInput = {
  newAlias: Scalars['String']['input'];
  newClientSecret?: InputMaybe<Scalars['String']['input']>;
  originalAlias?: InputMaybe<Scalars['String']['input']>;
};

export type OAuthSsoProviderClientSecretsUpdateInstructionsInput = {
  action: Scalars['String']['input'];
  data?: InputMaybe<Array<OAuthSsoProviderClientSecretInput>>;
};

/** OAuth SSO Provider demo secret item */
export type OAuthSsoProviderDemoSecretItem = {
  __typename?: 'OAuthSSOProviderDemoSecretItem';
  type: Scalars['String']['output'];
};

export enum Periodical {
  Monthly = 'MONTHLY',
  Weekly = 'WEEKLY'
}

export type PreviewUpdateSubscriptionInput = {
  /** App ID. */
  appID: Scalars['ID']['input'];
  /** Plan name. */
  planName: Scalars['String']['input'];
};

export type PreviewUpdateSubscriptionPayload = {
  __typename?: 'PreviewUpdateSubscriptionPayload';
  amountDue: Scalars['Int']['output'];
  currency: Scalars['String']['output'];
};

export type Query = {
  __typename?: 'Query';
  /** Active users chart dataset */
  activeUserChart?: Maybe<Chart>;
  /** The list of apps accessible to the current viewer */
  appList?: Maybe<Array<AppListItem>>;
  /** Check whether the viewer can accept the collaboration invitation */
  checkCollaboratorInvitation?: Maybe<CheckCollaboratorInvitationPayload>;
  /** Fetches an object given its ID */
  node?: Maybe<Node>;
  /** Lookup nodes by a list of IDs. */
  nodes: Array<Maybe<Node>>;
  /** Signup by methods dataset */
  signupByMethodsChart?: Maybe<Chart>;
  /** Signup conversion rate dashboard data */
  signupConversionRate?: Maybe<SignupConversionRate>;
  /** Available subscription plans */
  subscriptionPlans: Array<SubscriptionPlan>;
  /** Total users count chart dataset */
  totalUserCountChart?: Maybe<Chart>;
  /** The current viewer */
  viewer?: Maybe<Viewer>;
};


export type QueryActiveUserChartArgs = {
  appID: Scalars['ID']['input'];
  periodical: Periodical;
  rangeFrom: Scalars['Date']['input'];
  rangeTo: Scalars['Date']['input'];
};


export type QueryCheckCollaboratorInvitationArgs = {
  code: Scalars['String']['input'];
};


export type QueryNodeArgs = {
  id: Scalars['ID']['input'];
};


export type QueryNodesArgs = {
  ids: Array<Scalars['ID']['input']>;
};


export type QuerySignupByMethodsChartArgs = {
  appID: Scalars['ID']['input'];
  rangeFrom: Scalars['Date']['input'];
  rangeTo: Scalars['Date']['input'];
};


export type QuerySignupConversionRateArgs = {
  appID: Scalars['ID']['input'];
  rangeFrom: Scalars['Date']['input'];
  rangeTo: Scalars['Date']['input'];
};


export type QueryTotalUserCountChartArgs = {
  appID: Scalars['ID']['input'];
  rangeFrom: Scalars['Date']['input'];
  rangeTo: Scalars['Date']['input'];
};

/** SAML Identity Provider signing certificate */
export type SamlIdpSigningCertificate = {
  __typename?: 'SAMLIdpSigningCertificate';
  certificateFingerprint: Scalars['String']['output'];
  certificatePEM: Scalars['String']['output'];
  keyID: Scalars['String']['output'];
};

/** SAML Identity Provider signing secrets */
export type SamlIdpSigningSecrets = {
  __typename?: 'SAMLIdpSigningSecrets';
  certificates: Array<SamlIdpSigningCertificate>;
};

export type SamlIdpSigningSecretsDeleteDataInput = {
  keyIDs: Array<Scalars['String']['input']>;
};

export type SamlIdpSigningSecretsUpdateInstructionsInput = {
  action: Scalars['String']['input'];
  deleteData?: InputMaybe<SamlIdpSigningSecretsDeleteDataInput>;
};

/** SAML Service Provider signing secrets */
export type SamlSpSigningSecrets = {
  __typename?: 'SAMLSpSigningSecrets';
  certificates: Array<SamlSpSigningCertificate>;
  clientID: Scalars['String']['output'];
};

export type SamlSpSigningSecretsSetDataInput = {
  items: Array<SamlSpSigningSecretsSetDataInputItem>;
};

export type SamlSpSigningSecretsSetDataInputItem = {
  certificates: Array<Scalars['String']['input']>;
  clientID: Scalars['String']['input'];
};

export type SamlSpSigningSecretsUpdateInstructionsInput = {
  action: Scalars['String']['input'];
  setData?: InputMaybe<SamlSpSigningSecretsSetDataInput>;
};

export type SmsProviderConfigurationDenoInput = {
  script: Scalars['String']['input'];
  timeout?: InputMaybe<Scalars['Int']['input']>;
};

export type SmsProviderConfigurationInput = {
  /** Deno hook configuration */
  deno?: InputMaybe<SmsProviderConfigurationDenoInput>;
  /** Twilio configuration */
  twilio?: InputMaybe<SmsProviderConfigurationTwilioInput>;
  /** Webhook Configuration */
  webhook?: InputMaybe<SmsProviderConfigurationWebhookInput>;
};

export type SmsProviderConfigurationTwilioInput = {
  accountSID: Scalars['String']['input'];
  apiKeySID?: InputMaybe<Scalars['String']['input']>;
  apiKeySecret?: InputMaybe<Scalars['String']['input']>;
  authToken?: InputMaybe<Scalars['String']['input']>;
  credentialType: TwilioCredentialType;
  from?: InputMaybe<Scalars['String']['input']>;
  messagingServiceSID?: InputMaybe<Scalars['String']['input']>;
};

export type SmsProviderConfigurationWebhookInput = {
  timeout?: InputMaybe<Scalars['Int']['input']>;
  url: Scalars['String']['input'];
};

/** Custom SMS Provider configs */
export type SmsProviderCustomSmsProviderSecrets = {
  __typename?: 'SMSProviderCustomSMSProviderSecrets';
  timeout?: Maybe<Scalars['Int']['output']>;
  url: Scalars['String']['output'];
};

/** SMS Provider secrets */
export type SmsProviderSecrets = {
  __typename?: 'SMSProviderSecrets';
  customSMSProviderCredentials?: Maybe<SmsProviderCustomSmsProviderSecrets>;
  twilioCredentials?: Maybe<SmsProviderTwilioCredentials>;
};

export type SmsProviderSecretsSetDataInput = {
  customSMSProviderCredentials?: InputMaybe<CustomSmsProviderSecretsInput>;
  twilioCredentials?: InputMaybe<SmsProviderTwilioCredentialsInput>;
};

export type SmsProviderSecretsUpdateInstructionsInput = {
  action: Scalars['String']['input'];
  setData?: InputMaybe<SmsProviderSecretsSetDataInput>;
};

/** Twilio credentials */
export type SmsProviderTwilioCredentials = {
  __typename?: 'SMSProviderTwilioCredentials';
  accountSID: Scalars['String']['output'];
  apiKeySID?: Maybe<Scalars['String']['output']>;
  apiKeySecret?: Maybe<Scalars['String']['output']>;
  authToken?: Maybe<Scalars['String']['output']>;
  credentialType: TwilioCredentialType;
  from?: Maybe<Scalars['String']['output']>;
  messagingServiceSID?: Maybe<Scalars['String']['output']>;
};

export type SmsProviderTwilioCredentialsInput = {
  accountSID: Scalars['String']['input'];
  apiKeySID?: InputMaybe<Scalars['String']['input']>;
  apiKeySecret?: InputMaybe<Scalars['String']['input']>;
  authToken?: InputMaybe<Scalars['String']['input']>;
  credentialType: TwilioCredentialType;
  from?: InputMaybe<Scalars['String']['input']>;
  messagingServiceSID?: InputMaybe<Scalars['String']['input']>;
};

/** SMTP secret */
export type SmtpSecret = {
  __typename?: 'SMTPSecret';
  host: Scalars['String']['output'];
  password?: Maybe<Scalars['String']['output']>;
  port: Scalars['Int']['output'];
  sender?: Maybe<Scalars['String']['output']>;
  username: Scalars['String']['output'];
};

export type SmtpSecretInput = {
  host: Scalars['String']['input'];
  password?: InputMaybe<Scalars['String']['input']>;
  port: Scalars['Int']['input'];
  sender?: InputMaybe<Scalars['String']['input']>;
  username: Scalars['String']['input'];
};

export type SaveOnboardingSurveyInput = {
  /** Onboarding survey result JSON. */
  surveyJSON: Scalars['String']['input'];
};

export type SaveProjectWizardDataInput = {
  /** The project wizard data to save. */
  data?: InputMaybe<Scalars['ProjectWizardData']['input']>;
  /** ID of the app. */
  id: Scalars['String']['input'];
};

export type SaveProjectWizardDataPayload = {
  __typename?: 'SaveProjectWizardDataPayload';
  app: App;
};

/** The content of authgear.secrets.yaml */
export type SecretConfig = {
  __typename?: 'SecretConfig';
  adminAPISecrets?: Maybe<Array<AdminApiSecret>>;
  botProtectionProviderSecret?: Maybe<BotProtectionProviderSecret>;
  oauthClientSecrets?: Maybe<Array<OauthClientSecretItem>>;
  oauthSSOProviderClientSecrets?: Maybe<Array<OAuthSsoProviderClientSecret>>;
  samlIdpSigningSecrets?: Maybe<SamlIdpSigningSecrets>;
  samlSpSigningSecrets?: Maybe<Array<SamlSpSigningSecrets>>;
  smsProviderSecrets?: Maybe<SmsProviderSecrets>;
  smtpSecret?: Maybe<SmtpSecret>;
  webhookSecret?: Maybe<WebhookSecret>;
};

export type SecretConfigUpdateInstructionsInput = {
  adminAPIAuthKey?: InputMaybe<AdminApiAuthKeyUpdateInstructionInput>;
  botProtectionProviderSecret?: InputMaybe<BotProtectionProviderSecretUpdateInstructionsInput>;
  oauthClientSecrets?: InputMaybe<OAuthClientSecretsUpdateInstructionsInput>;
  oauthSSOProviderClientSecrets?: InputMaybe<OAuthSsoProviderClientSecretsUpdateInstructionsInput>;
  samlIdpSigningSecrets?: InputMaybe<SamlIdpSigningSecretsUpdateInstructionsInput>;
  samlSpSigningSecrets?: InputMaybe<SamlSpSigningSecretsUpdateInstructionsInput>;
  smsProviderSecrets?: InputMaybe<SmsProviderSecretsUpdateInstructionsInput>;
  smtpSecret?: InputMaybe<SmtpSecretUpdateInstructionsInput>;
};

export type SendTestSmsInput = {
  /** App ID to test. */
  appID: Scalars['ID']['input'];
  /** The SMS provider configuration. */
  providerConfiguration: SmsProviderConfigurationInput;
  /** The recipient phone number. */
  to: Scalars['String']['input'];
};

export type SetSubscriptionCancelledStatusInput = {
  /** Target app ID. */
  appID: Scalars['ID']['input'];
  /** Target app subscription cancellation status. */
  cancelled: Scalars['Boolean']['input'];
};

export type SetSubscriptionCancelledStatusPayload = {
  __typename?: 'SetSubscriptionCancelledStatusPayload';
  app: App;
};

/** Signup conversion rate dashboard data */
export type SignupConversionRate = {
  __typename?: 'SignupConversionRate';
  conversionRate: Scalars['Float']['output'];
  totalSignup: Scalars['Int']['output'];
  totalSignupUniquePageView: Scalars['Int']['output'];
};

export type SkipAppTutorialInput = {
  /** ID of the app. */
  id: Scalars['String']['input'];
};

export type SkipAppTutorialPayload = {
  __typename?: 'SkipAppTutorialPayload';
  app: App;
};

export type SkipAppTutorialProgressInput = {
  /** ID of the app. */
  id: Scalars['String']['input'];
  /** The progress to skip. */
  progress: Scalars['String']['input'];
};

export type SkipAppTutorialProgressPayload = {
  __typename?: 'SkipAppTutorialProgressPayload';
  app: App;
};

export type SmtpSecretUpdateInstructionsInput = {
  action: Scalars['String']['input'];
  data?: InputMaybe<SmtpSecretInput>;
};

export type Subscription = {
  __typename?: 'Subscription';
  cancelledAt?: Maybe<Scalars['DateTime']['output']>;
  createdAt: Scalars['DateTime']['output'];
  endedAt?: Maybe<Scalars['DateTime']['output']>;
  id: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type SubscriptionItemPrice = {
  __typename?: 'SubscriptionItemPrice';
  currency: Scalars['String']['output'];
  freeQuantity?: Maybe<Scalars['Int']['output']>;
  smsRegion: UsageSmsRegion;
  transformQuantityDivideBy?: Maybe<Scalars['Int']['output']>;
  transformQuantityRound: TransformQuantityRound;
  type: SubscriptionItemPriceType;
  unitAmount: Scalars['Int']['output'];
  usageType: UsageType;
  whatsappRegion: UsageWhatsappRegion;
};

export enum SubscriptionItemPriceType {
  Fixed = 'FIXED',
  Usage = 'USAGE'
}

export type SubscriptionPlan = {
  __typename?: 'SubscriptionPlan';
  name: Scalars['String']['output'];
  prices: Array<SubscriptionItemPrice>;
};

export type SubscriptionUsage = {
  __typename?: 'SubscriptionUsage';
  items: Array<SubscriptionUsageItem>;
  nextBillingDate: Scalars['DateTime']['output'];
};

export type SubscriptionUsageItem = {
  __typename?: 'SubscriptionUsageItem';
  currency?: Maybe<Scalars['String']['output']>;
  freeQuantity?: Maybe<Scalars['Int']['output']>;
  quantity: Scalars['Int']['output'];
  smsRegion: UsageSmsRegion;
  totalAmount?: Maybe<Scalars['Int']['output']>;
  transformQuantityDivideBy?: Maybe<Scalars['Int']['output']>;
  transformQuantityRound: TransformQuantityRound;
  type: SubscriptionItemPriceType;
  unitAmount?: Maybe<Scalars['Int']['output']>;
  usageType: UsageType;
  whatsappRegion: UsageWhatsappRegion;
};

export enum TransformQuantityRound {
  Down = 'DOWN',
  None = 'NONE',
  Up = 'UP'
}

/** Tutorial status of an app */
export type TutorialStatus = {
  __typename?: 'TutorialStatus';
  appID: Scalars['String']['output'];
  data: Scalars['TutorialStatusData']['output'];
};

export enum TwilioCredentialType {
  ApiKey = 'api_key',
  AuthToken = 'auth_token'
}

export type UpdateAppInput = {
  /** authgear.yaml in JSON. */
  appConfig?: InputMaybe<Scalars['AppConfig']['input']>;
  /** The checksum of appConfig. If provided, it will be used to detect conflict. */
  appConfigChecksum?: InputMaybe<Scalars['String']['input']>;
  /** App ID to update. */
  appID: Scalars['ID']['input'];
  /** update secret config instructions. */
  secretConfigUpdateInstructions?: InputMaybe<SecretConfigUpdateInstructionsInput>;
  /** The checksum of secretConfig. If provided, it will be used to detect conflict. */
  secretConfigUpdateInstructionsChecksum?: InputMaybe<Scalars['String']['input']>;
  /** Resource file updates. */
  updates?: InputMaybe<Array<AppResourceUpdate>>;
};

export type UpdateAppPayload = {
  __typename?: 'UpdateAppPayload';
  app: App;
};

export type UpdateSubscriptionInput = {
  /** App ID. */
  appID: Scalars['ID']['input'];
  /** Plan name. */
  planName: Scalars['String']['input'];
};

export type UpdateSubscriptionPayload = {
  __typename?: 'UpdateSubscriptionPayload';
  app: App;
};

export type Usage = {
  __typename?: 'Usage';
  items: Array<UsageItem>;
};

export type UsageItem = {
  __typename?: 'UsageItem';
  quantity: Scalars['Int']['output'];
  smsRegion: UsageSmsRegion;
  usageType: UsageType;
  whatsappRegion: UsageWhatsappRegion;
};

export enum UsageSmsRegion {
  None = 'NONE',
  NorthAmerica = 'NORTH_AMERICA',
  OtherRegions = 'OTHER_REGIONS'
}

export enum UsageType {
  Mau = 'MAU',
  None = 'NONE',
  Sms = 'SMS',
  Whatsapp = 'WHATSAPP'
}

export enum UsageWhatsappRegion {
  None = 'NONE',
  NorthAmerica = 'NORTH_AMERICA',
  OtherRegions = 'OTHER_REGIONS'
}

/** Portal User */
export type User = Node & {
  __typename?: 'User';
  email?: Maybe<Scalars['String']['output']>;
  /** The ID of an object */
  id: Scalars['ID']['output'];
};

export type VerifyDomainInput = {
  /** Target app ID. */
  appID: Scalars['ID']['input'];
  /** Domain ID. */
  domainID: Scalars['String']['input'];
};

export type VerifyDomainPayload = {
  __typename?: 'VerifyDomainPayload';
  app: App;
  domain: Domain;
};

/** The viewer */
export type Viewer = Node & {
  __typename?: 'Viewer';
  email?: Maybe<Scalars['String']['output']>;
  formattedName?: Maybe<Scalars['String']['output']>;
  geoIPCountryCode?: Maybe<Scalars['String']['output']>;
  /** The ID of an object */
  id: Scalars['ID']['output'];
  isOnboardingSurveyCompleted?: Maybe<Scalars['Boolean']['output']>;
  projectOwnerCount: Scalars['Int']['output'];
  projectQuota?: Maybe<Scalars['Int']['output']>;
};

/** Webhook secret */
export type WebhookSecret = {
  __typename?: 'WebhookSecret';
  secret?: Maybe<Scalars['String']['output']>;
};

export type CancelFailedSubscriptionInput = {
  /** Target app ID. */
  appID: Scalars['ID']['input'];
};

export type GenerateShortLivedAdminApiTokenInput = {
  /** App ID to generate token for. */
  appID: Scalars['ID']['input'];
  /** App secret visit token. */
  appSecretVisitToken: Scalars['String']['input'];
};

/** OAuth client secret item */
export type OauthClientSecretItem = {
  __typename?: 'oauthClientSecretItem';
  clientID: Scalars['String']['output'];
  keys?: Maybe<Array<OauthClientSecretKey>>;
};

/** OAuth client secret key */
export type OauthClientSecretKey = {
  __typename?: 'oauthClientSecretKey';
  createdAt?: Maybe<Scalars['DateTime']['output']>;
  key: Scalars['String']['output'];
  keyID: Scalars['String']['output'];
};

export type ReconcileCheckoutSession = {
  /** Target app ID. */
  appID: Scalars['ID']['input'];
  /** Checkout session ID. */
  checkoutSessionID: Scalars['String']['input'];
};

export type ReconcileCheckoutSessionPayload = {
  __typename?: 'reconcileCheckoutSessionPayload';
  app: App;
};

/** SAML Identity Provider signing certificate */
export type SamlSpSigningCertificate = {
  __typename?: 'samlSpSigningCertificate';
  certificateFingerprint: Scalars['String']['output'];
  certificatePEM: Scalars['String']['output'];
};

export type SendDenoHookInput = {
  /** App ID. */
  appID: Scalars['ID']['input'];
  /** The content of the hook. */
  content: Scalars['String']['input'];
};

export type SendTestSmtpConfigurationEmailInput = {
  /** App ID to test. */
  appID: Scalars['ID']['input'];
  /** SMTP Host. */
  smtpHost: Scalars['String']['input'];
  /** SMTP Password. */
  smtpPassword: Scalars['String']['input'];
  /** SMTP Port. */
  smtpPort: Scalars['Int']['input'];
  /** SMTP Sender. */
  smtpSender: Scalars['String']['input'];
  /** SMTP Username. */
  smtpUsername: Scalars['String']['input'];
  /** The recipient email address. */
  to: Scalars['String']['input'];
};
