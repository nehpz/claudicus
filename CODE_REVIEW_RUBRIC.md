
# AI Code Review Rubric

**Last updated: 2025-06-25**

Use this rubric to evaluate each AI‑generated code file.  
Score **each category 0‑5** using the descriptors, multiply by the weight, then sum for a final _0‑5_ total.  
Files scoring **< 3 in Functionality** **or** **< 2 in Testing** are auto‑rejected regardless of total.

| # | Category | Weight | 5 — Excellent | 4 — Good | 3 — Adequate | 2 — Poor | 1 — Deficient | 0 — Nil/Fail |
|---|----------|--------|---------------|----------|--------------|----------|---------------|--------------|
| **1** | **Functionality & Correctness** | **40 %** | Full feature coverage, flawless logic, robust edge‑case handling | Minor edge‑case gaps or negligible bug | Meets core reqs; some edge cases TBD | Fails req or notable bug | Major features missing | Won’t build/run |
| **2** | **Testing & Coverage** | **25 %** | ≥100 % critical funcs & ≥70 % overall; tests thorough & clean | Slightly < targets but still strong | Tests exist & pass; limited depth | Few tests; superficial cov. | Barely any tests | No tests |
| **3** | **Maintainability & Readability** | **15 %** | Clear structure; idiomatic style; great names, docs | Mostly clean; small style issues | Readable; some clutter/docs gaps | Hard to follow; big funcs | Spaghetti; unreadable | Unmaintainable |
| **4** | **Design Principles (SOLID/DRY)** | **10 %** | Strong SRP, minimal duplication, extensible | Minor duplication/tight coupling | Mostly OK; some smells | Noticeable duplication/rigid | Repeated violations | No discernible design |
| **5** | **Project Standards Adherence** | **10 %** | Perfectly matches patterns; reuses libs/utils | Tiny deviations fixable by lint | Acceptable; several inconsistencies | Ignores many conventions | Contradicts core conventions | Breaks build/lints |

## Grade Bands

|| Total (0‑5) | Action |
||-------------|--------|
|| **≥ 4.5** | **KEEP** outright |
|| **4.0 – 4.49** | Keep; minor cleanup OK |
|| **3.0 – 3.99** | Keep *only* if no higher file; list required fixes |
|| **< 3.0** | **DISCARD** |

### Example Evaluation (partial)

```text
File: user_service.ts
Category | W | S | W·S
1 (Func)        |0.40|4|1.60
2 (Tests)       |0.25|5|1.25
3 (Maintain)    |0.15|4|0.60
4 (Design)      |0.10|3|0.30
5 (Standards)   |0.10|4|0.40
TOTAL = 4.15 → KEEP  (add small refactor for SRP)
```

