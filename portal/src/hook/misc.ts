import { useState, useCallback } from "react";

export function useSimpleRPC<T>(
  f: (...args: any[]) => Promise<T>
): {
  loading: boolean;
  error: Error | null;
  rpc: (...args: unknown[]) => Promise<T | null>;
} {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const rpc = useCallback(
    async (...args: unknown[]) => {
      try {
        setLoading(true);
        const res = await f(...args);
        setError(null);
        setLoading(false);
        return res;
      } catch (err) {
        setError(err);
        setLoading(false);
        return null;
      }
    },
    [f]
  );

  return {
    loading,
    error,
    rpc,
  };
}
