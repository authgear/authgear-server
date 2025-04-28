import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react";
import { ToggleGroup, ToggleGroupProps } from "./ToggleGroup";
import loginEmailIcon from "../../../images/login_email.svg";
import loginPhoneIcon from "../../../images/login_phone.svg";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: ToggleGroup,
  tags: ["autodocs"],
  args: {
    items: [
      {
        value: "email",
        text: "Email",
        icon: <img src={loginEmailIcon} width={20} height={20} />,
      },
      {
        value: "phone",
        text: "Phone",
        icon: <img src={loginPhoneIcon} width={20} height={20} />,
      },
      {
        value: "phone2",
        text: "Phone",
        icon: <img src={loginPhoneIcon} width={20} height={20} />,
      },
      {
        value: "phone3",
        text: "Phone",
        icon: <img src={loginPhoneIcon} width={20} height={20} />,
      },
      {
        value: "phone4",
        text: "Phone",
        icon: <img src={loginPhoneIcon} width={20} height={20} />,
      },
      {
        value: "phone5",
        text: "Phone",
        icon: <img src={loginPhoneIcon} width={20} height={20} />,
      },
    ],
    values: [],
  },
  argTypes: {
    values: { table: { disable: true } },
    onValuesChange: { table: { disable: true } },
  },
  render: (args) => {
    return <Demo args={args} />;
  },
} satisfies Meta<typeof ToggleGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {},
};

function Demo({
  args: { values: _0, onValuesChange: _1, ...rest },
}: {
  args: ToggleGroupProps<string>;
}) {
  const [values, setValues] = useState<string[]>([]);

  return (
    <>
      <style>
        {`
      .toggleGroupDemo {
        width: 400px;
        max-height: 316px;
        display: flex;
        flex-direction: column;
      }
      `}
      </style>
      <div className="toggleGroupDemo [&>*]:flex-initial">
        <ToggleGroup {...rest} values={values} onValuesChange={setValues} />
      </div>
    </>
  );
}
