import React, { useRef, useLayoutEffect, RefObject } from "react";

export interface ScaleContainerProps {
  className?: string;
  children?: React.ReactNode;
  childrenRef: RefObject<HTMLElement>;
  mode: "fixed-width";
}

// ScaleContainer scales the children so that the children is visual the same as itself.
// The supported mode is fixed-width, which means the parent width is fixed.
// The children can of any size.
// The aspect ratio is controlled by the children.
// The parent height is then derived from the parent width (fixed) and the aspect ratio.
// Finally the children is scaled to match the parent.
const ScaleContainer: React.FC<ScaleContainerProps> = function ScaleContainer(
  props: ScaleContainerProps
) {
  const containerRef = useRef<HTMLElement | null>(null);
  const { className, children, childrenRef } = props;

  useLayoutEffect(() => {
    const parent = containerRef.current;
    const child = childrenRef.current;
    if (parent == null || child == null) {
      return;
    }

    const childWidth = child.offsetWidth;
    const childHeight = child.offsetHeight;
    const parentWidth = parent.offsetWidth;

    const aspectRatio = childWidth / childHeight;
    const parentHeight = parentWidth / aspectRatio;
    const scale = parentWidth / childWidth;

    // When we use useLayoutEffect, we opt-in to imperative DOM manipulation,
    // so we change the dom directly here, instead of using setState.
    parent.style.height = `${parentHeight}px`;
    child.style.transform = `scale(${scale})`;
  }, [childrenRef]);

  return (
    // @ts-expect-error
    <div ref={containerRef} className={className}>
      {children}
    </div>
  );
};

export default ScaleContainer;
