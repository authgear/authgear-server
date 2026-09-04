import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { SecondaryButton } from "../v2/Button/SecondaryButton/SecondaryButton";
import { UnauthenticatedDialog } from "./UnauthenticatedDialog";

function UnauthenticatedDialogDemo(): React.ReactElement {
  const [hidden, setHidden] = useState(false);

  return (
    <>
      <SecondaryButton
        size="2"
        text="Reopen dialog"
        onClick={() => setHidden(false)}
      />
      <UnauthenticatedDialog
        isHidden={hidden}
        onConfirm={() => setHidden(true)}
      />
    </>
  );
}

const meta = {
  component: UnauthenticatedDialogDemo,
  tags: ["autodocs"],
} satisfies Meta<typeof UnauthenticatedDialogDemo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const SessionExpired: Story = {
  name: "Session expired",
};
