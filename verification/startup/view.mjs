export const title = 'Open a document and exit';
export const state = v => v.phase;
export const step = action => ({BareDump:'choose dump output', CountedDump:'limit dump pages', File:'choose file', InvalidCount:'give an invalid count', FinishArguments:'finish arguments', Load:'load document', Quit:'quit reader', ProgramError:'terminal program fails', Done:'finished'})[action] ?? action;
