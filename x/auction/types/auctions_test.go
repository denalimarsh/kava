package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

const (
	TestInitiatorModuleName = "liquidator"
	TestLotDenom            = "usdx"
	TestLotAmount           = 100
	TestBidDenom            = "kava"
	TestBidAmount           = 5
	TestExtraEndTime        = 10000
	TestAuctionID           = 9999123
	TestAccAddress1         = "kava1qcfdf69js922qrdr4yaww3ax7gjml6pd39p8lj"
	TestAccAddress2         = "kava1pdfav2cjhry9k79nu6r8kgknnjtq6a7rcr0qlr"
)

func TestNewWeightedAddresses(t *testing.T) {

	tests := []struct {
		name       string
		addresses  []sdk.AccAddress
		weights    []sdk.Int
		expectpass bool
	}{
		{
			"normal",
			[]sdk.AccAddress{
				sdk.AccAddress([]byte(TestAccAddress1)),
				sdk.AccAddress([]byte(TestAccAddress2)),
			},
			[]sdk.Int{
				sdk.NewInt(6),
				sdk.NewInt(8),
			},
			true,
		},
		{
			"mismatched",
			[]sdk.AccAddress{
				sdk.AccAddress([]byte(TestAccAddress1)),
				sdk.AccAddress([]byte(TestAccAddress2)),
			},
			[]sdk.Int{
				sdk.NewInt(6),
			},
			false,
		},
		{
			"negativeWeight",
			[]sdk.AccAddress{
				sdk.AccAddress([]byte(TestAccAddress1)),
				sdk.AccAddress([]byte(TestAccAddress2)),
			},
			[]sdk.Int{
				sdk.NewInt(6),
				sdk.NewInt(-8),
			},
			false,
		},
	}

	// Run NewWeightedAdresses tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Attempt to instantiate new WeightedAddresses
			weightedAddresses, err := NewWeightedAddresses(tc.addresses, tc.weights)

			if tc.expectpass {
				// Confirm there is no error
				require.Nil(t, err)

				// Check addresses, weights
				require.Equal(t, tc.addresses, weightedAddresses.Addresses)
				require.Equal(t, tc.weights, weightedAddresses.Weights)
			} else {
				// Confirm that there is an error
				require.NotNil(t, err)

				switch tc.name {
				case "mismatched":
					require.Contains(t, err.Error(), "number of addresses doesn't match number of weights")
				case "negativeWeight":
					require.Contains(t, err.Error(), "weights contain a negative amount")
				default:
					// Unexpected error state
					t.Fail()
				}
			}
		})
	}
}

func TestNewForwardAuction(t *testing.T) {
	// TODO: Move to SetupTest
	endTime := time.Now().Add(TestExtraEndTime)

	// Create a new ForwardAuction
	forwardAuction := NewForwardAuction(
		TestInitiatorModuleName,
		c(TestLotDenom, TestLotAmount),
		TestBidDenom, endTime,
	)

	require.Equal(t, forwardAuction.Initiator, TestInitiatorModuleName)
	require.Equal(t, forwardAuction.Lot, c(TestLotDenom, TestLotAmount))
	// require.Equal(t, forwardAuction.Bidder, nilAccAddress)
	require.Equal(t, forwardAuction.Bid, c(TestBidDenom, 0))
	require.Equal(t, forwardAuction.EndTime, endTime)
	require.Equal(t, forwardAuction.MaxEndTime, endTime)

	// TODO: Does 'WithID' need to return BaseAuction instead of Auction?
	// forwardAuctionWithID := forwardAuction.WithID(TestAuctionID)
	// require.Equal(t, forwardAuctionWithID.ID, TestAuctionID)
}

// TODO: func TestAuctionGetID()

func TestBaseAuctionGetters(t *testing.T) {
	// TODO: Move to SetupTest
	endTime := time.Now().Add(TestExtraEndTime)

	// Create a new BaseAuction (via ForwardAuction)
	auction := NewForwardAuction(
		TestInitiatorModuleName,
		c(TestLotDenom, TestLotAmount),
		TestBidDenom, endTime,
	)

	auctionBidder := auction.GetBidder()
	auctionBid := auction.GetBid()
	auctionLot := auction.GetLot()
	auctionEndTime := auction.GetEndTime()

	require.Equal(t, auction.Bidder, auctionBidder)
	require.Equal(t, auction.Lot, auctionLot)
	require.Equal(t, auction.Bid, auctionBid)
	require.Equal(t, auction.EndTime, auctionEndTime)
}

func TestNewReverseAuction(t *testing.T) {
	// TODO: Move to SetupTest
	endTime := time.Now().Add(TestExtraEndTime)

	// Create a new ReverseAuction
	// TODO: Edit NewReverseAuction so the params line up...
	reverseAuction := NewReverseAuction(
		TestInitiatorModuleName,
		c(TestBidDenom, TestBidAmount),
		c(TestLotDenom, TestLotAmount),
		endTime,
	)

	require.Equal(t, reverseAuction.Initiator, TestInitiatorModuleName)
	require.Equal(t, reverseAuction.Lot, c(TestLotDenom, TestLotAmount))
	// require.Equal(t, forwardAuction.Bidder, nilAccAddress)
	require.Equal(t, reverseAuction.Bid, c(TestBidDenom, TestBidAmount))
	require.Equal(t, reverseAuction.EndTime, endTime)
	require.Equal(t, reverseAuction.MaxEndTime, endTime)
}

func TestNewForwardReverseAuction(t *testing.T) {
	// Setup WeightedAddresses
	addresses := []sdk.AccAddress{
		sdk.AccAddress([]byte(TestAccAddress1)),
		sdk.AccAddress([]byte(TestAccAddress2)),
	}

	weights := []sdk.Int{
		sdk.NewInt(6),
		sdk.NewInt(8),
	}

	weightedAddresses, _ := NewWeightedAddresses(addresses, weights)

	// TODO: Move to SetupTest
	endTime := time.Now().Add(TestExtraEndTime)

	forwardReverseAuction := NewForwardReverseAuction(
		TestInitiatorModuleName,
		c(TestLotDenom, TestLotAmount),
		endTime,
		c(TestBidDenom, TestBidAmount),
		weightedAddresses,
	)

	require.Equal(t, forwardReverseAuction.BaseAuction.Initiator, TestInitiatorModuleName)
	require.Equal(t, forwardReverseAuction.BaseAuction.Lot, c(TestLotDenom, TestLotAmount))
	// require.Equal(t, forwardAuction.Bidder, nilAccAddress)
	require.Equal(t, forwardReverseAuction.BaseAuction.Bid, c(TestBidDenom, 0))
	require.Equal(t, forwardReverseAuction.BaseAuction.EndTime, endTime)
	require.Equal(t, forwardReverseAuction.BaseAuction.MaxEndTime, endTime)
	require.Equal(t, forwardReverseAuction.MaxBid, c(TestBidDenom, TestBidAmount))
	require.Equal(t, forwardReverseAuction.LotReturns, weightedAddresses)

}

// defined to avoid cluttering test cases with long function name
func c(denom string, amount int64) sdk.Coin {
	return sdk.NewInt64Coin(denom, amount)
}
