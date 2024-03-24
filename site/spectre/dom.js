/**
 * DOM utilities to ease the pain of document.createElement.
 */

/**
 * A slightly relaxed Node type for my DOM utilities.
 * @typedef {(Node|string|null)} BNode
 *
 * One or more BNodes.
 * @typedef {(BNode|BNode[])} BNodes
 */

/**
 * Ensures a DOM Node.
 * @param {BNode} v
 * @returns {Node}
 */
function N(v) {
  if (typeof v === "string") {
    return document.createTextNode(v);
  }
  return v;
}

/**
 * Adds children to a DOM node.
 * @param {Node} n
 * @param {BNodes} children
 */
function addChildren(n, children) {
  if (Array.isArray(children)) {
    for (const child of children) {
      if (child) {
        n.appendChild(N(child));
      }
    }
  } else {
    if (children) {
      n.appendChild(N(children));
    }
  }
}

/**
 * Creates a DOM element.
 * @param {string} type The type of DOM element to create (e.g. `"div"`)
 * @param {string[]} [classes] Any classes to add to the element
 * @param {BNodes} [children] Any children to add to the element
 * @returns {HTMLElement}
 */
function E(type, classes, children) {
  const el = document.createElement(type);
  if (classes && classes.length > 0) {
    el.classList.add(...classes);
  }
  if (children) {
    addChildren(el, children);
  }
  return el;
}

/**
 * Creates a DOM element with a map of attributes.
 * @param {string} type The type of DOM element to create (e.g. `"div"`)
 * @param {string[]} [classes] Any classes to add to the element
 * @param {*} attributes Any attributes to add to the element
 * @param {BNodes} [children] Any children to add to the element
 */
function EAtts(type, classes, attributes, children) {
  const el = E(type, classes, children);
  for (const [name, value] of Object.entries(attributes)) {
    if (name === "style") {
      for (const [styleName, styleValue] of Object.entries(value)) {
        el.style[styleName] = styleValue;
      }
    } else {
      el.setAttribute(name, value);
    }
  }
  return el;
}

/**
 * Creates a DOM fragment.
 * @param {BNodes} children
 * @returns {DocumentFragment}
 */
function F(children) {
  const f = document.createDocumentFragment();
  addChildren(f, children);
  return f;
}
