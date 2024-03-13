import { useEffect, useState } from "react";

export function useSnapshotData<T>(data: T | null): T | null {
  const [snapshot, setSnapshot] = useState<T | null>(data);
  useEffect(() => {
    if (data !== null) {
      setSnapshot(data);
    }
  }, [data]);
  return snapshot;
}
