---------------- MODULE BlockContext ----------------
EXTENDS Naturals, Sequences
CONSTANT Mutate
VARIABLES first, prevCode, prevList, blankTail, seen, steps, wrongIndent
vars == <<first, prevCode, prevList, blankTail, seen, steps, wrongIndent>>
Init == /\ first = TRUE /\ prevCode = FALSE /\ prevList = FALSE
        /\ blankTail = TRUE /\ seen = FALSE /\ steps = 0 /\ wrongIndent = FALSE
Blank == /\ steps < 3
         /\ first' = IF Mutate THEN TRUE ELSE first
         /\ prevCode' = FALSE /\ prevList' = FALSE /\ blankTail' = TRUE
         /\ steps' = steps + 1 /\ UNCHANGED <<seen, wrongIndent>>
Block(kind) == /\ steps < 3
               /\ wrongIndent' = (wrongIndent \/ (kind = "prose" /\ seen /\ first))
               /\ first' = FALSE /\ seen' = TRUE /\ blankTail' = FALSE
               /\ prevCode' = (kind = "code") /\ prevList' = (kind = "list")
               /\ steps' = steps + 1
Prose == Block("prose")
Code == Block("code")
List == Block("list")
Heading == Block("heading")
Done == /\ steps = 3 /\ UNCHANGED vars
Next == Blank \/ Prose \/ Code \/ List \/ Heading \/ Done
LaterProseKeepsIndent == ~wrongIndent
PriorKindsExclusive == ~(prevCode /\ prevList)
======================================================
