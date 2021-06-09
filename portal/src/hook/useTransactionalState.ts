import { useState, useCallback } from "react";

export interface TransactionalState<T> {
  committedValue: T;
  uncommittedValue: T;
  setValue: (value: T) => void;
  setCommittedValue: (value: T) => void;
  commit: () => void;
  rollback: () => void;
}

export default function useTransactionalState<T>(
  value: T
): TransactionalState<T> {
  const [committedValue, setCommittedValue] = useState(value);
  const [uncommittedValue, setUncommittedValue] = useState(value);

  const setBoth = useCallback((value: T) => {
    setCommittedValue(value);
    setUncommittedValue(value);
  }, []);

  const commit = useCallback(() => {
    setCommittedValue(uncommittedValue);
  }, [uncommittedValue]);

  const rollback = useCallback(() => {
    setUncommittedValue(committedValue);
  }, [committedValue]);

  return {
    committedValue,
    uncommittedValue,
    setValue: setUncommittedValue,
    setCommittedValue: setBoth,
    commit,
    rollback,
  };
}
