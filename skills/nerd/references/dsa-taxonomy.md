# DSA Taxonomy — Deep Dive Reference

Loaded by the `nerd` skill when the human asks to zoom into a specific data structure or algorithm. Each entry has: **fundamental nature** (how it actually works, not the marketing), **invariant** (the condition that must hold before/during/after — where correctness and bugs live), **memory/conceptual layout**, **operations + cost**, **when it appears in code**, and **what to draw**.

---

## Data Structures

### Hash Map / Dict / Associative Array
- **Nature:** A function from key → value, backed by an array of "buckets." A hash function turns the key into an array index. Collisions (two keys, same index) are resolved by chaining (each bucket is a small list) or open addressing (probe to the next free slot).
- **Invariant:** Every key is located at `buckets[hash(key) % len]` (or reachable via its collision chain/probe sequence); no key exists in two places; the number of stored keys equals the distinct-key count. Violation → silent data loss or double-counting.
- **Layout:**
  ```
  hash(key) ──► index ──► buckets[ idx ]
                          ┌──────────────┐
  buckets:  [0] ────────► │ k:v │ k:v │   ← chain (linked list)
            [1] ─────────► │ k:v │
            [2] ─────────► (empty)
            [3] ─────────► │ k:v │ k:v │ k:v │
  ```
- **Cost:** get/put/delete **O(1) average**, O(n) worst (everything collides). Resize (rehash) is amortized O(1) but one insert can be O(n).
- **Appears as:** `map[K]V` (Go), `HashMap<K,V>` (Rust), `Map`/`Object` (JS), `dict` (Python).
- **Draw:** the buckets array with a couple of chains; annotate the hash→index arrow.

### Hash Set
- **Nature:** A hash map with no values (or values ignored). Existence is the only question.
- **Invariant:** No duplicate elements; `contains(x)` is true iff `x` was added and not removed.
- **Cost:** O(1) add/contains/remove.
- **Appears as:** `map[K]struct{}` (Go idiom), `HashSet<T>` (Rust), `Set` (JS/Python).
- **Draw:** same buckets array, but buckets hold only keys.

### Dynamic Array / Slice / Vector
- **Nature:** Contiguous memory with a length and a capacity. When full, allocate a bigger block (usually 2×) and copy. Appends are cheap "on average" because of amortization.
- **Invariant:** `0 ≤ len ≤ cap`; elements `a[0..len-1]` are live, `a[len..cap-1]` are uninitialized; `len` always reflects the count of live elements.
- **Layout:**
  ```
  len=5, cap=8
  ┌───┬───┬───┬───┬───┬───┬───┬───┐
  │ a │ b │ c │ d │ e │ · │ · │ · │   ← · = unused capacity
  └───┴───┴───┴───┴───┴───┴───┴───┘
  ```
- **Cost:** index O(1); append O(1) amortized (O(n) on resize); insert/remove middle O(n) (shift everything).
- **Appears as:** `[]T` (Go slice = {ptr, len, cap}), `Vec<T>` (Rust), `Array`/`[]` (JS), `list` (Python).
- **Draw:** contiguous slots with len/cap labels.

### Linked List
- **Nature:** Nodes scattered in memory, each holding a value + pointer to the next. No random access; you walk from the head.
- **Invariant:** Starting from `head` and following `next` pointers reaches each node exactly once and terminates at nil (no cycles, no orphans). For doubly-linked: every `node.prev.next == node` and `node.next.prev == node`.
- **Layout:**
  ```
  head ──► [a|•]──► [b|•]──► [c|∅]
  ```
- **Cost:** insert/remove at known node O(1); seek by index O(n).
- **Appears as:** rare in high-level code (Go/Rust stdlib has no general one); shows up in LRU caches, free lists, queue backs.
- **Draw:** box-with-arrow chain, annotate the head (and tail if doubly-linked).

### Ring Buffer / Circular Queue
- **Nature:** Fixed-size array where head and tail wrap around modulo size. Overwrites oldest when full (or blocks). No allocation after init.
- **Invariant:** `(head − tail) mod size == count` and `count ≤ size` (or `< size` if you reserve one empty slot to disambiguate full vs empty); every live element lies on the clockwise arc from tail to head.
- **Layout:**
  ```
        tail            head
         ▼               ▼
  ┌───┬───┬───┬───┬───┬───┐
  │ d │ e │ · │ · │ a │ b │ c │   (wraps)
  └───┴───┴───┴───┴───┴───┘
  ```
