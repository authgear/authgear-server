import { Declaration, Node, Root, Rule } from "postcss";

interface CssNodeWrapper {
  accept(visitor: CssNodeVisitor): void;
}

export class CssRootNodeWrapper implements CssNodeWrapper {
  readonly root: Root;
  readonly nodes: CssNodeWrapper[];

  constructor(root: Root) {
    this.root = root;
    this.nodes = root.nodes.map((n) => wrapNode(n));
  }

  accept(visitor: CssNodeVisitor): void {
    visitor.visitRoot(this);
  }
}

export class CssRuleNodeWrapper implements CssNodeWrapper {
  readonly rule: Rule;
  readonly nodes: CssNodeWrapper[];

  get selector(): string {
    return this.rule.selector;
  }

  constructor(rule: Rule) {
    this.rule = rule;
    this.nodes = rule.nodes.map((n) => wrapNode(n));
  }

  accept(visitor: CssNodeVisitor): void {
    visitor.visitRule(this);
  }
}

export class CssDeclarationNodeWrapper implements CssNodeWrapper {
  declaration: Declaration;

  constructor(declaration: Declaration) {
    this.declaration = declaration;
  }

  accept(visitor: CssNodeVisitor): void {
    visitor.visitDeclaration(this);
  }
}

export class CssOtherNodeWrapper implements CssNodeWrapper {
  node: Node;

  constructor(node: Node) {
    this.node = node;
  }

  accept(visitor: CssNodeVisitor): void {
    visitor.visitOther(this);
  }
}

function wrapNode(node: Node): CssNodeWrapper {
  if (node instanceof Root) {
    return new CssRootNodeWrapper(node);
  } else if (node instanceof Rule) {
    return new CssRuleNodeWrapper(node);
  } else if (node instanceof Declaration) {
    return new CssDeclarationNodeWrapper(node);
  }
  return new CssOtherNodeWrapper(node);
}

export abstract class CssNodeVisitor {
  abstract visitRoot(root: CssRootNodeWrapper): void;
  abstract visitRule(rule: CssRuleNodeWrapper): void;
  abstract visitDeclaration(declaration: CssDeclarationNodeWrapper): void;
  abstract visitOther(other: CssOtherNodeWrapper): void;
}
