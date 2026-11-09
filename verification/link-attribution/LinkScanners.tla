------------------------ MODULE LinkScanners ------------------------
EXTENDS Naturals, Sequences, Integers
CONSTANTS Mode, MutateRun
VARIABLES i, pending, output, opener, closed, active, groups, cursor, positions, valid
vars == <<i,pending,output,opener,closed,active,groups,cursor,positions,valid>>
Init == /\ i=0 /\ pending=FALSE /\ output= <<>> /\ opener=0 /\ closed=0
        /\ active=-1 /\ groups= <<>> /\ cursor=0 /\ positions= <<>> /\ valid=TRUE
AnchorLetter == /\ Mode="Anchor" /\ i<4
                /\ output'=IF pending /\ Len(output)>0 THEN Append(Append(output,"-"),"a") ELSE Append(output,"a")
                /\ pending'=FALSE /\ i'=i+1
                /\ UNCHANGED <<opener,closed,active,groups,cursor,positions,valid>>
AnchorSeparator == /\ Mode="Anchor" /\ i<4 /\ pending'=TRUE /\ i'=i+1
                   /\ UNCHANGED <<output,opener,closed,active,groups,cursor,positions,valid>>
AnchorPunctuation == /\ Mode="Anchor" /\ i<4 /\ i'=i+1
                     /\ UNCHANGED <<pending,output,opener,closed,active,groups,cursor,positions,valid>>
BacktickRun == /\ Mode="Inline" /\ i<4
               /\ \E run \in {1,2}: /\ opener'=IF opener=0 THEN run ELSE IF opener=run \/ MutateRun THEN 0 ELSE opener
                                    /\ closed'=IF opener=run \/ MutateRun THEN closed+1 ELSE closed
                                    /\ valid'=IF opener=run \/ MutateRun THEN valid /\ opener=run ELSE valid
               /\ i'=i+1
               /\ UNCHANGED <<pending,output,active,groups,cursor,positions>>
Prose == /\ Mode="Grouping" /\ i<4
         /\ active'=IF active<0 THEN i ELSE active
         /\ groups'=Append(groups,IF active<0 THEN i ELSE active) /\ i'=i+1
         /\ UNCHANGED <<pending,output,opener,closed,cursor,positions,valid>>
Boundary == /\ Mode="Grouping" /\ i<4
            /\ active'=-1 /\ groups'=Append(groups,-1) /\ i'=i+1
            /\ UNCHANGED <<pending,output,opener,closed,cursor,positions,valid>>
Setext == /\ Mode="Grouping" /\ i<4
          /\ active'=IF active>=0 THEN -1 ELSE i
          /\ groups'=Append(groups,active') /\ i'=i+1
          /\ UNCHANGED <<pending,output,opener,closed,cursor,positions,valid>>
MatchSource == /\ Mode="SourceCursor" /\ i<3
               /\ \E p \in 0..29: /\ (IF Len(positions)=0 THEN TRUE ELSE p>=positions[Len(positions)])
                                  /\ positions'=Append(positions,p)
                                  /\ cursor'=IF p>=20 THEN 2 ELSE IF p>=10 THEN 1 ELSE 0
               /\ i'=i+1
               /\ UNCHANGED <<pending,output,opener,closed,active,groups,valid>>
Done == /\ i=(IF Mode="SourceCursor" THEN 3 ELSE 4) /\ UNCHANGED vars
Next == AnchorLetter \/ AnchorSeparator \/ AnchorPunctuation \/ BacktickRun \/ Prose \/ Boundary \/ Setext \/ MatchSource \/ Done
Normalized == /\ \A n \in 1..Len(output): output[n] \in {"a","-"}
              /\ \A n \in 1..Len(output): output[n]="-" => n>1 /\ n<Len(output) /\ output[n-1]="a" /\ output[n+1]="a"
ExactRunScope == valid /\ opener \in {0,1,2} /\ closed<=i\div 2
GroupsValid == \A n \in 1..Len(groups): groups[n]=-1 \/ groups[n] \in 0..(n-1)
CursorOwnsMatch == Len(positions)>0 => cursor*10<=positions[Len(positions)] /\ (cursor=2 \/ (cursor+1)*10>positions[Len(positions)])
=============================================================================
