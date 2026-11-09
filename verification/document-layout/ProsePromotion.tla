---------------- MODULE ProsePromotion ----------------
EXTENDS Naturals
VARIABLES line, collected, previousKind, promoted, stopped
vars == <<line,collected,previousKind,promoted,stopped>>
Init == /\ line = 0 /\ collected = 0 /\ previousKind = "start"
        /\ promoted = FALSE /\ stopped = FALSE
ReadLine == /\ line < 3 /\ ~stopped
            /\ \E kind \in {"text","underline","boundary"}:
                /\ promoted' = (collected > 0 /\ kind = "underline")
                /\ stopped' = (kind = "boundary" \/ promoted')
                /\ collected' = IF kind = "text" THEN collected+1 ELSE collected
                /\ previousKind' = kind
            /\ line' = line+1
Done == /\ (line = 3 \/ stopped) /\ UNCHANGED vars
Next == ReadLine \/ Done
AdjacentOnly == promoted => (collected > 0 /\ previousKind = "underline" /\ stopped)
======================================================
