------------------------ MODULE LinkScanners ------------------------
EXTENDS Naturals, Sequences
VARIABLES i,pending,output
vars == <<i,pending,output>>
Init == /\ i=0 /\ pending=FALSE /\ output= <<>>
Letter == /\ i<4
          /\ output'=IF pending /\ Len(output)>0 THEN Append(Append(output,"-"),"a") ELSE Append(output,"a")
          /\ pending'=FALSE /\ i'=i+1
Separator == /\ i<4 /\ pending'=TRUE /\ i'=i+1 /\ UNCHANGED output
Punctuation == /\ i<4 /\ i'=i+1 /\ UNCHANGED <<pending,output>>
Done == /\ i=4 /\ UNCHANGED vars
Next == Letter \/ Separator \/ Punctuation \/ Done
Normalized == \A n \in 1..Len(output): output[n]="-" => n>1 /\ n<Len(output) /\ output[n-1]="a" /\ output[n+1]="a"
=============================================================================
