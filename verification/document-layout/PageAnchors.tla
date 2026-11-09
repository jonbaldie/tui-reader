---------------- MODULE PageAnchors ----------------
EXTENDS Integers
VARIABLES page, line, firstHeading, firstContent, last, carry, anchor, chosen
vars == <<page,line,firstHeading,firstContent,last,carry,anchor,chosen>>
Init == /\ page = 0 /\ line = 0 /\ firstHeading = -1 /\ firstContent = -1
        /\ last = 0 /\ carry = 0 /\ anchor = 0 /\ chosen = FALSE
Scan == /\ page < 3 /\ line < 2 /\ ~chosen
        /\ \E kind \in {"blank","text","heading"}:
           /\ firstHeading' = IF kind = "heading" /\ firstHeading = -1 THEN page*2+line ELSE firstHeading
           /\ firstContent' = IF kind # "blank" /\ firstContent = -1 THEN page*2+line ELSE firstContent
           /\ last' = IF kind # "blank" THEN page*2+line ELSE last
        /\ line' = line+1 /\ UNCHANGED <<page,carry,anchor,chosen>>
Choose == /\ page < 3 /\ line = 2 /\ ~chosen
          /\ anchor' = IF page = 0 THEN IF firstContent >= 0 THEN firstContent ELSE 0
                        ELSE IF firstHeading >= 0 THEN firstHeading ELSE IF firstContent >= 0 THEN firstContent ELSE carry
          /\ chosen' = TRUE
          /\ UNCHANGED <<page,line,firstHeading,firstContent,last,carry>>
NextPage == /\ page < 3 /\ chosen /\ page' = page+1 /\ line' = 0
            /\ carry' = IF page = 0 THEN anchor ELSE last
            /\ last' = IF page = 0 THEN anchor ELSE last
            /\ firstHeading' = -1 /\ firstContent' = -1 /\ chosen' = FALSE /\ UNCHANGED anchor
Done == /\ page = 3 /\ UNCHANGED vars
Next == Scan \/ Choose \/ NextPage \/ Done
AnchorPriority == chosen => IF page = 0 THEN anchor = (IF firstContent >= 0 THEN firstContent ELSE 0)
                           ELSE anchor = (IF firstHeading >= 0 THEN firstHeading ELSE IF firstContent >= 0 THEN firstContent ELSE carry)
======================================================
