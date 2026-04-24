import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { MemoryRouter } from "react-router-dom";
import Widget from "../../Widget";
import {
  LoginMethodSelectLoginMethodsSection,
  LoginMethodFirstLevelOptionsGrid,
  type LoginMethod,
  type LoginMethodFirstLevelOption,
} from "./LoginMethodConfigurationScreen";
import styles from "./LoginMethodConfigurationScreen.module.css";

const meta = {
  title: "components/v1/LoginMethodChooser",
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

/**
 * Full first-level chooser: title + description under each icon (no “Select login methods” `WidgetTitle` in this story).
 */
function LoginMethodChooserWithLabelsDefaultRender() {
  const [first, setFirst] = useState<LoginMethodFirstLevelOption | null>(null);
  return (
    <Widget className={styles.widget}>
      <LoginMethodFirstLevelOptionsGrid
        phoneLoginIDDisabled={false}
        firstLevelOption={first}
        onChangeFirstLevelOption={setFirst}
        showWidgetTitle={false}
        iconOnly={false}
      />
    </Widget>
  );
}

function LoginMethodChooserWithLabelsSelectedRender() {
  const [loginMethod, setLoginMethod] = useState<LoginMethod>("oauth");
  return (
    <Widget className={styles.widget}>
      <LoginMethodSelectLoginMethodsSection
        phoneLoginIDDisabled={false}
        loginMethod={loginMethod}
        onChangeLoginMethod={setLoginMethod}
        showWidgetTitle={false}
      />
    </Widget>
  );
}

export const Default: Story = {
  name: "Default",
  render: () => <LoginMethodChooserWithLabelsDefaultRender />,
};

/**
 * Same as production wiring, except the section `WidgetTitle` is hidden in Storybook.
 */
export const Selected: Story = {
  name: "Selected",
  render: () => <LoginMethodChooserWithLabelsSelectedRender />,
};
