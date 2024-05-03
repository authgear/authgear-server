import { createContext, useContext } from "react";

export interface UnauthenticatedDialogContextValue {
  setDisplayUnauthenticatedDialog: (shouldDisplay: boolean) => void;
}

export const UnauthenticatedDialogContext = createContext<
  UnauthenticatedDialogContextValue | undefined
>(undefined);

export function useUnauthenticatedDialogContext(): UnauthenticatedDialogContextValue {
  const ctx = useContext(UnauthenticatedDialogContext);
  if (ctx === undefined) {
    throw new Error("UnauthenticatedDialogContext is not provided");
  }
  return ctx;
}
