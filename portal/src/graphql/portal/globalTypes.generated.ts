export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  /** The `AppConfig` scalar type represents an app config JSON object */
  AppConfig: GQL_AppConfig;
  /** The `Date` scalar type represents a Date. The Date is serialized in ISO 8601 format */
  Date: GQL_Date;
  /** The `DateTime` scalar type represents a DateTime. The DateTime is serialized as an RFC 3339 quoted string */
  DateTime: GQL_DateTime;
  /** The `FeatureConfig` scalar type represents an feature config JSON object */
  FeatureConfig: GQL_FeatureConfig;
  /** The `TutorialStatusData` scalar type represents tutorial status data */
  TutorialStatusData: GQL_TutorialStatusData;
};

export type AcceptCollaboratorInvitationInput = {
  /** Invitation code. */
  code: Scalars['String'];
};

export type AcceptCollaboratorInvitationPayload = {
  __typename?: 'AcceptCollaboratorInvitationPayload';
  app: App;
};

/** Admin API secret */
export type AdminApiSecret = {
  __typename?: 'AdminAPISecret';
  createdAt?: Maybe<Scalars['DateTime']>;
  keyID: Scalars['String'];
  privateKeyPEM?: Maybe<Scalars['String']>;
  publicKeyPEM: Scalars['String'];
};

/** Authgear app */
export type App = Node & {
  __typename?: 'App';
  collaboratorInvitations: Array<CollaboratorInvitation>;
  collaborators: Array<Collaborator>;
  domains: Array<Domain>;
  effectiveAppConfig: Scalars['AppConfig'];
  effectiveFeatureConfig: Scalars['FeatureConfig'];
  /** The ID of an object */
  id: Scalars['ID'];
  isProcessingSubscription: Scalars['Boolean'];
  nftCollections: Array<NftCollection>;
  planName: Scalars['String'];
  rawAppConfig: Scalars['AppConfig'];
  resources: Array<AppResource>;
  secretConfig: SecretConfig;
  subscription?: Maybe<Subscription>;
  subscriptionUsage?: Maybe<SubscriptionUsage>;
  tutorialStatus: TutorialStatus;
  viewer: Collaborator;
};


/** Authgear app */
export type AppResourcesArgs = {
  paths?: InputMaybe<Array<Scalars['String']>>;
};


/** Authgear app */
export type AppSubscriptionUsageArgs = {
  date: Scalars['DateTime'];
};

/** A connection to a list of items. */
export type AppConnection = {
  __typename?: 'AppConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<AppEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']>;
};

/** An edge in a connection */
export type AppEdge = {
  __typename?: 'AppEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of the edge */
  node?: Maybe<App>;
};

/** Resource file for an app */
export type AppResource = {
  __typename?: 'AppResource';
  data?: Maybe<Scalars['String']>;
  effectiveData?: Maybe<Scalars['String']>;
  languageTag?: Maybe<Scalars['String']>;
  path: Scalars['String'];
};

/** Update to resource file. */
export type AppResourceUpdate = {
  /** New data of the resource file. Set to null to remove it. */
  data?: InputMaybe<Scalars['String']>;
  /** Path of the resource file to update. */
  path: Scalars['String'];
};

export type Chart = {
  __typename?: 'Chart';
  dataset: Array<Maybe<DataPoint>>;
};

export type CheckCollaboratorInvitationPayload = {
  __typename?: 'CheckCollaboratorInvitationPayload';
  appID: Scalars['String'];
  isInvitee: Scalars['Boolean'];
};

/** Collaborator of an app */
export type Collaborator = {
  __typename?: 'Collaborator';
  createdAt: Scalars['DateTime'];
  id: Scalars['String'];
  role: CollaboratorRole;
  user: User;
};

/** Collaborator invitation of an app */
export type CollaboratorInvitation = {
  __typename?: 'CollaboratorInvitation';
  createdAt: Scalars['DateTime'];
  expireAt: Scalars['DateTime'];
  id: Scalars['String'];
  invitedBy: User;
  inviteeEmail: Scalars['String'];
};

export enum CollaboratorRole {
  Editor = 'EDITOR',
  Owner = 'OWNER'
}

export type CreateAppInput = {
  /** ID of the new app. */
  id: Scalars['String'];
};

export type CreateAppPayload = {
  __typename?: 'CreateAppPayload';
  app: App;
};

export type CreateCheckoutSessionInput = {
  /** App ID. */
  appID: Scalars['ID'];
  /** Plan name. */
  planName: Scalars['String'];
};

export type CreateCheckoutSessionPayload = {
  __typename?: 'CreateCheckoutSessionPayload';
  url: Scalars['String'];
};

