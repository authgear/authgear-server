import { useState, useEffect, Dispatch, SetStateAction } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";

export function useSearchParamsState<T extends string>(
  key: string,
  initialValue: T
): [state: T, setState: Dispatch<SetStateAction<T>>] {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const initialValueFromParams = searchParams.get(key) ?? initialValue;

  const [state, setState] = useState<T>(initialValueFromParams as T);

  useEffect(() => {
    const newParams = new URLSearchParams(searchParams);
    if (state !== initialValue) {
      newParams.set(key, state);
    } else {
      newParams.delete(key);
    }
    navigate(
      {
        search: newParams.toString(),
        hash: location.hash,
        pathname: location.pathname,
      },
      { replace: true }
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state, key, initialValue]);

  return [state, setState];
}
