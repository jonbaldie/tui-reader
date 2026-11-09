export const title = "Document layout context";
export const state = v => v.first ? "Before first block" : v.prevCode ? "After code" : v.prevList ? "After list" : v.blankTail ? "After separator" : "After text";
export const step = (a,from,to) => a === "Block" ? to.prevCode ? "read code" : to.prevList ? "read list" : "read text" : a === "Blank" ? "blank line" : "end of input";
