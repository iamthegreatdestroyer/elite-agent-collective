// Package proposal implements L2 of the AgentOS authority ladder.
//
// §2.1 defines L2 as: "Signed, structured, Echo-logged proposals — human-approved."
// Three of those four properties are implemented here. The fourth is not, and the gap is stated
// rather than labelled over:
//
//	structured    ✅ a typed object with validated fields, not prose
//	Echo-logged   ✅ every creation and decision appends to the audit chain
//	human-approved ⚠️ PARTIAL — see "WHAT APPROVAL DOES AND DOES NOT MEAN" below
//	signed        ❌ NOT IMPLEMENTED — see "WHY NOTHING IS SIGNED" below
//
// # WHY NOTHING IS SIGNED
//
// A signature needs a key, and a key needs custody. There is no key-custody model in this service:
// no §5A row, no rotation, no protection statement, nowhere for a private key to live that is not
// simply a file beside the code. A signature produced by an unmanaged key on the same host that
// serves the API proves nothing an attacker who reached that host could not forge — it would be
// decorative, and decorative cryptography is worse than none because it invites reliance.
//
// What IS provided instead, and it is weaker in a specific way worth understanding: every proposal
// carries a COMMITMENT — a BLAKE2b-256 digest over its canonical content — and that commitment is
// appended to the Echo chain at creation. Because the chain is append-only and hash-linked, a
// proposal's content cannot be altered after the fact without breaking the chain. That is
// TAMPER-EVIDENCE, not AUTHENTICITY: it proves the content is the content that was recorded, and
// says nothing about who produced it. Do not read the commitment as a signature.
//
// # WHAT APPROVAL DOES AND DOES NOT MEAN
//
// Approval records a human decision. It does NOT execute anything, and cannot: the backend has no
// tool surface (`exec.Command` appears nowhere outside comments) and L3 is not implemented. An
// approved proposal is a decision on the record, not an action taken.
//
// ⚠️ And approval is NOT currently authenticated. Auth is enabled only when OIDC_CLIENT_ID is set,
// and it is unset on the deployment — so the Authenticate middleware passes every request through.
// A Decision therefore records `Authenticated: false`, carrying its own weakness in the record
// rather than in a comment nobody reads. When auth is enabled the same field records true and the
// approver's identity becomes meaningful. Until then "human-approved" means "somebody with network
// access to this endpoint approved it".
package proposal

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iamthegreatdestroyer/elite-agent-collective/backend/internal/echo"
)

// Schema names the canonical encoding. Changing the field list REQUIRES a new schema name: the
// schema is hashed first, so a silent field change would produce commitments indistinguishable
// from the old shape's.
const Schema = "proposal.v2"

// Status is a proposal's lifecycle state.
type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

// Step is one action a proposal proposes. It is a description, never an instruction that anything
// executes — nothing in this service can act on it.
type Step struct {
	Order  int    `json:"order"`
	Action string `json:"action"`
	Target string `json:"target"`
	// ReversibleClaimed is the MODEL'S claim that this step could be undone. It is not verified
	// and must not be read as a fact.
	//
	// §7 rates "agent error at machine speed" Medium on the strength of "Reversible L3" — a
	// rating resting on a property that, as a bare `reversible: true`, was supplied by the model
	// and checked by nobody. That is a claim carrying the authority of a measurement, which is
	// the defect this whole series keeps finding in other people's documents.
	//
	// The name carries the provenance, exactly as Decision.Authenticated does. Authoritative
	// reversibility will come from the CAPABILITY a step exercises — our statement about our own
	// tool, verifiable — once L3 capabilities exist. Until then this field informs the human
	// reviewer and grants nothing.
	ReversibleClaimed bool `json:"reversible_claimed"`
}

// Decision records a human ruling on a proposal.
type Decision struct {
	Status Status    `json:"status"`
	At     time.Time `json:"at"`
	// Approver is the caller's identity when authentication is enabled, else "unauthenticated".
	Approver string `json:"approver"`
	// Authenticated is false whenever auth is disabled. It exists so a reader of the stored
	// record can tell a verified approval from an unverified one WITHOUT knowing the service's
	// configuration at the time — provenance travels with the record, not beside it.
	Authenticated bool   `json:"authenticated"`
	Note          string `json:"note,omitempty"`
}

// Proposal is an L2 artifact: what an agent proposes, for a human to rule on.
type Proposal struct {
	ID        string    `json:"id"`
	Agent     string    `json:"agent"`
	Title     string    `json:"title"`
	Rationale string    `json:"rationale"`
	Steps     []Step    `json:"steps"`
	Risks     []string  `json:"risks"`
	CreatedAt time.Time `json:"created_at"`

	// Commitment is a hex BLAKE2b-256 digest over the canonical CONTENT above. It deliberately
	// excludes Status and Decision, so ruling on a proposal does not invalidate the commitment
	// recorded when it was created — the content stays verifiable across its whole lifecycle.
	Commitment string `json:"commitment"`

	Status   Status    `json:"status"`
	Decision *Decision `json:"decision,omitempty"`
}

var (
	ErrInvalid  = errors.New("proposal: invalid")
	ErrDecided  = errors.New("proposal: already decided")
	ErrNotFound = errors.New("proposal: not found")
)

