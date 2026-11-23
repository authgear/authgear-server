// A checkbox label: text with links, the only markup the server keeps in a
// boolean field's label.

export type LabelSegment = { text: string } | { text: string; href: string };

// The label's text and links; null when it has any other markup, so the
// editor never rewrites what it cannot show.
export function parseLinkLabel(html: string): LabelSegment[] | null {
  const doc = new DOMParser().parseFromString(
    `<body>${html}</body>`,
    "text/html"
  );
  const out: LabelSegment[] = [];
  for (const node of Array.from(doc.body.childNodes)) {
    if (node.nodeType === Node.TEXT_NODE) {
      out.push({ text: node.textContent ?? "" });
      continue;
    }
    if (!(node instanceof HTMLAnchorElement)) {
      return null;
    }
    const href = node.getAttribute("href");
    if (
      href == null ||
      // A stored link the builder would not make is left for Custom Text.
      !isAllowedLinkURL(href) ||
      node.attributes.length !== 1 ||
      Array.from(node.childNodes).some((c) => c.nodeType !== Node.TEXT_NODE)
    ) {
      return null;
    }
    out.push({ text: node.textContent, href });
  }
  return out;
}

function escapeHTML(text: string): string {
  return text
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

export function serializeLinkLabel(segments: LabelSegment[]): string {
  let out = "";
  for (const s of segments) {
    if (s.text === "") {
      continue;
    }
    out +=
      "href" in s
        ? `<a href="${escapeHTML(s.href)}">${escapeHTML(s.text)}</a>`
        : escapeHTML(s.text);
  }
  return out;
}

export function isAllowedLinkURL(url: string): boolean {
  try {
    const parsed = new URL(url);
    return parsed.protocol === "https:" || parsed.protocol === "http:";
  } catch {
    return false;
  }
}
