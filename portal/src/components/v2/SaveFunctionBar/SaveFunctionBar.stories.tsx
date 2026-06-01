import React, { useCallback, useRef, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import {
  FormContainerBase,
  FormContainerBaseProps,
} from "../../../FormContainerBase";
import { SaveFunctionBar } from "./SaveFunctionBar";

function SaveFunctionBarStoryHarness({
  startClean = false,
}: {
  startClean?: boolean;
}): React.ReactElement {
  const anchorRef = useRef<HTMLDivElement>(null);
  const [value, setValue] = useState(startClean ? "hello" : "hello!");
  const [savedValue, setSavedValue] = useState("hello");
  const isDirty = value !== savedValue;

  const form: FormContainerBaseProps["form"] = {
    isDirty,
    isUpdating: false,
    updateError: null,
    reset: useCallback(() => {
      setValue(savedValue);
    }, [savedValue]),
    save: useCallback(async () => {
      setSavedValue(value);
    }, [value]),
  };

  return (
    <FormContainerBase form={form}>
      <div style={{ padding: 24, paddingBottom: 120, width: 752 }}>
        <div ref={anchorRef} aria-hidden style={{ height: 0 }} />
        <label>
          Edit to show save bar
          <input
            value={value}
            onChange={(e) => setValue(e.target.value)}
            style={{ display: "block", marginTop: 8, width: "100%" }}
          />
        </label>
      </div>
      <SaveFunctionBar anchorRef={anchorRef} />
    </FormContainerBase>
  );
}

const meta = {
  component: SaveFunctionBar,
  tags: ["autodocs"],
  decorators: [
    (Story) => (
      <div style={{ minHeight: 240 }}>
        <Story />
      </div>
    ),
  ],
} satisfies Meta<typeof SaveFunctionBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <SaveFunctionBarStoryHarness />,
};

export const HiddenWhenClean: Story = {
  render: () => <SaveFunctionBarStoryHarness startClean />,
};
