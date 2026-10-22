import React, { useMemo, lazy, Suspense } from "react";
import { useParams, Navigate } from "react-router-dom";
import { ApolloProvider, useQuery } from "@apollo/client";

import { makeClient } from "./graphql/adminapi/apollo";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import ScreenLayout from "./ScreenLayout";
import ShowLoading from "./ShowLoading";
import { useUnauthenticatedDialogContext } from "./components/auth/UnauthenticatedDialogContext";
import { useSystemConfig } from "./context/SystemConfigContext";
import {
  ScreenNavQueryDocument,
  ScreenNavQueryQuery,
} from "./graphql/portal/query/screenNavQuery.generated";
import { usePortalClient } from "./graphql/portal/apollo";
import RequireAppFeature from "./RequireAppFeature";
import { AppRoute, AppRoutes } from "./AppRoutes";
import RequireUser from "./RequireUser";
import RequireAPIResource from "./RequireAPIResource";
import RequireRole from "./RequireRole";
import RequireGroup from "./RequireGroup";
import { PortalAPIFeatureConfig } from "./types";
import { isIntegrationsAvailable } from "./util/integrations";

// Feature-config predicates for the route-level RequireAppFeature guards. Each
// returns whether the project may access the gated section. Kept module-level so
// their identity is stable across renders.
const isApp2AppAvailable = (fc: PortalAPIFeatureConfig | null): boolean =>
  fc?.oauth?.client?.app2app_enabled ?? false;

const RolesScreen = lazy(async () => import("./graphql/adminapi/RolesScreen"));
const AddRoleScreen = lazy(
  async () => import("./graphql/adminapi/AddRoleScreen")
);
const RoleDetailsScreen = lazy(
  async () => import("./graphql/adminapi/RoleDetailsScreen")
);
const GroupsScreen = lazy(
  async () => import("./graphql/adminapi/GroupsScreen")
);
const AddGroupScreen = lazy(
  async () => import("./graphql/adminapi/AddGroupScreen")
);
const GroupDetailsScreen = lazy(
  async () => import("./graphql/adminapi/GroupDetailsScreen")
);
const UsersRedirectScreen = lazy(async () => import("./UsersRedirectScreen"));
const UsersScreen = lazy(async () => import("./graphql/adminapi/UsersScreen"));
const AddUserScreen = lazy(
  async () => import("./graphql/adminapi/AddUserScreen")
);
const UserDetailsScreen = lazy(
  async () => import("./graphql/adminapi/UserDetailsScreen")
);
const ChangePasswordScreen = lazy(
  async () => import("./graphql/adminapi/ChangePasswordScreen")
);
const AddPasswordScreen = lazy(
  async () => import("./graphql/adminapi/AddPasswordScreen")
);

const AuditLogScreen = lazy(
  async () => import("./graphql/adminapi/AuditLogScreen")
);
const AuditLogEntryScreen = lazy(
  async () => import("./graphql/adminapi/AuditLogEntryScreen")
);

