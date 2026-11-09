-------------------------- MODULE Styling --------------------------
EXTENDS Naturals, Sequences, FiniteSets
VARIABLES tokens, cursor, occurrence, styled
vars == <<tokens,cursor,occurrence,styled>>
Init == /\ tokens \in [1..3 -> {"code","unattached","attached"}]
        /\ cursor = 1 /\ occurrence = 0 /\ styled = <<>>
CopyLiteral == /\ cursor <= 3 /\ tokens[cursor] # "attached"
               /\ cursor' = cursor+1 /\ UNCHANGED <<tokens,occurrence,styled>>
StyleLabel == /\ cursor <= 3 /\ tokens[cursor] = "attached"
              /\ styled' = Append(styled,occurrence) /\ occurrence'=occurrence+1
              /\ cursor'=cursor+1 /\ UNCHANGED tokens
Done == /\ cursor=4 /\ UNCHANGED vars
Next == CopyLiteral \/ StyleLabel \/ Done
OccurrenceAccounting == occurrence = Cardinality({j \in 1..(cursor-1): tokens[j]="attached"})
NoDuplicateOccurrence == \A j \in 1..Len(styled): styled[j]=j-1
Spec == Init /\ [][Next]_vars
====================================================================
