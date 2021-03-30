import React, { useMemo } from "react";
import { Routes, Route, useParams, Navigate } from "react-router-dom";
import { ApolloProvider } from "@apollo/client";

import { makeClient } from "./graphql/adminapi/apollo";
import { useAppConfigQuery } from "./graphql/portal/query/appConfigQuery";
import ScreenLayout from "./ScreenLayout";
import ShowLoading from "./ShowLoading";

import UsersScreen from "./graphql/adminapi/UsersScreen";
import AddUserScreen from "./graphql/adminapi/AddUserScreen";
import UserDetailsScreen from "./graphql/adminapi/UserDetailsScreen";
import AddEmailScreen from "./graphql/adminapi/AddEmailScreen";
import AddPhoneScreen from "./graphql/adminapi/AddPhoneScreen";
import AddUsernameScreen from "./graphql/adminapi/AddUsernameScreen";
import ResetPasswordScreen from "./graphql/adminapi/ResetPasswordScreen";

import AnonymousUsersConfigurationScreen from "./graphql/portal/AnonymousUsersConfigurationScreen";
import SingleSignOnConfigurationScreen from "./graphql/portal/SingleSignOnConfigurationScreen";
import PasswordPolicyConfigurationScreen from "./graphql/portal/PasswordPolicyConfigurationScreen";
import ForgotPasswordConfigurationScreen from "./graphql/portal/ForgotPasswordConfigurationScreen";
import OAuthClientConfigurationScreen from "./graphql/portal/OAuthClientConfigurationScreen";
import CreateOAuthClientScreen from "./graphql/portal/CreateOAuthClientScreen";
import EditOAuthClientScreen from "./graphql/portal/EditOAuthClientScreen";
import CustomDomainListScreen from "./graphql/portal/CustomDomainListScreen";
import VerifyDomainScreen from "./graphql/portal/VerifyDomainScreen";
import UISettingsScreen from "./graphql/portal/UISettingsScreen";
import LocalizationConfigurationScreen from "./graphql/portal/LocalizationConfigurationScreen";
import InviteAdminScreen from "./graphql/portal/InviteAdminScreen";
import PortalAdminsSettings from "./graphql/portal/PortalAdminsSettings";
import SessionConfigurationScreen from "./graphql/portal/SessionConfigurationScreen";
import WebhookConfigurationScreen from "./graphql/portal/WebhookConfigurationScreen";
import CORSConfigurationScreen from "./graphql/portal/CORSConfigurationScreen";
import LoginIDConfigurationScreen from "./graphql/portal/LoginIDConfigurationScreen";
import AuthenticatorConfigurationScreen from "./graphql/portal/AuthenticatorConfigurationScreen";
import VerificationConfigurationScreen from "./graphql/portal/VerificationConfigurationScreen";
import BiometricConfigurationScreen from "./graphql/portal/BiometricConfigurationScreen";

const AppRoot: React.FC = function AppRoot() {
  const { appID } = useParams();
  const client = useMemo(() => {
    return makeClient(appID);
  }, [appID]);

  // NOTE: check if appID actually exist in authorized app list
  const { effectiveAppConfig, loading, error } = useAppConfigQuery(appID);
  if (loading) {
    return <ShowLoading />;
  }

  // if node is null after loading without error, treat this as invalid
  // request, frontend cannot distinguish between inaccessible and not found
  const isInvalidAppID = error == null && effectiveAppConfig == null;

  // redirect to app list if app id is invalid
  if (isInvalidAppID) {
    return <Navigate to="/apps" replace={true} />;
  }

  return (
    <ApolloProvider client={client}>
      <ScreenLayout>
        <Routes>
          <Route path="/" element={<Navigate to="users/" replace={true} />} />
          <Route path="/users/" element={<UsersScreen />} />
          <Route path="/users/add-user/" element={<AddUserScreen />} />
          <Route
            path="/users/:userID/"
            element={<Navigate to="details/" replace={true} />}
          />
          <Route
            path="/users/:userID/details/"
            element={<UserDetailsScreen />}
          />
          <Route
            path="/users/:userID/details/add-email"
            element={<AddEmailScreen />}
          />
          <Route
            path="/users/:userID/details/add-phone"
            element={<AddPhoneScreen />}
          />
          <Route
            path="/users/:userID/details/add-username"
            element={<AddUsernameScreen />}
          />
          <Route
            path="/users/:userID/details/reset-password"
            element={<ResetPasswordScreen />}
          />
          <Route
            path="/configuration/authentication/login-id"
            element={<LoginIDConfigurationScreen />}
          />
          <Route
            path="/configuration/authentication/authenticators"
            element={<AuthenticatorConfigurationScreen />}
          />
          <Route
            path="/configuration/authentication/verification"
            element={<VerificationConfigurationScreen />}
          />
          <Route
            path="/configuration/anonymous-users"
            element={<AnonymousUsersConfigurationScreen />}
          />
          <Route
            path="/configuration/biometric"
            element={<BiometricConfigurationScreen />}
          />
          <Route
            path="/configuration/single-sign-on"
            element={<SingleSignOnConfigurationScreen />}
          />
          <Route
            path="/configuration/passwords/policy"
            element={<PasswordPolicyConfigurationScreen />}
          />
          <Route
            path="/configuration/passwords/forgot-password"
            element={<ForgotPasswordConfigurationScreen />}
          />
          <Route
            path="/configuration/apps/cors"
            element={<CORSConfigurationScreen />}
          />
          <Route
            path="/configuration/apps/oauth"
            element={<OAuthClientConfigurationScreen />}
          />
          <Route
            path="/configuration/apps/oauth/add"
            element={<CreateOAuthClientScreen />}
          />
          <Route
            path="/configuration/apps/oauth/:clientID/edit"
            element={<EditOAuthClientScreen />}
          />
          <Route
            path="/configuration/dns/custom-domains"
            element={<CustomDomainListScreen />}
          />
          <Route
            path="/configuration/dns/custom-domains/:domainID/verify"
            element={<VerifyDomainScreen />}
          />
          <Route
            path="/configuration/ui-settings"
            element={<UISettingsScreen />}
          />
          <Route
            path="/configuration/localization"
            element={<LocalizationConfigurationScreen />}
          />
          <Route
            path="/configuration/settings/portal-admins"
            element={<PortalAdminsSettings />}
          />
          <Route
            path="/configuration/settings/portal-admins/invite"
            element={<InviteAdminScreen />}
          />
          <Route
            path="/configuration/settings/sessions"
            element={<SessionConfigurationScreen />}
          />
          <Route
            path="/configuration/settings/web-hooks"
            element={<WebhookConfigurationScreen />}
          />
        </Routes>
      </ScreenLayout>
    </ApolloProvider>
  );
};

export default AppRoot;
