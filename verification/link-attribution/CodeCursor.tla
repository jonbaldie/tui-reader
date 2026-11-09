------------------------ MODULE CodeCursor ------------------------
EXTENDS Naturals, Sequences
VARIABLES start, finish, first, last, checked
vars == <<start,finish,first,last,checked>>
Spans == <<<<1,3>>,<<4,6>>>>
Init == /\ start=0 /\ finish=0 /\ first=1 /\ last=1 /\ checked=TRUE
Scan == /\ finish<7
        /\ \E a \in start..6: \E b \in (a+1)..7:
              /\ b>=finish
              /\ LET f == IF first=3 \/ a>=6 THEN 3 ELSE IF first=2 \/ a>=3 THEN 2 ELSE 1
                     l == IF last=3 \/ b>4 THEN 3 ELSE IF last=2 \/ b>1 THEN 2 ELSE 1
                     actual == IF f>=l THEN FALSE ELSE Spans[f][1]<a \/ Spans[l-1][2]>b
                     reference == \E n \in 1..2: Spans[n][1]<b /\ a<Spans[n][2] /\ ~(a<=Spans[n][1] /\ Spans[n][2]<=b)
                 IN /\ first'=f /\ last'=l /\ checked'=(actual=reference)
              /\ start'=a /\ finish'=b
Done == /\ finish=7 /\ UNCHANGED vars
Next == Scan \/ Done
OverlapEquivalent == checked
=============================================================================
