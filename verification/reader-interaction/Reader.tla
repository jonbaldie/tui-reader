--------------------------- MODULE Reader ---------------------------
EXTENDS Naturals, Integers, Sequences, TLC
CONSTANT BadReset
VARIABLES page, selected, history, geometry, quitting, error
vars == <<page, selected, history, geometry, quitting, error>>
Last == IF geometry = 0 THEN 1 ELSE 0
Links == IF page = 0 THEN 1 ELSE 0
Init == /\ page = 0 /\ selected = -1 /\ history = <<>>
        /\ geometry = 0 /\ quitting = FALSE /\ error \in BOOLEAN
Move == /\ ~error /\ page < Last /\ page' = page + 1
        /\ selected' = (IF BadReset THEN selected ELSE -1)
        /\ UNCHANGED <<history,geometry,quitting,error>>
Previous == /\ ~error /\ page > 0 /\ page' = page - 1 /\ selected' = -1
            /\ UNCHANGED <<history,geometry,quitting,error>>
Home == /\ page' = 0 /\ selected' = -1
        /\ UNCHANGED <<history,geometry,quitting,error>>
End == /\ page' = (IF error THEN page ELSE Last) /\ selected' = -1
       /\ UNCHANGED <<history,geometry,quitting,error>>
Focus == /\ ~error /\ Links > 0 /\ selected' = 0
         /\ UNCHANGED <<page,history,geometry,quitting,error>>
Follow == /\ ~error /\ selected = 0 /\ Len(history) < 2
          /\ history' = Append(history,page) /\ page' = Last /\ selected' = -1
          /\ UNCHANGED <<geometry,quitting,error>>
Back == /\ Len(history) > 0 /\ page' = history[Len(history)]
        /\ history' = SubSeq(history,1,Len(history)-1) /\ selected' = -1
        /\ UNCHANGED <<geometry,quitting,error>>
Resize == /\ geometry' = 1-geometry
          /\ page' = (IF geometry'=1 THEN 0 ELSE page)
          /\ history' = [i \in 1..Len(history) |-> IF geometry'=1 THEN 0 ELSE history[i]]
          /\ selected' = (IF error THEN selected ELSE -1)
          /\ UNCHANGED <<quitting,error>>
Quit == /\ quitting' = TRUE /\ UNCHANGED <<page,selected,history,geometry,error>>
Idle == UNCHANGED vars
Next == Move \/ Previous \/ Home \/ End \/ Focus \/ Follow \/ Back \/ Resize \/ Quit \/ Idle
TypeOK == /\ page \in 0..Last /\ selected \in {-1,0}
          /\ geometry \in {0,1} /\ quitting \in BOOLEAN /\ error \in BOOLEAN
SelectionValid == selected = -1 \/ (~error /\ selected < Links)
HistoryValid == \A i \in 1..Len(history): history[i] \in 0..Last
Spec == Init /\ [][Next]_vars
=====================================================================
