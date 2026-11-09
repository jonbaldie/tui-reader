----------------------- MODULE LinkAttribution -----------------------
EXTENDS Naturals, Sequences
CONSTANTS MutateFence, MutateQueue
VARIABLES phase, inFence, scanned, sources, next, attached, cached
vars == <<phase, inFence, scanned, sources, next, attached, cached>>
Init == /\ phase = "Extract" /\ inFence = FALSE /\ scanned = 0
        /\ sources = <<>> /\ next = 0 /\ attached = <<>> /\ cached = FALSE
OpenFence == /\ phase = "Extract" /\ scanned < 3 /\ ~inFence
             /\ inFence' = TRUE /\ scanned' = scanned + 1
             /\ UNCHANGED <<phase, sources, next, attached, cached>>
CloseFence == /\ phase = "Extract" /\ scanned < 3 /\ inFence
              /\ inFence' = FALSE /\ scanned' = scanned + 1
              /\ UNCHANGED <<phase, sources, next, attached, cached>>
ReadLink == /\ phase = "Extract" /\ scanned < 3
            /\ sources' = IF ~inFence \/ MutateFence
                           THEN Append(sources, [raw |-> scanned, code |-> inFence])
                           ELSE sources
            /\ scanned' = scanned + 1
            /\ UNCHANGED <<phase, inFence, next, attached, cached>>
CacheSource == /\ phase = "Extract" /\ scanned = 3
               /\ phase' = "Attach" /\ cached' = TRUE
               /\ UNCHANGED <<inFence, scanned, sources, next, attached>>
ConsumeOccurrence == /\ phase = "Attach" /\ next < Len(sources)
                     /\ attached' = Append(attached, sources[next+1].raw)
                     /\ next' = IF MutateQueue THEN next ELSE next+1
                     /\ UNCHANGED <<phase, inFence, scanned, sources, cached>>
Finish == /\ phase = "Attach" /\ next = Len(sources)
          /\ phase' = "Done"
          /\ UNCHANGED <<inFence, scanned, sources, next, attached, cached>>
Done == /\ phase = "Done" /\ UNCHANGED vars
Next == OpenFence \/ CloseFence \/ ReadLink \/ CacheSource \/ ConsumeOccurrence \/ Finish \/ Done
NoCodeLinks == \A i \in 1..Len(sources): ~sources[i].code
ConsumedOnce == Len(attached) = next
AttachmentOrder == \A i \in 1..Len(attached): attached[i] = sources[i].raw
CacheBeforeAttach == phase \in {"Attach", "Done"} => cached
Complete == phase = "Done" => Len(attached) = Len(sources)
=============================================================================
