export const title = 'Extract links outside code, then attach each occurrence';
export const state = v => v.phase === 'Extract' ? (v.inFence ? 'Inside code' : 'Reading Markdown') : v.phase === 'Attach' ? 'Matching link occurrences' : 'Links ready';
export const step = a => ({OpenFence:'code block starts', CloseFence:'code block ends', ReadLink:'read an internal link', CacheSource:'finish reading the document', ConsumeOccurrence:'place the next link', Finish:'finish matching', Done:'finished'})[a] ?? a;
