import React, { useMemo } from "react";
import { Routes, Route, useParams, Navigate } from "react-router-dom";
import { ApolloProvider } from "@apollo/client";

import { makeClient } from "./graphql/adminapi/apollo";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import ScreenLayout from "./ScreenLayout";
import ShowLoading from "./ShowLoading";

import UsersScreen from "./graphql/adminapi/UsersScreen";
import AddUserScreen from "./graphql/adminapi/AddUserScreen";
import UserDetailsScreen from "./graphql/adminapi/UserDetailsScreen";
import AddEmailScreen from "./graphql/adminapi/AddEmailScreen";
import AddPhoneScreen from "./graphql/adminapi/AddPhoneScreen";
import AddUsernameScreen from "./graphql/adminapi/AddUsernameScreen";
import ResetPasswordScreen from "./graphql/adminapi/ResetPasswordScreen";
import EditPictureScreen from "./graphql/adminapi/EditPictureScreen";

import AuditLogScreen from "./graphql/adminapi/AuditLogScreen";
import AuditLogEntryScreen from "./graphql/adminapi/AuditLogEntryScreen";

import ProjectRootScreen from "./graphql/portal/ProjectRootScreen";
import GetStartedScreen from "./graphql/portal/GetStartedScreen";
import AnonymousUsersConfigurationScreen from "./graphql/portal/AnonymousUsersConfigurationScreen";
import SingleSignOnConfigurationScreen from "./graphql/portal/SingleSignOnConfigurationScreen";
import PasswordPolicyConfigurationScreen from "./graphql/portal/PasswordPolicyConfigurationScreen";
import ForgotPasswordConfigurationScreen from "./graphql/portal/ForgotPasswordConfigurationScreen";
import ApplicationsConfigurationScreen from "./graphql/portal/ApplicationsConfigurationScreen";
import CreateOAuthClientScreen from "./graphql/portal/CreateOAuthClientScreen";
import EditOAuthClientScreen from "./graphql/portal/EditOAuthClientScreen";
import CustomDomainListScreen from "./graphql/portal/CustomDomainListScreen";
import VerifyDomainScreen from "./graphql/portal/VerifyDomainScreen";
import UISettingsScreen from "./graphql/portal/UISettingsScreen";
import LocalizationConfigurationScreen from "./graphql/portal/LocalizationConfigurationScreen";
import InviteAdminScreen from "./graphql/portal/InviteAdminScreen";
import PortalAdminsSettings from "./graphql/portal/PortalAdminsSettings";
import WebhookConfigurationScreen from "./graphql/portal/WebhookConfigurationScreen";
import AdminAPIConfigurationScreen from "./graphql/portal/AdminAPIConfigurationScreen";
import LoginIDConfigurationScreen from "./graphql/portal/LoginIDConfigurationScreen";
import AuthenticatorConfigurationScreen from "./graphql/portal/AuthenticatorConfigurationScreen";
import VerificationConfigurationScreen from "./graphql/portal/VerificationConfigurationScreen";
import BiometricConfigurationScreen from "./graphql/portal/BiometricConfigurationScreen";
import SubscriptionScreen from "./graphql/portal/SubscriptionScreen";
import SMTPConfigurationScreen from "./graphql/portal/SMTPConfigurationScreen";
import StandardAttributesConfigurationScreen from "./graphql/portal/StandardAttributesConfigurationScreen";
import CustomAttributesConfigurationScreen from "./graphql/portal/CustomAttributesConfigurationScreen";
import EditCustomAttributeScreen from "./graphql/portal/EditCustomAttributeScreen";
import CreateCustomAttributeScreen from "./graphql/portal/CreateCustomAttributeScreen";
import AccountDeletionConfigurationScreen from "./graphql/portal/AccountDeletionConfigurationScreen";
import AnalyticsScreen from "./graphql/portal/AnalyticsScreen";
import IntegrationsConfigurationScreen from "./graphql/portal/IntegrationsConfigurationScreen";
import GoogleTagManagerConfigurationScreen from "./graphql/portal/GoogleTagManagerConfigurationScreen";

