import React, { useContext, useEffect, useRef, useState } from "react";
import {
  Button,
  Checkbox,
  Flex,
  IconButton,
  Popover,
  Text,
  TextField,
} from "@radix-ui/themes";
import { Link2Icon } from "@radix-ui/react-icons";
import { Context } from "../../../../intl";
import {
  LabelSegment,
  isAllowedLinkURL,
  parseLinkLabel,
  serializeLinkLabel,
} from "./linkLabel";
import styles from "./LinkLabelEditor.module.css";

export interface LinkLabelEditorProps {
  // The label as stored, unescaped; "" when unset.
  value: string;
  placeholder: string;
  disabled: boolean;
  // The new label; "" when cleared.
  onChange: (html: string) => void;
}

// The editor's content as segments: text, and the links in it. Anything
// else the browser inserts is read as its text.
function readSegments(root: HTMLElement): LabelSegment[] {
  const out: LabelSegment[] = [];
  for (const node of Array.from(root.childNodes)) {
    const text = node.textContent ?? "";
    if (node instanceof HTMLAnchorElement) {
      const href = node.getAttribute("href") ?? "";
      out.push(href !== "" ? { text, href } : { text });
    } else {
      out.push({ text });
    }
  }
  return out;
}

function renderSegments(root: HTMLElement, segments: LabelSegment[]): void {
  root.replaceChildren(
    ...segments.map((s) => {
      if (!("href" in s)) {
        return document.createTextNode(s.text);
      }
      const a = document.createElement("a");
      a.setAttribute("href", s.href);
      a.textContent = s.text;
      return a;
    })
  );
}

interface LinkTarget {
  range: Range;
  // The link being edited; null when making a link of the selection.
  anchor: HTMLAnchorElement | null;
}

// A checkbox label with links: select words and click Link, or click a link
// to change or remove it. Stored as text with <a href>, nothing else.
export function LinkLabelEditor({
  value,
  placeholder,
  disabled,
  onChange,
}: LinkLabelEditorProps): React.ReactElement {
  const { renderToString } = useContext(Context);
  const editorRef = useRef<HTMLDivElement>(null);
  const [segments, setSegments] = useState<LabelSegment[]>(
    () => parseLinkLabel(value) ?? []
  );
  const [target, setTarget] = useState<LinkTarget | null>(null);
  const [url, setURL] = useState("");

  // Rendered once; the browser owns the content while it is edited, and
  // the component is remounted for another key or language.
  useEffect(() => {
    if (editorRef.current != null) {
      renderSegments(editorRef.current, segments);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const commit = () => {
    const root = editorRef.current;
    if (root == null) {
      return;
    }
    const next = readSegments(root);
    setSegments(next);
    onChange(serializeLinkLabel(next));
  };

  const openForSelection = () => {
    const selection = window.getSelection();
    const root = editorRef.current;
    if (selection == null || selection.rangeCount === 0 || root == null) {
      return;
    }
    const range = selection.getRangeAt(0);
    if (range.collapsed || !root.contains(range.commonAncestorContainer)) {
      return;
    }
    setTarget({ range: range.cloneRange(), anchor: null });
    setURL("");
  };

  const onEditorClick = (e: React.MouseEvent) => {
    if (!(e.target instanceof HTMLAnchorElement)) {
      return;
    }
    // A link in the editor is edited, never followed, read-only included.
    e.preventDefault();
    if (disabled) {
      return;
    }
    const range = document.createRange();
    range.selectNodeContents(e.target);
    setTarget({ range, anchor: e.target });
    setURL(e.target.getAttribute("href") ?? "");
  };

  const apply = () => {
    if (target == null || !isAllowedLinkURL(url)) {
      return;
    }
    if (target.anchor != null) {
      target.anchor.setAttribute("href", url);
    } else {
      const a = document.createElement("a");
      a.setAttribute("href", url);
      // Links hold text only; markup inside the selection becomes text.
      a.textContent = target.range.toString();
      target.range.deleteContents();
      target.range.insertNode(a);
    }
    setTarget(null);
    commit();
  };

  const removeLink = () => {
    const anchor = target?.anchor;
    if (anchor == null) {
      return;
    }
    anchor.replaceWith(document.createTextNode(anchor.textContent));
    setTarget(null);
    commit();
  };

  return (
    <Flex direction="column" gap="2">
      <Popover.Root
        open={target != null}
        onOpenChange={(open) => {
          if (!open) {
            setTarget(null);
          }
        }}
      >
        <Flex gap="2" align="start">
          {/* A contenteditable box: a textarea cannot show links. */}
          <div
            ref={editorRef}
            className={styles.editor}
            contentEditable={!disabled}
            role="textbox"
            aria-multiline="false"
            aria-disabled={disabled}
            tabIndex={disabled ? -1 : 0}
            data-placeholder={placeholder}
            data-empty={segments.every((s) => s.text === "")}
            onInput={commit}
            onClick={onEditorClick}
            onKeyDown={(e) => {
              // A label is one line.
              if (e.key === "Enter") {
                e.preventDefault();
              }
            }}
            onPaste={(e) => {
              e.preventDefault();
              const text = e.clipboardData.getData("text/plain");
              const selection = window.getSelection();
              if (selection != null && selection.rangeCount > 0) {
                const range = selection.getRangeAt(0);
                range.deleteContents();
                range.insertNode(document.createTextNode(text));
                range.collapse(false);
              }
              commit();
            }}
            suppressContentEditableWarning={true}
          />
          <Popover.Trigger>
            <IconButton
              type="button"
              variant="soft"
              color="gray"
              disabled={disabled}
              aria-label={renderToString("AuthFlowsScreen.form-text.link")}
              // Keeps the selection in the editor.
              onMouseDown={(e) => e.preventDefault()}
              onClick={(e) => {
                e.preventDefault();
                openForSelection();
              }}
            >
              <Link2Icon />
            </IconButton>
          </Popover.Trigger>
        </Flex>
        <Popover.Content size="1" maxWidth="320px">
          <Flex direction="column" gap="2">
            <Text
              as="label"
              size="1"
              weight="medium"
              className={styles.urlLabel}
            >
              {renderToString("AuthFlowsScreen.form-text.link-url")}
              <TextField.Root
                size="1"
                value={url}
                placeholder="https://"
                onChange={(e) => setURL(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    apply();
                  }
                }}
              />
            </Text>
            <Flex gap="2" justify="end">
              {target?.anchor != null ? (
                <Button
                  type="button"
                  size="1"
                  variant="soft"
                  color="red"
                  onClick={removeLink}
                >
                  {renderToString("AuthFlowsScreen.form-text.link-remove")}
                </Button>
              ) : null}
              <Button
                type="button"
                size="1"
                disabled={!isAllowedLinkURL(url)}
                onClick={apply}
              >
                {renderToString("AuthFlowsScreen.form-text.link-apply")}
              </Button>
            </Flex>
          </Flex>
        </Popover.Content>
      </Popover.Root>
      <Text as="div" size="1" color="gray">
        {renderToString("AuthFlowsScreen.form-text.preview")}
      </Text>
      <Flex gap="2" align="start" className={styles.preview}>
        <Checkbox size="1" disabled={true} />
        <Text size="2">
          {segments.every((s) => s.text === "")
            ? placeholder
            : segments.map((s, i) =>
                "href" in s ? (
                  <a
                    key={i}
                    href={isAllowedLinkURL(s.href) ? s.href : undefined}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {s.text}
                  </a>
                ) : (
                  <React.Fragment key={i}>{s.text}</React.Fragment>
                )
              )}
        </Text>
      </Flex>
    </Flex>
  );
}
