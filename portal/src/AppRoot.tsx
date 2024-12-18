import React, { useMemo } from "react";
import { Routes, Route, useParams, Navigate } from "react-router-dom";
import { ApolloProvider } from "@apollo/client";

import { makeClient } from "./graphql/adminapi/apollo";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import ScreenLayout from "./ScreenLayout";
import ShowLoading from "./ShowLoading";
import { useUnauthenticatedDialogContext } from "./components/auth/UnauthenticatedDialogContext";
import { useUIImplementation } from "./hook/useUIImplementation";
import FlavoredErrorBoundSuspense from "./FlavoredErrorBoundSuspense";

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

  const uiImplementation = useUIImplementation(
    effectiveAppConfig?.ui?.implementation
  );

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

  const useAuthUIV2 = uiImplementation === "authflowv2";

  return (
    <ApolloProvider client={client}>
      <ScreenLayout>
        <Routes>
          <Route
            index={true}
            element={
              <FlavoredErrorBoundSuspense
                factory={async () =>
                  import("./graphql/portal/ProjectRootScreen")
                }
              >
                {(ProjectRootScreen) => <ProjectRootScreen />}
              </FlavoredErrorBoundSuspense>
            }
          />

          <Route path="getting-started">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/GetStartedScreen")
                  }
                >
                  {(GetStartedScreen) => <GetStartedScreen />}
                </FlavoredErrorBoundSuspense>
              }
            />
          </Route>

          <Route path="analytics">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/AnalyticsScreen")
                  }
                >
                  {(AnalyticsScreen) => <AnalyticsScreen />}
                </FlavoredErrorBoundSuspense>
              }
            />
          </Route>
          <Route
            path="users/*"
            element={
              <FlavoredErrorBoundSuspense
                factory={async () => import("./UsersRedirectScreen")}
              >
                {(UsersRedirectScreen) => <UsersRedirectScreen />}
              </FlavoredErrorBoundSuspense>
            }
          ></Route>

          <Route path="user-management">
            <Route path="roles">
              <Route
                index={true}
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/adminapi/RolesScreen")
                    }
                  >
                    {(RolesScreen) => <RolesScreen />}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route
                path="add-role"
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/adminapi/AddRoleScreen")
                    }
                  >
                    {(AddRoleScreen) => <AddRoleScreen />}
                  </FlavoredErrorBoundSuspense>
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
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/RoleDetailsScreen")
                        }
                      >
                        {(RoleDetailsScreen) => <RoleDetailsScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                </Route>
              </Route>
            </Route>

            <Route path="groups">
              <Route
                index={true}
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/adminapi/GroupsScreen")
                    }
                  >
                    {(GroupsScreen) => <GroupsScreen />}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route
                path="add-group"
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/adminapi/AddGroupScreen")
                    }
                  >
                    {(AddGroupScreen) => <AddGroupScreen />}
                  </FlavoredErrorBoundSuspense>
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
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/GroupDetailsScreen")
                        }
                      >
                        {(GroupDetailsScreen) => <GroupDetailsScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                </Route>
              </Route>
            </Route>

            <Route path="users">
              <Route
                index={true}
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/adminapi/UsersScreen")
                    }
                  >
                    {(UsersScreen) => <UsersScreen />}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route
                path="add-user"
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/adminapi/AddUserScreen")
                    }
                  >
                    {(AddUserScreen) => <AddUserScreen />}
                  </FlavoredErrorBoundSuspense>
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
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/UserDetailsScreen")
                        }
                      >
                        {(UserDetailsScreen) => <UserDetailsScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="add-email"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/EmailScreen")
                        }
                      >
                        {(EmailScreen) => <EmailScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="edit-email/:identityID"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/EmailScreen")
                        }
                      >
                        {(EmailScreen) => <EmailScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="add-phone"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/PhoneScreen")
                        }
                      >
                        {(PhoneScreen) => <PhoneScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="edit-phone/:identityID"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/PhoneScreen")
                        }
                      >
                        {(PhoneScreen) => <PhoneScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="add-username"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/UsernameScreen")
                        }
                      >
                        {(UsernameScreen) => <UsernameScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="edit-username/:identityID"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/UsernameScreen")
                        }
                      >
                        {(UsernameScreen) => <UsernameScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="change-password"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/ChangePasswordScreen")
                        }
                      >
                        {(ChangePasswordScreen) => <ChangePasswordScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="edit-picture"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/EditPictureScreen")
                        }
                      >
                        {(EditPictureScreen) => <EditPictureScreen />}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="add-2fa-phone"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/Add2FAScreen")
                        }
                      >
                        {(Add2FAScreen) => (
                          <Add2FAScreen authenticatorType="oob_otp_sms" />
                        )}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="add-2fa-email"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/Add2FAScreen")
                        }
                      >
                        {(Add2FAScreen) => (
                          <Add2FAScreen authenticatorType="oob_otp_email" />
                        )}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                  <Route
                    path="add-2fa-password"
                    element={
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/adminapi/Add2FAScreen")
                        }
                      >
                        {(Add2FAScreen) => (
                          <Add2FAScreen authenticatorType="password" />
                        )}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                </Route>
              </Route>
            </Route>
          </Route>

          <Route path="branding">
            <Route
              index={true}
              element={<Navigate to="design" replace={true} />}
            />
            <Route
              path="design"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/UISettingsScreen")
                  }
                >
                  {(UISettingsScreen) => (
                    <FlavoredErrorBoundSuspense
                      factory={async () =>
                        import("./graphql/portal/DesignScreen/DesignScreen")
                      }
                    >
                      {(DesignScreen) =>
                        useAuthUIV2 ? <DesignScreen /> : <UISettingsScreen />
                      }
                    </FlavoredErrorBoundSuspense>
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="localization"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/LocalizationConfigurationScreen")
                  }
                >
                  {(LocalizationConfigurationScreen) => (
                    <LocalizationConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route path="custom-domains">
              <Route
                index={true}
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/portal/CustomDomainListScreen")
                    }
                  >
                    {(CustomDomainListScreen) => <CustomDomainListScreen />}
                  </FlavoredErrorBoundSuspense>
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
                    <FlavoredErrorBoundSuspense
                      factory={async () =>
                        import("./graphql/portal/VerifyDomainScreen")
                      }
                    >
                      {(VerifyDomainScreen) => <VerifyDomainScreen />}
                    </FlavoredErrorBoundSuspense>
                  }
                />
              </Route>
            </Route>
            <Route
              path="custom-text"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/CustomTextConfigurationScreen")
                  }
                >
                  {(CustomTextConfigurationScreen) => (
                    <CustomTextConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
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
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/portal/LoginMethodConfigurationScreen")
                    }
                  >
                    {(LoginMethodConfigurationScreen) => (
                      <LoginMethodConfigurationScreen />
                    )}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route path="external-oauth">
                <Route
                  index={true}
                  element={
                    <FlavoredErrorBoundSuspense
                      factory={async () =>
                        import(
                          "./graphql/portal/SingleSignOnConfigurationScreen"
                        )
                      }
                    >
                      {(SingleSignOnConfigurationScreen) => (
                        <SingleSignOnConfigurationScreen />
                      )}
                    </FlavoredErrorBoundSuspense>
                  }
                />
                <Route
                  path="add"
                  element={
                    <FlavoredErrorBoundSuspense
                      factory={async () =>
                        import(
                          "./graphql/portal/AddSingleSignOnConfigurationScreen"
                        )
                      }
                    >
                      {(AddSingleSignOnConfigurationScreen) => (
                        <AddSingleSignOnConfigurationScreen />
                      )}
                    </FlavoredErrorBoundSuspense>
                  }
                />
                <Route
                  path="edit/:provider"
                  element={
                    <FlavoredErrorBoundSuspense
                      factory={async () =>
                        import(
                          "./graphql/portal/EditSingleSignOnConfigurationScreen"
                        )
                      }
                    >
                      {(EditSingleSignOnConfigurationScreen) => (
                        <EditSingleSignOnConfigurationScreen />
                      )}
                    </FlavoredErrorBoundSuspense>
                  }
                />
              </Route>
              <Route
                path="biometric"
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/portal/BiometricConfigurationScreen")
                    }
                  >
                    {(BiometricConfigurationScreen) => (
                      <BiometricConfigurationScreen />
                    )}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route
                path="2fa"
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/portal/MFAConfigurationScreen")
                    }
                  >
                    {(MFAConfigurationScreen) => <MFAConfigurationScreen />}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route
                path="anonymous-users"
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import(
                        "./graphql/portal/AnonymousUsersConfigurationScreen"
                      )
                    }
                  >
                    {(AnonymousUsersConfigurationScreen) => (
                      <AnonymousUsersConfigurationScreen />
                    )}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route
                path="app2app"
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/portal/App2AppConfigurationScreen")
                    }
                  >
                    {(App2AppConfigurationScreen) => (
                      <App2AppConfigurationScreen />
                    )}
                  </FlavoredErrorBoundSuspense>
                }
              />
            </Route>
            <Route path="apps">
              <Route
                index={true}
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/portal/ApplicationsConfigurationScreen")
                    }
                  >
                    {(ApplicationsConfigurationScreen) => (
                      <ApplicationsConfigurationScreen />
                    )}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route
                path="add"
                element={
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/portal/CreateOAuthClientScreen")
                    }
                  >
                    {(CreateOAuthClientScreen) => <CreateOAuthClientScreen />}
                  </FlavoredErrorBoundSuspense>
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
                    <FlavoredErrorBoundSuspense
                      factory={async () =>
                        import("./graphql/portal/EditOAuthClientScreen")
                      }
                    >
                      {(EditOAuthClientScreen) => <EditOAuthClientScreen />}
                    </FlavoredErrorBoundSuspense>
                  }
                />
              </Route>
            </Route>
            <Route
              path="languages"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/LanguagesConfigurationScreen")
                  }
                >
                  {(LanguagesConfigurationScreen) => (
                    <LanguagesConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
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
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import(
                        "./graphql/portal/StandardAttributesConfigurationScreen"
                      )
                    }
                  >
                    {(StandardAttributesConfigurationScreen) => (
                      <StandardAttributesConfigurationScreen />
                    )}
                  </FlavoredErrorBoundSuspense>
                }
              />
              <Route path="custom-attributes">
                <Route
                  index={true}
                  element={
                    <FlavoredErrorBoundSuspense
                      factory={async () =>
                        import(
                          "./graphql/portal/CustomAttributesConfigurationScreen"
                        )
                      }
                    >
                      {(CustomAttributesConfigurationScreen) => (
                        <CustomAttributesConfigurationScreen />
                      )}
                    </FlavoredErrorBoundSuspense>
                  }
                />
                <Route
                  path="add"
                  element={
                    <FlavoredErrorBoundSuspense
                      factory={async () =>
                        import("./graphql/portal/CreateCustomAttributeScreen")
                      }
                    >
                      {(CreateCustomAttributeScreen) => (
                        <CreateCustomAttributeScreen />
                      )}
                    </FlavoredErrorBoundSuspense>
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
                      <FlavoredErrorBoundSuspense
                        factory={async () =>
                          import("./graphql/portal/EditCustomAttributeScreen")
                        }
                      >
                        {(EditCustomAttributeScreen) => (
                          <EditCustomAttributeScreen />
                        )}
                      </FlavoredErrorBoundSuspense>
                    }
                  />
                </Route>
              </Route>
            </Route>
          </Route>
          <Route path="bot-protection">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/BotProtectionConfigurationScreen")
                  }
                >
                  {(BotProtectionConfigurationScreen) => (
                    <BotProtectionConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
          </Route>

          <Route path="integrations">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/IntegrationsConfigurationScreen")
                  }
                >
                  {(IntegrationsConfigurationScreen) => (
                    <IntegrationsConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="google-tag-manager"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import(
                      "./graphql/portal/GoogleTagManagerConfigurationScreen"
                    )
                  }
                >
                  {(GoogleTagManagerConfigurationScreen) => (
                    <GoogleTagManagerConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
          </Route>

          <Route path="billing">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/SubscriptionScreen")
                  }
                >
                  {(SubscriptionScreen) => <SubscriptionScreen />}
                </FlavoredErrorBoundSuspense>
              }
            />
          </Route>

          <Route path="billing-redirect">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/SubscriptionRedirect")
                  }
                >
                  {(SubscriptionRedirect) => <SubscriptionRedirect />}
                </FlavoredErrorBoundSuspense>
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
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/HookConfigurationScreen")
                  }
                >
                  {(HookConfigurationScreen) => <HookConfigurationScreen />}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="admin-api"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/AdminAPIConfigurationScreen")
                  }
                >
                  {(AdminAPIConfigurationScreen) => (
                    <AdminAPIConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="account-deletion"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import(
                      "./graphql/portal/AccountDeletionConfigurationScreen"
                    )
                  }
                >
                  {(AccountDeletionConfigurationScreen) => (
                    <AccountDeletionConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="account-anonymization"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import(
                      "./graphql/portal/AccountAnonymizationConfigurationScreen"
                    )
                  }
                >
                  {(AccountAnonymizationConfigurationScreen) => (
                    <AccountAnonymizationConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="session"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/CookieLifetimeConfigurationScreen")
                  }
                >
                  {(CookieLifetimeConfigurationScreen) => (
                    <CookieLifetimeConfigurationScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="smtp"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/SMTPConfigurationScreen")
                  }
                >
                  {(SMTPConfigurationScreen) => <SMTPConfigurationScreen />}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="endpoint-direct-access"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/EndpointDirectAccessScreen")
                  }
                >
                  {(EndpointDirectAccessScreen) => (
                    <EndpointDirectAccessScreen />
                  )}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="saml-certificate"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/SAMLCertificateScreen")
                  }
                >
                  {(SAMLCertificateScreen) => <SAMLCertificateScreen />}
                </FlavoredErrorBoundSuspense>
              }
            />
          </Route>

          <Route path="audit-log">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/adminapi/AuditLogScreen")
                  }
                >
                  {(AuditLogScreen) => <AuditLogScreen />}
                </FlavoredErrorBoundSuspense>
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
                  <FlavoredErrorBoundSuspense
                    factory={async () =>
                      import("./graphql/adminapi/AuditLogEntryScreen")
                    }
                  >
                    {(AuditLogEntryScreen) => <AuditLogEntryScreen />}
                  </FlavoredErrorBoundSuspense>
                }
              />
            </Route>
          </Route>

          <Route path="portal-admins">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/PortalAdminsSettings")
                  }
                >
                  {(PortalAdminsSettings) => <PortalAdminsSettings />}
                </FlavoredErrorBoundSuspense>
              }
            />
            <Route
              path="invite"
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/InviteAdminScreen")
                  }
                >
                  {(InviteAdminScreen) => <InviteAdminScreen />}
                </FlavoredErrorBoundSuspense>
              }
            />
          </Route>

          {/* This screen is not shown in nav bar, which is intentional to prevent normal users from accessing it */}
          <Route path="edit-config">
            <Route
              index={true}
              element={
                <FlavoredErrorBoundSuspense
                  factory={async () =>
                    import("./graphql/portal/EditConfigurationScreen")
                  }
                >
                  {(EditConfigurationScreen) => <EditConfigurationScreen />}
                </FlavoredErrorBoundSuspense>
              }
            />
          </Route>
        </Routes>
      </ScreenLayout>
    </ApolloProvider>
  );
};

export default AppRoot;
