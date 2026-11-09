----------------------------- MODULE Startup -----------------------------
EXTENDS Integers, Sequences, TLC
VARIABLES phase, dump, limit, path, count, loadFailed, exitCode
vars == <<phase, dump, limit, path, count, loadFailed, exitCode>>
Init == /\ phase = "Arguments" /\ dump = FALSE /\ limit = 0
        /\ path = FALSE /\ count = 0 /\ loadFailed = FALSE /\ exitCode = -1
BareDump == /\ phase = "Arguments" /\ count < 2
            /\ dump' = TRUE /\ count' = count + 1
            /\ UNCHANGED <<phase,limit,path,loadFailed,exitCode>>
CountedDump == /\ phase = "Arguments" /\ count < 2
               /\ dump' = TRUE /\ limit' \in {1,2} /\ count' = count + 1
               /\ UNCHANGED <<phase,path,loadFailed,exitCode>>
File == /\ phase = "Arguments" /\ count < 2
        /\ path' = TRUE /\ count' = count + 1
        /\ UNCHANGED <<phase,dump,limit,loadFailed,exitCode>>
InvalidCount == /\ phase = "Arguments" /\ phase' = "Usage error" /\ exitCode' = 1
                /\ UNCHANGED <<dump,limit,path,count,loadFailed>>
FinishArguments == /\ phase = "Arguments"
                   /\ phase' = IF path THEN "Loading" ELSE "Usage error"
                   /\ exitCode' = IF path THEN -1 ELSE 1
                   /\ UNCHANGED <<dump,limit,path,count,loadFailed>>
Load == /\ phase = "Loading" /\ loadFailed' \in BOOLEAN
        /\ phase' = IF dump THEN IF loadFailed' THEN "Load error" ELSE "Dump complete"
                     ELSE IF loadFailed' THEN "Error screen" ELSE "Reading"
        /\ exitCode' = IF dump THEN IF loadFailed' THEN 1 ELSE 0 ELSE -1
        /\ UNCHANGED <<dump,limit,path,count>>
Quit == /\ phase \in {"Reading","Error screen"}
        /\ phase' = "Exited" /\ exitCode' = IF loadFailed THEN 1 ELSE 0
        /\ UNCHANGED <<dump,limit,path,count,loadFailed>>
ProgramError == /\ phase \in {"Reading","Error screen"}
                /\ phase' = "Program error" /\ exitCode' = 1
                /\ UNCHANGED <<dump,limit,path,count,loadFailed>>
Done == /\ phase \in {"Usage error","Load error","Dump complete","Exited","Program error"}
        /\ UNCHANGED vars
Next == BareDump \/ CountedDump \/ File \/ InvalidCount \/ FinishArguments \/ Load \/ Quit \/ ProgramError \/ Done
TypeOK == /\ dump \in BOOLEAN /\ path \in BOOLEAN /\ loadFailed \in BOOLEAN
          /\ limit \in 0..2 /\ count \in 0..2 /\ exitCode \in {-1,0,1}
LoadFailureCannotSucceed == (loadFailed /\ phase = "Exited") => exitCode = 1
DumpCannotRead == dump => phase \notin {"Reading","Error screen"}
UsageFails == phase = "Usage error" => exitCode = 1
=============================================================================
