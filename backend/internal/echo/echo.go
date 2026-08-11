// Package echo implements the Ouroboros Echo chain — an append-only, hash-chained
// evidence log — per ouroboros-labyrinth v1.19 §6.3 and the §3.2 PRIMITIVE LAW.
//
// SPEC, transcribed verbatim from v1.19 (do not "simplify" these; R25-C2 records that
// leaving H, LP and the domain separators unspecified was a MAJOR review finding):
//
//	H(x)  = BLAKE2b-256(x)                     // NOT SHA-256
//	LP(x) = u32be(len(x)) ‖ x                  // every variable-length field
//	payload_hash = H( "ouroboros-payload-v1" ‖ LP(canonical(payload)) )
//
//	echo_event[0] = H( "ouroboros-echo-genesis-v1" ‖ LP(K_f_commitment)
//	                   ‖ LP(u64be_unix_ns(created_at)) )
//	echo_event[n] = H( "ouroboros-echo-chain-v1" ‖ LP(echo_event[n-1])
//	                   ‖ LP(u64be_unix_ns(timestamp)) ‖ LP(trigger) ‖ LP(payload_hash) )
//
// WHY GENESIS EXISTS (v1.19, R29-4): it is the chain's base case, binding vault identity
// and creation time. A verifier walking backwards TERMINATES there — which is what makes a
// truncated chain distinguishable from a merely young one. Without it, lopping off the first
// k records yields a chain that verifies perfectly.
//
// WHY LENGTH-PREFIXING (R25-C2): H(a ‖ b) collides trivially across field boundaries when
// fields are variable-length — ("ab","c") and ("a","bc") hash identically. u32be(len) ‖ x
// makes the encoding injective, so a forged event cannot be assembled by re-splitting a real
// one's bytes.
//
// # WHAT THIS PACKAGE DELIBERATELY DOES NOT DO
//
//   - canonical() — deterministic CBOR, RFC 8949 (§3.2). NOT implemented. Callers pass an
//     already-canonical byte string to PayloadHash. Inventing a second, non-CBOR canonical
//     form here would produce hashes that no Ouroboros verifier can reproduce, which is worse
//     than the honest gap. Closing it needs a deterministic-CBOR dependency and is a separate,
//     deliberate decision.
//   - Per-event public-key encryption of records, the three-aggregator replication
//     (hub/pi/studio), commitments export, and the alert surface (§6.4 — whose channel, the
//     spec states plainly, does not exist yet). This is the LOCAL chain mechanism only.
//
// # WHAT THIS CHAIN CANNOT DETECT — stated plainly, because an audit log trusted beyond
// # its guarantee is worse than none
//
// TAIL TRUNCATION IS UNDETECTABLE from the log alone. Deleting the last k records leaves a
// file byte-for-byte identical to what a younger chain would look like; nothing inside it
// distinguishes "these events never happened" from "these events were erased". Head
// truncation IS caught, because genesis anchors the walk (R29-4).
//
// Closing the tail gap requires a head commitment held somewhere the writer cannot reach —
// which is exactly what §6.3's "commitments (hashes only): replicated to B/C" provides, and
// which this local-only package does not implement. Until then: this chain proves that the
// records it still holds are unaltered and correctly ordered. It does not prove they are all
// of them. TestTruncation_HeadDetected_TailNotDetectable pins that limitation, and fails if
// it ever silently changes.
//
// # THE STORAGE INVARIANT
//
// Records carry metadata and hashes ONLY — never prompt text, never keys, never payloads.
// A leaked Echo log must embarrass the attacker, not the owner (v1.19 §6.3). The chain is
// append-only by construction, so a bug that writes plaintext CANNOT be un-logged; that is
// why TestNoPlaintextReachesDisk exists and why it runs before this is wired to anything.
package echo

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/blake2b"
)

// Domain separators. Distinct per chain so an event from one context can never be replayed
// as an event in another (§3.2, per-chain domain separation).
const (
	domGenesis = "ouroboros-echo-genesis-v1"
	domChain   = "ouroboros-echo-chain-v1"
	domPayload = "ouroboros-payload-v1"
)

// HashLen is the digest size in bytes: BLAKE2b-256.
const HashLen = 32

// h implements H(x) = BLAKE2b-256(x).
func h(parts ...[]byte) []byte {
	d, err := blake2b.New256(nil)
	if err != nil {
		// Only returns an error for an invalid key size; nil key cannot fail.
		panic("echo: blake2b.New256: " + err.Error())
	}
	for _, p := range parts {
		d.Write(p)
	}
	return d.Sum(nil)
}

// lp implements LP(x) = u32be(len(x)) ‖ x.
func lp(b []byte) []byte {
	out := make([]byte, 4+len(b))
	binary.BigEndian.PutUint32(out[:4], uint32(len(b)))
	copy(out[4:], b)
	return out
}

// u64ns implements u64be_unix_ns(t).
func u64ns(unixNano int64) []byte {
	out := make([]byte, 8)
	binary.BigEndian.PutUint64(out, uint64(unixNano))
	return out
}

