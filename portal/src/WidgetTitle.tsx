import React, { useEffect, useRef } from "react";
import cn from "classnames";
import { Text } from "@fluentui/react";
import { useLocation } from "react-router-dom";

export interface WidgetTitleProps {
  className?: string;
  children?: React.ReactNode;
  id?: string;
}

const WidgetTitle: React.VFC<WidgetTitleProps> = function WidgetTitle(
  props: WidgetTitleProps
) {
  const { className, children, id } = props;
  const location = useLocation();
  const anchorRef = useRef<HTMLAnchorElement | null>(null);

  useEffect(() => {
    // Scroll to the section if the current fragment matches the id
    // It is needed because usually we have loading states in the screen
    // therefore this component won't be mounted initially
    if (id && location.hash === `#${id}`) {
      requestAnimationFrame(() => {
        if (anchorRef.current != null) {
          anchorRef.current.scrollIntoView();
        }
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const element = (
    <Text
      as="h2"
      variant="xLarge"
      block={true}
      styles={{
        root: {
          // See Widget.
          lineHeight: "28px",
        },
      }}
    >
      {children}
    </Text>
  );

  if (id != null) {
    return (
      <a
        id={id}
        href={"#" + id}
        className={cn(className, "block")}
        ref={anchorRef}
      >
        {element}
      </a>
    );
  }

  return <div className={className}>{element}</div>;
};

export default WidgetTitle;