const ProjectRootScreen = lazy(
  async () => import("./graphql/portal/ProjectRootScreen")
);
const GetStartedScreen = lazy(
  async () => import("./graphql/portal/GetStartedScreen/GetStartedScreen")
);
const AnonymousUsersConfigurationScreen = lazy(
  async () => import("./graphql/portal/AnonymousUsersConfigurationScreen")
);
const App2AppConfigurationScreen = lazy(
  async () => import("./graphql/portal/App2AppConfigurationScreen")
);
const SingleSignOnConfigurationScreen = lazy(
  async () => import("./graphql/portal/SingleSignOnConfigurationScreen")
);
const AddSingleSignOnConfigurationScreen = lazy(
  async () => import("./graphql/portal/AddSingleSignOnConfigurationScreen")
);
const EditSingleSignOnConfigurationScreen = lazy(
  async () => import("./graphql/portal/EditSingleSignOnConfigurationScreen")
);
const CreateOAuthClientScreen = lazy(
  async () => import("./graphql/portal/CreateOAuthClientScreen")
);
const CreateM2MClientScreen = lazy(
  async () => import("./graphql/portal/CreateM2MClientScreen")
);
const EditOAuthClientScreen = lazy(
  async () => import("./graphql/portal/EditOAuthClientScreen")
);
const CustomDomainListScreen = lazy(
  async () => import("./graphql/portal/CustomDomainListScreen")
);
const VerifyDomainScreen = lazy(
  async () => import("./graphql/portal/VerifyDomainScreen")
);
const DesignScreen = lazy(
  async () => import("./graphql/portal/DesignScreen/DesignScreen")
);
const LocalizationConfigurationScreen = lazy(
  async () => import("./graphql/portal/LocalizationConfigurationScreen")
);
const CustomTextConfigurationScreen = lazy(
  async () => import("./graphql/portal/CustomTextConfigurationScreen")
);
const LanguagesConfigurationScreen = lazy(
  async () => import("./graphql/portal/LanguagesConfigurationScreen")
);
const PortalAdminsSettings = lazy(
  async () => import("./graphql/portal/PortalAdminsSettings")
);
const HookConfigurationScreen = lazy(
  async () => import("./graphql/portal/HookConfigurationScreen")
);
const AdminAPIConfigurationScreen = lazy(
  async () => import("./graphql/portal/AdminAPIConfigurationScreen")
);
const LoginMethodConfigurationScreen = lazy(
  async () => import("./graphql/portal/LoginMethodConfigurationScreen")
);
const BiometricConfigurationScreen = lazy(
  async () => import("./graphql/portal/BiometricConfigurationScreen")
);
const MFAConfigurationScreen = lazy(
  async () => import("./graphql/portal/MFAConfigurationScreen")
);
const SubscriptionScreen = lazy(
  async () => import("./graphql/portal/SubscriptionScreen")
);
const LicenseScreen = lazy(
  async () => import("./graphql/portal/LicenseScreen")
);
const SMTPConfigurationScreen = lazy(
  async () => import("./graphql/portal/SMTPConfigurationScreen")
);
const SMSProviderConfigurationScreen = lazy(
  async () => import("./graphql/portal/SMSProviderConfigurationScreen")
);
const StandardAttributesConfigurationScreen = lazy(
  async () => import("./graphql/portal/StandardAttributesConfigurationScreen")
);
const CustomAttributesConfigurationScreen = lazy(
  async () => import("./graphql/portal/CustomAttributesConfigurationScreen")
);
const EditCustomAttributeScreen = lazy(
  async () => import("./graphql/portal/EditCustomAttributeScreen")
);
const CreateCustomAttributeScreen = lazy(
  async () => import("./graphql/portal/CreateCustomAttributeScreen")
);
const AccountDeletionConfigurationScreen = lazy(
  async () => import("./graphql/portal/AccountDeletionConfigurationScreen")
);
const AccountAnonymizationConfigurationScreen = lazy(
  async () => import("./graphql/portal/AccountAnonymizationConfigurationScreen")
);
const AnalyticsScreen = lazy(
  async () => import("./graphql/portal/AnalyticsScreen")
);
const IntegrationsConfigurationScreen = lazy(
  async () => import("./graphql/portal/IntegrationsConfigurationScreen")
);
const BotProtectionConfigurationScreen = lazy(
  async () => import("./graphql/portal/BotProtectionConfigurationScreen")
);
const FraudProtectionConfigurationScreen = lazy(
  async () =>
    import("./screens/fraud-protection/FraudProtectionConfigurationScreen")
);
const FraudProtectionLogEntryScreen = lazy(
  async () => import("./screens/fraud-protection/FraudProtectionLogEntryScreen")
);
const IPBlocklistScreen = lazy(
  async () => import("./graphql/portal/IPBlocklistScreen")
);
const AccountLockoutScreen = lazy(
  async () => import("./graphql/portal/AccountLockoutScreen")
);
const SubscriptionRedirect = lazy(
  async () => import("./graphql/portal/SubscriptionRedirect")
);
const EndpointDirectAccessScreen = lazy(
  async () => import("./graphql/portal/EndpointDirectAccessScreen")
);
const SAMLCertificateScreen = lazy(
  async () => import("./graphql/portal/SAMLCertificateScreen")
);
const CookieLifetimeConfigurationScreen = lazy(
  async () => import("./graphql/portal/CookieLifetimeConfigurationScreen")
);
const EditConfigurationScreen = lazy(
  async () => import("./graphql/portal/EditConfigurationScreen")
);
const APIResourcesScreen = lazy(
  async () => import("./screens/api-resources/APIResourcesScreen")
);
const CreateAPIResourceScreen = lazy(
  async () => import("./screens/api-resources/CreateAPIResourceScreen")
);
const APIResourceDetailsScreen = lazy(
  async () => import("./screens/api-resources/APIResourceDetailsScreen")
);
const EditScopeScreen = React.lazy(
  async () => import("./screens/api-resources/EditScopeScreen")
);
const EditApplicationScopesScreen = lazy(
  async () => import("./screens/api-resources/EditApplicationScopesScreen")
);
const EditOAuthClientResourceScopeScreen = lazy(
  async () =>
    import("./screens/api-resources/EditOAuthClientResourceScopeScreen")
);
const DynamicClientListScreen = lazy(
  async () => import("./screens/dynamic-clients/DynamicClientListScreen")
);
const ClientApplicationsScreen = lazy(
  async () => import("./screens/dynamic-clients/ClientApplicationsScreen")
);
const AIAgentsScreen = lazy(
  async () => import("./screens/dynamic-clients/AIAgentsScreen")
);

