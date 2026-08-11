package echo

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	tsGenesis = int64(1700000000000000000)
	tsEvent   = int64(1700000000123456789)
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("bad hex: %v", err)
	}
	return b
}

// ─────────────────────────────────────────────────────────────────────────────
// CORRECTNESS. Round 30 finding R30-1 against the Ouroboros harness was that its
// eight divergence tests contained NO correctness test — they proved the chain
// detected damage without ever proving it computed the right thing undamaged.
// These vectors were produced by a SEPARATE implementation written from the spec
// text (scratchpad/vectors.py, BLAKE2b via Python hashlib). Expectations derived
// from this package's own output could only ever confirm themselves.
// ─────────────────────────────────────────────────────────────────────────────

func TestKnownAnswerVectors(t *testing.T) {
	g := GenesisHash([]byte("test-commitment"), time.Unix(0, tsGenesis))
	const wantGenesis = "adfdf659374c40087d5ee7f67d75d16b75248b9b4d41999bab570a0d8e5dcf3d"
	if got := hex.EncodeToString(g); got != wantGenesis {
		t.Errorf("GenesisHash\n got  %s\n want %s", got, wantGenesis)
	}

	const wantPayload = "09df2012db95f084aff8489e6edd168cb6a9d9394610d234414e01522258ed38"
	if got := hex.EncodeToString(PayloadHash([]byte("hello"))); got != wantPayload {
		t.Errorf("PayloadHash\n got  %s\n want %s", got, wantPayload)
	}

	ph := bytes.Repeat([]byte{0xAB}, 32)
	e1 := EventHash(g, time.Unix(0, tsEvent), "ladder.decision", ph)
	const wantE1 = "00c07f7ecfd5650a71db9afc420cd1ec17e7d7c3af6b357b54f31aeeea5e5183"
	if got := hex.EncodeToString(e1); got != wantE1 {
		t.Errorf("EventHash[1]\n got  %s\n want %s", got, wantE1)
	}

	e2 := EventHash(e1, time.Unix(0, tsEvent+1), "agent.dispatch", bytes.Repeat([]byte{0xCD}, 32))
	const wantE2 = "df745a7cfcc0e767ccad35557fdc97ee79001db33adb1d8035a54eca6eefa017"
	if got := hex.EncodeToString(e2); got != wantE2 {
		t.Errorf("EventHash[2]\n got  %s\n want %s", got, wantE2)
	}
}

// H must be BLAKE2b-256, not SHA-256. Getting this wrong yields a chain that is
// internally consistent and interoperates with nothing.
func TestHashIsBlake2b256(t *testing.T) {
	if n := len(GenesisHash([]byte("x"), time.Unix(0, 0))); n != 32 {
		t.Fatalf("digest is %d bytes, want 32", n)
	}
	// SHA-256("") = e3b0c442...; BLAKE2b-256 must not produce it for the same input.
	sha256Empty := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hex.EncodeToString(h([]byte(""))) == sha256Empty {
		t.Error("H appears to be SHA-256; spec §3.2 pins BLAKE2b-256")
	}
}

// The property length-prefixing exists to provide: concatenation must be injective.
// Without LP, H("x"‖"ab"‖"c") == H("x"‖"a"‖"bc") and an attacker can re-split a real
// event's bytes into a different, equally valid one.
func TestLengthPrefixRemovesAmbiguity(t *testing.T) {
	a := h([]byte("x"), lp([]byte("ab")), lp([]byte("c")))
	b := h([]byte("x"), lp([]byte("a")), lp([]byte("bc")))
	if bytes.Equal(a, b) {
		t.Error("LP failed to disambiguate field boundaries")
	}
	// And the naive unprefixed form really is ambiguous — proving the guard is load-bearing.
	na := h([]byte("x"), []byte("ab"), []byte("c"))
	nb := h([]byte("x"), []byte("a"), []byte("bc"))
	if !bytes.Equal(na, nb) {
		t.Error("precondition failed: unprefixed concatenation was expected to collide")
	}
}

