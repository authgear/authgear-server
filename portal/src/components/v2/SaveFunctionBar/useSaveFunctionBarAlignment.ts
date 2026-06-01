import { useLayoutEffect, useState } from "react";

function findScrollParent(element: HTMLElement | null): HTMLElement | null {
  let parent = element?.parentElement ?? null;
  while (parent != null) {
    const { overflowY } = getComputedStyle(parent);
    if (overflowY === "auto" || overflowY === "scroll") {
      return parent;
    }
    parent = parent.parentElement;
  }
  return null;
}

export function useSaveFunctionBarAlignment(
  anchorRef: React.RefObject<HTMLElement | null> | undefined
): React.CSSProperties | undefined {
  const [style, setStyle] = useState<React.CSSProperties | undefined>(() =>
    anchorRef != null ? { visibility: "hidden" } : undefined
  );

  useLayoutEffect(() => {
    const anchor = anchorRef?.current;
    if (anchor == null) {
      setStyle(undefined);
      return;
    }

    const update = () => {
      const rect = anchor.getBoundingClientRect();
      setStyle({
        left: rect.left,
        width: rect.width,
        visibility: "visible",
      });
    };

    update();

    const resizeObserver = new ResizeObserver(update);
    resizeObserver.observe(anchor);
    window.addEventListener("resize", update);

    const scrollParent = findScrollParent(anchor);
    scrollParent?.addEventListener("scroll", update, { passive: true });

    return () => {
      resizeObserver.disconnect();
      window.removeEventListener("resize", update);
      scrollParent?.removeEventListener("scroll", update);
    };
  }, [anchorRef]);

  return style;
}
