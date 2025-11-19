import React, { useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { FormField, FormFieldProps } from "./FormField";
import { RadioCards } from "../RadioCards/RadioCards";
import { TextField } from "../TextField/TextField";
import { ToggleGroup } from "../ToggleGroup/ToggleGroup";
import { ImageIcon } from "@radix-ui/react-icons";

// More on how to set up stories at: https://storybook.js.org/docs/writing-stories#default-export
const meta = {
  component: FormField,
  tags: ["autodocs"],
  argTypes: {
    error: {
      control: {
        type: "text",
      },
    },
    hint: {
      control: {
        type: "text",
      },
    },
  },
  args: {
    size: "3",
    label: "Label",
  },
} satisfies Meta<typeof FormField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    error: "Some error",
    hint: "Some hint",
  },

  render: (args) => {
    return (
      <div style={{ width: "400px" }}>
        <FormField {...args}>
          <p className="border border-solid">Any Input Here</p>
        </FormField>
      </div>
    );
  },
};

export const WithRadioCards: Story = {
  args: {},

  render: (args) => {
    return <RadioCardsDemo {...args} />;
  },
};

function RadioCardsDemo({ ...args }: FormFieldProps) {
  const [value, setValue] = useState<string>("1");

  return (
    <div style={{ width: "max-content" }}>
      <FormField {...args}>
        <RadioCards
          size="3"
          options={[
            {
              value: "1",
              title: "Option 1",
            },
            {
              value: "2",
              title: "Option 2",
            },
          ]}
          value={value}
          onValueChange={setValue}
          numberOfColumns={2}
        />
      </FormField>
    </div>
  );
}

export const WithTextField: Story = {
  args: {},

  render: (args) => {
    return <TextFieldDemo {...args} />;
  },
};

function TextFieldDemo({ ...args }: FormFieldProps) {
  return (
    <div style={{ width: "300px" }}>
      <FormField {...args}>
        <TextField size={args.size} />
      </FormField>
    </div>
  );
}

export const WithToggleGroup: Story = {
  args: {},

  render: (args) => {
    return <ToggleGroupDemo {...args} />;
  },
};

function ToggleGroupDemo({ ...args }: FormFieldProps) {
  const [values, setValues] = useState<string[]>([]);
  return (
    <div style={{ width: "300px" }}>
      <FormField {...args}>
        <ToggleGroup
          items={[
            {
              value: "1",
              text: "Option 1",
              icon: <ImageIcon width={20} height={20} />,
            },
            {
              value: "2",
              text: "Option 2",
              icon: <ImageIcon width={20} height={20} />,
            },
          ]}
          values={values}
          onValuesChange={setValues}
        />
      </FormField>
    </div>
  );
}
