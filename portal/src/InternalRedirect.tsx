import React, { useEffect } from "react";
import { useNavigate, useLocation } from "react-router-dom";

export interface InternalRedirectState {
  originalPath: string;
  state: unknown;
}

export default function InternalRedirect(): React.ReactElement | null {
  const location = useLocation();
  const navigate = useNavigate();
  const state: InternalRedirectState = location.state as InternalRedirectState;

  useEffect(() => {
    navigate(state.originalPath, {
      state: state.state,
      replace: true,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return null;
}
