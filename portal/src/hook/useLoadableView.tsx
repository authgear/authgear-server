import React, { useCallback } from "react";
import ShowError from "../ShowError";
import ShowLoading from "../ShowLoading";

export interface Loadable {
  isLoading: boolean;
  loadError: unknown;
  reload: () => void;
}

export function useLoadableView<T extends readonly Loadable[]>({
  loadables,
  render,
  isLoading,
}: {
  loadables: T;
  render: (loadables: T) => React.ReactElement | null;
  isLoading?: boolean;
}): React.ReactElement | null {
  const reloadAll = useCallback(() => {
    for (const it of loadables) {
      it.reload();
    }
  }, [loadables]);

  if (loadables.some((it) => it.isLoading) || isLoading) {
    return <ShowLoading />;
  }

  if (loadables.some((it) => it.loadError != null)) {
    return (
      <ShowError
        error={loadables.find((it) => it.loadError != null)?.loadError}
        onRetry={reloadAll}
      />
    );
  }

  return render(loadables);
}
