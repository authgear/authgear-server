import { useState, useEffect, useCallback } from "react";

interface FormModel {
  state: unknown;
  save: () => Promise<void>;
}

export function useDelayedSave(form: FormModel): () => void {
  const [delaySave, setDelaySave] = useState(false);

  useEffect(() => {
    if (!delaySave) {
      return;
    }

    setDelaySave(false);

    void form.save();
  }, [form, delaySave]);

  const triggerSave = useCallback(() => {
    setDelaySave(true);
  }, []);

  return triggerSave;
}