const AppRoot: React.FC = function AppRoot() {
  const { appID } = useParams() as { appID: string };
  const client = useMemo(() => {
    return makeClient(appID);
  }, [appID]);

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

  return (
    <ApolloProvider client={client}>
      <ScreenLayout>
        <Routes>
          <Route index={true} element={<ProjectRootScreen />} />

          <Route path="getting-started">
            <Route index={true} element={<GetStartedScreen />} />
          </Route>

          <Route path="analytics">
            <Route index={true} element={<AnalyticsScreen />} />
          </Route>

          <Route path="users">
            <Route index={true} element={<UsersScreen />} />
            <Route path="add-user" element={<AddUserScreen />} />
            <Route path=":userID">
              <Route
                index={true}
                element={<Navigate to="details" replace={true} />}
              />
              <Route path="details">
                <Route index={true} element={<UserDetailsScreen />} />
                <Route path="add-email" element={<AddEmailScreen />} />
                <Route path="add-phone" element={<AddPhoneScreen />} />
                <Route path="add-username" element={<AddUsernameScreen />} />
                <Route
                  path="reset-password"
                  element={<ResetPasswordScreen />}
                />
                <Route path="edit-picture" element={<EditPictureScreen />} />
              </Route>
            </Route>
          </Route>

          <Route path="custom-domains">
            <Route index={true} element={<CustomDomainListScreen />} />
            <Route path=":domainID">
              <Route
                index={true}
                element={<Navigate to="verify" replace={true} />}
              />
              <Route path="verify" element={<VerifyDomainScreen />} />
            </Route>
          </Route>

          <Route path="configuration">
            <Route path="authentication">
              <Route path="login-id" element={<LoginIDConfigurationScreen />} />
              <Route
                path="authenticators"
                element={<AuthenticatorConfigurationScreen />}
              />
              <Route
                path="verification"
                element={<VerificationConfigurationScreen />}
              />
            </Route>
            <Route
              path="anonymous-users"
              element={<AnonymousUsersConfigurationScreen />}
            />
            <Route
              path="biometric"
              element={<BiometricConfigurationScreen />}
            />
            <Route
              path="single-sign-on"
              element={<SingleSignOnConfigurationScreen />}
            />
            <Route
              path="password-policy"
              element={<PasswordPolicyConfigurationScreen />}
            />
            <Route path="apps">
              <Route
                index={true}
                element={<ApplicationsConfigurationScreen />}
              />
              <Route path="add" element={<CreateOAuthClientScreen />} />
              <Route path=":clientID">
                <Route
                  index={true}
                  element={<Navigate to="edit" replace={true} />}
                />
                <Route path="edit" element={<EditOAuthClientScreen />} />
              </Route>
            </Route>
            <Route path="smtp" element={<SMTPConfigurationScreen />} />
            <Route path="ui-settings" element={<UISettingsScreen />} />
            <Route
              path="localization"
              element={<LocalizationConfigurationScreen />}
            />
            <Route path="user-profile">
              <Route
                path="standard-attributes"
                element={<StandardAttributesConfigurationScreen />}
              />
              <Route path="custom-attributes">
                <Route
                  index={true}
                  element={<CustomAttributesConfigurationScreen />}
                />
                <Route path="add" element={<CreateCustomAttributeScreen />} />
                <Route path=":index">
                  <Route
                    index={true}
                    element={<Navigate to="edit" replace={true} />}
                  />
                  <Route path="edit" element={<EditCustomAttributeScreen />} />
                </Route>
              </Route>
            </Route>
          </Route>

          <Route path="integrations">
            <Route index={true} element={<IntegrationsConfigurationScreen />} />
            <Route
              path="google-tag-manager"
              element={<GoogleTagManagerConfigurationScreen />}
            />
          </Route>

          <Route path="billing">
            <Route index={true} element={<SubscriptionScreen />} />
          </Route>

          <Route path="advanced">
            <Route path="webhooks" element={<WebhookConfigurationScreen />} />
            <Route path="admin-api" element={<AdminAPIConfigurationScreen />} />
            <Route
              path="account-deletion"
              element={<AccountDeletionConfigurationScreen />}
            />
            <Route
              path="password-reset-code"
              element={<ForgotPasswordConfigurationScreen />}
            />
          </Route>

          <Route path="audit-log">
            <Route index={true} element={<AuditLogScreen />} />
            <Route path=":logID">
              <Route
                index={true}
                element={<Navigate to="details" replace={true} />}
              />
              <Route path="details" element={<AuditLogEntryScreen />} />
            </Route>
          </Route>

          <Route path="portal-admins">
            <Route index={true} element={<PortalAdminsSettings />} />
            <Route path="invite" element={<InviteAdminScreen />} />
          </Route>
        </Routes>
      </ScreenLayout>
    </ApolloProvider>
  );
};

export default AppRoot;
