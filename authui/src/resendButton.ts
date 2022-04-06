export function setupResendButton(): () => void {
  const el = document.querySelector("#resend-button") as HTMLButtonElement;
  if (el == null) {
    return () => {};
  }

  const scheduledAt = new Date();
  const cooldown = Number(el.getAttribute("data-cooldown")) * 1000;
  const label = el.getAttribute("data-label");
  const labelUnit = el.getAttribute("data-label-unit")!;
  let animHandle: number | null = null;

  function tick() {
    const now = new Date();
    const timeElapsed = now.getTime() - scheduledAt.getTime();

    let displaySeconds = 0;
    if (timeElapsed <= cooldown) {
      displaySeconds = Math.round((cooldown - timeElapsed) / 1000);
    }

    if (displaySeconds === 0) {
      el.disabled = false;
      el.textContent = label;
      animHandle = null;
    } else {
      el.disabled = true;
      el.textContent = labelUnit.replace("%d", String(displaySeconds));
      animHandle = requestAnimationFrame(tick);
    }
  }

  animHandle = requestAnimationFrame(tick);

  return () => {
    if (animHandle != null) {
      cancelAnimationFrame(animHandle);
    }
  };
}