const AppRoot: React.VFC = function AppRoot() {
  const { appID } = useParams() as { appID: string };
  const { setDisplayUnauthenticatedDialog } = useUnauthenticatedDialogContext();
  const { isAuthgearOnce } = useSystemConfig();
  const portalClient = usePortalClient();
  const client = useMemo(() => {
    const onLogout = () => {
      setDisplayUnauthenticatedDialog(true);
    };
    return makeClient(appID, onLogout);
  }, [appID, setDisplayUnauthenticatedDialog]);

  // NOTE: check if appID actually exist in authorized app list
  const {
    effectiveAppConfig,
    isLoading: loading,
    loadError: error,
  } = useAppAndSecretConfigQuery(appID);

  const screenNavQuery = useQuery<ScreenNavQueryQuery>(ScreenNavQueryDocument, {
    client: portalClient,
    variables: {
      id: appID,
    },
  });
  const projectWizardData =
    screenNavQuery.data?.node?.__typename === "App"
      ? screenNavQuery.data.node.tutorialStatus.data.project_wizard
      : null;

  if (loading || screenNavQuery.loading) {
    return <ShowLoading />;
  }

  // if node is null after loading without error, treat this as invalid
  // request, frontend cannot distinguish between inaccessible and not found
  const isInvalidAppID = error == null && effectiveAppConfig == null;

  // redirect to app list if app id is invalid
  if (isInvalidAppID) {
    return <Navigate to="/projects" replace={true} />;
  }

  /* Handle project wizard redirection */
  // Continue project wizard if it is not completed
  // null means completed
  // undefined means not started
  // any other value means in-progress

  // Project wizard in-progress, continue
  if (projectWizardData != null && !projectWizardData.completed) {
    return (
      <Navigate
        to={`/project/${encodeURIComponent(appID)}/wizard`}
        replace={true}
      />
    );
  }

  // In ONCE, we always force the user to complete the wizard if it is not completed or not started
  if (projectWizardData == null && isAuthgearOnce) {
    return (
      <Navigate
        to={`/project/${encodeURIComponent(appID)}/wizard`}
        replace={true}
      />
    );
  }

  // In other cases, skip the wizard if it is not started

  return (
    <ApolloProvider client={client}>
      <ScreenLayout>
        <AppRoutes>
          <AppRoute
            feature={null}
            index={true}
            element={
              <Suspense fallback={<ShowLoading />}>
                <ProjectRootScreen />
              </Suspense>
            }
          />

          <AppRoute feature={null} path="getting-started">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <GetStartedScreen />
                </Suspense>
              }
            />
          </AppRoute>

          <AppRoute feature="analytics" path="analytics">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <AnalyticsScreen />
                </Suspense>
              }
            />
          </AppRoute>
          <AppRoute
            feature={null}
            path="users/*"
            element={
              <Suspense fallback={<ShowLoading />}>
                <UsersRedirectScreen />
              </Suspense>
            }
          ></AppRoute>

          <AppRoute feature={null} path="user-management">
            <AppRoute
              feature={null}
              index={true}
              element={<Navigate to="users" replace={true} />}
            />

            <AppRoute feature={null} path="roles">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <RolesScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="add-role"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AddRoleScreen />
                  </Suspense>
                }
              />
              <AppRoute feature={null} path=":roleID" element={<RequireRole />}>
                <AppRoute
                  feature={null}
                  index={true}
                  element={<Navigate to="details" replace={true} />}
                />
                <AppRoute feature={null} path="details">
                  <AppRoute
                    feature={null}
                    index={true}
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <RoleDetailsScreen />
                      </Suspense>
                    }
                  />
                </AppRoute>
              </AppRoute>
            </AppRoute>

            <AppRoute feature={null} path="groups">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <GroupsScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="add-group"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AddGroupScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path=":groupID"
                element={<RequireGroup />}
              >
                <AppRoute
                  feature={null}
                  index={true}
                  element={<Navigate to="details" replace={true} />}
                />
                <AppRoute feature={null} path="details">
                  <AppRoute
                    feature={null}
                    index={true}
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <GroupDetailsScreen />
                      </Suspense>
                    }
                  />
                </AppRoute>
              </AppRoute>
            </AppRoute>

            <AppRoute feature={null} path="users">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <UsersScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="add-user"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AddUserScreen />
                  </Suspense>
                }
              />
              <AppRoute feature={null} path=":userID" element={<RequireUser />}>
                <AppRoute
                  feature={null}
                  index={true}
                  element={<Navigate to="details" replace={true} />}
                />
                <AppRoute feature={null} path="details">
                  <AppRoute
                    feature={null}
                    index={true}
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <UserDetailsScreen />
                      </Suspense>
                    }
                  />
                  <AppRoute
                    feature={null}
                    path="change-password"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <ChangePasswordScreen />
                      </Suspense>
                    }
                  />
                  <AppRoute
                    feature={null}
                    path="add-password"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <AddPasswordScreen />
                      </Suspense>
                    }
                  />
                </AppRoute>
              </AppRoute>
            </AppRoute>
          </AppRoute>

          <AppRoute feature={null} path="api-resources">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <APIResourcesScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="create"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <CreateAPIResourceScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path=":resourceID"
              element={<RequireAPIResource />}
            >
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <APIResourceDetailsScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="scopes/:scopeID"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <EditScopeScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="applications/:clientID/scopes"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <EditApplicationScopesScreen />
                  </Suspense>
                }
              />
            </AppRoute>
          </AppRoute>

          <AppRoute feature={null} path="branding">
            <AppRoute
              feature={null}
              index={true}
              element={<Navigate to="design" replace={true} />}
            />
            <AppRoute
              feature={null}
              path="design"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <DesignScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="localization"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <LocalizationConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute feature={null} path="custom-domains">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <CustomDomainListScreen />
                  </Suspense>
                }
              />
              <AppRoute feature={null} path=":domainID">
                <AppRoute
                  feature={null}
                  index={true}
                  element={<Navigate to="verify" replace={true} />}
                />
                <AppRoute
                  feature={null}
                  path="verify"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <VerifyDomainScreen />
                    </Suspense>
                  }
                />
              </AppRoute>
            </AppRoute>
            <AppRoute
              feature={null}
              path="custom-text"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <CustomTextConfigurationScreen />
                </Suspense>
              }
            />
          </AppRoute>

          <AppRoute feature={null} path="configuration">
            <AppRoute
              feature={null}
              index={true}
              element={<Navigate to="authentication" replace={true} />}
            />
            <AppRoute feature={null} path="authentication">
              <AppRoute
                feature={null}
                index={true}
                element={<Navigate to="login-methods" replace={true} />}
              />
              <AppRoute
                feature={null}
                path="login-methods"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <LoginMethodConfigurationScreen />
                  </Suspense>
                }
              />
              <AppRoute feature={null} path="external-oauth">
                <AppRoute
                  feature={null}
                  index={true}
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <SingleSignOnConfigurationScreen />
                    </Suspense>
                  }
                />
                <AppRoute
                  feature={null}
                  path="add"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <AddSingleSignOnConfigurationScreen />
                    </Suspense>
                  }
                />
                <AppRoute
                  feature={null}
                  path="edit/:provider/:alias"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <EditSingleSignOnConfigurationScreen />
                    </Suspense>
                  }
                />
              </AppRoute>
              <AppRoute
                feature={null}
                path="biometric"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <BiometricConfigurationScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="2fa"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <MFAConfigurationScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="anonymous-users"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AnonymousUsersConfigurationScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                element={<RequireAppFeature isAvailable={isApp2AppAvailable} />}
              >
                <AppRoute
                  feature={null}
                  path="app2app"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <App2AppConfigurationScreen />
                    </Suspense>
                  }
                />
              </AppRoute>
            </AppRoute>
            <AppRoute feature={null} path="apps">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <ClientApplicationsScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="agents"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AIAgentsScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="add"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <CreateOAuthClientScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="add-m2m"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <CreateM2MClientScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="dynamic-clients"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <DynamicClientListScreen />
                  </Suspense>
                }
              />
              <AppRoute feature={null} path=":clientID">
                <AppRoute
                  feature={null}
                  index={true}
                  element={<Navigate to="edit" replace={true} />}
                />
                <AppRoute
                  feature={null}
                  path="edit"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <EditOAuthClientScreen />
                    </Suspense>
                  }
                />
                <AppRoute
                  feature={null}
                  path="edit/resource/:resourceID/scopes"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <EditOAuthClientResourceScopeScreen />
                    </Suspense>
                  }
                />
              </AppRoute>
            </AppRoute>
            <AppRoute
              feature={null}
              path="languages"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <LanguagesConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute feature={null} path="user-profile">
              <AppRoute
                feature={null}
                index={true}
                element={<Navigate to="standard-attributes" replace={true} />}
              />
              <AppRoute
                feature={null}
                path="standard-attributes"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <StandardAttributesConfigurationScreen />
                  </Suspense>
                }
              />
              <AppRoute feature={null} path="custom-attributes">
                <AppRoute
                  feature={null}
                  index={true}
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <CustomAttributesConfigurationScreen />
                    </Suspense>
                  }
                />
                <AppRoute
                  feature={null}
                  path="add"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <CreateCustomAttributeScreen />
                    </Suspense>
                  }
                />
                <AppRoute feature={null} path=":index">
                  <AppRoute
                    feature={null}
                    index={true}
                    element={<Navigate to="edit" replace={true} />}
                  />
                  <AppRoute
                    feature={null}
                    path="edit"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <EditCustomAttributeScreen />
                      </Suspense>
                    }
                  />
                </AppRoute>
              </AppRoute>
            </AppRoute>
          </AppRoute>

          <AppRoute feature={null} path="attack-protection">
            <AppRoute feature={null} path="account-lockout">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AccountLockoutScreen />
                  </Suspense>
                }
              />
            </AppRoute>
            <AppRoute feature={null} path="bot-protection">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <BotProtectionConfigurationScreen />
                  </Suspense>
                }
              />
            </AppRoute>
            <AppRoute feature="fraudProtection" path="fraud-protection">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <FraudProtectionConfigurationScreen />
                  </Suspense>
                }
              />
              <AppRoute
                feature={null}
                path="logs/:logID"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <FraudProtectionLogEntryScreen />
                  </Suspense>
                }
              />
            </AppRoute>
            <AppRoute feature={null} path="ip-blocklist">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <IPBlocklistScreen />
                  </Suspense>
                }
              />
            </AppRoute>
          </AppRoute>

          <AppRoute
            feature={null}
            element={
              <RequireAppFeature isAvailable={isIntegrationsAvailable} />
            }
          >
            <AppRoute feature={null} path="integrations">
              <AppRoute
                feature={null}
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <IntegrationsConfigurationScreen />
                  </Suspense>
                }
              />
            </AppRoute>
          </AppRoute>

          <AppRoute feature={null} path="billing">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <SubscriptionScreen />
                </Suspense>
              }
            />
          </AppRoute>

          <AppRoute feature={null} path="license">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <LicenseScreen />
                </Suspense>
              }
            />
          </AppRoute>

          <AppRoute feature={null} path="billing-redirect">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <SubscriptionRedirect />
                </Suspense>
              }
            />
          </AppRoute>

          <AppRoute feature={null} path="advanced">
            <AppRoute
              feature={null}
              index={true}
              element={<Navigate to="password-reset-code" replace={true} />}
            />
            <AppRoute
              feature={null}
              path="hooks"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <HookConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="admin-api"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <AdminAPIConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="account-deletion"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <AccountDeletionConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="account-anonymization"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <AccountAnonymizationConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="session"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <CookieLifetimeConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="smtp"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <SMTPConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="sms-gateway"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <SMSProviderConfigurationScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="endpoint-direct-access"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <EndpointDirectAccessScreen />
                </Suspense>
              }
            />
            <AppRoute
              feature={null}
              path="saml-certificate"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <SAMLCertificateScreen />
                </Suspense>
              }
            />
          </AppRoute>

          <AppRoute feature="auditLog" path="audit-log">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <AuditLogScreen />
                </Suspense>
              }
            />
            <AppRoute feature={null} path=":logID">
              <AppRoute
                feature={null}
                index={true}
                element={<Navigate to="details" replace={true} />}
              />
              <AppRoute
                feature={null}
                path="details"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AuditLogEntryScreen />
                  </Suspense>
                }
              />
            </AppRoute>
          </AppRoute>

          <AppRoute feature={null} path="portal-admins">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <PortalAdminsSettings />
                </Suspense>
              }
            />
          </AppRoute>

          {/* This screen is not shown in nav bar, which is intentional to prevent normal users from accessing it */}
          <AppRoute feature={null} path="edit-config">
            <AppRoute
              feature={null}
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <EditConfigurationScreen />
                </Suspense>
              }
            />
          </AppRoute>
        </AppRoutes>
      </ScreenLayout>
    </ApolloProvider>
  );
};

export default AppRoot;
