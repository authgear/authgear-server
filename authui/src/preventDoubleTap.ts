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
export function setupPreventDoubleTap(): () => void {
  function listener(e: Event) {
    e.preventDefault();
    e.stopPropagation();
  }
  document.addEventListener("dblclick", listener);
  return () => {
    document.removeEventListener("dblclick", listener);
  };
}
