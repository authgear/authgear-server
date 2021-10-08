enum State {
  SLASH = 0,
  ESCAPE = 1,
  CHAR = 2,
}

class StringBuilder {
  private buf: string;
  constructor() {
    this.buf = "";
  }
  writeString(s: string) {
    this.buf += s;
  }
  build(): string {
    return this.buf;
  }
}

// eslint-disable-next-line complexity
export function parseJSONPointer(pointer: string): string[] {
  const output = [];
  let state: State = State.SLASH;
  let w: StringBuilder | null = null;
  for (const r of pointer) {
    switch (state) {
      case State.SLASH: {
        if (r !== "/") {
          throw new Error("expecting / but found: " + r);
        }
        w = new StringBuilder();
        state = State.CHAR;
        break;
      }
      case State.ESCAPE: {
        switch (r) {
          case "0":
            w?.writeString("~");
            break;
          case "1":
            w?.writeString("/");
            break;
          default:
            throw new Error("expecting 0 or 1 but found: " + r);
        }
        break;
      }
      case State.CHAR: {
        switch (r) {
          case "~":
            state = State.ESCAPE;
            break;
          case "/":
            if (w != null) {
              output.push(w.build());
            }
            w = new StringBuilder();
            break;
          default:
            w?.writeString(r);
            break;
        }
        break;
      }
    }
  }

  if (state === State.ESCAPE) {
    throw new Error("expecting 0 or 1 but found: EOF");
  }

  if (w != null) {
    output.push(w.build());
  }

  return output;
}

export function jsonPointerToString(t: string[]): string {
  let buf = "";
  for (const token of t) {
    buf += "/";
    for (const r of token) {
      switch (r) {
        case "~":
          buf += "~0";
          break;
        case "/":
          buf += "~1";
          break;
        default:
          buf += r;
          break;
      }
    }
  }
  return buf;
}

export function parseJSONPointerIntoParentChild(
  pointer: string
): [string, string] | null {
  const t = parseJSONPointer(pointer);
  if (t.length === 0) {
    return null;
  }

  const parent = t.slice(0, t.length - 1);
  const child = t[t.length - 1];

  return [jsonPointerToString(parent), child];
}

export function parentChildToJSONPointer(
  parent: string,
  child: string
): string {
  const pointer = parseJSONPointer(parent);
  pointer.push(child);
  return jsonPointerToString(pointer);
}

export function matchParentChild(
  pointer: string,
  parent: string | RegExp,
  child: string
): boolean {
  const parentChild = parseJSONPointerIntoParentChild(pointer);
  if (parentChild == null) {
    return false;
  }

  if (typeof parent === "string") {
    return parentChild[0] === parent && parentChild[1] === child;
  }

  return parent.test(parentChild[0]) && parentChild[1] === child;
}
