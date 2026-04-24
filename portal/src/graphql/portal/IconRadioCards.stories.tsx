import React, { useCallback, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { MemoryRouter } from "react-router-dom";
import Widget from "../../Widget";
import {
  LoginMethodAuthenticationSection,
  type LoginMethod,
  type LoginMethodFirstLevelOption,
  type LoginMethodSecondLevelOption,
  loginMethodToFirstLevelOption,
  loginMethodToSecondLevelOption,
} from "./LoginMethodConfigurationScreen";
import styles from "./LoginMethodConfigurationScreen.module.css";

/**
 * v1 “icon radio” row used on **Authentication** (Passwordless / Enter password):
 * `LoginMethodAuthenticationSection` (`ChoiceButton` / `AuthenticationButton` on
 * `LoginMethodConfigurationScreen`).
 */
const meta = {
  title: "components/v1/IconRadioCards",
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <MemoryRouter>
        <div style={{ maxWidth: 920, padding: 16, boxSizing: "border-box" }}>
          <Story />
        </div>
      </MemoryRouter>
    ),
  ],
} satisfies Meta;

export default meta;
type Story = StoryObj<typeof meta>;

const firstLevelForDemo: LoginMethodFirstLevelOption = "email";

/**
 * No second-level value selected; both option cards are unchecked. (This state is
 * for documentation; the live login method model always has a value once email/phone
 * is chosen.)
 */
function IconRadioCardsDefaultRender() {
  const [secondLevelOption, setSecondLevelOption] =
    useState<LoginMethodSecondLevelOption | null>(null);
  return (
    <Widget className={styles.widget}>
      <LoginMethodAuthenticationSection
        firstLevelOption={firstLevelForDemo}
        secondLevelOption={secondLevelOption}
        onChangeSecondLevelOption={setSecondLevelOption}
        showSubtitle={false}
      />
    </Widget>
  );
}

function IconRadioCardsWithSelectionRender() {
  const [loginMethod, setLoginMethod] =
    useState<LoginMethod>("passwordless-email");
  const firstLevelOption = loginMethodToFirstLevelOption(loginMethod);
  const secondLevelOption = loginMethodToSecondLevelOption(loginMethod);

  const onChangeSecondLevelOption = useCallback(
    (opt: LoginMethodSecondLevelOption) => {
      if (
        firstLevelOption !== "oauth" &&
        firstLevelOption !== "custom" &&
        firstLevelOption !== "username"
      ) {
        setLoginMethod(`${opt}-${firstLevelOption}` as LoginMethod);
      }
    },
    [firstLevelOption]
  );

  return (
    <Widget className={styles.widget}>
      {secondLevelOption != null ? (
        <LoginMethodAuthenticationSection
          firstLevelOption={firstLevelOption}
          secondLevelOption={secondLevelOption}
          onChangeSecondLevelOption={onChangeSecondLevelOption}
          showSubtitle={false}
        />
      ) : null}
    </Widget>
  );
}

export const Default: Story = {
  name: "Default",
  render: () => <IconRadioCardsDefaultRender />,
};

export const Selected: Story = {
  name: "Selected",
  render: () => <IconRadioCardsWithSelectionRender />,
};
