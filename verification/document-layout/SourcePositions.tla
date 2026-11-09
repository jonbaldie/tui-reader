---------------- MODULE SourcePositions ----------------
EXTENDS Integers
CONSTANT Mutate
VARIABLES raw, offset, previousRaw, previousCount, steps, wrongOffset
vars == <<raw, offset, previousRaw, previousCount, steps, wrongOffset>>
Init == /\ raw = -1 /\ offset = 0 /\ previousRaw = -1 /\ previousCount = 0 /\ steps = 0 /\ wrongOffset = FALSE
Line == /\ steps < 3
        /\ \E r \in {-1,0,1}, count \in 1..2, hasOffset \in BOOLEAN, sourceOffset \in 0..2:
           LET fallback == IF r >= 0 /\ r = raw THEN offset+previousCount ELSE 0
               expected == IF hasOffset THEN sourceOffset ELSE fallback
               actual == IF Mutate /\ r # raw THEN offset+previousCount ELSE expected
           IN /\ raw' = r /\ previousRaw' = raw /\ offset' = actual
              /\ previousCount' = count /\ wrongOffset' = (wrongOffset \/ (actual # expected))
        /\ steps' = steps+1
Done == /\ steps = 3 /\ UNCHANGED vars
Next == Line \/ Done
OffsetsResetAtSourceBoundary == ~wrongOffset
======================================================
