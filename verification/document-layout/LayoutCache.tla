---------------- MODULE LayoutCache ----------------
EXTENDS Naturals
CONSTANT Mutate
VARIABLES cache, geometry, pagesGeometry, mapsGeometry, calls
vars == <<cache, geometry, pagesGeometry, mapsGeometry, calls>>
Init == /\ cache = FALSE /\ geometry = 0 /\ pagesGeometry = 0 /\ mapsGeometry = 0 /\ calls = 0
Layout == /\ calls < 3
          /\ \E g \in 1..2:
             /\ geometry' = g /\ pagesGeometry' = g
             /\ mapsGeometry' = IF Mutate THEN mapsGeometry ELSE g
          /\ cache' = TRUE /\ calls' = calls + 1
Done == /\ calls = 3 /\ UNCHANGED vars
Next == Layout \/ Done
LayoutConsistent == geometry = pagesGeometry /\ geometry = mapsGeometry
CacheReady == calls > 0 => cache
======================================================
