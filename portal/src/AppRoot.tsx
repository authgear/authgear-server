import React, { useMemo, lazy, Suspense } from "react";
import { Routes, Route, useParams, Navigate } from "react-router-dom";
import { ApolloProvider } from "@apollo/client";

import { makeClient } from "./graphql/adminapi/apollo";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import ScreenLayout from "./ScreenLayout";
import ShowLoading from "./ShowLoading";
import CookieLifetimeConfigurationScreen from "./graphql/portal/CookieLifetimeConfigurationScreen";
import { AppContext } from "./context/AppContext";

const UsersScreen = lazy(async () => import("./graphql/adminapi/UsersScreen"));
const AddUserScreen = lazy(
  async () => import("./graphql/adminapi/AddUserScreen")
);
const UserDetailsScreen = lazy(
  async () => import("./graphql/adminapi/UserDetailsScreen")
);
const AddEmailScreen = lazy(
  async () => import("./graphql/adminapi/AddEmailScreen")
);
const AddPhoneScreen = lazy(
  async () => import("./graphql/adminapi/AddPhoneScreen")
);
const AddUsernameScreen = lazy(
  async () => import("./graphql/adminapi/AddUsernameScreen")
);
const ResetPasswordScreen = lazy(
  async () => import("./graphql/adminapi/ResetPasswordScreen")
);
const EditPictureScreen = lazy(
  async () => import("./graphql/adminapi/EditPictureScreen")
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
  async () => import("./graphql/portal/GetStartedScreen")
);
const AnonymousUsersConfigurationScreen = lazy(
  async () => import("./graphql/portal/AnonymousUsersConfigurationScreen")
);
const SingleSignOnConfigurationScreen = lazy(
  async () => import("./graphql/portal/SingleSignOnConfigurationScreen")
);
const PasswordPolicyConfigurationScreen = lazy(
  async () => import("./graphql/portal/PasswordPolicyConfigurationScreen")
);
const ForgotPasswordConfigurationScreen = lazy(
  async () => import("./graphql/portal/ForgotPasswordConfigurationScreen")
);
const ApplicationsConfigurationScreen = lazy(
  async () => import("./graphql/portal/ApplicationsConfigurationScreen")
);
const CreateOAuthClientScreen = lazy(
  async () => import("./graphql/portal/CreateOAuthClientScreen")
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
const UISettingsScreen = lazy(
  async () => import("./graphql/portal/UISettingsScreen")
);
const LocalizationConfigurationScreen = lazy(
  async () => import("./graphql/portal/LocalizationConfigurationScreen")
);
const InviteAdminScreen = lazy(
  async () => import("./graphql/portal/InviteAdminScreen")
);
const PortalAdminsSettings = lazy(
  async () => import("./graphql/portal/PortalAdminsSettings")
);
const WebhookConfigurationScreen = lazy(
  async () => import("./graphql/portal/WebhookConfigurationScreen")
);
const AdminAPIConfigurationScreen = lazy(
  async () => import("./graphql/portal/AdminAPIConfigurationScreen")
);
const LoginIDConfigurationScreen = lazy(
  async () => import("./graphql/portal/LoginIDConfigurationScreen")
);
const AuthenticatorConfigurationScreen = lazy(
  async () => import("./graphql/portal/AuthenticatorConfigurationScreen")
);
const VerificationConfigurationScreen = lazy(
  async () => import("./graphql/portal/VerificationConfigurationScreen")
);
const Web3ConfigurationScreen = lazy(
  async () => import("./graphql/portal/Web3ConfigurationScreen")
);
const PasskeyConfigurationScreen = lazy(
  async () => import("./graphql/portal/PasskeyConfigurationScreen")
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
const SMTPConfigurationScreen = lazy(
  async () => import("./graphql/portal/SMTPConfigurationScreen")
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
const AnalyticsScreen = lazy(
  async () => import("./graphql/portal/AnalyticsScreen")
);
const IntegrationsConfigurationScreen = lazy(
  async () => import("./graphql/portal/IntegrationsConfigurationScreen")
);
const GoogleTagManagerConfigurationScreen = lazy(
  async () => import("./graphql/portal/GoogleTagManagerConfigurationScreen")
);
const SubscriptionRedirect = lazy(
  async () => import("./graphql/portal/SubscriptionRedirect")
);

const AppRoot: React.VFC = function AppRoot() {
  const { appID } = useParams() as { appID: string };
  const client = useMemo(() => {
    return makeClient(appID);
  }, [appID]);

  // NOTE: check if appID actually exist in authorized app list
  const { effectiveAppConfig, viewer, loading, error } =
    useAppAndSecretConfigQuery(appID);

  const appContextValue = useMemo(() => {
    return {
      appID: effectiveAppConfig?.id ?? "",
      currentCollaboratorRole: viewer?.role ?? "",
    };
  }, [effectiveAppConfig, viewer?.role]);

  if (loading) {
    return <ShowLoading />;
  }

  // if node is null after loading without error, treat this as invalid
  // request, frontend cannot distinguish between inaccessible and not found
  const isInvalidAppID = error == null && effectiveAppConfig == null;

  // redirect to app list if app id is invalid
  if (isInvalidAppID) {
    return <Navigate to="/projects" replace={true} />;
  }

  return (
    <ApolloProvider client={client}>
      <AppContext.Provider value={appContextValue}>
        <ScreenLayout>
          <Routes>
            <Route
              index={true}
              element={
                <Suspense fallback={<ShowLoading />}>
                  <ProjectRootScreen />
                </Suspense>
              }
            />

            <Route path="getting-started">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <GetStartedScreen />
                  </Suspense>
                }
              />
            </Route>

            <Route path="analytics">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AnalyticsScreen />
                  </Suspense>
                }
              />
            </Route>

            <Route path="users">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <UsersScreen />
                  </Suspense>
                }
              />
              <Route
                path="add-user"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AddUserScreen />
                  </Suspense>
                }
              />
              <Route path=":userID">
                <Route
                  index={true}
                  element={<Navigate to="details" replace={true} />}
                />
                <Route path="details">
                  <Route
                    index={true}
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <UserDetailsScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="add-email"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <AddEmailScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="add-phone"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <AddPhoneScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="add-username"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <AddUsernameScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="reset-password"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <ResetPasswordScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="edit-picture"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <EditPictureScreen />
                      </Suspense>
                    }
                  />
                </Route>
              </Route>
            </Route>

            <Route path="custom-domains">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <CustomDomainListScreen />
                  </Suspense>
                }
              />
              <Route path=":domainID">
                <Route
                  index={true}
                  element={<Navigate to="verify" replace={true} />}
                />
                <Route
                  path="verify"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <VerifyDomainScreen />
                    </Suspense>
                  }
                />
              </Route>
            </Route>

            <Route path="configuration">
              <Route
                index={true}
                element={<Navigate to="authentication" replace={true} />}
              />
              <Route path="authentication">
                <Route
                  index={true}
                  element={<Navigate to="login-id" replace={true} />}
                />
                <Route
                  path="login-id"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <LoginIDConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="authenticators"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <AuthenticatorConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="verification"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <VerificationConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="external-oauth"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <SingleSignOnConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="passkey"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <PasskeyConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="biometric"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <BiometricConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="2fa"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <MFAConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="web3"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <Web3ConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="anonymous-users"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <AnonymousUsersConfigurationScreen />
                    </Suspense>
                  }
                />
              </Route>
              <Route
                path="password-policy"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <PasswordPolicyConfigurationScreen />
                  </Suspense>
                }
              />
              <Route path="apps">
                <Route
                  index={true}
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <ApplicationsConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route
                  path="add"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <CreateOAuthClientScreen />
                    </Suspense>
                  }
                />
                <Route path=":clientID">
                  <Route
                    index={true}
                    element={<Navigate to="edit" replace={true} />}
                  />
                  <Route
                    path="edit"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <EditOAuthClientScreen />
                      </Suspense>
                    }
                  />
                </Route>
              </Route>
              <Route
                path="smtp"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <SMTPConfigurationScreen />
                  </Suspense>
                }
              />
              <Route
                path="ui-settings"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <UISettingsScreen />
                  </Suspense>
                }
              />
              <Route
                path="localization"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <LocalizationConfigurationScreen />
                  </Suspense>
                }
              />
              <Route path="user-profile">
                <Route
                  index={true}
                  element={<Navigate to="standard-attributes" replace={true} />}
                />
                <Route
                  path="standard-attributes"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <StandardAttributesConfigurationScreen />
                    </Suspense>
                  }
                />
                <Route path="custom-attributes">
                  <Route
                    index={true}
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <CustomAttributesConfigurationScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="add"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <CreateCustomAttributeScreen />
                      </Suspense>
                    }
                  />
                  <Route path=":index">
                    <Route
                      index={true}
                      element={<Navigate to="edit" replace={true} />}
                    />
                    <Route
                      path="edit"
                      element={
                        <Suspense fallback={<ShowLoading />}>
                          <EditCustomAttributeScreen />
                        </Suspense>
                      }
                    />
                  </Route>
                </Route>
              </Route>
            </Route>

            <Route path="integrations">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <IntegrationsConfigurationScreen />
                  </Suspense>
                }
              />
              <Route
                path="google-tag-manager"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <GoogleTagManagerConfigurationScreen />
                  </Suspense>
                }
              />
            </Route>

            <Route path="billing">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <SubscriptionScreen />
                  </Suspense>
                }
              />
            </Route>

            <Route path="billing-redirect">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <SubscriptionRedirect />
                  </Suspense>
                }
              />
            </Route>

            <Route path="advanced">
              <Route
                index={true}
                element={<Navigate to="password-reset-code" replace={true} />}
              />
              <Route
                path="password-reset-code"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <ForgotPasswordConfigurationScreen />
                  </Suspense>
                }
              />
              <Route
                path="webhooks"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <WebhookConfigurationScreen />
                  </Suspense>
                }
              />
              <Route
                path="admin-api"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AdminAPIConfigurationScreen />
                  </Suspense>
                }
              />
              <Route
                path="account-deletion"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AccountDeletionConfigurationScreen />
                  </Suspense>
                }
              />
              <Route
                path="session"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <CookieLifetimeConfigurationScreen />
                  </Suspense>
                }
              />
            </Route>

            <Route path="audit-log">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AuditLogScreen />
                  </Suspense>
                }
              />
              <Route path=":logID">
                <Route
                  index={true}
                  element={<Navigate to="details" replace={true} />}
                />
                <Route
                  path="details"
                  element={
                    <Suspense fallback={<ShowLoading />}>
                      <AuditLogEntryScreen />
                    </Suspense>
                  }
                />
              </Route>
            </Route>

            <Route path="portal-admins">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <PortalAdminsSettings />
                  </Suspense>
                }
              />
              <Route
                path="invite"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <InviteAdminScreen />
                  </Suspense>
                }
              />
            </Route>
          </Routes>
        </ScreenLayout>
      </AppContext.Provider>
    </ApolloProvider>
  );
};

export default AppRoot;
