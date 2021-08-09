/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: SendTestEmailMutation
// ====================================================

export interface SendTestEmailMutation {
  /**
   * Send test STMP configuration email
   */
  sendTestSMTPConfigurationEmail: boolean | null;
}

export interface SendTestEmailMutationVariables {
  appID: string;
  smtpHost: string;
  smtpPort: number;
  smtpUsername: string;
  smtpPassword: string;
  to: string;
}