// LP(x) must be u32be(len(x)) ‖ x exactly. Every other test uses short fields, where a
// 16-bit and a 32-bit prefix are indistinguishable, so a mutation narrowing the width
// passed the entire suite. Asserting the prefix bytes directly is what pins it — a
// black-box hash comparison cannot, because truncating the length still leaves the data
// itself in the stream.
func TestLengthPrefixIsU32BE(t *testing.T) {
	cases := []struct {
		n    int
		want []byte
	}{
		{0, []byte{0x00, 0x00, 0x00, 0x00}},
		{1, []byte{0x00, 0x00, 0x00, 0x01}},
		{65535, []byte{0x00, 0x00, 0xFF, 0xFF}},
		{65536, []byte{0x00, 0x01, 0x00, 0x00}}, // wraps to 0000 under u16be
		{1 << 20, []byte{0x00, 0x10, 0x00, 0x00}},
	}
	for _, c := range cases {
		got := lp(bytes.Repeat([]byte{0x5A}, c.n))
		if !bytes.Equal(got[:4], c.want) {
			t.Errorf("lp(%d bytes) prefix = % x, want % x", c.n, got[:4], c.want)
		}
		if len(got) != 4+c.n {
			t.Errorf("lp(%d bytes) total length = %d, want %d", c.n, len(got), 4+c.n)
		}
	}
}

// Seq is NOT an input to the chain hash (§6.3 covers prev, timestamp, trigger, payload_hash
// only), so record ordering rests entirely on this positional check rather than on
// cryptography. Removing it passed every other test, which is exactly why it needs its own.
func TestSeqRenumberingIsRejected(t *testing.T) {
	path, c := newChainWith(t, 3)
	c.Close()
	lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")

	var g Record
	if err := json.Unmarshal([]byte(lines[0]), &g); err != nil {
		t.Fatal(err)
	}
	g.Seq = 7 // hash is untouched and still valid — only the position claim is a lie
	edited, _ := json.Marshal(g)
	lines[0] = string(edited)
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)

	if err := VerifyFile(path); err == nil {
		t.Error("a renumbered genesis record verified successfully")
	}
}

// Each domain separator must yield a different chain for identical inputs.
func TestDomainSeparation(t *testing.T) {
	if bytes.Equal(PayloadHash([]byte("a")), h([]byte(domChain), lp([]byte("a")))) {
		t.Error("payload and chain domains produce colliding hashes")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// STORAGE INVARIANT
// ─────────────────────────────────────────────────────────────────────────────

// The one that must pass before this is wired to anything. The chain is append-only,
// so plaintext written by mistake CANNOT be un-logged.
func TestNoPlaintextReachesDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "echo.log")
	c, err := Open(path, []byte("vault-id"), time.Unix(0, tsGenesis))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	secret := "SUPERSECRET-PROMPT-TEXT-hunter2"
	if _, err := c.Append("ladder.decision", PayloadHash([]byte(secret))); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{secret, "hunter2", "SUPERSECRET"} {
		if bytes.Contains(raw, []byte(forbidden)) {
			t.Errorf("plaintext %q reached the append-only log — it cannot be un-logged", forbidden)
		}
	}
	// Every field on disk must be a hash, a number, or a trigger label.
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var r Record
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			t.Fatalf("unparseable record: %v", err)
		}
		if r.PayloadHash != "" {
			if b, err := hex.DecodeString(r.PayloadHash); err != nil || len(b) > HashLen {
				t.Errorf("payload_hash is not a 32-byte hex digest: %q", r.PayloadHash)
			}
		}
	}
}

// A non-32-byte payload hash is refused — the guard against a caller passing raw content.
func TestAppendRejectsNonDigest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "echo.log")
	c, _ := Open(path, []byte("vault-id"), time.Unix(0, tsGenesis))
	defer c.Close()
	if _, err := c.Append("t", []byte("this is obviously not a digest")); err == nil {
		t.Error("Append accepted a non-digest payload hash")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DIVERGENCE — each is a distinct way a chain can be wrong, per the v0.4 review:
// crash-before-flush, duplicated record, truncation, in-place tampering.
// ─────────────────────────────────────────────────────────────────────────────

func newChainWith(t *testing.T, n int) (string, *Chain) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "echo.log")
	c, err := Open(path, []byte("vault-id"), time.Unix(0, tsGenesis))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		if _, err := c.Append("agent.dispatch", PayloadHash([]byte{byte(i)})); err != nil {
			t.Fatal(err)
		}
	}
	return path, c
}

