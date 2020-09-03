import { useLocation, useResolvedPath } from "react-router-dom";

// useNavIsActive is the hook version of NavLink.
// The implementation is copied from https://github.com/ReactTraining/react-router/blob/v6.0.0-beta.0/packages/react-router-dom/index.tsx#L268
function useNavIsActive(to: string): boolean {
  const location = useLocation();
  const path = useResolvedPath(to);
  const locationPathname = location.pathname;
  const toPathname = path.pathname;
  const isActive = locationPathname === toPathname;
  return isActive;
}

export default useNavIsActive;
