import { visit as turboVisit } from "@hotwired/turbo";

export function localVisit(u: string, newSearchParams?: URLSearchParams) {
  const currentURL = new URL(window.location.href);
  const newURL = new URL(u, currentURL);

  const searchParams = new URLSearchParams();

  // Preserve query except q_*
  for (const [key, value] of currentURL.searchParams) {
    console.log("louis#key", key);
    console.log("louis#value", value);
    if (!key.startsWith("q_")) {
      searchParams.set(key, value);
    }
  }

  if (newSearchParams != null) {
    for (const [key, value] of newSearchParams) {
      searchParams.set(key, value);
    }
  }

  const search = searchParams.toString();
  if (search !== "") {
    newURL.search = "?" + search;
  }

  turboVisit(newURL);
}
