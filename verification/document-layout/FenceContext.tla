---------------- MODULE FenceContext ----------------
EXTENDS Naturals
CONSTANT Mutate
VARIABLES fence, steps, hiddenHeading
vars == <<fence, steps, hiddenHeading>>
Init == /\ fence = "none" /\ steps = 0 /\ hiddenHeading = FALSE
Open == /\ steps < 3 /\ fence = "none"
        /\ fence' \in {"backticks", "tildes"}
        /\ steps' = steps+1 /\ UNCHANGED hiddenHeading
Close == /\ steps < 3 /\ fence # "none"
         /\ fence' = "none" /\ steps' = steps+1 /\ UNCHANGED hiddenHeading
OtherDelimiter == /\ steps < 3 /\ fence # "none"
                  /\ steps' = steps+1 /\ UNCHANGED <<fence, hiddenHeading>>
Heading == /\ steps < 3 /\ steps' = steps+1
           /\ hiddenHeading' = (hiddenHeading \/ (Mutate /\ fence # "none"))
           /\ UNCHANGED fence
Done == /\ steps = 3 /\ UNCHANGED vars
Next == Open \/ Close \/ OtherDelimiter \/ Heading \/ Done
CodeHeadingNotRegistered == ~hiddenHeading
======================================================
