import { Controller } from "@hotwired/stimulus";
// Disable double tap zoom
// There are rumours on the web claiming that touch-action: manipulation can do that.
// However, I tried
// * {
//   touch-action: manipulation;
// }
// and
// * {
//   touch-action: pan-y;
// }
// None of them work.
export class PreventDoubleTapController extends Controller {
  action(e: Event) {
    e.preventDefault();
    e.stopPropagation();
  }
}
