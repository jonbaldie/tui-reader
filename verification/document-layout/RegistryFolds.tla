---------------- MODULE RegistryFolds ----------------
EXTENDS Integers, Sequences
VARIABLES history, headings, rawPages, query, answer
vars == <<history, headings, rawPages, query, answer>>
Keys == {0,1}
Init == /\ history = <<>> /\ headings = [k \in Keys |-> -1]
        /\ rawPages = [k \in Keys |-> -1] /\ query = -1 /\ answer = -1
Register == /\ Len(history) < 3
            /\ \E key \in Keys:
                /\ history' = Append(history,key)
                /\ headings' = [headings EXCEPT ![key] = Len(history)]
                /\ rawPages' = IF rawPages[key] = -1 THEN [rawPages EXCEPT ![key] = Len(history)] ELSE rawPages
            /\ UNCHANGED <<query,answer>>
Lookup == /\ Len(history) = 3 /\ query = -1
          /\ query' \in Keys /\ answer' = rawPages[query']
          /\ UNCHANGED <<history,headings,rawPages>>
Done == /\ query # -1 /\ UNCHANGED vars
Next == Register \/ Lookup \/ Done
FirstLast == \A key \in Keys:
  LET found == {i \in 1..Len(history) : history[i] = key}
  IN IF found = {} THEN headings[key] = -1 /\ rawPages[key] = -1
     ELSE /\ \E i \in found: /\ rawPages[key] = i-1 /\ \A j \in found: i <= j
          /\ \E i \in found: /\ headings[key] = i-1 /\ \A j \in found: i >= j
LookupRetainsFirst == query # -1 => answer = rawPages[query]
======================================================
