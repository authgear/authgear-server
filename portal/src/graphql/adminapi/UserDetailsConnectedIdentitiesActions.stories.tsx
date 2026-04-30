import React, { useContext, useMemo } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import type { IContextualMenuProps } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import DefaultButton from "../../DefaultButton";
import { SystemConfigContext } from "../../context/SystemConfigContext";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "../../system-config";

const systemConfig = instantiateSystemConfig(defaultSystemConfig);

/** Row type → same **Actions** button, different `menuProps` (see Controls → `variant`). */
export type RowActionsMenuVariant =
  | "loginIdUnverified"
  | "loginIdVerified"
  | "oauth";

const rowActionsMenuVariantOptions = [
  "loginIdUnverified",
  "loginIdVerified",
  "oauth",
] as const satisfies readonly RowActionsMenuVariant[];

interface ConnectedIdentitiesActionsStoryArgs {
  variant?: RowActionsMenuVariant;
}

/**
 * **ContextualMenuButton** pattern: per-row **Actions** from User details → Connected
 * identities (`#connected-identities`): `DefaultButton` + `menuProps`, matching
 * `UserDetailsConnectedIdentities.tsx`. Storybook: `components/v1/Button/ContextualMenuButton`
 * (sibling **ContextualMenuButtonWithIcon** for primary + icon + menu).
 */
const meta = {
  title: "components/v1/Button/ContextualMenuButton",
  tags: ["autodocs"],
  args: {
    variant: "loginIdUnverified" satisfies RowActionsMenuVariant,
  },
  argTypes: {
    variant: {
      control: {
        type: "select",
        labels: {
          loginIdUnverified: "Login ID row (unverified)",
          loginIdVerified: "Login ID row (verified)",
          oauth: "OAuth row",
        },
      },
      options: [...rowActionsMenuVariantOptions],
      description: "Switches the contextual menu (login ID vs OAuth).",
    },
  },
  decorators: [
    (Story) => (
      <SystemConfigContext.Provider value={systemConfig}>
        <div style={{ maxWidth: 720, padding: "8px 0" }}>
          <Story />
        </div>
      </SystemConfigContext.Provider>
    ),
  ],
} satisfies Meta<ConnectedIdentitiesActionsStoryArgs>;

export default meta;
type Story = StoryObj<typeof meta>;

function LoginIdRowActionsMenu(props: {
  verified: boolean;
}): React.ReactElement {
  const { verified } = props;
  const { themes } = systemConfig;
  const { renderToString } = useContext(Context);

  const menuProps = useMemo<IContextualMenuProps>(
    () => ({
      shouldFocusOnMount: true,
      items: [
        {
          key: "verify",
          text: renderToString(
            verified ? "make-as-unverified" : "make-as-verified"
          ),
          onClick: () => {},
        },
        {
          key: "edit",
          text: renderToString("edit"),
          onClick: () => {},
        },
        {
          key: "remove",
          text: renderToString("remove"),
          onClick: () => {},
        },
      ],
    }),
    [verified, renderToString]
  );

  return (
    <DefaultButton
      theme={themes.main}
      text={<FormattedMessage id="action" />}
      menuProps={menuProps}
    />
  );
}

function OauthRowActionsMenu(): React.ReactElement {
  const { themes } = systemConfig;
  const { renderToString } = useContext(Context);
  const menuProps = useMemo<IContextualMenuProps>(
    () => ({
      shouldFocusOnMount: true,
      items: [
        {
          key: "verify",
          text: renderToString("make-as-verified"),
          onClick: () => {},
        },
        {
          key: "disconnect",
          text: renderToString("disconnect"),
          onClick: () => {},
        },
      ],
    }),
    [renderToString]
  );

  return (
    <DefaultButton
      theme={themes.main}
      text={<FormattedMessage id="action" />}
      menuProps={menuProps}
    />
  );
}

export const RowActions: Story = {
  name: "Actions",
  render: (args) => {
    const variant = args.variant ?? "loginIdUnverified";
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "flex-end",
          padding: "16px 0",
        }}
      >
        {variant === "loginIdUnverified" ? (
          <LoginIdRowActionsMenu verified={false} />
        ) : variant === "loginIdVerified" ? (
          <LoginIdRowActionsMenu verified={true} />
        ) : (
          <OauthRowActionsMenu />
        )}
      </div>
    );
  },
};
