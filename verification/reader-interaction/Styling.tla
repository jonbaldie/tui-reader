-------------------------- MODULE Styling --------------------------
EXTENDS Naturals, Integers, Sequences
VARIABLES spans, selected, cursor, last, rendered, highlighted
vars == <<spans,selected,cursor,last,rendered,highlighted>>
Kinds == {"first","next","continuation","opener","other"}
Start(k) == IF k \in {"next","continuation"} THEN 1 ELSE 0
End(k) == Start(k)+1
Owner(k) == IF k = "next" THEN 1 ELSE IF k = "continuation" THEN -1 ELSE 0
OtherLine(k) == k = "other"
IsOpener(k) == k = "opener"
Init == /\ spans \in [1..2 -> Kinds] /\ selected \in {-1,0,1}
        /\ cursor=1 /\ last=0 /\ rendered= <<>> /\ highlighted= <<>>
Eligible(k) == ~OtherLine(k) /\ Start(k)>=last /\ (~IsOpener(k) \/ Owner(k)=selected)
Scan == /\ cursor<=2
        /\ LET k == spans[cursor]
           IN IF Eligible(k) THEN
                /\ rendered'=Append(rendered,k) /\ last'=End(k)
                /\ highlighted'=IF Owner(k)=selected /\ selected>=0 THEN Append(highlighted,k) ELSE highlighted
              ELSE UNCHANGED <<rendered,last,highlighted>>
        /\ cursor'=cursor+1 /\ UNCHANGED <<spans,selected>>
Done == /\ cursor=3 /\ UNCHANGED vars
Next == Scan \/ Done
OrderedRanges == /\ last = (IF Len(rendered)=0 THEN 0 ELSE End(rendered[Len(rendered)]))
                 /\ \A n \in 2..Len(rendered): Start(rendered[n])>=End(rendered[n-1])
SelectedOwner == \A n \in 1..Len(highlighted): Owner(highlighted[n])=selected /\ selected>=0
OpenerSelection == \A n \in 1..Len(rendered): ~IsOpener(rendered[n]) \/ Owner(rendered[n])=selected
LineIsolation == \A n \in 1..Len(rendered): ~OtherLine(rendered[n])
Spec == Init /\ [][Next]_vars
====================================================================
