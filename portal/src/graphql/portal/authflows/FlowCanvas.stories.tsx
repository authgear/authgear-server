import React, { useMemo } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { MemoryRouter } from "react-router-dom";
import { FlowCanvas } from "./FlowCanvas";
import { FlowSketchContext, FlowSketchContextValue } from "./FlowSketchContext";
import { flowToTree } from "./tree";
import { FlowType } from "./model";
import {
  customSignupWithForm,
  generatedEmailOAuthPasskey2FA,
  generatedUsernameOnly,
} from "./__fixtures__/flows";
import {
  AuthenticationFlowConfig,
  AuthenticationFlowsConfig,
} from "../../../types";

interface CanvasStoryProps {
  flowType: FlowType;
  flow: AuthenticationFlowConfig;
  custom?: AuthenticationFlowsConfig;
  generated: AuthenticationFlowsConfig;
}

function CanvasStory({
  flowType,
  flow,
  custom,
  generated,
}: CanvasStoryProps): React.ReactElement {
  const context = useMemo<FlowSketchContextValue>(
    () => ({
      flowType,
      oauthProviders: [{ alias: "google", type: "google" }],
      customAttributes: [],
      phoneOTPMode: "sms",
      custom,
      generated,
    }),
    [flowType, custom, generated]
  );
  return (
    <MemoryRouter>
      <FlowSketchContext.Provider value={context}>
        <div style={{ display: "flex", height: 640 }}>
          <FlowCanvas
            tree={flowToTree(flow)}
            selectedStepID={null}
            onSelectStep={() => {}}
          />
        </div>
      </FlowSketchContext.Provider>
    </MemoryRouter>
  );
}

const meta = {
  title: "components/v1/authflows/FlowCanvas",
  component: CanvasStory,
} satisfies Meta<typeof CanvasStory>;

export default meta;
type Story = StoryObj<typeof meta>;

function first(
  flows: AuthenticationFlowConfig[] | undefined
): AuthenticationFlowConfig {
  if (flows == null || flows.length === 0) {
    throw new Error("fixture has no flow");
  }
  return flows[0];
}

const g = generatedEmailOAuthPasskey2FA;

export const GeneratedLogin: Story = {
  args: { flowType: "login", flow: first(g.login_flows), generated: g },
};

export const GeneratedSignupOrLogin: Story = {
  args: {
    flowType: "signup_login",
    flow: first(g.signup_login_flows),
    generated: g,
  },
};

export const GeneratedReauth: Story = {
  args: { flowType: "reauth", flow: first(g.reauth_flows), generated: g },
};

export const GeneratedAccountRecovery: Story = {
  args: {
    flowType: "account_recovery",
    flow: first(g.account_recovery_flows),
    generated: g,
  },
};

export const CustomSignupWithForm: Story = {
  args: {
    flowType: "signup",
    flow: first(customSignupWithForm.signup_flows),
    custom: customSignupWithForm,
    generated: g,
  },
};

export const UsernameOnlyAccountRecovery: Story = {
  args: {
    flowType: "account_recovery",
    flow: first(generatedUsernameOnly.account_recovery_flows),
    generated: generatedUsernameOnly,
  },
};

export const EveryStepType: Story = {
  args: {
    flowType: "signup",
    generated: g,
    flow: {
      name: "default",
      steps: [
        {
          type: "identify",
          one_of: [{ identification: "phone" }, { identification: "ldap" }],
        },
        { type: "verify" },
        {
          type: "create_authenticator",
          one_of: [
            { authentication: "primary_password" },
            { authentication: "primary_passkey" },
          ],
        },
        {
          type: "authenticate",
          one_of: [
            { authentication: "secondary_oob_otp_sms" },
            { authentication: "recovery_code" },
            { authentication: "device_token" },
          ],
        },
        { type: "view_recovery_code" },
        {
          type: "fill_in_user_profile",
          user_profile: [{ pointer: "/gender", required: true }],
        },
        { type: "change_password" },
        {
          type: "select_destination",
          allowed_channels: [{ channel: "email" }, { channel: "sms" }],
        },
        { type: "verify_account_recovery_code" },
        { type: "reset_password" },
        { type: "prompt_create_passkey" },
        { type: "check_account_status" },
        { type: "future_step" },
      ],
    },
  },
};
