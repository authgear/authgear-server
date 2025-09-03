const cssVarsToAttrs = {
  "--alignment-logo": "alignment-logo",
  "--alignment-card": "alignment-card",
  "--alignment-content": "alignment-content",
};

export function injectCSSAttrs(el: HTMLElement) {
  const fn = () => {
    for (const [cssVar, attr] of Object.entries(cssVarsToAttrs)) {
      const value = getComputedStyle(el).getPropertyValue(cssVar).trim();
      el.setAttribute(attr, value);
    }
  };

  // Once fn() is invoked, the page becomes visible.
  // In order to be compatible with our hack of safe area inset on Android,
  // we want to ensure fn() is invoked after
  // the safe area insets are set on :root.
  const fnWithDelay = () => {
    window.setTimeout(() => {
      fn();
    }, 0);
  };

  switch (document.readyState) {
    case "complete":
      fnWithDelay();
      break;
    default:
      window.addEventListener("load", fnWithDelay);
      break;
  }
}
