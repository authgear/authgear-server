import { useLayoutEffect, useRef, useState } from "react";

// Refs are normally populated before layout effects run, so the anchor is
// available immediately. This is only a safety net for an anchor that mounts a
// few frames late; we stop retrying afterwards to avoid a runaway rAF loop.
const MAX_ANCHOR_WAIT_FRAMES = 60;

export function useSaveFunctionBarAlignment(
  anchorRef: React.RefObject<HTMLElement | null> | undefined
): React.CSSProperties | undefined {
  const [style, setStyle] = useState<React.CSSProperties | undefined>(() =>
    anchorRef != null ? { visibility: "hidden" } : undefined
  );
  const lastRectRef = useRef<{ left: number; width: number } | null>(null);

  useLayoutEffect(() => {
    if (anchorRef == null) {
      return;
    }

    let frame: number | null = null;
    let waitedFrames = 0;
    let observedAnchor: HTMLElement | null = null;

    function schedule() {
      if (frame != null) {
        return;
      }
      frame = requestAnimationFrame(() => {
        frame = null;
        apply();
      });
    }

    const resizeObserver = new ResizeObserver(schedule);

    function apply() {
      const anchor = anchorRef!.current;
      if (anchor == null) {
        // Anchor not mounted yet: retry on the next frame, up to a cap.
        if (waitedFrames < MAX_ANCHOR_WAIT_FRAMES) {
          waitedFrames += 1;
          schedule();
        }
        return;
      }
      waitedFrames = 0;
      if (observedAnchor !== anchor) {
        if (observedAnchor != null) {
          resizeObserver.unobserve(observedAnchor);
        }
        resizeObserver.observe(anchor);
        observedAnchor = anchor;
      }
      const rect = anchor.getBoundingClientRect();
      const last = lastRectRef.current;
      if (last?.left === rect.left && last.width === rect.width) {
        return;
      }
      lastRectRef.current = { left: rect.left, width: rect.width };
      setStyle({ left: rect.left, width: rect.width, visibility: "visible" });
    }

    // Measure synchronously when the anchor is already present (the common
    // case) to avoid a one-frame hidden flash; otherwise wait for it.
    if (anchorRef.current != null) {
      apply();
    } else {
      schedule();
    }

    window.addEventListener("resize", schedule);
    // Capture phase receives scroll events from any ancestor scroll container,
    // including the window/document, without having to locate the scroller.
    window.addEventListener("scroll", schedule, {
      capture: true,
      passive: true,
    });

    return () => {
      if (frame != null) {
        cancelAnimationFrame(frame);
      }
      resizeObserver.disconnect();
      window.removeEventListener("resize", schedule);
      window.removeEventListener("scroll", schedule, { capture: true });
    };
  }, [anchorRef]);

  return style;
}
