import React from "react";
import { addons, types } from "storybook/manager-api";
import { AddonPanel } from "storybook/internal/components";
import { CommentsPanel } from "./CommentsPanel";
import { ADDON_ID, PANEL_ID } from "./constants";
import { refreshUnresolvedCounts } from "./refreshCounts";
import { SidebarCommentLabel } from "./SidebarCommentLabel";
import { ensureCommentsStyles } from "./styles";

ensureCommentsStyles();

addons.register(ADDON_ID, () => {
  addons.setConfig({
    sidebar: {
      renderLabel: (item, api) => (
        <SidebarCommentLabel item={item} api={api} />
      ),
    },
  });

  void refreshUnresolvedCounts();
  setInterval(() => void refreshUnresolvedCounts(), 15_000);

  addons.add(PANEL_ID, {
    type: types.PANEL,
    title: "Comments",
    match: ({ viewMode }) => !viewMode || viewMode === "story",
    render: ({ active }) => (
      <AddonPanel active={Boolean(active)}>
        <CommentsPanel />
      </AddonPanel>
    ),
  });
});
