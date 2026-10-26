import { useEffect } from "react";

// Asks the browser to confirm a refresh, a closed tab or an outside link
// while getIsDirty() says there are unsaved changes. In-app navigation is
// NavigationBlockerDialog's job.
export function useLeavePageWarning(getIsDirty: () => boolean): void {
  useEffect(() => {
    const onBeforeUnload = (e: BeforeUnloadEvent) => {
      if (getIsDirty()) {
        e.preventDefault();
        // Older browsers show the prompt only when returnValue is set.
        e.returnValue = "";
      }
    };
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [getIsDirty]);
}
