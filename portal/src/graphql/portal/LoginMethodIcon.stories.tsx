import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { MemoryRouter } from "react-router-dom";
import {
  LoginMethodFirstLevelOptionsGrid,
  type LoginMethodFirstLevelOption,
} from "./LoginMethodConfigurationScreen";
import styles from "./LoginMethodConfigurationScreen.module.css";

const meta = {
  title: "components/v1/LoginMethodChooser/LoginMethodIcon",
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
 * First-level **icons only** (no `WidgetTitle`, no labels under each icon), for layout reference.
 */
function LoginMethodIconDefaultRender() {
  const [first, setFirst] = useState<LoginMethodFirstLevelOption | null>(null);
  return (
    <div className={styles.widget}>
      <LoginMethodFirstLevelOptionsGrid
        phoneLoginIDDisabled={false}
        firstLevelOption={first}
        onChangeFirstLevelOption={setFirst}
        showWidgetTitle={false}
        iconOnly={true}
      />
    </div>
  );
}

function LoginMethodIconSelectedRender() {
  const [first, setFirst] = useState<LoginMethodFirstLevelOption | null>(
    "oauth"
  );
  return (
    <div className={styles.widget}>
      <LoginMethodFirstLevelOptionsGrid
        phoneLoginIDDisabled={false}
        firstLevelOption={first}
        onChangeFirstLevelOption={setFirst}
        showWidgetTitle={false}
        iconOnly={true}
      />
    </div>
  );
}

export const Default: Story = {
  name: "Default",
  render: () => <LoginMethodIconDefaultRender />,
};

export const Selected: Story = {
  name: "Selected",
  render: () => <LoginMethodIconSelectedRender />,
};
