// disableAllButtons remembers the disabled state of all buttons.
// It disables all button and returns a function to revert to original state.
export function disableAllButtons(): () => void {
  const buttons = document.querySelectorAll("button");
  const original: [HTMLButtonElement, boolean][] = [];
  for (let i = 0; i < buttons.length; i++) {
    const button = buttons[i];
    const disabled = button.disabled;
    const state: [HTMLButtonElement, boolean] = [button, disabled];
    button.disabled = true;
    original.push(state);
  }

  return () => {
    for (let i = 0; i < original.length; i++) {
      const [button, disabled] = original[i];
      button.disabled = disabled;
    }
  };
}