// PayloadHash implements payload_hash = H( "ouroboros-payload-v1" ‖ LP(canonical(payload)) ).
//
// `canonical` MUST already be the caller's canonical encoding of the payload — this function
// does not and cannot canonicalise for you. Passing a Go map's JSON encoding, whose key order
// is stable but whose spec is not, will produce a hash that is reproducible here and nowhere
// else.
func PayloadHash(canonical []byte) []byte {
	return h([]byte(domPayload), lp(canonical))
}

// GenesisHash implements echo_event[0].
func GenesisHash(identityCommitment []byte, createdAt time.Time) []byte {
	return h([]byte(domGenesis), lp(identityCommitment), lp(u64ns(createdAt.UnixNano())))
}

// EventHash implements echo_event[n] for n > 0.
func EventHash(prev []byte, timestamp time.Time, trigger string, payloadHash []byte) []byte {
	return h([]byte(domChain), lp(prev), lp(u64ns(timestamp.UnixNano())), lp([]byte(trigger)), lp(payloadHash))
}

// Record is one line of the on-disk log.
//
// Field names are part of the on-disk format: changing them breaks every existing chain,
// because Verify re-reads them. The hash is NOT computed over this JSON — it is computed
// over the spec's byte construction — so the JSON is a transport, not a canonical form.
//
// ⚠️ Seq is NOT an input to the chain hash. §6.3 covers prev, timestamp, trigger and
// payload_hash only, so a renumbered record still carries a cryptographically valid hash.
// Ordering is enforced by verify's positional check (seq must equal the record's index),
// not by the hash — a weaker guarantee than the chaining, and worth knowing before relying
// on Seq for anything.
type Record struct {
	Seq uint64 `json:"seq"`
	TS  int64  `json:"ts"` // unix nanoseconds
	// Trigger is empty for the genesis record.
	Trigger string `json:"trigger"`
	// PayloadHash is hex. For the genesis record this field carries the identity
	// commitment instead, which is genesis's corresponding input.
	PayloadHash string `json:"payload_hash"`
	Hash        string `json:"hash"`
	Prev        string `json:"prev"` // empty for genesis
}

// Chain is an append-only Echo log backed by one file.
type Chain struct {
	mu   sync.Mutex
	path string
	f    *os.File
	last []byte // raw hash of the newest record
	seq  uint64
	// syncFn is the durability barrier, injectable so the "head does not advance on a
	// failed sync" guarantee can actually be tested. Without a seam, that promise is
	// only a comment — and a durability claim no test can falsify is exactly the kind
	// that quietly stops being true.
	syncFn func(*os.File) error
}

var (
	// ErrBroken means the on-disk chain does not verify. It is deliberately distinct from
	// an I/O error: a chain that cannot be read is an access failure, while a chain that
	// reads fine and does not verify is evidence of tampering or corruption.
	ErrBroken = errors.New("echo: chain verification failed")
	// ErrBadHashLen guards against a caller passing a non-BLAKE2b-256 digest.
	ErrBadHashLen = errors.New("echo: payload hash must be 32 bytes (BLAKE2b-256)")
)

