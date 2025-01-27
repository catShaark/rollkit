package types

import (
	"bytes"
	"encoding"
	"fmt"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmtypes "github.com/cometbft/cometbft/types"

	"github.com/celestiaorg/go-header"
	cmbytes "github.com/cometbft/cometbft/libs/bytes"
)

// Hash is a 32-byte array which is used to represent a hash result.
type Hash = header.Hash

// BaseHeader contains the most basic data of a header
type BaseHeader struct {
	// Height represents the block height (aka block number) of a given header
	Height uint64
	// Time contains Unix nanotime of a block
	Time uint64
	// The Chain ID
	ChainID string
}

// Header defines the structure of Rollkit block header.
type Header struct {
	BaseHeader
	// Block and App version
	Version Version

	// prev block info
	LastHeaderHash Hash

	// hashes of block data
	LastCommitHash Hash // commit from aggregator(s) from the last block
	DataHash       Hash // Block.Data root aka Transactions
	ConsensusHash  Hash // consensus params for current block
	AppHash        Hash // state after applying txs from the current block

	// compablity with light client
	ValidatorHash Hash

	// Root hash of all results from the txs from the previous block.
	// This is ABCI specific but smart-contract chains require some way of committing
	// to transaction receipts/results.
	LastResultsHash Hash

	// Note that the address can be derived from the pubkey which can be derived
	// from the signature when using secp256k.
	// We keep this in case users choose another signature format where the
	// pubkey can't be recovered by the signature (e.g. ed25519).
	ProposerAddress []byte // original proposer of the block
}

// New creates a new Header.
func (h *Header) New() *Header {
	return new(Header)
}

// IsZero returns true if the header is nil.
func (h *Header) IsZero() bool {
	return h == nil
}

// ChainID returns chain ID of the header.
func (h *Header) ChainID() string {
	return h.BaseHeader.ChainID
}

// Height returns height of the header.
func (h *Header) Height() uint64 {
	return h.BaseHeader.Height
}

// LastHeader returns last header hash of the header.
func (h *Header) LastHeader() Hash {
	return h.LastHeaderHash[:]
}

// Time returns timestamp as unix time with nanosecond precision
func (h *Header) Time() time.Time {
	return time.Unix(0, int64(h.BaseHeader.Time))
}

// Verify verifies the header.
func (h *Header) Verify(untrstH *Header) error {
	if !bytes.Equal(untrstH.ProposerAddress, h.ProposerAddress) {
		return &header.VerifyError{
			Reason: fmt.Errorf("expected proposer (%X) got (%X)",
				h.ProposerAddress,
				untrstH.ProposerAddress,
			),
		}
	}
	return nil
}

// Validate performs basic validation of a header.
func (h *Header) Validate() error {
	return h.ValidateBasic()
}

// ValidateBasic performs basic validation of a header.
func (h *Header) ValidateBasic() error {
	if len(h.ProposerAddress) == 0 {
		return ErrNoProposerAddress
	}

	return nil
}

// MakeCometBFTVote make a cometBFT consensus vote for the sequencer to commit
// we have the sequencer signs cometBFT consensus vote for compability with cometBFT client
func (h *Header) MakeCometBFTVote() []byte {
	vote := cmtproto.Vote{
		Type:   cmtproto.PrecommitType,
		Height: int64(h.Height()),
		Round:  0,
		// Header hash = block hash in rollkit
		BlockID: cmtproto.BlockID{
			Hash:          cmbytes.HexBytes(h.Hash()),
			PartSetHeader: cmtproto.PartSetHeader{},
		},
		Timestamp: h.Time(),
		// proposerAddress = sequencer = validator
		ValidatorAddress: h.ProposerAddress,
		ValidatorIndex:   0,
	}
	chainID := h.ChainID()
	consensusVoteBytes := cmtypes.VoteSignBytes(chainID, &vote)

	return consensusVoteBytes
}

var _ header.Header[*Header] = &Header{}
var _ encoding.BinaryMarshaler = &Header{}
var _ encoding.BinaryUnmarshaler = &Header{}
