------------------------- MODULE LinkAttribution -------------------------
EXTENDS Naturals, Integers, Sequences, FiniteSets
VARIABLES height, starts, offsets, next, attached, spans
vars == <<height,starts,offsets,next,attached,spans>>
Init == /\ height \in {1,2}
        /\ starts \in [1..2 -> 0..2] /\ starts[1]<=starts[2]
        /\ offsets \in [1..2 -> {0,1}]
        /\ next=1 /\ attached= <<>> /\ spans={}
Owner(i) == starts[i] \div height
Local(i) == Cardinality({j \in 1..(i-1): Owner(j)=Owner(i)})
Expected(i) == LET labelPage == (starts[i]+offsets[i]) \div height
              IN {[id |-> i, page |-> labelPage, index |-> IF labelPage=Owner(i) THEN Local(i) ELSE -1, opener |-> FALSE]}
                 \cup IF labelPage=Owner(i) THEN {} ELSE {[id |-> i, page |-> Owner(i), index |-> Local(i), opener |-> TRUE]}
Place == /\ next<=2
         /\ attached'=Append(attached,[id |-> next,page |-> Owner(next),index |-> Local(next)])
         /\ spans'=spans \cup Expected(next)
         /\ next'=next+1
         /\ UNCHANGED <<height,starts,offsets>>
Done == /\ next=3 /\ UNCHANGED vars
Next == Place \/ Done
OccurrenceOrder == attached= [i \in 1..(next-1) |-> [id |-> i,page |-> Owner(i),index |-> Local(i)]]
SpanOwnership == \A s \in spans: ~s.opener => s.index=IF s.page=Owner(s.id) THEN Local(s.id) ELSE -1
OwnerFallback == \A i \in 1..(next-1): (\E s \in spans: s.id=i /\ s.opener) \equiv ((starts[i]+offsets[i]) \div height # Owner(i))
CompleteSpans == spans=UNION {Expected(i): i \in 1..(next-1)}
=============================================================================
