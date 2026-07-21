import React, { useCallback, useRef, useState } from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { createMemoryRouter, RouterProvider } from "react-router-dom";
import {
  FormContainerBase,
  FormContainerBaseProps,
} from "../../../FormContainerBase";
import { SystemConfigContext } from "../../../context/SystemConfigContext";
import {
  defaultSystemConfig,
  instantiateSystemConfig,
} from "../../../system-config";
import { SaveFunctionBar } from "./SaveFunctionBar";

const systemConfig = instantiateSystemConfig(defaultSystemConfig);

function SaveFunctionBarStoryHarness({
  startClean = false,
}: {
  startClean?: boolean;
}): React.ReactElement {
  const anchorRef = useRef<HTMLDivElement>(null);
  const [value, setValue] = useState(startClean ? "hello" : "hello!");
  const [savedValue, setSavedValue] = useState("hello");
  const isDirty = value !== savedValue;

  const reset = useCallback(() => {
    setValue(savedValue);
  }, [savedValue]);

  const save = useCallback(async () => {
    setSavedValue(value);
  }, [value]);

  const getIsDirty = useCallback(() => isDirty, [isDirty]);

  const form: FormContainerBaseProps["form"] = {
    getIsDirty,
    isUpdating: false,
    updateError: null,
    reset,
    save,
  };

  return (
    <FormContainerBase form={form}>
      <div style={{ padding: 24, paddingBottom: 120, width: 752 }}>
        <div ref={anchorRef} aria-hidden={true} style={{ height: 0 }} />
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
    (Story) => {
      const router = createMemoryRouter(
        [
          {
            path: "/",
            element: (
              <SystemConfigContext.Provider value={systemConfig}>
                <div style={{ minHeight: 320, position: "relative" }}>
                  <Story />
                </div>
              </SystemConfigContext.Provider>
            ),
          },
        ],
        { initialEntries: ["/"] }
      );

      return <RouterProvider router={router} />;
    },
  ],
} satisfies Meta<typeof SaveFunctionBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <SaveFunctionBarStoryHarness />,
};

export const HiddenWhenClean: Story = {
  render: () => <SaveFunctionBarStoryHarness startClean={true} />,
};
