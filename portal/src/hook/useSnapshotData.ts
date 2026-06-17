import { useEffect, useState } from "react";

export function useSnapshotData<T>(data: T | null): T | null {
  const [snapshot, setSnapshot] = useState<T | null>(data);
  useEffect(() => {
    if (data !== null) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setSnapshot(data);
    }
  }, [data]);
  return snapshot;
}
