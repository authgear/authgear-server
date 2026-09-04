import React from "react";
import { Navigate, Outlet } from "react-router-dom";
import ShowLoading from "./ShowLoading";
import ShowError from "./ShowError";

export interface EntityRouteGuardProps {
  loading: boolean;
  error: unknown;
  refetch: () => unknown;
  entity: unknown;
  redirectTo: string;
}

// EntityRouteGuard is the shared body of the Require* route-level guards
// (RequireUser, RequireAPIResource, RequireRole, RequireGroup). Each guard
// loads its project-scoped entity and delegates here: while loading show the
// spinner, on error show it with retry, and when the entity does not exist in
// the current project (e.g. after switching projects with a stale URL)
// redirect to the entity's list page; otherwise render the matched child
// route via <Outlet>.
export const EntityRouteGuard: React.VFC<EntityRouteGuardProps> =
  function EntityRouteGuard({ loading, error, refetch, entity, redirectTo }) {
    if (loading) {
      return <ShowLoading />;
    }

    if (error != null) {
      return (
        <ShowError
          error={error}
          onRetry={() => {
            refetch();
          }}
        />
      );
    }

    if (entity == null) {
      return <Navigate to={redirectTo} replace={true} />;
    }

    return <Outlet />;
  };