- **Cost:** push/pop O(1). Bounded memory — ideal for logs, audio, metrics spools.
- **Draw:** array with head/tail pointers and the wrap.

### Stack (LIFO)
- **Nature:** Push and pop from one end (the top). Last in, first out.
- **Invariant:** The element removed by `pop` is exactly the most recent `push` not yet popped — removal order is the strict reverse of insertion order.
- **Cost:** push/pop/peek O(1).
- **Where:** call stacks, undo history, expression evaluation, DFS frontier, "bracket matching."
- **Draw:** vertical pile, push/pop arrows on top.

### Queue (FIFO)
- **Nature:** Push at back, pop from front. First in, first out.
- **Invariant:** Elements are removed in the exact order they were inserted (no reordering, no skipping).
- **Cost:** enqueue/dequeue O(1).
- **Where:** task/job queues, BFS frontier, "pending work" buffers.
- **Draw:** horizontal tube, in on right, out on left.

### Priority Queue / Heap
- **Nature:** Always returns the "highest priority" (min or max) element. Usually backed by a binary heap: an array where each parent outranks its children.
- **Invariant (heap property):** Every parent ranks ≥ (max-heap) or ≤ (min-heap) both its children — the root is always the extremum. After any push/pop, the property is restored before the operation returns.
- **Layout (binary min-heap as a tree vs its array):**
  ```
       tree:            array (level order):
          1             ┌───┬───┬───┬───┬───┬───┐
        /   \           │ 1 │ 3 │ 2 │ 7 │ 4 │ 5 │
       3     2          └───┴───┴───┴───┴───┴───┘
      / \   /           parent(i) = i/2
     7   4 5            children(i) = 2i, 2i+1
  ```
- **Cost:** push O(log n), pop-min/max O(log n), peek O(1).
- **Where:** schedulers, Dijkstra/A*, "next best item," timers.
- **Draw:** both the tree and the array, show the parent/child index math.

### Binary Search Tree (BST) / Balanced (RB, AVL)
- **Nature:** Ordered tree — left subtree < node < right subtree. Lookup prunes half the tree each step. Balanced variants keep height ~log n via rotations.
- **Invariant:** For every node N: all keys in N's left subtree < N.key < all keys in N's right subtree. (Balanced variants add: height of subtrees differs by ≤1.) This is what makes search O(log n).
- **Cost:** search/insert/delete O(log n) if balanced, O(n) if degenerate (a linked list).
- **Where:** `std::map`/`TreeMap` (ordered-key maps), range queries.
- **Draw:** tree with the left<root<right rule labeled.

### Trie (Prefix Tree)
- **Nature:** Tree keyed by the **characters** of a string, one edge per character. Shared prefixes share a path. A node may be marked "end of word."
- **Invariant:** A string is present iff there is a root-to-node path spelling it and that node is marked end; all strings sharing a prefix share that prefix's path (no duplicate prefix edges).
- **Layout (for "cat", "car", "card"):**
  ```
          (root)
            │ c
            ▼
            ●──a──► ●──r──► ●(end: "car")
                      │       │ d
                      ▼       ▼
                     ●(t)●(end: "card")
                   (end: "cat")
  ```
- **Cost:** insert/lookup O(L) where L = key length (independent of how many keys stored).
- **Where:** IP routing tables, autocomplete, dictionary filters, command parsing.
- **Draw:** the branching character-path with end-nodes marked.

### Graph
- **Nature:** Nodes (vertices) connected by edges. Two storage styles:
  - **Adjacency list:** each node → list of its neighbors. Sparse-friendly. Space O(V+E).
  - **Adjacency matrix:** V×V grid of booleans/weights. Dense-friendly. Space O(V²).
- **Invariant:** The stored representation faithfully encodes the edge set — for every edge (u,v), `adj[u]` contains v (and `matrix[u][v]` is set); there are no phantom edges and no missing real ones.
- **Layout:**
  ```
  adjacency list:        A ──► [B, C]
    A: [B,C]             B ──► [A, D]
    B: [A,D]             C ──► [A]
    C: [A]               D ──► [B]
    D: [B]
  ```
