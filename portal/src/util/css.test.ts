/* global describe, it, expect */
import { setCSS, getCSS } from "./css";

const COMMENT = "AUTHGEAR THEME CSS. DO NOT EDIT!";

const CSS = `.a {
  color: green;
}

.b {
  color: green;
}
`;

describe("setCSS", () => {
  it("sets the CSS at the beginning of the file", () => {
    const ROOT = `.a {
  color: red;
}

@some(thing: great) {
  .a {
    color: blue;
  }
}
`;

    const expected = `/* AUTHGEAR THEME CSS. DO NOT EDIT! */

.a {
  color: green;
}

.b {
  color: green;
}

/* AUTHGEAR THEME CSS. DO NOT EDIT! */

.a {
  color: red;
}

@some(thing: great) {
  .a {
    color: blue;
  }
}
`;

    const actual = setCSS(ROOT, CSS, COMMENT);
    expect(actual).toEqual(expected);
  });

  it("sets the CSS", () => {
    const ROOT = `.a {
  color: red;
}

/* AUTHGEAR THEME CSS. DO NOT EDIT! */

.c {
  color: blue;
}

/* AUTHGEAR THEME CSS. DO NOT EDIT! */

@some(thing: great) {
  .a {
    color: blue;
  }
}
`;

    const expected = `.a {
  color: red;
}

/* AUTHGEAR THEME CSS. DO NOT EDIT! */.a {
  color: green;
}

.b {
  color: green;
}

/* AUTHGEAR THEME CSS. DO NOT EDIT! */

@some(thing: great) {
  .a {
    color: blue;
  }
}
`;

    const actual = setCSS(ROOT, CSS, COMMENT);
    expect(actual).toEqual(expected);
  });
});

describe("getCSS", () => {
  it("returns empty array", () => {
    const ROOT = `.a {
  color: red;
}

@some(thing: great) {
  .a {
    color: blue;
  }
}
`;

    const actual = getCSS(ROOT, COMMENT);
    expect(actual).toEqual([]);
  });

  it("returns a list of nodes", () => {
    const ROOT = `.a {
  color: red;
}

/* AUTHGEAR THEME CSS. DO NOT EDIT! */

.c {
  color: blue;
}

/* AUTHGEAR THEME CSS. DO NOT EDIT! */

@some(thing: great) {
  .a {
    color: blue;
  }
}
`;

    const actual = getCSS(ROOT, COMMENT)
      .map((n) => n.toString())
      .join("\n");

    expect(actual).toEqual(`.c {
  color: blue;
}`);
  });
});
