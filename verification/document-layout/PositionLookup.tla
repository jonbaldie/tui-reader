---------------- MODULE PositionLookup ----------------
EXTENDS Integers, FiniteSets
VARIABLES cursor, best, eligible, query, mode, stopped
vars == <<cursor,best,eligible,query,mode,stopped>>
Raw == <<0,0,2>>
Offset == <<0,2,0>>
Page == <<0,1,2>>
Init == /\ cursor = 1 /\ best = -1 /\ eligible = {}
        /\ query \in 0..3 /\ mode \in {"position","raw"} /\ stopped = FALSE
Visit == /\ cursor <= 3 /\ ~stopped
         /\ LET ok == IF mode = "position" THEN Raw[cursor] = 0 /\ Offset[cursor] <= query
                      ELSE Raw[cursor] < query
            IN /\ eligible' = IF ok THEN eligible \cup {Page[cursor]} ELSE eligible
               /\ best' = IF ok THEN IF mode = "position" THEN Page[cursor] ELSE IF Page[cursor] > best THEN Page[cursor] ELSE best ELSE best
         /\ stopped' = (mode = "position" /\ Raw[cursor] > 0)
         /\ cursor' = cursor+1 /\ UNCHANGED <<query,mode>>
Done == /\ (cursor = 4 \/ stopped) /\ UNCHANGED vars
Next == Visit \/ Done
GreatestCandidate == IF eligible = {} THEN best = -1 ELSE best \in eligible /\ \A n \in eligible: best >= n
======================================================
