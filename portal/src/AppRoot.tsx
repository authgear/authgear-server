import React, { useMemo, lazy, Suspense } from "react";
import { Routes, Route, useParams, Navigate } from "react-router-dom";
import { ApolloProvider } from "@apollo/client";

import { makeClient } from "./graphql/adminapi/apollo";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import ScreenLayout from "./ScreenLayout";
import ShowLoading from "./ShowLoading";
import CookieLifetimeConfigurationScreen from "./graphql/portal/CookieLifetimeConfigurationScreen";
import { useUnauthenticatedDialogContext } from "./components/auth/UnauthenticatedDialogContext";

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
const EmailScreen = lazy(async () => import("./graphql/adminapi/EmailScreen"));
const PhoneScreen = lazy(async () => import("./graphql/adminapi/PhoneScreen"));
const UsernameScreen = lazy(
  async () => import("./graphql/adminapi/UsernameScreen")
);
const ResetPasswordScreen = lazy(
  async () => import("./graphql/adminapi/ResetPasswordScreen")
);
const EditPictureScreen = lazy(
  async () => import("./graphql/adminapi/EditPictureScreen")
);
const Add2FAScreen = lazy(
  async () => import("./graphql/adminapi/Add2FAScreen")
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
const App2AppConfigurationScreen = lazy(
  async () => import("./graphql/portal/App2AppConfigurationScreen")
);
const SingleSignOnConfigurationScreen = lazy(
  async () => import("./graphql/portal/SingleSignOnConfigurationScreen")
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
const DesignScreen = lazy(async () => import("./graphql/portal/DesignScreen"));
const LocalizationConfigurationScreen = lazy(
  async () => import("./graphql/portal/LocalizationConfigurationScreen")
);
const CustomTextConfigurationScreen = lazy(
  async () => import("./graphql/portal/CustomTextConfigurationScreen")
);
const LanguagesConfigurationScreen = lazy(
  async () => import("./graphql/portal/LanguagesConfigurationScreen")
);
const InviteAdminScreen = lazy(
  async () => import("./graphql/portal/InviteAdminScreen")
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
const Web3ConfigurationScreen = lazy(
  async () => import("./graphql/portal/Web3ConfigurationScreen")
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
const AccountAnonymizationConfigurationScreen = lazy(
  async () => import("./graphql/portal/AccountAnonymizationConfigurationScreen")
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
  const { setDisplayUnauthenticatedDialog } = useUnauthenticatedDialogContext();
  const client = useMemo(() => {
    const onLogout = () => {
      setDisplayUnauthenticatedDialog(true);
    };
    return makeClient(appID, onLogout);
  }, [appID, setDisplayUnauthenticatedDialog]);

  // NOTE: check if appID actually exist in authorized app list
  const { effectiveAppConfig, loading, error } =
    useAppAndSecretConfigQuery(appID);

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

  const useAuthUIV2 = effectiveAppConfig?.ui?.implementation === "authflowv2";

  return (
    <ApolloProvider client={client}>
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
          <Route
            path="users/*"
            element={
              <Suspense fallback={<ShowLoading />}>
                <UsersRedirectScreen />
              </Suspense>
            }
          ></Route>

          <Route path="user-management">
            <Route path="roles">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <RolesScreen />
                  </Suspense>
                }
              />
              <Route
                path="add-role"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AddRoleScreen />
                  </Suspense>
                }
              />
              <Route path=":roleID">
                <Route
                  index={true}
                  element={<Navigate to="details" replace={true} />}
                />
                <Route path="details">
                  <Route
                    index={true}
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <RoleDetailsScreen />
                      </Suspense>
                    }
                  />
                </Route>
              </Route>
            </Route>

            <Route path="groups">
              <Route
                index={true}
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <GroupsScreen />
                  </Suspense>
                }
              />
              <Route
                path="add-group"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <AddGroupScreen />
                  </Suspense>
                }
              />
              <Route path=":groupID">
                <Route
                  index={true}
                  element={<Navigate to="details" replace={true} />}
                />
                <Route path="details">
                  <Route
                    index={true}
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <GroupDetailsScreen />
                      </Suspense>
                    }
                  />
                </Route>
              </Route>
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
                        <EmailScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="edit-email/:identityID"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <EmailScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="add-phone"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <PhoneScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="edit-phone/:identityID"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <PhoneScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="add-username"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <UsernameScreen />
                      </Suspense>
                    }
                  />
                  <Route
                    path="edit-username/:identityID"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <UsernameScreen />
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
                  <Route
                    path="add-2fa-phone"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <Add2FAScreen authenticatorType="oob_otp_sms" />
                      </Suspense>
                    }
                  />
                  <Route
                    path="add-2fa-email"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <Add2FAScreen authenticatorType="oob_otp_email" />
                      </Suspense>
                    }
                  />
                  <Route
                    path="add-2fa-password"
                    element={
                      <Suspense fallback={<ShowLoading />}>
                        <Add2FAScreen authenticatorType="password" />
                      </Suspense>
                    }
                  />
                </Route>
              </Route>
            </Route>
          </Route>

          <Route path="branding">
            <Route
              index={true}
              element={
                <Navigate
                  to={useAuthUIV2 ? "design" : "ui-settings"}
                  replace={true}
                />
              }
            />
            {useAuthUIV2 ? (
              <Route
                path="design"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <DesignScreen />
                  </Suspense>
                }
              />
            ) : (
              <Route
                path="ui-settings"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <UISettingsScreen />
                  </Suspense>
                }
              />
            )}
            <Route
              path="localization"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <LocalizationConfigurationScreen />
                </Suspense>
              }
            />
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
            <Route
              path="custom-text"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <CustomTextConfigurationScreen />
                </Suspense>
              }
            />
          </Route>

          <Route path="configuration">
            <Route
              index={true}
              element={<Navigate to="authentication" replace={true} />}
            />
            <Route path="authentication">
              <Route
                index={true}
                element={<Navigate to="login-methods" replace={true} />}
              />
              <Route
                path="login-methods"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <LoginMethodConfigurationScreen />
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
              <Route
                path="app2app"
                element={
                  <Suspense fallback={<ShowLoading />}>
                    <App2AppConfigurationScreen />
                  </Suspense>
                }
              />
            </Route>
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
              path="languages"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <LanguagesConfigurationScreen />
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
              path="hooks"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <HookConfigurationScreen />
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
              path="account-anonymization"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <AccountAnonymizationConfigurationScreen />
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
            <Route
              path="smtp"
              element={
                <Suspense fallback={<ShowLoading />}>
                  <SMTPConfigurationScreen />
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
    </ApolloProvider>
  );
};

export default AppRoot;
