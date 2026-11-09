export const title = "Each link occurrence owns one navigation entry";
export const state = v => Number(v.next) === 3 ? "Both occurrences attached" : Number(v.next) === 2 ? "First occurrence attached" : "Layout placements ready";
export const step = a => a === "Place" ? "attach next occurrence and its label pieces" : "finished";
