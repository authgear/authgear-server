import React, { useContext, useMemo } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import type { IContextualMenuProps } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import PrimaryButton from "../../PrimaryButton";
import { SystemConfigContext } from "../../context/SystemConfigContext";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "../../system-config";

const systemConfig = instantiateSystemConfig(defaultSystemConfig);

export type AddPrimaryMenuKind = "addIdentity" | "add2fa";

const addPrimaryMenuKindOptions = [
  "addIdentity",
  "add2fa",
] as const satisfies readonly AddPrimaryMenuKind[];

type AddPrimaryMenuStoryArgs = {
  menuKind: AddPrimaryMenuKind;
  disabled: boolean;
};

/**
 * **ContextualMenuButtonWithIcon**: `PrimaryButton` + icon + `menuProps` (same pattern as
 * User details — **Add identity** in `UserDetailsConnectedIdentities.tsx` or **Add 2FA** in
 * `UserDetailsAccountSecurity.tsx`). Storybook: `components/v1/Button/…`. Switch source with
 * Controls → `menuKind`.
 */
const meta = {
  title: "components/v1/Button/ContextualMenuButtonWithIcon",
  tags: ["autodocs"],
  args: {
    menuKind: "addIdentity" satisfies AddPrimaryMenuKind,
    disabled: false,
  },
  argTypes: {
    menuKind: {
      control: {
        type: "select",
        labels: {
          addIdentity: "Add identity",
          add2fa: "Add 2FA",
        },
      },
      options: [...addPrimaryMenuKindOptions],
      description: "Which header action the button represents (copy and menu items follow production).",
    },
    disabled: {
      control: "boolean",
      description:
        "When on, matches no options / nothing to add: disabled with an empty menu.",
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
} satisfies Meta<AddPrimaryMenuStoryArgs>;

export default meta;
type Story = StoryObj<typeof meta>;

function useAddIdentityMenuProps(): IContextualMenuProps {
  const { renderToString } = useContext(Context);
  return useMemo(
    () => ({
      directionalHintFixed: true,
      items: [
        {
          key: "email",
          text: renderToString("UserDetails.connected-identities.email"),
          iconProps: { iconName: "Mail" },
          onClick: () => {},
        },
        {
          key: "phone",
          text: renderToString("UserDetails.connected-identities.phone"),
          iconProps: { iconName: "CellPhone" },
          onClick: () => {},
        },
        {
          key: "username",
          text: renderToString("UserDetails.connected-identities.username"),
          iconProps: { iconName: "Accounts" },
          onClick: () => {},
        },
      ],
    }),
    [renderToString]
  );
}

/** Menu shape matches `add2FAMenuProps` in `UserDetailsAccountSecurity.tsx` (noop navigation). */
function useAdd2FAMenuProps(): IContextualMenuProps {
  const { renderToString } = useContext(Context);
  return useMemo(
    () => ({
      directionalHintFixed: true,
      items: [
        {
          key: "password",
          text: renderToString("AuthenticatorType.secondary.password"),
          iconProps: { iconName: "Accounts" },
          onClick: () => {},
        },
        {
          key: "oob_otp_email",
          text: renderToString("AuthenticatorType.secondary.oob-otp-email"),
          iconProps: { iconName: "Mail" },
          onClick: () => {},
        },
        {
          key: "oob_otp_sms",
          text: renderToString("AuthenticatorType.secondary.oob-otp-phone"),
          iconProps: { iconName: "CellPhone" },
          onClick: () => {},
        },
      ],
    }),
    [renderToString]
  );
}

function AddPrimaryMenuSection(props: AddPrimaryMenuStoryArgs): React.ReactElement {
  const { menuKind, disabled } = props;
  const identityMenu = useAddIdentityMenuProps();
  const add2faMenu = useAdd2FAMenuProps();

  const menuProps = disabled
    ? { items: [] }
    : menuKind === "addIdentity"
      ? identityMenu
      : add2faMenu;

  const labelId =
    menuKind === "addIdentity"
      ? "UserDetails.connected-identities.add-identity"
      : "UserDetails.account-security.secondary.add";

  return (
    <section
      style={{
        display: "flex",
        flexDirection: "row",
        alignItems: "center",
        justifyContent: "flex-end",
        paddingTop: 12,
      }}
    >
      <PrimaryButton
        disabled={disabled}
        iconProps={{ iconName: "CirclePlus" }}
        menuProps={menuProps}
        styles={{
          menuIcon: { paddingLeft: "3px" },
          icon: { paddingRight: "3px" },
        }}
        text={<FormattedMessage id={labelId} />}
      />
    </section>
  );
}

export const Default: Story = {
  render: (args) => (
    <AddPrimaryMenuSection
      menuKind={args.menuKind ?? "addIdentity"}
      disabled={args.disabled ?? false}
    />
  ),
};

/** Same canvas with **disabled** on by default (still override with Controls). */
export const Disabled: Story = {
  args: {
    disabled: true,
  },
  render: (args) => (
    <AddPrimaryMenuSection
      menuKind={args.menuKind ?? "addIdentity"}
      disabled={args.disabled ?? true}
    />
  ),
};
