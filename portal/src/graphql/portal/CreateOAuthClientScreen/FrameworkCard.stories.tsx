import type { Meta, StoryObj } from "@storybook/react-vite";
import { FrameworkCard } from "./FrameworkCard";
import { findFramework } from "./frameworks";

const meta = {
  title: "portal/CreateOAuthClient/FrameworkCard",
  component: FrameworkCard,
  args: {
    framework: findFramework("react")!,
    selected: false,
    onSelect: () => {},
  },
} satisfies Meta<typeof FrameworkCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Selected: Story = {
  args: {
    selected: true,
  },
};

export const Flask: Story = {
  args: {
    framework: findFramework("flask")!,
  },
};

export const NextJs: Story = {
  args: {
    framework: findFramework("nextjs")!,
    selected: true,
  },
};

export const ReactNative: Story = {
  args: {
    framework: findFramework("react-native")!,
  },
};
