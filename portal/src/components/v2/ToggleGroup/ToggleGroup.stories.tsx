import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { ToggleGroup, ToggleGroupProps } from "./ToggleGroup";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: ToggleGroup,
  tags: ["autodocs"],
  args: {
    items: [
      {
        value: "email",
        text: "Email",
        icon: <DemoIcon />,
      },
      {
        value: "phone",
        text: "Phone",
        icon: <DemoIcon />,
      },
      {
        value: "phone2",
        text: "Phone",
        icon: <DemoIcon />,
      },
      {
        value: "phone3",
        text: "Phone",
        icon: <DemoIcon />,
      },
      {
        value: "phone4",
        text: "Phone",
        icon: <DemoIcon />,
      },
      {
        value: "phone5",
        text: "Phone",
        icon: <DemoIcon />,
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

function DemoIcon() {
  return (
    <div className="h-5 w-5 grid items-center justify-center">
      <i className={"fa fa-envelope text-xl"} />
    </div>
  );
}
