import type { Meta, StoryObj } from "@storybook/react-vite";
import { AuthMethodChoiceComponent } from "./AuthMethodChoice";

const meta = {
  title: "portal/CreateOAuthClient/AuthMethodChoice",
  component: AuthMethodChoiceComponent,
  args: {
    value: null,
    onChange: () => {},
    nginxDocsHref: "https://docs.authgear.com/get-started/backend-api/nginx",
  },
} satisfies Meta<typeof AuthMethodChoiceComponent>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const TokenSelected: Story = {
  args: {
    value: "token",
  },
};

export const CookieSelected: Story = {
  args: {
    value: "cookie",
  },
};