export type CreateCollaboratorInvitationInput = {
  /** Target app ID. */
  appID: Scalars['ID'];
  /** Invitee email address. */
  inviteeEmail: Scalars['String'];
};

export type CreateCollaboratorInvitationPayload = {
  __typename?: 'CreateCollaboratorInvitationPayload';
  app: App;
  collaboratorInvitation: CollaboratorInvitation;
};

export type CreateDomainInput = {
  /** Target app ID. */
  appID: Scalars['ID'];
  /** Domain name. */
  domain: Scalars['String'];
};

export type CreateDomainPayload = {
  __typename?: 'CreateDomainPayload';
  app: App;
  domain: Domain;
};

export type DataPoint = {
  __typename?: 'DataPoint';
  data: Scalars['Float'];
  label: Scalars['String'];
};

export type DeleteCollaboratorInput = {
  /** Collaborator ID. */
  collaboratorID: Scalars['String'];
};

export type DeleteCollaboratorInvitationInput = {
  /** Collaborator invitation ID. */
  collaboratorInvitationID: Scalars['String'];
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
  appID: Scalars['ID'];
  /** Domain ID. */
  domainID: Scalars['String'];
};

export type DeleteDomainPayload = {
  __typename?: 'DeleteDomainPayload';
  app: App;
};

/** DNS domain of an app */
export type Domain = {
  __typename?: 'Domain';
  apexDomain: Scalars['String'];
  cookieDomain: Scalars['String'];
  createdAt: Scalars['DateTime'];
  domain: Scalars['String'];
  id: Scalars['String'];
  isCustom: Scalars['Boolean'];
  isVerified: Scalars['Boolean'];
  verificationDNSRecord: Scalars['String'];
};

export type GenerateStripeCustomerPortalSessionInput = {
  /** Target app ID. */
  appID: Scalars['ID'];
};

export type GenerateStripeCustomerPortalSessionPayload = {
  __typename?: 'GenerateStripeCustomerPortalSessionPayload';
  url: Scalars['String'];
};

export type Mutation = {
  __typename?: 'Mutation';
  /** Accept collaborator invitation to the target app. */
  acceptCollaboratorInvitation: AcceptCollaboratorInvitationPayload;
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
  /** Generate Stripe customer portal session */
  generateStripeCustomerPortalSession: GenerateStripeCustomerPortalSessionPayload;
  /** Preview update subscription */
  previewUpdateSubscription: PreviewUpdateSubscriptionPayload;
  /** Reconcile the completed checkout session */
  reconcileCheckoutSession: ReconcileCheckoutSessionPayload;
  /** Send test STMP configuration email */
  sendTestSMTPConfigurationEmail?: Maybe<Scalars['Boolean']>;
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
  /** Start watching a batch of NFT Collections */
  watchNFTCollections: WatchNftCollectionsPayload;
};