func TestVerifyAcceptsHealthyChain(t *testing.T) {
	path, c := newChainWith(t, 5)
	defer c.Close()
	if err := c.Verify(); err != nil {
		t.Fatalf("healthy chain failed to verify: %v", err)
	}
	if err := VerifyFile(path); err != nil {
		t.Fatalf("VerifyFile disagreed with Verify: %v", err)
	}
	if _, seq := c.Head(); seq != 5 {
		t.Errorf("head seq = %d, want 5", seq)
	}
}

// Truncation splits into two cases with OPPOSITE outcomes, and conflating them under one
// name would overstate what this package guarantees.
func TestTruncation_HeadDetected_TailNotDetectable(t *testing.T) {
	path, c := newChainWith(t, 5)
	c.Close()
	lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")

	// HEAD truncation (genesis removed) IS caught — the attack genesis exists to defeat.
	os.WriteFile(path, []byte(strings.Join(lines[1:], "\n")+"\n"), 0o600)
	if err := VerifyFile(path); err == nil {
		t.Error("a chain with its genesis record removed verified successfully")
	}

	// TAIL truncation is NOT caught, and cannot be by the chain alone: dropping the last
	// k records leaves a shorter chain byte-for-byte identical to what a younger chain
	// would look like. Nothing INSIDE the log distinguishes "never happened" from "erased".
	//
	// This assertion pins the limitation deliberately. If an external head commitment is
	// added later (§6.3 replicates commitments to B/C for exactly this reason), this test
	// will FAIL — which is the point: the guarantee will have changed and the package
	// comment must change with it. A silent upgrade of a stated limitation is how stale
	// claims survive.
	os.WriteFile(path, []byte(strings.Join(lines[:3], "\n")+"\n"), 0o600)
	if err := VerifyFile(path); err != nil {
		t.Fatalf("tail truncation is now DETECTED (%v) — the limitation documented in the "+
			"package comment is out of date and must be updated", err)
	}
}

func TestDetectsTamperedRecord(t *testing.T) {
	path, c := newChainWith(t, 4)
	c.Close()
	lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")

	var r Record
	json.Unmarshal([]byte(lines[2]), &r)
	r.Trigger = "something.else" // change contents, keep the stored hash
	edited, _ := json.Marshal(r)
	lines[2] = string(edited)
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)

	if err := VerifyFile(path); err == nil {
		t.Error("an altered record verified successfully")
	}
}

// A SPLICE: a record that is internally perfect — its stored hash correctly recomputes from
// its own stated Prev — but which points at the wrong parent.
//
// This is the case the explicit linkage comparison exists for, and nothing else catches it.
// verify recomputes each hash from the record's OWN claimed Prev, so an attacker who edits
// Prev and recomputes the hash to match passes the hash check completely. Only comparing the
// claimed parent against the actual preceding record's hash detects the fork. Removing that
// comparison left the entire suite green before this test existed.
func TestDetectsSplicedParent(t *testing.T) {
	path, c := newChainWith(t, 4)
	c.Close()
	lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")

	var genesis, victim Record
	if err := json.Unmarshal([]byte(lines[0]), &genesis); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(lines[3]), &victim); err != nil {
		t.Fatal(err)
	}

	// Re-parent record 3 onto genesis, and recompute its hash so it is self-consistent.
	victim.Prev = genesis.Hash
	victim.Hash = hex.EncodeToString(EventHash(
		mustHex(t, genesis.Hash), time.Unix(0, victim.TS), victim.Trigger, mustHex(t, victim.PayloadHash)))

	// Precondition: the spliced record really does pass a hash-only check.
	selfCheck := hex.EncodeToString(EventHash(
		mustHex(t, victim.Prev), time.Unix(0, victim.TS), victim.Trigger, mustHex(t, victim.PayloadHash)))
	if selfCheck != victim.Hash {
		t.Fatal("precondition failed: spliced record is not self-consistent")
	}

	edited, _ := json.Marshal(victim)
	lines[3] = string(edited)
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)

	if err := VerifyFile(path); err == nil {
		t.Error("a self-consistent record spliced onto the wrong parent verified successfully")
	}
}

func TestDetectsDuplicatedRecord(t *testing.T) {
	path, c := newChainWith(t, 3)
	c.Close()
	lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")
	dup := append([]string{}, lines...)
	dup = append(dup[:2], append([]string{lines[1]}, dup[2:]...)...) // replay record 1
	os.WriteFile(path, []byte(strings.Join(dup, "\n")+"\n"), 0o600)

	if err := VerifyFile(path); err == nil {
		t.Error("a chain containing a duplicated record verified successfully")
	}
}

