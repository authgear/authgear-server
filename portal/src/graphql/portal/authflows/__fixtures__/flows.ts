import { AuthenticationFlowsConfig } from "../../../../types";

const generatedSignupFlows: AuthenticationFlowsConfig["signup_flows"] = [
  {
    name: "default",
    steps: [
      {
        name: "signup_identify",
        type: "identify",
        one_of: [
          {
            identification: "email",
            steps: [
              { type: "verify", target_step: "signup_identify" },
              {
                name: "authenticate_primary_email",
                type: "create_authenticator",
                one_of: [{ authentication: "primary_password" }],
              },
            ],
          },
          { identification: "oauth" },
        ],
      },
      { type: "prompt_create_passkey" },
    ],
  },
];

// Generated for email login ID + Google + passkey, password, 2FA (TOTP, email OTP) if_exists.
export const generatedEmailOAuthPasskey2FA: AuthenticationFlowsConfig = {
  signup_flows: generatedSignupFlows,
  promote_flows: generatedSignupFlows,
  login_flows: [
    {
      name: "default",
      steps: [
        {
          name: "login_identify",
          type: "identify",
          one_of: [
            { identification: "select_account" },
            {
              identification: "email",
              steps: [
                {
                  name: "authenticate_primary_email",
                  type: "authenticate",
                  one_of: [
                    {
                      authentication: "primary_password",
                      steps: [
                        {
                          name: "authenticate_secondary_email",
                          type: "authenticate",
                          optional: true,
                          one_of: [
                            { authentication: "secondary_totp" },
                            { authentication: "secondary_oob_otp_email" },
                            { authentication: "recovery_code" },
                            { authentication: "device_token" },
                          ],
                        },
                        {
                          type: "change_password",
                          target_step: "authenticate_primary_email",
                        },
                      ],
                    },
                    { authentication: "primary_passkey" },
                  ],
                },
              ],
            },
            { identification: "oauth" },
            { identification: "passkey" },
          ],
        },
        { type: "check_account_status" },
        { type: "terminate_other_sessions" },
        { type: "prompt_create_passkey" },
      ],
    },
  ],
  signup_login_flows: [
    {
      name: "default",
      steps: [
        {
          name: "signup_login_identify",
          type: "identify",
          one_of: [
            { identification: "select_account", login_flow: "default" },
            {
              identification: "email",
              login_flow: "default",
              signup_flow: "default",
            },
            {
              identification: "oauth",
              login_flow: "default",
              signup_flow: "default",
            },
            { identification: "passkey", login_flow: "default" },
          ],
        },
      ],
    },
  ],
  reauth_flows: [
    {
      name: "default",
      steps: [
        {
          name: "reauth_identify",
          type: "identify",
          one_of: [{ identification: "id_token" }],
        },
        {
          name: "reauthenticate",
          type: "authenticate",
          one_of: [
            { authentication: "primary_password" },
            { authentication: "primary_passkey" },
            { authentication: "secondary_totp" },
            { authentication: "secondary_oob_otp_email" },
          ],
        },
      ],
    },
  ],
  account_recovery_flows: [
    {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [
            {
              identification: "email",
              on_failure: "ignore",
              steps: [
                {
                  type: "select_destination",
                  allowed_channels: [{ channel: "email", otp_form: "link" }],
                },
              ],
            },
          ],
        },
        { type: "verify_account_recovery_code" },
        { type: "reset_password" },
      ],
    },
  ],
};

// Generated for a username-only project. Account recovery has no options,
// so the generator omits one_of entirely.
export const generatedUsernameOnly: AuthenticationFlowsConfig = {
  signup_flows: [
    {
      name: "default",
      steps: [
        {
          name: "signup_identify",
          type: "identify",
          one_of: [
            {
              identification: "username",
              steps: [
                {
                  name: "authenticate_primary_username",
                  type: "create_authenticator",
                  one_of: [{ authentication: "primary_password" }],
                },
              ],
            },
          ],
        },
      ],
    },
  ],
  account_recovery_flows: [
    {
      name: "default",
      steps: [
        { type: "identify" },
        { type: "verify_account_recovery_code" },
        { type: "reset_password" },
      ],
    },
  ],
};

// cousin-4bx3ay's custom signup flow: email, verify, password, then a form.
export const customSignupWithForm: AuthenticationFlowsConfig = {
  signup_flows: [
    {
      name: "default",
      steps: [
        {
          name: "signup_identify",
          type: "identify",
          one_of: [
            {
              identification: "email",
              steps: [
                { type: "verify", target_step: "signup_identify" },
                {
                  name: "authenticate_primary_email",
                  type: "create_authenticator",
                  one_of: [{ authentication: "primary_password" }],
                },
              ],
            },
          ],
        },
        {
          type: "fill_form",
          form: {
            name: "about_you",
            fields: [
              { user_profile: { pointer: "/given_name", required: true } },
              { user_profile: { pointer: "/family_name", required: true } },
              { user_profile: { pointer: "/birthdate" } },
            ],
          },
        },
      ],
    },
  ],
};