- **Traversals:** BFS (queue, shortest in edges), DFS (stack/recursion, connectivity/cycles).
- **Where:** dependency graphs, networks, peer meshes, call graphs.
- **Draw:** the nodes-as-circles edges-as-lines, plus the adjacency representation beside it.

### DAG (Directed Acyclic Graph)
- **Nature:** A graph with directed edges and no cycles. Implies a **topological order** — a sequence where every edge points forward.
- **Invariant:** No directed cycle exists (equivalently, a valid topological ordering exists). If you can close a directed loop, it is NOT a DAG and topo sort is impossible.
- **Cost:** topo sort O(V+E); a valid order proves acyclicity.
- **Where:** build systems (make/bazel), git history, task dependencies, package graphs, state machines without loops.
- **Draw:** layered left-to-right, edges only point right; if you can draw a cycle, it's not a DAG.

### Tree (general hierarchy)
- **Nature:** One root, every non-root has exactly one parent, connected, acyclic. Depth = root-to-node; height = deepest node.
- **Invariant:** Exactly one root; every non-root has exactly one parent; unique path from root to any node; n nodes ⇒ exactly n−1 edges.
- **Where:** DOM, ASTs, filesystems, org charts, Merkle trees.
- **Draw:** root at top, children below; annotate depth.

### Bloom Filter
- **Nature:** A bit array + k hash functions. To add: set k bits. To query: check k bits; all set → "maybe present," any unset → "definitely absent." Never false-negative, can false-positive.
- **Invariant:** A bit, once set, is never cleared; `contains(x)` returns "definitely absent" only when at least one of `x`'s k bits is unset. (No false negatives, ever.)
- **Cost:** add/query O(k), constant space.
- **Where:** "have I seen this URL/ID before" cheaply; cache front-ends, anti-spam.
- **Draw:** bit array with k arrows from a key to k positions.

---

## Algorithm Classes

### Reconcile / Converge Loop (control loop)
- **Nature:** Repeatedly compare **observed** state to **desired** state, apply the diff, repeat. Converges because each pass shrinks the gap. Self-healing: if reality drifts, the next pass fixes it.
- **Invariant:** The desired state is a fixed point — once observed == desired, the loop makes no further changes (idempotent); and across passes the observed state monotonically approaches the desired state (the gap never grows except by external change).
- **Examples:** kubernetes controllers, our mesh orchestrator Sync, autoscalers.
- **Draw:** a loop arrow with "observed ─► diff ─► apply ─► (back to observed)."

### Voting / Consensus
- **Nature:** Gather N independent observations; decide by majority/agreement. Tolerates disagreement up to a threshold.
- **Invariant:** The decision is consistent with every observation that didn't dissent beyond the threshold — it never contradicts a strict majority of fresh inputs. (For Raft-style: at most one leader per term; a committed value is never lost.)
- **Examples:** NAT classification, Raft/Paxos quorum, anomaly voting.
- **Draw:** N observers → a tally box → one decision.

### Lookup / Dispatch Table
- **Nature:** A pure mapping from inputs to output, no I/O, no state. Often a switch or a map. Easiest to test exhaustively.
- **Invariant:** Purity — identical inputs always yield identical output with no side effects; the mapping covers every input the caller can produce (total) or explicitly defines the uncovered cases (error).
- **Examples:** path selection, routing, state-machine transitions.
- **Draw:** truth table (inputs → output).

### Upsert (insert-or-update)
- **Nature:** Keyed write: if key exists, overwrite fields; else insert. The key is the stable identity; other fields are mutable.
- **Invariant:** At most one record per key (identity never duplicates); after the call, the stored record for that key matches the written fields.
- **Examples:** announce/registry, "save by id."
- **Draw:** decision (key in map?) → either overwrite-value or insert-new.

### Atomic Swap
- **Nature:** Two related mutations (invalidate old + register new) that must succeed together or not at all. The old becomes toxic instantly.
- **Invariant:** At no instant are both old and new valid, and at no instant are both invalid — exactly one of {old key, new key} is live throughout the swap.
- **Examples:** key rotation, blue/green flip, pointer swap.
- **Draw:** old-record and new-record with a "single transaction" bracket.

### Introduction / Signaling
- **Nature:** A broker that knows both parties gives each the other's address, then removes itself from the data path. The parties then talk directly.
- **Invariant:** After introduction completes, the broker is absent from the data path — subsequent bytes flow peer-to-peer; the broker only re-enters on re-introduction. Also: each party receives the other's *current* (not stale) address.
- **Examples:** rendezvous, ICE signaling, SIP.
- **Draw:** A and B both connect to broker; broker hands each the other's address; then A↔B directly, broker idle.

