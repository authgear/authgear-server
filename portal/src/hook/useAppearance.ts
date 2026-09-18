import { useSyncExternalStore } from "react";
import {
  Appearance,
  AppearanceState,
  getAppearanceState,
  setAppearance,
  subscribeAppearance,
} from "../util/appearance";

export interface UseAppearanceResult extends AppearanceState {
  setPreference: (preference: Appearance) => void;
}

export function useAppearance(): UseAppearanceResult {
  const state = useSyncExternalStore(
    subscribeAppearance,
    getAppearanceState,
    getAppearanceState
  );
  return {
    preference: state.preference,
    resolved: state.resolved,
    setPreference: setAppearance,
  };
}
