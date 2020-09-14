import { useState, useCallback } from "react";

export function useSimpleRPC<Arg extends any[], Ret>(
  f: (...args: Arg) => Promise<Ret>
): {
  loading: boolean;
  error: unknown;
  rpc: (...args: Arg) => Promise<Ret | null>;
} {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const rpc = useCallback(
    async (...args: Arg) => {
      try {
        setLoading(true);
        const res = await f(...args);
        setError(null);
        setLoading(false);
        return res;
      } catch (err: unknown) {
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
