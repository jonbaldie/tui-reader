---------------- MODULE RegistryFolds ----------------
EXTENDS Integers, Sequences
VARIABLES history, headings, rawPages, query, answer
vars == <<history, headings, rawPages, query, answer>>
Keys == {0,1}
Init == /\ history = <<>> /\ headings = [k \in Keys |-> -1]
        /\ rawPages = [k \in Keys |-> -1] /\ query = -1 /\ answer = -1
Register == /\ Len(history) < 3
            /\ \E key \in Keys, end \in Keys:
                /\ end >= key
                /\ history' = Append(history,[first |-> key, last |-> end])
                /\ headings' = [headings EXCEPT ![key] = Len(history)]
                /\ rawPages' = [k \in Keys |-> IF k >= key /\ k <= end /\ rawPages[k] = -1 THEN Len(history) ELSE rawPages[k]]
            /\ UNCHANGED <<query,answer>>
Lookup == /\ Len(history) = 3 /\ query = -1
          /\ query' \in Keys /\ answer' = rawPages[query']
          /\ UNCHANGED <<history,headings,rawPages>>
Done == /\ query # -1 /\ UNCHANGED vars
Next == Register \/ Lookup \/ Done
FirstLast == \A key \in Keys:
  LET headingFound == {i \in 1..Len(history) : history[i].first = key}
      rawFound == {i \in 1..Len(history) : history[i].first <= key /\ key <= history[i].last}
  IN /\ IF headingFound = {} THEN headings[key] = -1
         ELSE \E i \in headingFound: /\ headings[key] = i-1 /\ \A j \in headingFound: i >= j
     /\ IF rawFound = {} THEN rawPages[key] = -1
         ELSE \E i \in rawFound: /\ rawPages[key] = i-1 /\ \A j \in rawFound: i <= j

LookupRetainsFirst == query # -1 => answer = rawPages[query]
======================================================