// A torn final line — the crash-before-flush case. The partial record was never durable,
// so the chain must still verify, treating it as absent rather than as corruption.
func TestSurvivesTornFinalLine(t *testing.T) {
	path, c := newChainWith(t, 3)
	c.Close()
	raw := readFile(t, path)
	os.WriteFile(path, []byte(raw+`{"seq":4,"ts":17000000,"trig`), 0o600)

	if err := VerifyFile(path); err != nil {
		t.Errorf("a torn final line was treated as corruption: %v", err)
	}
}

// Reopening must refuse to extend a broken chain: appending onto a break launders it
// behind valid-looking successors.
func TestOpenRefusesBrokenChain(t *testing.T) {
	path, c := newChainWith(t, 3)
	c.Close()
	lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")
	var r Record
	json.Unmarshal([]byte(lines[1]), &r)
	r.PayloadHash = hex.EncodeToString(bytes.Repeat([]byte{0x99}, 32))
	edited, _ := json.Marshal(r)
	lines[1] = string(edited)
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)

	if _, err := Open(path, []byte("vault-id"), time.Unix(0, tsGenesis)); err == nil {
		t.Error("Open extended a chain that does not verify")
	}
}

// Reopening a healthy chain must continue it, not restart it.
func TestReopenContinuesChain(t *testing.T) {
	path, c := newChainWith(t, 2)
	headBefore, seqBefore := c.Head()
	c.Close()

	c2, err := Open(path, []byte("vault-id"), time.Unix(0, tsGenesis))
	if err != nil {
		t.Fatal(err)
	}
	defer c2.Close()
	head2, seq2 := c2.Head()
	if head2 != headBefore || seq2 != seqBefore {
		t.Errorf("reopen moved the head: %s/%d -> %s/%d", short(headBefore), seqBefore, short(head2), seq2)
	}
	if _, err := c2.Append("agent.dispatch", PayloadHash([]byte("x"))); err != nil {
		t.Fatal(err)
	}
	if err := c2.Verify(); err != nil {
		t.Errorf("chain broke across reopen: %v", err)
	}
}

func readFile(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Genesis contents must be verified, not merely present. Tampering with genesis while
// leaving Seq=0 was caught by nothing until this test: the head-truncation case is
// intercepted by the sequence check first, so the genesis hash comparison itself was
// never exercised.
func TestDetectsTamperedGenesis(t *testing.T) {
	path, c := newChainWith(t, 3)
	c.Close()
	lines := strings.Split(strings.TrimSpace(readFile(t, path)), "\n")

	var g Record
	if err := json.Unmarshal([]byte(lines[0]), &g); err != nil {
		t.Fatal(err)
	}
	g.TS += 1 // Seq stays 0; only the bound creation time moves
	edited, _ := json.Marshal(g)
	lines[0] = string(edited)
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)

	if err := VerifyFile(path); err == nil {
		t.Error("a genesis record with an altered timestamp verified successfully")
	}

	// And the commitment itself must be bound.
	json.Unmarshal([]byte(lines[0]), &g)
	g.TS -= 1
	g.PayloadHash = hex.EncodeToString([]byte("different-vault"))
	edited, _ = json.Marshal(g)
	lines[0] = string(edited)
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)
	if err := VerifyFile(path); err == nil {
		t.Error("a genesis record rebound to a different vault identity verified successfully")
	}
}

// The documented durability guarantee: if the sync fails, the head must NOT advance, so the
// next append still chains from the last DURABLE record. An event that never reached disk
// must not become the parent of one that does.
func TestHeadDoesNotAdvanceOnSyncFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "echo.log")
	c, err := Open(path, []byte("vault-id"), time.Unix(0, tsGenesis))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if _, err := c.Append("agent.dispatch", PayloadHash([]byte("one"))); err != nil {
		t.Fatal(err)
	}
	headBefore, seqBefore := c.Head()

	c.syncFn = func(*os.File) error { return errors.New("simulated fsync failure") }
	if _, err := c.Append("agent.dispatch", PayloadHash([]byte("two"))); err == nil {
		t.Error("Append reported success despite a failed sync")
	}

	headAfter, seqAfter := c.Head()
	if headAfter != headBefore || seqAfter != seqBefore {
		t.Errorf("head advanced past a non-durable record: %s/%d -> %s/%d",
			short(headBefore), seqBefore, short(headAfter), seqAfter)
	}

	// Recovery: once sync works again, the chain continues from the durable head and
	// still verifies end to end.
	c.syncFn = (*os.File).Sync
	if _, err := c.Append("agent.dispatch", PayloadHash([]byte("three"))); err != nil {
		t.Fatal(err)
	}
	if err := c.Verify(); err != nil {
		t.Errorf("chain did not recover after a failed sync: %v", err)
	}
}

