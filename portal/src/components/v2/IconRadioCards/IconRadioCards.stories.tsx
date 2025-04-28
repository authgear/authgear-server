import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { IconRadioCards, IconRadioCardsProps } from "./IconRadioCards";
import passwordlessIcon from "../../../images/passwordless_icon.svg";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: IconRadioCards,
  tags: ["autodocs"],
  argTypes: {
    value: { table: { disable: true } },
    onValueChange: { table: { disable: true } },
  },
  args: {
    options: new Array(2).fill(null).map((_, idx) => ({
      value: `${idx + 1}`,
      icon: <img src={passwordlessIcon} width={40} height={40} />,
      title: `Passwordless`,
      subtitle: `One-Time-Password (OTP)`,
    })),
    value: null,
    onValueChange: () => {},
  },
  render: (args) => {
    return <Demo args={args} />;
  },
} satisfies Meta<typeof IconRadioCards>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    size: "2",
  },
};

export const WithTooltip: Story = {
  args: {
    size: "2",
    options: [
      {
        value: `1`,
        icon: <img src={passwordlessIcon} width={40} height={40} />,
        title: `Passwordless`,
        subtitle: `One-Time-Password (OTP)`,
        tooltip: "This is a tooltip",
      },
      {
        value: `2`,
        icon: <img src={passwordlessIcon} width={40} height={40} />,
        title: `Passwordless`,
        subtitle: `One-Time-Password (OTP)`,
        tooltip: "This is a tooltip",
        disabled: true,
      },
    ],
  },
};

function Demo({ args }: { args: IconRadioCardsProps<string> }) {
  const { value: _0, onValueChange: _1, ...rest } = args;

  const [value, setValue] = useState<string | null>(null);

  return (
    <div style={{ width: 600 }}>
      <IconRadioCards value={value} onValueChange={setValue} {...rest} />
    </div>
  );
}
