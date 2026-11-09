---------------- MODULE ScannerDecisions ----------------
EXTENDS Integers
VARIABLES machine, steps, pending, emittedSpace, hadSpace, fits, accepted,
         cursor, previousCursor, found, heading, content, result, column,
         digits, suffix, previousSpace, titled, wasLF, consumedCRLF, previousRaw, previousCount
vars == <<machine, steps, pending, emittedSpace, hadSpace, fits, accepted,
         cursor, previousCursor, found, heading, content, result, column,
         digits, suffix, previousSpace, titled, wasLF, consumedCRLF, previousRaw, previousCount>>
Kinds == {"token", "indent", "provenance", "columns", "digits", "title", "lines"}
Init == /\ machine \in Kinds /\ steps = 0 /\ pending = FALSE
        /\ emittedSpace = FALSE /\ hadSpace = FALSE /\ fits = FALSE /\ accepted = FALSE
        /\ cursor = 0 /\ previousCursor = 0 /\ found = FALSE
        /\ heading = FALSE /\ content = FALSE /\ result = "carry" /\ column = 0
        /\ digits = FALSE /\ suffix = FALSE /\ previousSpace = TRUE /\ titled = FALSE
        /\ wasLF = FALSE /\ consumedCRLF = FALSE /\ previousRaw = 0 /\ previousCount = 0
Token == /\ machine = "token" /\ steps < 3 /\ steps' = steps+1
         /\ \E kind \in {"space","word","link"}:
            /\ pending' = (kind = "space") /\ hadSpace' = pending
            /\ emittedSpace' = (kind # "space" /\ pending)
            /\ found' = (kind = "link") /\ result' = kind
            /\ previousCursor' = cursor
            /\ cursor' = cursor + IF kind = "link" THEN 2 ELSE 1
         /\ UNCHANGED <<machine,fits,accepted,heading,content,column,digits,suffix,previousSpace,titled,wasLF,consumedCRLF,previousRaw,previousCount>>
Indent == /\ machine = "indent" /\ cursor < 2 /\ steps' = steps+1
          /\ IF cursor = 0 THEN
                /\ fits' \in BOOLEAN /\ accepted' = FALSE /\ cursor' = 1 /\ result' = "candidate"
              ELSE
                /\ fits' = fits /\ accepted' = fits /\ cursor' = 2
                /\ result' = IF fits THEN "indented" ELSE "full width"
          /\ UNCHANGED <<machine,pending,emittedSpace,hadSpace,previousCursor,found,heading,content,column,digits,suffix,previousSpace,titled,wasLF,consumedCRLF,previousRaw,previousCount>>
Provenance == /\ machine = "provenance" /\ steps < 3 /\ steps' = steps+1
              /\ \E success \in BOOLEAN:
                  /\ found' = success /\ previousCursor' = cursor
                  /\ cursor' = IF success THEN cursor+1 ELSE cursor
                  /\ previousCount' = previousRaw
                  /\ previousRaw' = IF success THEN previousRaw+1 ELSE previousRaw
              /\ UNCHANGED <<machine,pending,emittedSpace,hadSpace,fits,accepted,heading,content,result,column,digits,suffix,previousSpace,titled,wasLF,consumedCRLF>>
Columns == /\ machine = "columns" /\ steps < 3 /\ column < 4 /\ ~suffix /\ steps' = steps+1
           /\ \E ch \in {"space","tab","other"}:
               /\ column' = IF ch = "space" THEN column+1 ELSE IF ch = "tab" THEN column+(4-(column % 4)) ELSE column
               /\ suffix' = (ch = "other") /\ accepted' = (column' >= 4)
           /\ UNCHANGED <<machine,pending,emittedSpace,hadSpace,fits,cursor,previousCursor,found,heading,content,result,digits,previousSpace,titled,wasLF,consumedCRLF,previousRaw,previousCount>>
Digits == /\ machine = "digits" /\ steps < 3 /\ ~suffix /\ steps' = steps+1
          /\ \E ch \in {"digit","suffix","other","eof"}:
             /\ digits' = (digits \/ ch = "digit") /\ suffix' = (ch # "digit")
             /\ accepted' = (digits /\ ch = "suffix") /\ result' = ch
          /\ UNCHANGED <<machine,pending,emittedSpace,hadSpace,fits,cursor,previousCursor,found,heading,content,column,previousSpace,titled,wasLF,consumedCRLF,previousRaw,previousCount>>
Title == /\ machine = "title" /\ steps < 3 /\ steps' = steps+1
         /\ previousSpace' \in BOOLEAN /\ titled' = previousSpace /\ hadSpace' = previousSpace
         /\ UNCHANGED <<machine,pending,emittedSpace,fits,accepted,cursor,previousCursor,found,heading,content,result,column,digits,suffix,wasLF,consumedCRLF,previousRaw,previousCount>>
Lines == /\ machine = "lines" /\ steps < 3 /\ steps' = steps+1
         /\ \E ch \in {"CR","LF","other"}:
             /\ previousCursor' = cursor /\ cursor' = cursor+1
             /\ hadSpace' = wasLF /\ accepted' = (ch = "LF")
             /\ previousRaw' = previousCount
             /\ previousCount' = previousCount + IF ch = "CR" \/ (ch = "LF" /\ ~wasLF) THEN 1 ELSE 0
             /\ consumedCRLF' = (ch = "LF" /\ wasLF)
             /\ wasLF' = (ch = "CR")
         /\ UNCHANGED <<machine,pending,emittedSpace,fits,found,heading,content,result,column,digits,suffix,previousSpace,titled>>
Done == /\ (steps = 3 \/ (machine = "columns" /\ (column >= 4 \/ suffix)) \/ (machine = "indent" /\ cursor = 2) \/ (machine = "digits" /\ suffix)) /\ UNCHANGED vars
Next == Token \/ Indent \/ Provenance \/ Columns \/ Digits \/ Title \/ Lines \/ Done
ObservedRule == CASE machine = "token" -> ((emittedSpace => hadSpace) /\ (found => result = "link" /\ cursor = previousCursor+2))
                 [] machine = "indent" -> (cursor = 2 => (accepted = fits /\ result = IF fits THEN "indented" ELSE "full width"))
                 [] machine = "provenance" -> (cursor >= previousCursor /\ (~found => cursor = previousCursor) /\ previousRaw >= previousCount /\ (~found => previousRaw = previousCount))
                 [] machine = "columns" -> (accepted = (column >= 4))
                 [] machine = "digits" -> (accepted => digits /\ suffix /\ result = "suffix")
                 [] machine = "title" -> titled = hadSpace
                 [] OTHER -> (consumedCRLF => previousCount = previousRaw)
======================================================