export type MutationAcceptCollaboratorInvitationArgs = {
  input: AcceptCollaboratorInvitationInput;
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


export type MutationGenerateStripeCustomerPortalSessionArgs = {
  input: GenerateStripeCustomerPortalSessionInput;
};


export type MutationPreviewUpdateSubscriptionArgs = {
  input: PreviewUpdateSubscriptionInput;
};


export type MutationReconcileCheckoutSessionArgs = {
  input: ReconcileCheckoutSession;
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


export type MutationWatchNftCollectionsArgs = {
  input: WatchNftCollectionsInput;
};

/** Web3 NFT Collection */
export type NftCollection = {
  __typename?: 'NFTCollection';
  blockHeight: Scalars['Int'];
  blockchain: Scalars['String'];
  contractAddress: Scalars['String'];
  createdAt: Scalars['DateTime'];
  name: Scalars['String'];
  network: Scalars['String'];
  tokenType: Scalars['String'];
  totalSupply: Scalars['Int'];
};

/** Web3 NFT ContractMetadata */
export type NftContractMetadata = {
  __typename?: 'NFTContractMetadata';
  address: Scalars['String'];
  name: Scalars['String'];
  symbol: Scalars['String'];
  tokenType: Scalars['String'];
  totalSupply: Scalars['String'];
};

/** An object with an ID */
export type Node = {
  /** The id of the object */
  id: Scalars['ID'];
};

/** OAuth client secret */
export type OAuthClientSecret = {
  __typename?: 'OAuthClientSecret';
  alias: Scalars['String'];
  clientSecret: Scalars['String'];
};

export type OauthClientSecretInput = {
  alias: Scalars['String'];
  clientSecret: Scalars['String'];
};

/** Information about pagination in a connection. */
export type PageInfo = {
  __typename?: 'PageInfo';
  /** When paginating forwards, the cursor to continue. */
  endCursor?: Maybe<Scalars['String']>;
  /** When paginating forwards, are there more items? */
  hasNextPage: Scalars['Boolean'];
  /** When paginating backwards, are there more items? */
  hasPreviousPage: Scalars['Boolean'];
  /** When paginating backwards, the cursor to continue. */
  startCursor?: Maybe<Scalars['String']>;
};

export enum Periodical {
  Monthly = 'MONTHLY',
  Weekly = 'WEEKLY'
}

export type PreviewUpdateSubscriptionInput = {
  /** App ID. */
  appID: Scalars['ID'];
  /** Plan name. */
  planName: Scalars['String'];
};

export type PreviewUpdateSubscriptionPayload = {
  __typename?: 'PreviewUpdateSubscriptionPayload';
  amountDue: Scalars['Int'];
  currency: Scalars['String'];
};

export type Query = {
  __typename?: 'Query';
  /** Active users chart dataset */
  activeUserChart?: Maybe<Chart>;
  /** All apps accessible by the viewer */
  apps?: Maybe<AppConnection>;
  /** Check whether the viewer can accept the collaboration invitation */
  checkCollaboratorInvitation?: Maybe<CheckCollaboratorInvitationPayload>;
  /** Fetch NFT Contract Metadata */
  nftContractMetadata?: Maybe<NftContractMetadata>;
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
  viewer?: Maybe<User>;
};


export type QueryActiveUserChartArgs = {
  appID: Scalars['ID'];
  periodical: Periodical;
  rangeFrom: Scalars['Date'];
  rangeTo: Scalars['Date'];
};


export type QueryAppsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


export type QueryCheckCollaboratorInvitationArgs = {
  code: Scalars['String'];
};


export type QueryNftContractMetadataArgs = {
  appID?: InputMaybe<Scalars['ID']>;
  contractID?: InputMaybe<Scalars['String']>;
};


export type QueryNodeArgs = {
  id: Scalars['ID'];
};


export type QueryNodesArgs = {
  ids: Array<Scalars['ID']>;
};


export type QuerySignupByMethodsChartArgs = {
  appID: Scalars['ID'];
  rangeFrom: Scalars['Date'];
  rangeTo: Scalars['Date'];
};


export type QuerySignupConversionRateArgs = {
  appID: Scalars['ID'];
  rangeFrom: Scalars['Date'];
  rangeTo: Scalars['Date'];
};


export type QueryTotalUserCountChartArgs = {
  appID: Scalars['ID'];
  rangeFrom: Scalars['Date'];
  rangeTo: Scalars['Date'];
};

/** SMTP secret */
export type SmtpSecret = {
  __typename?: 'SMTPSecret';
  host: Scalars['String'];
  password?: Maybe<Scalars['String']>;
  port: Scalars['Int'];
  username: Scalars['String'];
};

export type SmtpSecretInput = {
  host: Scalars['String'];
  password?: InputMaybe<Scalars['String']>;
  port: Scalars['Int'];
  username: Scalars['String'];
};

/** The content of authgear.secrets.yaml */
export type SecretConfig = {
  __typename?: 'SecretConfig';
  adminAPISecrets?: Maybe<Array<AdminApiSecret>>;
  oauthClientSecrets?: Maybe<Array<OAuthClientSecret>>;
  smtpSecret?: Maybe<SmtpSecret>;
  webhookSecret?: Maybe<WebhookSecret>;
};

export type SecretConfigInput = {
  oauthClientSecrets?: InputMaybe<Array<OauthClientSecretInput>>;
  smtpSecret?: InputMaybe<SmtpSecretInput>;
};

export type SetSubscriptionCancelledStatusInput = {
  /** Target app ID. */
  appID: Scalars['ID'];
  /** Target app subscription cancellation status. */
  cancelled: Scalars['Boolean'];
};

export type SetSubscriptionCancelledStatusPayload = {
  __typename?: 'SetSubscriptionCancelledStatusPayload';
  app: App;
};

/** Signup conversion rate dashboard data */
export type SignupConversionRate = {
  __typename?: 'SignupConversionRate';
  conversionRate: Scalars['Float'];
  totalSignup: Scalars['Int'];
  totalSignupUniquePageView: Scalars['Int'];
};

export type SkipAppTutorialInput = {
  /** ID of the app. */
  id: Scalars['String'];
};

export type SkipAppTutorialPayload = {
  __typename?: 'SkipAppTutorialPayload';
  app: App;
};

export type SkipAppTutorialProgressInput = {
  /** ID of the app. */
  id: Scalars['String'];
  /** The progress to skip. */
  progress: Scalars['String'];
};

export type SkipAppTutorialProgressPayload = {
  __typename?: 'SkipAppTutorialProgressPayload';
  app: App;
};

export type Subscription = {
  __typename?: 'Subscription';
  cancelledAt?: Maybe<Scalars['DateTime']>;
  createdAt: Scalars['DateTime'];
  endedAt?: Maybe<Scalars['DateTime']>;
  id: Scalars['String'];
  updatedAt: Scalars['DateTime'];
};

export type SubscriptionItemPrice = {
  __typename?: 'SubscriptionItemPrice';
  currency: Scalars['String'];
  freeQuantity?: Maybe<Scalars['Int']>;
  smsRegion: SubscriptionItemPriceSmsRegion;
  transformQuantityDivideBy?: Maybe<Scalars['Int']>;
  transformQuantityRound: TransformQuantityRound;
  type: SubscriptionItemPriceType;
  unitAmount: Scalars['Int'];
  usageType: SubscriptionItemPriceUsageType;
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
  Sms = 'SMS'
}

export type SubscriptionPlan = {
  __typename?: 'SubscriptionPlan';
  name: Scalars['String'];
  prices: Array<SubscriptionItemPrice>;
};

export type SubscriptionUsage = {
  __typename?: 'SubscriptionUsage';
  items: Array<SubscriptionUsageItem>;
  nextBillingDate: Scalars['DateTime'];
};

export type SubscriptionUsageItem = {
  __typename?: 'SubscriptionUsageItem';
  currency?: Maybe<Scalars['String']>;
  freeQuantity?: Maybe<Scalars['Int']>;
  quantity: Scalars['Int'];
  smsRegion: SubscriptionItemPriceSmsRegion;
  totalAmount?: Maybe<Scalars['Int']>;
  transformQuantityDivideBy?: Maybe<Scalars['Int']>;
  transformQuantityRound: TransformQuantityRound;
  type: SubscriptionItemPriceType;
  unitAmount?: Maybe<Scalars['Int']>;
  usageType: SubscriptionItemPriceUsageType;
};

export enum TransformQuantityRound {
  Down = 'DOWN',
  None = 'NONE',
  Up = 'UP'
}

/** Tutorial status of an app */
export type TutorialStatus = {
  __typename?: 'TutorialStatus';
  appID: Scalars['String'];
  data: Scalars['TutorialStatusData'];
};

export type UpdateAppInput = {
  /** authgear.yaml in JSON. */
  appConfig?: InputMaybe<Scalars['AppConfig']>;
  /** App ID to update. */
  appID: Scalars['ID'];
  /** secrets to update. */
  secretConfig?: InputMaybe<SecretConfigInput>;
  /** Resource file updates. */
  updates?: InputMaybe<Array<AppResourceUpdate>>;
};

export type UpdateAppPayload = {
  __typename?: 'UpdateAppPayload';
  app: App;
};

export type UpdateSubscriptionInput = {
  /** App ID. */
  appID: Scalars['ID'];
  /** Plan name. */
  planName: Scalars['String'];
};

export type UpdateSubscriptionPayload = {
  __typename?: 'UpdateSubscriptionPayload';
  app: App;
};

/** Portal User */
export type User = Node & {
  __typename?: 'User';
  email?: Maybe<Scalars['String']>;
  /** The ID of an object */
  id: Scalars['ID'];
};

export type VerifyDomainInput = {
  /** Target app ID. */
  appID: Scalars['ID'];
  /** Domain ID. */
  domainID: Scalars['String'];
};

export type VerifyDomainPayload = {
  __typename?: 'VerifyDomainPayload';
  app: App;
  domain: Domain;
};

export type WatchNftCollectionsInput = {
  contractIDs: Array<Scalars['String']>;
  /** ID of the app. */
  id: Scalars['String'];
};

export type WatchNftCollectionsPayload = {
  __typename?: 'WatchNFTCollectionsPayload';
  app: App;
};

/** Webhook secret */
export type WebhookSecret = {
  __typename?: 'WebhookSecret';
  secret?: Maybe<Scalars['String']>;
};

export type ReconcileCheckoutSession = {
  /** Target app ID. */
  appID: Scalars['ID'];
  /** Checkout session ID. */
  checkoutSessionID: Scalars['String'];
};

export type ReconcileCheckoutSessionPayload = {
  __typename?: 'reconcileCheckoutSessionPayload';
  app: App;
};

export type SendTestSmtpConfigurationEmailInput = {
  /** App ID to test. */
  appID: Scalars['ID'];
  /** SMTP Host. */
  smtpHost: Scalars['String'];
  /** SMTP Password. */
  smtpPassword: Scalars['String'];
  /** SMTP Port. */
  smtpPort: Scalars['Int'];
  /** SMTP Username. */
  smtpUsername: Scalars['String'];
  /** The recipient email address. */
  to: Scalars['String'];
};
