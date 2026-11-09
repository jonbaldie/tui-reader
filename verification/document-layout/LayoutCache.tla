---------------- MODULE LayoutCache ----------------
EXTENDS Naturals
CONSTANT Mutate
VARIABLES geometry, pagesGeometry, mapsGeometry, calls
vars == <<geometry, pagesGeometry, mapsGeometry, calls>>
Init == geometry = 0 /\ pagesGeometry = 0 /\ mapsGeometry = 0 /\ calls = 0
Layout == /\ calls < 3
          /\ \E g \in 1..2:
             /\ geometry' = g /\ pagesGeometry' = g
             /\ mapsGeometry' = IF Mutate THEN mapsGeometry ELSE g
          /\ calls' = calls + 1
Done == /\ calls = 3 /\ UNCHANGED vars
Next == Layout \/ Done
LayoutConsistent == geometry = pagesGeometry /\ geometry = mapsGeometry
======================================================
