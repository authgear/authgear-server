function findReference(scope, identifier) {
  let current = scope;
  while (current != null) {
    const ref = current.references.find((r) => r.identifier === identifier);
    if (ref != null) {
      return ref;
    }
    current = current.upper;
  }
  return null;
}

function isEventTypeAnnotation(typeAnnotation) {
  if (typeAnnotation == null) {
    return false;
  }
  return /Event\b/.test(typeAnnotation);
}

function getIdentifierTypeAnnotationText(sourceCode, param) {
  if (param.type !== "Identifier" || param.typeAnnotation == null) {
    return null;
  }
  return sourceCode.getText(param.typeAnnotation.typeAnnotation);
}

function isEventParameter(sourceCode, param) {
  if (param.type !== "Identifier") {
    return false;
  }
  if (/^(e|event|evt)$/i.test(param.name)) {
    return true;
  }
  const annotationText = getIdentifierTypeAnnotationText(sourceCode, param);
  return isEventTypeAnnotation(annotationText);
}

module.exports = {
  meta: {
    type: "problem",
    docs: {
      description:
        "Disallow using React event objects in nested callbacks where they may become invalid.",
    },
    schema: [],
    messages: {
      unsafeEventReference:
        "Avoid using event object '{{name}}' inside nested callbacks. Capture needed values first in the handler.",
    },
  },
  create(context) {
    const sourceCode = context.sourceCode;
    const trackedFrames = [];
    let functionDepth = 0;

    function enterFunction(node) {
      functionDepth += 1;

      const scope = sourceCode.getScope(node);
      const trackedVariables = new Set();
      for (const param of node.params) {
        if (!isEventParameter(sourceCode, param) || param.type !== "Identifier") {
          continue;
        }
        const variable = scope.variables.find(
          (v) =>
            v.name === param.name &&
            v.defs.some((def) => def.type === "Parameter")
        );
        if (variable != null) {
          trackedVariables.add(variable);
        }
      }

      if (trackedVariables.size > 0) {
        trackedFrames.push({
          depth: functionDepth,
          trackedVariables,
        });
      }
    }

    function exitFunction() {
      if (
        trackedFrames.length > 0 &&
        trackedFrames[trackedFrames.length - 1].depth === functionDepth
      ) {
        trackedFrames.pop();
      }
      functionDepth -= 1;
    }

    return {
      ":function": enterFunction,
      ":function:exit": exitFunction,
      Identifier(node) {
        if (trackedFrames.length === 0) {
          return;
        }

        const scope = sourceCode.getScope(node);
        const reference = findReference(scope, node);
        if (reference == null || !reference.isRead() || reference.resolved == null) {
          return;
        }

        for (let i = trackedFrames.length - 1; i >= 0; i -= 1) {
          const frame = trackedFrames[i];
          if (functionDepth <= frame.depth) {
            continue;
          }
          if (!frame.trackedVariables.has(reference.resolved)) {
            continue;
          }
          context.report({
            node,
            messageId: "unsafeEventReference",
            data: {
              name: node.name,
            },
          });
          break;
        }
      },
    };
  },
};