// Limits. Generous but bounded: an unbounded field in a stored, hashed object is a memory and
// disk amplification path from an untrusted model response.
const (
	MaxTitle     = 200
	MaxRationale = 4000
	MaxSteps     = 32
	MaxRisks     = 32
	MaxStepField = 500
	MaxRisk      = 500
)

// Validate rejects anything that is not a well-formed proposal.
//
// A model that fails to produce a valid proposal produces NO proposal. There is deliberately no
// partial-credit path that wraps prose in a proposal-shaped envelope and calls it structured —
// that would make "L2" a label on the same L1 answer, which is the defect this package exists to
// avoid.
func (p *Proposal) Validate() error {
	switch {
	case strings.TrimSpace(p.Agent) == "":
		return fmt.Errorf("%w: agent is empty", ErrInvalid)
	case strings.TrimSpace(p.Title) == "":
		return fmt.Errorf("%w: title is empty", ErrInvalid)
	case len(p.Title) > MaxTitle:
		return fmt.Errorf("%w: title exceeds %d bytes", ErrInvalid, MaxTitle)
	case strings.TrimSpace(p.Rationale) == "":
		return fmt.Errorf("%w: rationale is empty — a proposal without a why is prose", ErrInvalid)
	case len(p.Rationale) > MaxRationale:
		return fmt.Errorf("%w: rationale exceeds %d bytes", ErrInvalid, MaxRationale)
	case len(p.Steps) == 0:
		return fmt.Errorf("%w: no steps — a proposal that proposes nothing is not a proposal", ErrInvalid)
	case len(p.Steps) > MaxSteps:
		return fmt.Errorf("%w: more than %d steps", ErrInvalid, MaxSteps)
	case len(p.Risks) > MaxRisks:
		return fmt.Errorf("%w: more than %d risks", ErrInvalid, MaxRisks)
	}
	for i, s := range p.Steps {
		if strings.TrimSpace(s.Action) == "" {
			return fmt.Errorf("%w: step %d has no action", ErrInvalid, i)
		}
		if len(s.Action) > MaxStepField || len(s.Target) > MaxStepField {
			return fmt.Errorf("%w: step %d exceeds %d bytes", ErrInvalid, i, MaxStepField)
		}
	}
	for i, r := range p.Risks {
		if strings.TrimSpace(r) == "" {
			return fmt.Errorf("%w: risk %d is empty", ErrInvalid, i)
		}
		if len(r) > MaxRisk {
			return fmt.Errorf("%w: risk %d exceeds %d bytes", ErrInvalid, i, MaxRisk)
		}
	}
	return nil
}

// commitmentFields returns the canonical field sequence hashed into the commitment.
//
// The COUNTS of the variable-length lists are hashed before their contents. Without them,
// (steps=[a,b], risks=[]) and (steps=[a], risks=[b]) would present the same field sequence and
// collide — length-prefixing each field is not sufficient when the number of fields itself varies.
func (p *Proposal) commitmentFields() [][]byte {
	u32 := func(n int) []byte {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(n))
		return b
	}
	u64 := func(n int64) []byte {
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(n))
		return b
	}
	boolByte := func(v bool) []byte {
		if v {
			return []byte{1}
		}
		return []byte{0}
	}

	f := [][]byte{
		[]byte(p.Agent),
		[]byte(p.Title),
		[]byte(p.Rationale),
		u64(p.CreatedAt.UTC().UnixNano()),
		u32(len(p.Steps)),
	}
	for _, s := range p.Steps {
		f = append(f, u32(s.Order), []byte(s.Action), []byte(s.Target), boolByte(s.ReversibleClaimed))
	}
	f = append(f, u32(len(p.Risks)))
	for _, r := range p.Risks {
		f = append(f, []byte(r))
	}
	return f
}

// Commit computes the content commitment and the content-addressed ID.
//
// The ID is derived FROM the commitment rather than randomly generated: it needs no entropy
// source, it is reproducible by anyone holding the content, and two proposals with identical
// content are the same proposal — which is the correct semantics, not a collision.
func (p *Proposal) Commit() {
	sum := echo.AgentPayloadHash(Schema, p.commitmentFields()...)
	p.Commitment = hex.EncodeToString(sum)
	p.ID = p.Commitment[:16]
}

// VerifyCommitment recomputes the commitment from the content and compares.
//
// It recomputes rather than trusting the stored value; comparing the stored commitment against
// itself would be a check that cannot fail.
func (p *Proposal) VerifyCommitment() error {
	want := hex.EncodeToString(echo.AgentPayloadHash(Schema, p.commitmentFields()...))
	if p.Commitment != want {
		return fmt.Errorf("%w: commitment mismatch — content was altered after creation", ErrInvalid)
	}
	if p.ID != want[:16] {
		return fmt.Errorf("%w: id does not derive from the commitment", ErrInvalid)
	}
	return nil
}

// Decide applies a human ruling. It returns ErrDecided if one already exists: a proposal is ruled
// on once, and a second ruling is a new proposal, not an edit of the old one.
func (p *Proposal) Decide(d Decision) error {
	if p.Status != StatusPending {
		return fmt.Errorf("%w: %s is already %s", ErrDecided, p.ID, p.Status)
	}
	if d.Status != StatusApproved && d.Status != StatusRejected {
		return fmt.Errorf("%w: decision must be %q or %q", ErrInvalid, StatusApproved, StatusRejected)
	}
	if d.Approver == "" {
		d.Approver = "unauthenticated"
	}
	p.Decision = &d
	p.Status = d.Status
	return nil
}