### Simultaneous Send / Rendezvous Puncture
- **Nature:** Two peers both send a packet to the other's public address within the same window, so each NAT opens a hole expecting replies. Neither can succeed alone.
- **Invariant:** A direct path is used only if both NAT holes were opened within the same token-validity window; if either side fails, the path is abandoned (no half-open channel is trusted).
- **Examples:** UDP hole punching, NAT-PMP/PCP-assisted punch.
- **Draw:** two arrows crossing (A→B and B→A) inside a time window.

### Blind Relay / Pipe
- **Nature:** Forward bytes between two parties without understanding them. May inspect just enough to validate (e.g., first byte = protocol type) but never decrypts.
- **Invariant:** The relay never learns plaintext and never alters payload bytes — integrity and confidentiality are preserved end-to-end (what A sent is byte-identical to what B receives).
- **Examples:** DERP, TURN, transparent proxies.
- **Draw:** A → relay → B, with a "cannot read payload" label; bytes flow, relay is dumb.

### Two-Phase / Gate-Then-Act
- **Nature:** Check a precondition first; only proceed if it holds. Used for safety, idempotency, and to make negative tests honest (they can't pass against an "always-fail" stub).
- **Invariant:** The action executes only if the gate passed, and the gate's truth is re-checked or held under a lock so it cannot change between check and act (TOCTOU-safe).
- **Examples:** honest-red tests, transactions, "double-checked locking."
- **Draw:** gate box → (pass) act, or (fail) abort.

### Backpressure / Spill
- **Nature:** When a consumer can't keep up, the producer is slowed, buffered, or dropped. Prevents unbounded memory growth. Strategies: block, drop-oldest, drop-newest, spill to disk.
- **Invariant:** The buffer never exceeds its declared bound — overflow is handled by the chosen strategy (block/drop/spill) rather than by silent unbounded growth or crash.
- **Examples:** bounded queues, ring buffers, spool-to-disk on overflow.
- **Draw:** producer → [bounded buffer] → consumer, with overflow arrows (drop/spill).

---

## How to Extract an Invariant from Any Code

The canonical invariants above cover the textbook structures, but real code has bespoke invariants. When zooming into a novel algorithm, derive its invariant with these prompts:

1. **What must be true before the loop runs** so the first iteration is safe? (initialization)
2. **What does each iteration preserve?** Find the sentence that's true at the top of every iteration if it was true at the top of the previous one. (maintenance)
3. **When the loop ends, what does that sentence tell you about the answer?** (termination)

Phrase the result as an assertion the human could add as a runtime check or a property test. If you can write `assert(...)` for it, it's a real invariant.

### Loop Invariant (the classic form)
A statement that holds at the start of every iteration. Example — binary search:
```
// invariant: if target exists in the array, it lies within a[lo..hi]
while lo <= hi:
    mid = (lo+hi)/2
    if a[mid] == target: return mid
    if a[mid] <  target: lo = mid+1   // preserves: target still in [lo..hi]
    if a[mid] >  target: hi = mid-1   // preserves: target still in [lo..hi]
// loop ends → target not in [lo..hi] → not present
```
Draw the invariant as a label above the loop and annotate each branch with "preserves invariant."

### Invariant vs. Postcondition
- **Invariant** — holds *throughout* (the safety net).
- **Postcondition** — holds *at the end* (what the caller gets). The invariant at termination usually *implies* the postcondition. State both when they differ; the invariant is the deeper truth.

---

## Complexity Quick Reference (Big-O)

| Structure/Op | Average | Worst |
|---|---|---|
| Hash map get/put | O(1) | O(n) |
| Array index | O(1) | O(1) |
| Array insert middle | O(n) | O(n) |
| Sorted array binary search | O(log n) | O(log n) |
| BST search (balanced) | O(log n) | O(n) |
| Heap push/pop | O(log n) | O(log n) |
| Trie lookup | O(L) | O(L) |
| Graph BFS/DFS | O(V+E) | O(V+E) |
| Topo sort (DAG) | O(V+E) | O(V+E) |

When drawing a flow, annotate any O(n²) or worse hot path — that's usually where the problem hides.
