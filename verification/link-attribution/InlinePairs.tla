------------------------- MODULE InlinePairs -------------------------
EXTENDS Naturals, Sequences, FiniteSets
CONSTANT MutatePair
VARIABLES runs, cursor, pairs
vars == <<runs,cursor,pairs>>
Init == /\ runs \in [1..4 -> {1,2}] /\ cursor=1 /\ pairs= <<>>
Candidates == {j \in (cursor+1)..4: runs[j]=runs[cursor]}
First(s) == CHOOSE j \in s: \A k \in s: j<=k
Scan == /\ cursor<=4
        /\ LET candidates == IF MutatePair THEN (cursor+1)..4 ELSE Candidates
               close == IF candidates={} THEN 0 ELSE First(candidates)
           IN /\ cursor'=IF close=0 THEN cursor+1 ELSE close+1
              /\ pairs'=IF close=0 THEN pairs ELSE Append(pairs,<<cursor,close>>)
        /\ UNCHANGED runs
Done == /\ cursor>4 /\ UNCHANGED vars
Next == Scan \/ Done
ExactPairs == \A n \in 1..Len(pairs): runs[pairs[n][1]]=runs[pairs[n][2]]
OrderedPairs == \A n \in 1..Len(pairs): pairs[n][1]<pairs[n][2] /\ (IF n>1 THEN pairs[n-1][2]<pairs[n][1] ELSE TRUE)
EOFRecovery == (cursor>4 /\ runs= <<1,2,2,2>>) => pairs= <<<<2,3>>>>
=============================================================================
