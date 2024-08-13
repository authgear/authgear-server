import { useState, useCallback, useMemo } from "react";

export interface ConfirmationDialogStore {
  visible: boolean;
  loading?: boolean;
  show: () => void;
  dismiss: () => void;
  confirm: () => void;
}

export function useConfirmationDialog(): ConfirmationDialogStore {
  const [visible, setVisible] = useState(false);
  const [loading, setLoading] = useState(false);

  const show = useCallback(() => {
    setVisible(true);
  }, []);

  const dismiss = useCallback(() => {
    setVisible(false);
  }, []);

  const confirm = useCallback(() => {
    setLoading(true);
  }, []);

  return useMemo(() => {
    return {
      visible,
      loading,
      show,
      dismiss,
      confirm,
    };
  }, [visible, loading, show, dismiss, confirm]);
}