// Open opens or creates a chain at path.
//
// A new file is seeded with the genesis record. An existing file is VERIFIED IN FULL before
// it is opened for appending — refusing to extend a chain that does not verify, because an
// append onto a broken chain launders the break behind valid-looking successors.
func Open(path string, identityCommitment []byte, createdAt time.Time) (*Chain, error) {
	c := &Chain{path: path}

	recs, err := readAll(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if len(recs) == 0 {
		g := Record{
			Seq:         0,
			TS:          createdAt.UnixNano(),
			PayloadHash: hex.EncodeToString(identityCommitment),
			Hash:        hex.EncodeToString(GenesisHash(identityCommitment, createdAt)),
		}
		if err := appendLine(path, g); err != nil {
			return nil, err
		}
		recs = []Record{g}
	} else if err := verify(recs); err != nil {
		return nil, err
	}

	last := recs[len(recs)-1]
	raw, err := hex.DecodeString(last.Hash)
	if err != nil {
		return nil, fmt.Errorf("echo: unreadable head hash: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	c.f, c.last, c.seq = f, raw, last.Seq
	c.syncFn = (*os.File).Sync
	return c, nil
}

// Append adds one event and returns its record.
//
// The write is fsynced BEFORE the in-memory head advances. If the sync fails, the head is
// left untouched so the next append still chains from the last durable record — the
// crash-before-flush case the v0.4 review called out. An event that is not durable must not
// become the parent of one that is.
func (c *Chain) Append(trigger string, payloadHash []byte) (Record, error) {
	if len(payloadHash) != HashLen {
		return Record{}, ErrBadHashLen
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	r := Record{
		Seq:         c.seq + 1,
		TS:          now.UnixNano(),
		Trigger:     trigger,
		PayloadHash: hex.EncodeToString(payloadHash),
		Prev:        hex.EncodeToString(c.last),
		Hash:        hex.EncodeToString(EventHash(c.last, now, trigger, payloadHash)),
	}

	line, err := json.Marshal(r)
	if err != nil {
		return Record{}, err
	}

	// Remember the last durable size. Writing happens BEFORE syncing, so a failed sync
	// leaves a perfectly parseable record on disk that was never durable. Leaving it there
	// and declining to advance the head produces a file with two records claiming the same
	// sequence number — which is how the first version of this function behaved, and what
	// TestHeadDoesNotAdvanceOnSyncFailure caught.
	st, err := c.f.Stat()
	if err != nil {
		return Record{}, err
	}
	durable := st.Size()

	// rollback returns the file to its last durable state so the on-disk chain and the
	// in-memory head agree. If the rollback ITSELF fails, the file may retain a record the
	// head does not know about — verify() then rejects the chain on next Open (seq
	// mismatch), which is the fail-closed outcome and better than silently continuing.
	rollback := func() {
		if terr := c.f.Truncate(durable); terr == nil {
			_ = c.syncFn(c.f)
		}
	}

	if _, err := c.f.Write(append(line, '\n')); err != nil {
		rollback()
		return Record{}, err
	}
	if err := c.syncFn(c.f); err != nil {
		rollback()
		return Record{}, fmt.Errorf("echo: fsync failed, record rolled back, head not advanced: %w", err)
	}

	raw, _ := hex.DecodeString(r.Hash)
	c.last, c.seq = raw, r.Seq
	return r, nil
}

// Head returns the current head hash (hex) and sequence number.
func (c *Chain) Head() (string, uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return hex.EncodeToString(c.last), c.seq
}

// Close closes the underlying file.
func (c *Chain) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.f == nil {
		return nil
	}
	err := c.f.Close()
	c.f = nil
	return err
}

// Verify re-reads the whole chain from disk and recomputes every link.
func (c *Chain) Verify() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	recs, err := readAll(c.path)
	if err != nil {
		return err
	}
	return verify(recs)
}

// VerifyFile verifies a chain without opening it for append.
func VerifyFile(path string) error {
	recs, err := readAll(path)
	if err != nil {
		return err
	}
	return verify(recs)
}

// verify recomputes every link from the genesis record forward.
//
// It walks FORWARD from genesis rather than backward from the head on purpose: a backward
// walk that stops when it runs out of records accepts any suffix of a real chain as valid.
// Anchoring at genesis is what makes truncation detectable (R29-4).
func verify(recs []Record) error {
	if len(recs) == 0 {
		return fmt.Errorf("%w: empty chain (no genesis record)", ErrBroken)
	}

	g := recs[0]
	if g.Seq != 0 {
		return fmt.Errorf("%w: first record has seq %d, want 0 — chain is truncated at the head", ErrBroken, g.Seq)
	}
	commitment, err := hex.DecodeString(g.PayloadHash)
	if err != nil {
		return fmt.Errorf("%w: genesis commitment unreadable: %v", ErrBroken, err)
	}
	want := hex.EncodeToString(GenesisHash(commitment, time.Unix(0, g.TS)))
	if g.Hash != want {
		return fmt.Errorf("%w: genesis hash mismatch (have %s, computed %s)", ErrBroken, short(g.Hash), short(want))
	}

	prev := g.Hash
	for i := 1; i < len(recs); i++ {
		r := recs[i]
		if r.Seq != uint64(i) {
			return fmt.Errorf("%w: record %d has seq %d — a record was inserted, removed, or duplicated", ErrBroken, i, r.Seq)
		}
		if r.Prev != prev {
			return fmt.Errorf("%w: record %d prev=%s does not link to %s", ErrBroken, i, short(r.Prev), short(prev))
		}
		ph, err := hex.DecodeString(r.PayloadHash)
		if err != nil {
			return fmt.Errorf("%w: record %d payload hash unreadable: %v", ErrBroken, i, err)
		}
		prevRaw, err := hex.DecodeString(r.Prev)
		if err != nil {
			return fmt.Errorf("%w: record %d prev unreadable: %v", ErrBroken, i, err)
		}
		want := hex.EncodeToString(EventHash(prevRaw, time.Unix(0, r.TS), r.Trigger, ph))
		if r.Hash != want {
			return fmt.Errorf("%w: record %d hash mismatch — contents were altered after writing", ErrBroken, i)
		}
		prev = r.Hash
	}
	return nil
}

func readAll(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []Record
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for n := 0; sc.Scan(); n++ {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var r Record
		if err := json.Unmarshal(line, &r); err != nil {
			// A torn final line is the crash-before-flush case: the record was never
			// durable, so treat it as absent rather than as corruption.
			if n == 0 {
				return nil, fmt.Errorf("%w: record 0 is unparseable: %v", ErrBroken, err)
			}
			break
		}
		out = append(out, r)
	}
	if err := sc.Err(); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}
	return out, nil
}

func appendLine(path string, r Record) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	line, err := json.Marshal(r)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	return f.Sync()
}

func short(hexStr string) string {
	if len(hexStr) > 12 {
		return hexStr[:12] + "…"
	}
	return hexStr
}
