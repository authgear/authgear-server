import Turbolinks from "turbolinks";

let pathnameBeforeOnPopState = window.location.pathname;

function onPopState(_e: Event) {
  // When this event handler runs, location reflects the latest change.
  // So window.location is useless to us here.
  handleBack(pathnameBeforeOnPopState);
}

export function attachPopStateListener() {
  pathnameBeforeOnPopState = window.location.pathname;
  window.addEventListener("popstate", onPopState);
  return () => {
    window.removeEventListener("popstate", onPopState);
  };
}

function handleBack(pathname: string): boolean {
  const pathComponents = getPathComponents(pathname);
  if (isPathComponentsHierarchical(pathComponents)) {
    const newPathname = "/" + pathComponents.slice(0, pathComponents.length - 1).join("/");
    Turbolinks.visit(newPathname, { action: "replace" });
    return true;
  }
  return false;
}

function getPathComponents(pathname: string): string[] {
  const pathComponents = pathname.split("/").filter(c => c !== "");
  return pathComponents;
}

function isPathComponentsHierarchical(pathComponents: string[]): boolean {
  return pathComponents.length > 1 && pathComponents[0] === "settings";
}

function onClickBackButton(e: Event) {
  e.preventDefault();
  e.stopPropagation();
  const handled = handleBack(window.location.pathname);
  if (handled) {
    return;
  }
  window.history.back();
}

export function attachBackButtonListener() {
  const elems = document.querySelectorAll(".back-btn");
  for (let i = 0; i < elems.length; i++) {
    elems[i].addEventListener("click", onClickBackButton);
  }
  return () => {
    for (let i = 0; i < elems.length; i++) {
      elems[i].removeEventListener("click", onClickBackButton);
    }
  };
}

export function toggleBackButtonVisibility() {
  const elems = document.querySelectorAll(".back-btn");
  for (let i = 0; i < elems.length; i++) {
    const element = elems[i];
    const value = element.getAttribute("data-should-show");
    let display;
    if (value !== "true") {
      const pathComponents = getPathComponents(window.location.pathname);
      const hierarchical = isPathComponentsHierarchical(pathComponents);
      if (!hierarchical) {
        display = "none";
      }
    }
    if (display != null) {
      if (element instanceof HTMLElement) {
        element.style.display = display;
      }
    }
  }
}
