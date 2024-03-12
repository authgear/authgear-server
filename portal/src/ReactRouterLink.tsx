import React, { useCallback } from "react";
import { To, createPath } from "history";
import {
  useHref,
  useNavigate,
  useLocation,
  useResolvedPath,
} from "react-router-dom";

function isModifiedEvent(event: React.MouseEvent) {
  return !!(event.metaKey || event.altKey || event.ctrlKey || event.shiftKey);
}

export interface ReactRouterLinkPropsBase {
  children?: React.ReactNode;
  replace?: boolean;
  state?: any;
  preserveDefault?: boolean;
  to: To;
}

export interface ReactRouterLinkProps
  extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, "href">,
    ReactRouterLinkPropsBase {}

function useHandleClick({
  to,
  onClick,
  target,
  preserveDefault,
  replace: replaceProp,
  state,
}: {
  to: To;
  onClick?: (e: React.MouseEvent<any>) => void;
  target?: React.HTMLAttributeAnchorTarget;
  preserveDefault?: boolean;
  replace?: boolean;
  state?: unknown;
}) {
  const navigate = useNavigate();
  const location = useLocation();
  const path = useResolvedPath(to);

  const handleClick = useCallback(
    (event: React.MouseEvent<HTMLAnchorElement>) => {
      if (onClick) onClick(event);
      if (
        !event.defaultPrevented && // onClick prevented default
        event.button === 0 && // Ignore everything but left clicks
        (!target || target === "_self") && // Let browser handle "target=_blank" etc.
        !isModifiedEvent(event) // Ignore clicks with modifier keys
      ) {
        if (!preserveDefault) {
          // We do not want to prevent default in some cases, such as scrolling to a section by fragment.
          event.preventDefault();
        }

        // If the URL hasn't changed, a regular <a> will do a replace instead of
        // a push, so do the same here.
        const replace =
          !!replaceProp || createPath(location) === createPath(path);

        navigate(to, { replace, state });
      }
    },
    [
      onClick,
      target,
      preserveDefault,
      replaceProp,
      location,
      path,
      navigate,
      to,
      state,
    ]
  );
  return handleClick;
}

// ReactRouterLink is identical to Link from react-router-dom,
// except that it supports the `component` prop.
// In react-router-dom@5, the Link component does support the `component` prop.
// However, the support is gone in react-router-dom@6 :(
export const ReactRouterLink = React.forwardRef<
  HTMLAnchorElement,
  ReactRouterLinkProps
>(function LinkWithRef(props, ref) {
  const { onClick, state, target, to, preserveDefault, ...rest } = props;
  const href = useHref(to);

  const handleClick = useHandleClick(props);

  return (
    <a {...rest} href={href} onClick={handleClick} ref={ref} target={target} />
  );
});

interface BasicComponentProps {
  onClick?: ((ev: React.MouseEvent<any>) => void) | undefined;
  href?: string | undefined;
}

// Typesafe implementation of using a non-anchor based custom component
export function ReactRouterLinkComponent<P extends BasicComponentProps>(
  props: P & ReactRouterLinkPropsBase & { component: React.ComponentType<P> }
): React.ReactElement | null {
  const { onClick, state, to, component, preserveDefault, ...rest } = props;
  const href = useHref(to);

  const handleClick = useHandleClick(props);

  const Component = component as React.ComponentType<BasicComponentProps>;

  return <Component {...rest} href={href} onClick={handleClick} />;
}

export default ReactRouterLink;
