---------------- MODULE WidthAccumulator ----------------
EXTENDS Naturals
CONSTANT Mutate
VARIABLES current, emitted, consumed, steps
vars == <<current, emitted, consumed, steps>>
Init == /\ current = 0 /\ emitted = 0 /\ consumed = 0 /\ steps = 0
Append == /\ steps < 3
          /\ \E w \in 1..2:
            LET overflow == current > 0 /\ current+w > 2
                base == IF overflow THEN 0 ELSE current
                flushed == IF overflow THEN current ELSE 0
                full == base+w >= 2
            IN /\ current' = IF full THEN 0 ELSE base+w
               /\ emitted' = emitted + flushed + IF full /\ ~Mutate THEN base+w ELSE 0
               /\ consumed' = consumed+w
          /\ steps' = steps+1
Done == /\ steps = 3 /\ UNCHANGED vars
Next == Append \/ Done
NoLostColumns == current+emitted = consumed
BufferFits == current < 2
======================================================
