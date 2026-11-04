import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  StringConstraintsFields,
  StringConstraintsFieldsProps,
} from "./StringConstraintsFields";
import { draftFromConstraints } from "./util/stringConstraints";

const meta = {
  title: "components/StringConstraintsFields",
  component: StringConstraintsFields,
  tags: ["autodocs"],
  argTypes: {
    draft: { table: { disable: true } },
    onChange: { table: { disable: true } },
    patternMessage: { table: { disable: true } },
    onPatternMessageChange: { table: { disable: true } },
  },
  args: {
    draft: draftFromConstraints(undefined),
    onChange: () => {},
    patternMessage: "",
    onPatternMessageChange: () => {},
    parentJSONPointer: "/user_profile/custom_attributes/attributes/0",
  },
  render: (args) => <Demo args={args} />,
} satisfies Meta<typeof StringConstraintsFields>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};

export const AllSet: Story = {
  args: {
    draft: draftFromConstraints({
      min_length: 3,
      max_length: 20,
      allowed_characters: ["letters", "digits"],
      pattern: "^[A-Z]{2}[0-9]+$",
    }),
    patternMessage: "Start with two capital letters, then digits.",
  },
};

function Demo({ args }: { args: StringConstraintsFieldsProps }) {
  const [draft, setDraft] = useState(args.draft);
  const [message, setMessage] = useState(args.patternMessage);
  return (
    <div style={{ maxWidth: 480 }}>
      <StringConstraintsFields
        {...args}
        draft={draft}
        onChange={setDraft}
        patternMessage={message}
        onPatternMessageChange={setMessage}
      />
    </div>
  );
}
