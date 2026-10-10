# Regenerates the 3.5 MB large-book fixture (not committed): python3 -I gen-novel.py > novel.md
import random
random.seed(1)
words = "the quick brown fox jumps over a lazy dog while reading terminal pages".split()
out = ["# Contents", ""]
for c in range(1, 201):
    out.append(f"- [Chapter {c}](#chapter-{c})")
out.append("")
for c in range(1, 201):
    out += [f"# Chapter {c}", ""]
    for p in range(40):
        link = f" See [Chapter {c % 200 + 1}](#chapter-{c % 200 + 1})." if p % 10 == 0 else ""
        out.append(" ".join(random.choice(words) for _ in range(80)) + link)
        out.append("")
print("\n".join(out), end="")
