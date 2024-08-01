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
  /** The ID of an object */
  id: Scalars['ID']['output'];
  isProcessingSubscription: Scalars['Boolean']['output'];
  lastStripeError?: Maybe<Scalars['StripeError']['output']>;
  nftCollections: Array<NftCollection>;
  planName: Scalars['String']['output'];
  rawAppConfig: Scalars['AppConfig']['output'];
  rawAppConfigChecksum: Scalars['AppConfig']['output'];
  resources: Array<AppResource>;
  secretConfig: SecretConfig;
  secretConfigChecksum: Scalars['AppConfig']['output'];
  subscription?: Maybe<Subscription>;
  subscriptionUsage?: Maybe<SubscriptionUsage>;
  tutorialStatus: TutorialStatus;
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
  /** Generate Stripe customer portal session */
  generateStripeCustomerPortalSession: GenerateStripeCustomerPortalSessionPayload;
  /** Generate a token for tester */
  generateTesterToken: GenerateTestTokenPayload;
  /** Preview update subscription */
  previewUpdateSubscription: PreviewUpdateSubscriptionPayload;
  /** Probes a NFT Collection to see whether it is a large collection */
  probeNFTCollection: ProbeNftCollectionsPayload;
  /** Reconcile the completed checkout session */
  reconcileCheckoutSession: ReconcileCheckoutSessionPayload;
  /** Updates the current user's custom attribute with 'survey' key */
  saveOnboardingSurvey?: Maybe<Scalars['Boolean']['output']>;
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


export type MutationGenerateStripeCustomerPortalSessionArgs = {
  input: GenerateStripeCustomerPortalSessionInput;
};


export type MutationGenerateTesterTokenArgs = {
  input: GenerateTestTokenInput;
};


export type MutationPreviewUpdateSubscriptionArgs = {
  input: PreviewUpdateSubscriptionInput;
};


export type MutationProbeNftCollectionArgs = {
  input: ProbeNftCollectionInput;
};


export type MutationReconcileCheckoutSessionArgs = {
  input: ReconcileCheckoutSession;
};


export type MutationSaveOnboardingSurveyArgs = {
  input: SaveOnboardingSurveyInput;
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

/** Web3 NFT Collection */
export type NftCollection = {
  __typename?: 'NFTCollection';
  blockchain: Scalars['String']['output'];
  contractAddress: Scalars['String']['output'];
  createdAt: Scalars['DateTime']['output'];
  name: Scalars['String']['output'];
  network: Scalars['String']['output'];
  tokenType: Scalars['String']['output'];
  totalSupply?: Maybe<Scalars['String']['output']>;
};

/** An object with an ID */
export type Node = {
  /** The id of the object */
  id: Scalars['ID']['output'];
};

export type OAuthClientSecretsCleanupDataInput = {
  keepClientIDs: Array<Scalars['String']['input']>;
};

export type OAuthClientSecretsGenerateDataInput = {
  clientID: Scalars['String']['input'];
};

export type OAuthClientSecretsUpdateInstructionsInput = {
  action: Scalars['String']['input'];
  cleanupData?: InputMaybe<OAuthClientSecretsCleanupDataInput>;
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

export type ProbeNftCollectionInput = {
  contractID: Scalars['String']['input'];
};

export type ProbeNftCollectionsPayload = {
  __typename?: 'ProbeNFTCollectionsPayload';
  isLargeCollection: Scalars['Boolean']['output'];
};

export type Query = {
  __typename?: 'Query';
  /** Active users chart dataset */
  activeUserChart?: Maybe<Chart>;
  /** The list of apps accessible to the current viewer */
  appList?: Maybe<Array<AppListItem>>;
  /** Check whether the viewer can accept the collaboration invitation */
  checkCollaboratorInvitation?: Maybe<CheckCollaboratorInvitationPayload>;
  /** Fetch NFT Contract Metadata */
  nftContractMetadata?: Maybe<NftCollection>;
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


export type QueryNftContractMetadataArgs = {
  contractID: Scalars['String']['input'];
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

/** SMTP secret */
export type SmtpSecret = {
  __typename?: 'SMTPSecret';
  host: Scalars['String']['output'];
  password?: Maybe<Scalars['String']['output']>;
  port: Scalars['Int']['output'];
  username: Scalars['String']['output'];
};

export type SmtpSecretInput = {
  host: Scalars['String']['input'];
  password?: InputMaybe<Scalars['String']['input']>;
  port: Scalars['Int']['input'];
  username: Scalars['String']['input'];
};

export type SaveOnboardingSurveyInput = {
  /** Onboarding survey result JSON. */
  surveyJSON: Scalars['String']['input'];
};

/** The content of authgear.secrets.yaml */
export type SecretConfig = {
  __typename?: 'SecretConfig';
  adminAPISecrets?: Maybe<Array<AdminApiSecret>>;
  botProtectionProviderSecret?: Maybe<BotProtectionProviderSecret>;
  oauthClientSecrets?: Maybe<Array<OauthClientSecretItem>>;
  oauthSSOProviderClientSecrets?: Maybe<Array<OAuthSsoProviderClientSecret>>;
  smtpSecret?: Maybe<SmtpSecret>;
  webhookSecret?: Maybe<WebhookSecret>;
};

export type SecretConfigUpdateInstructionsInput = {
  adminAPIAuthKey?: InputMaybe<AdminApiAuthKeyUpdateInstructionInput>;
  botProtectionProviderSecret?: InputMaybe<BotProtectionProviderSecretUpdateInstructionsInput>;
  oauthClientSecrets?: InputMaybe<OAuthClientSecretsUpdateInstructionsInput>;
  oauthSSOProviderClientSecrets?: InputMaybe<OAuthSsoProviderClientSecretsUpdateInstructionsInput>;
  smtpSecret?: InputMaybe<SmtpSecretUpdateInstructionsInput>;
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
  smsRegion: SubscriptionItemPriceSmsRegion;
  transformQuantityDivideBy?: Maybe<Scalars['Int']['output']>;
  transformQuantityRound: TransformQuantityRound;
  type: SubscriptionItemPriceType;
  unitAmount: Scalars['Int']['output'];
  usageType: SubscriptionItemPriceUsageType;
  whatsappRegion: SubscriptionItemPriceWhatsappRegion;
};

export enum SubscriptionItemPriceSmsRegion {
  None = 'NONE',
  NorthAmerica = 'NORTH_AMERICA',
  OtherRegions = 'OTHER_REGIONS'
}

export enum SubscriptionItemPriceType {
  Fixed = 'FIXED',
  Usage = 'USAGE'
}

export enum SubscriptionItemPriceUsageType {
  Mau = 'MAU',
  None = 'NONE',
  Sms = 'SMS',
  Whatsapp = 'WHATSAPP'
}

export enum SubscriptionItemPriceWhatsappRegion {
  None = 'NONE',
  NorthAmerica = 'NORTH_AMERICA',
  OtherRegions = 'OTHER_REGIONS'
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
  smsRegion: SubscriptionItemPriceSmsRegion;
  totalAmount?: Maybe<Scalars['Int']['output']>;
  transformQuantityDivideBy?: Maybe<Scalars['Int']['output']>;
  transformQuantityRound: TransformQuantityRound;
  type: SubscriptionItemPriceType;
  unitAmount?: Maybe<Scalars['Int']['output']>;
  usageType: SubscriptionItemPriceUsageType;
  whatsappRegion: SubscriptionItemPriceWhatsappRegion;
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
  /** SMTP Username. */
  smtpUsername: Scalars['String']['input'];
  /** The recipient email address. */
  to: Scalars['String']['input'];
};
