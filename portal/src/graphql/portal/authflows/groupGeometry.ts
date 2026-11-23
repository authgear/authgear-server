// Fixed geometry of the grouped identify node: a "Continue as" page stacked
// above the other options' page, with the use-another-account action between
// them. The continue page has fixed row heights, so the lines inside the group
// can be drawn without measuring the DOM.
// Every value is in px, relative to the group's top-left corner.

// Title and step name, stacked.
const HEADER_HEIGHT = 56;
const PADDING = 16;
// Room left of the pages for the line that forks into them.
const FORK_SPACE = 28;
const FORK_X = 12;
const ROW_GAP = 12;
const PAGE_WIDTH = 240;
const PAGE_PADDING = 10;
const CAPTION_HEIGHT = 14;
const TITLE_HEIGHT = 18;
const BUTTON_ROW_HEIGHT = 34;
const ENTRY_OFFSET = 24;
// The use-another-account action, a small node between the two pages.
const ACTION_WIDTH = 150;
const ACTION_HEIGHT = 26;
// Room above and below the action for the arrows into and out of it.
const ACTION_MARGIN = 20;
const PAGE_GAP = ACTION_MARGIN + ACTION_HEIGHT + ACTION_MARGIN;

export interface IdentifyGroupGeometry {
  width: number;
  headerHeight: number;
  padding: number;
  // Where the step is entered: the target handle, between the two pages.
  entryY: number;
  forkX: number;
  continueLeft: number;
  continueTop: number;
  continueWidth: number;
  continueHeight: number;
  // Where the dashed "if signed in" line enters the continue page.
  continueEntryY: number;
  continueButtonY: number;
  // The use-another-account action between the pages.
  actionLeft: number;
  actionTop: number;
  actionWidth: number;
  actionHeight: number;
  loginLeft: number;
  loginTop: number;
  loginWidth: number;
  // Where the "otherwise" line enters the page below.
  loginEntryY: number;
  // Fixed row heights the continue page renders with.
  rows: {
    pagePadding: number;
    caption: number;
    title: number;
    button: number;
  };
}

export function identifyGroupGeometry(): IdentifyGroupGeometry {
  const pageLeft = FORK_SPACE;
  const continueTop = HEADER_HEIGHT + ROW_GAP;
  const buttonTop = continueTop + PAGE_PADDING + CAPTION_HEIGHT + TITLE_HEIGHT;
  const continueHeight =
    2 * PAGE_PADDING + CAPTION_HEIGHT + TITLE_HEIGHT + BUTTON_ROW_HEIGHT;
  const continueBottom = continueTop + continueHeight;
  const loginTop = continueBottom + PAGE_GAP;
  return {
    width: pageLeft + PAGE_WIDTH + PADDING,
    headerHeight: HEADER_HEIGHT,
    padding: PADDING,
    entryY: continueBottom + PAGE_GAP / 2,
    forkX: FORK_X,
    continueLeft: pageLeft,
    continueTop,
    continueWidth: PAGE_WIDTH,
    continueHeight,
    continueEntryY: continueTop + ENTRY_OFFSET,
    continueButtonY: buttonTop + BUTTON_ROW_HEIGHT / 2,
    actionLeft: pageLeft + (PAGE_WIDTH - ACTION_WIDTH) / 2,
    actionTop: continueBottom + ACTION_MARGIN,
    actionWidth: ACTION_WIDTH,
    actionHeight: ACTION_HEIGHT,
    loginLeft: pageLeft,
    loginTop,
    loginWidth: PAGE_WIDTH,
    loginEntryY: loginTop + ENTRY_OFFSET,
    rows: {
      pagePadding: PAGE_PADDING,
      caption: CAPTION_HEIGHT,
      title: TITLE_HEIGHT,
      button: BUTTON_ROW_HEIGHT,
    },
  };
}
