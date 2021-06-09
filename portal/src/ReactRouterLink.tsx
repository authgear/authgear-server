import React, { useCallback } from "react";
import { State, To, createPath } from "history";
import {
  useHref,
  useNavigate,
  useLocation,
  useResolvedPath,
} from "react-router";

function isModifiedEvent(event: React.MouseEvent) {
  return !!(event.metaKey || event.altKey || event.ctrlKey || event.shiftKey);
}

export interface ReactRouterLinkProps
  extends Omit<React.AnchorHTMLAttributes<HTMLAnchorElement>, "href"> {
  replace?: boolean;
  state?: State;
  component?: React.ElementType;
  to: To;
}

// ReactRouterLink is identical to Link from react-router-dom,
// except that it supports the `component` prop.
// In react-router-dom@5, the Link component does support the `component` prop.
// However, the support is gone in react-router-dom@6 :(
export const ReactRouterLink = React.forwardRef<
  HTMLAnchorElement,
  ReactRouterLinkProps
>(function LinkWithRef(
  {
    onClick,
    replace: replaceProp = false,
    state,
    target,
    to,
    component,
    ...rest
  },
  ref
) {
  const href = useHref(to);
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
        event.preventDefault();

        // If the URL hasn't changed, a regular <a> will do a replace instead of
        // a push, so do the same here.
        const replace =
          !!replaceProp || createPath(location) === createPath(path);

        navigate(to, { replace, state });
      }
    },
    [onClick, location, navigate, path, replaceProp, state, target, to]
  );

  const C = component ? component : "a";

  return (
    <C {...rest} href={href} onClick={handleClick} ref={ref} target={target} />
  );
});

export default ReactRouterLink;
