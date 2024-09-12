import { Controller } from "@hotwired/stimulus";
import { dispatchDialogOpen } from "./dialog";

/**
 * Dispatch a custom event to set settings dialog open
 */
export function dispatchSettingsDialogOpen(dialogID: string) {
  dispatchDialogOpen(dialogID);
}

export class SettingsDialogController extends Controller {
  open = (e: Event) => {
    const dialogID: string = (e as any).params.dialogid;
    dispatchSettingsDialogOpen(dialogID);
  };
}
