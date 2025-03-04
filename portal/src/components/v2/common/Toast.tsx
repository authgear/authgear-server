import React, {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
} from "react";
import { Toast } from "radix-ui";
import styles from "./Toast.module.css";

export interface ToastProviderContext {
  registerToast: (el: React.ReactElement) => string;
  removeToast: (id: string) => void;
}

const ProviderCtx = createContext<ToastProviderContext | undefined>(undefined);

export interface ToastProviderProps {
  children?: React.ReactChild;
}

let toastID = 0;
function nextToastID(): string {
  toastID += 1;
  return `${toastID}`;
}

export function ToastProvider({
  children,
}: ToastProviderProps): React.ReactElement {
  const [toasts, setToasts] = useState<Map<string, React.ReactElement>>(
    new Map()
  );

  const registerToast = useCallback((el: React.ReactElement): string => {
    const id = nextToastID();
    setToasts((prev) => {
      const newToasts = new Map(prev);
      newToasts.set(id, el);
      return newToasts;
    });
    return id;
  }, []);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => {
      const newToasts = new Map(prev);
      newToasts.delete(id);
      return newToasts;
    });
  }, []);

  const context = useMemo<ToastProviderContext>(() => {
    return {
      registerToast,
      removeToast,
    };
  }, [registerToast, removeToast]);

  return (
    <ProviderCtx.Provider value={context}>
      <Toast.Provider swipeDirection="right">
        {children}
        {Array.from(toasts.entries()).map(([id, el]) => (
          <ToastRoot key={id} id={id}>
            {el}
          </ToastRoot>
        ))}
        <Toast.Viewport className={styles.ToastViewport} />
      </Toast.Provider>
    </ProviderCtx.Provider>
  );
}

export function useToastProviderContext(): ToastProviderContext {
  const ctx = useContext(ProviderCtx);
  if (ctx == null) {
    throw new Error("ToastProviderContext not found");
  }
  return ctx;
}

export interface ToastContext {
  open: boolean;
  setOpen: (value: boolean) => void;
}

const ToastCtx = createContext<ToastContext | undefined>(undefined);

export function useToastContext(): ToastContext {
  const ctx = useContext(ToastCtx);
  if (ctx == null) {
    throw new Error("ToastContext not found");
  }
  return ctx;
}

function ToastRoot({
  id,
  children,
}: {
  id: string;
  children?: React.ReactChild | null;
}): React.ReactElement {
  const { removeToast } = useToastProviderContext();
  const [open, setOpen] = useState(true);
  const onOpenChange = useCallback(
    (value: boolean) => {
      setOpen(value);
      if (!value) {
        // remove it after 1 seconds, to allow it to finish the animation
        setTimeout(() => {
          removeToast(id);
        }, 1000);
      }
    },
    [id, removeToast]
  );

  const ctxValues = useMemo<ToastContext>(
    () => ({
      open,
      setOpen: onOpenChange,
    }),
    [onOpenChange, open]
  );

  return (
    <ToastCtx.Provider value={ctxValues}>
      <Toast.Root
        className={styles.ToastRoot}
        open={open}
        onOpenChange={onOpenChange}
      >
        {children}
      </Toast.Root>
    </ToastCtx.Provider>
  );
}