// R30-3 applied: a record that lands CORRUPT must be caught at append time, not at the next
// Open. Round 30 found the commit protocol marks an epoch current without ever reading the
// re-sealed envelope back, so a completed-but-wrong write surfaces only at the next read —
// after the previous good state is gone.
//
// The corruption is injected through syncFn, which runs between the write and the read-back,
// so no extra seam is needed. This is the analogue of R30-1's seal-then-corrupt-one-byte
// test: a commit that RUNS TO COMPLETION and produces a wrong result — the case the eight
// divergence tests never covered, because all of them stop the commit rather than spoiling it.
func TestCorruptedRecordIsCaughtAtAppendTime(t *testing.T) {
	path := filepath.Join(t.TempDir(), "echo.log")
	c, err := Open(path, []byte("vault-id"), time.Unix(0, tsGenesis))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	if _, err := c.Append("agent.dispatch", PayloadHash([]byte("good"))); err != nil {
		t.Fatal(err)
	}
	headBefore, seqBefore := c.Head()

	// Sync "succeeds", but the bytes on disk are not the bytes we wrote: flip one character
	// inside the last record. The write path reports success all the way through.
	// ONE-SHOT. rollback() also calls syncFn to make its truncation durable, so a fake that
	// mutates the file fires twice — the second time on the already-truncated file, damaging
	// the previous good record. The instrument has to stop acting once it has acted.
	corrupted := false
	c.syncFn = func(f *os.File) error {
		if corrupted {
			return nil
		}
		corrupted = true
		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		i := bytes.LastIndex(raw, []byte(`"trigger":"`))
		if i >= 0 {
			raw[i+11] = 'X' // corrupt the trigger, leaving the stored hash intact
			os.WriteFile(path, raw, 0o600)
		}
		return nil
	}

	if _, err := c.Append("agent.dispatch", PayloadHash([]byte("bad"))); err == nil {
		t.Error("a record that landed corrupt was accepted — a completed-but-wrong commit went undetected")
	}

	headAfter, seqAfter := c.Head()
	if headAfter != headBefore || seqAfter != seqBefore {
		t.Errorf("head advanced onto a corrupt record: %s/%d -> %s/%d",
			short(headBefore), seqBefore, short(headAfter), seqAfter)
	}

	// And the chain must be left usable: the corrupt record was rolled back, so a later
	// append still chains from the last good record.
	c.syncFn = (*os.File).Sync
	if _, err := c.Append("agent.dispatch", PayloadHash([]byte("recovered"))); err != nil {
		t.Fatalf("chain unusable after a corrupt write: %v", err)
	}
	if err := c.Verify(); err != nil {
		t.Errorf("chain does not verify after recovering from a corrupt write: %v", err)
	}
}

// PAYLOAD IDENTITY (R30-1's third recommended test): what is recovered after an append is
// bit-identical to what was committed. Distinct from tamper detection — this asserts the
// undamaged path is right, which is precisely what the eight divergence tests never did.
func TestAppendedRecordIsRecoveredIdentically(t *testing.T) {
	path := filepath.Join(t.TempDir(), "echo.log")
	c, err := Open(path, []byte("vault-id"), time.Unix(0, tsGenesis))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	written := make([]Record, 0, 5)
	for i := 0; i < 5; i++ {
		r, err := c.Append("agent.dispatch", PayloadHash([]byte{byte(i)}))
		if err != nil {
			t.Fatal(err)
		}
		written = append(written, r)
	}

	recs, err := readAll(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != len(written)+1 { // +1 for genesis
		t.Fatalf("recovered %d records, want %d", len(recs), len(written)+1)
	}
	for i, w := range written {
		got := recs[i+1]
		if got != w {
			t.Errorf("record %d differs after round-trip:\n written %+v\n read    %+v", i, w, got)
		}
	}
}
