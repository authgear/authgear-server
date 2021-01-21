import { parse, Comment, Node, Root } from "postcss";

// Locate the start index and the end index delimited by the comment.
function locate(
  root: Root,
  comment: string
): {
  root: Root;
  start: number;
  end: number;
} {
  let start = -1;
  let end = -1;
  for (let i = 0; i < root.nodes.length; i++) {
    const currentNode = root.nodes[i];
    if (currentNode instanceof Comment) {
      if (currentNode.text === comment) {
        if (start === -1 && end === -1) {
          // The first time we saw the comment.
          // Remember the index as start.
          start = i;
        } else if (start !== -1 && end === -1) {
          // The second time we saw the comment.
          // Remember the index as end.
          end = i;
        } else {
          // Otherwise we the comment appeared more then twice :(
          // We do not know how to handle that programmatically.
          throw new Error(
            "The given CSS file has special comment appeared more than twice. Please review the file manually."
          );
        }
      }
    }
  }

  // When we reach here, we have 3 outcomes.
  //
  // start === -1 && end === -1
  // start !== -1 && end === -1
  // start !== -1 && end !== -1
  //
  // The 1st and the 3rd case are normal.
  // The 2nd case is also broken.

  if (start !== -1 && end === -1) {
    throw new Error(
      "The given CSS file has special comment appeared only once. Please review the file manually."
    );
  }

  return {
    root,
    start,
    end,
  };
}

// eslint-disable-next-line complexity
export function setCSS(
  rootString: string,
  cssString: string,
  comment: string
): string {
  const root = parse(rootString);
  const { start, end } = locate(root, comment);

  // The file does not have the special comment.
  // Add the css at the beginning of the file.
  if (start === -1 && end === -1) {
    const css = parse(cssString);
    const commentStart = new Comment({ text: comment });
    const commentEnd = new Comment({ text: comment });

    root.prepend(commentEnd);
    for (let i = css.nodes.length - 1; i >= 0; i--) {
      root.prepend(css.nodes[i].clone());
    }
    root.prepend(commentStart);
  }

  // The file has the special comment.
  if (start !== -1 && end !== -1) {
    const css = parse(cssString);
    const newNodes = [];
    for (const node of css.nodes) {
      newNodes.push(node.clone());
    }
    const deleteCount = end - start - 1;
    root.nodes.splice(start + 1, deleteCount, ...newNodes);
  }

  return root.toResult().css;
}

export function getCSS(rootString: string, comment: string): Node[] {
  const root = parse(rootString);
  const { start, end } = locate(root, comment);

  // The file does not have the special comment.
  // Return empty array.
  if (start === -1 && end === -1) {
    return [];
  }

  const output = [];
  for (let i = start + 1; i < end; i++) {
    const node = root.nodes[i];
    output.push(node.clone());
  }

  return output;
}
