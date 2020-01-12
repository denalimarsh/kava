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

// // TODO can this be less verbose? Should PlaceBid() be split into smaller functions?
// // It would be possible to combine all auction tests into one test runner.
// func TestForwardAuction_PlaceBid(t *testing.T) {
// seller := sdk.AccAddress([]byte("a_seller"))
// buyer1 := sdk.AccAddress([]byte("buyer1"))
// buyer2 := sdk.AccAddress([]byte("buyer2"))
// end := EndTime(10000)
// now := EndTime(10)

// type args struct {
// 	currentBlockHeight EndTime
// 	bidder             sdk.AccAddress
// 	lot                sdk.Coin
// 	bid                sdk.Coin
// }
// tests := []struct {
// 	name            string
// 	auction         ForwardAuction
// 	args            args
// 	expectedOutputs []BankOutput
// 	expectedInputs  []BankInput
// 	expectedEndTime EndTime
// 	expectedBidder  sdk.AccAddress
// 	expectedBid     sdk.Coin
// 	expectpass      bool
// }{
// 	{
// 		"normal",
// 		ForwardAuction{BaseAuction{
// 			Initiator:  seller,
// 			Lot:        c("usdx", 100),
// 			Bidder:     buyer1,
// 			Bid:        c("kava", 6),
// 			EndTime:    end,
// 			MaxEndTime: end,
// 		}},
// 		args{now, buyer2, c("usdx", 100), c("kava", 10)},
// 		[]BankOutput{{buyer2, c("kava", 10)}},
// 		[]BankInput{{buyer1, c("kava", 6)}, {seller, c("kava", 4)}},
// 		now + DefaultMaxBidDuration,
// 		buyer2,
// 		c("kava", 10),
// 		true,
// 	},
// 	{
// 		"lowBid",
// 		ForwardAuction{BaseAuction{
// 			Initiator:  seller,
// 			Lot:        c("usdx", 100),
// 			Bidder:     buyer1,
// 			Bid:        c("kava", 6),
// 			EndTime:    end,
// 			MaxEndTime: end,
// 		}},
// 		args{now, buyer2, c("usdx", 100), c("kava", 5)},
// 		[]BankOutput{},
// 		[]BankInput{},
// 		end,
// 		buyer1,
// 		c("kava", 6),
// 		false,
// 	},
// 		{
// 			"equalBid",
// 			ForwardAuction{BaseAuction{
// 				Initiator:  seller,
// 				Lot:        c("usdx", 100),
// 				Bidder:     buyer1,
// 				Bid:        c("kava", 6),
// 				EndTime:    end,
// 				MaxEndTime: end,
// 			}},
// 			args{now, buyer2, c("usdx", 100), c("kava", 6)},
// 			[]BankOutput{},
// 			[]BankInput{},
// 			end,
// 			buyer1,
// 			c("kava", 6),
// 			false,
// 		},
// 		{
// 			"timeout",
// 			ForwardAuction{BaseAuction{
// 				Initiator:  seller,
// 				Lot:        c("usdx", 100),
// 				Bidder:     buyer1,
// 				Bid:        c("kava", 6),
// 				EndTime:    end,
// 				MaxEndTime: end,
// 			}},
// 			args{end + 1, buyer2, c("usdx", 100), c("kava", 10)},
// 			[]BankOutput{},
// 			[]BankInput{},
// 			end,
// 			buyer1,
// 			c("kava", 6),
// 			false,
// 		},
// 		{
// 			"hitMaxEndTime",
// 			ForwardAuction{BaseAuction{
// 				Initiator:  seller,
// 				Lot:        c("usdx", 100),
// 				Bidder:     buyer1,
// 				Bid:        c("kava", 6),
// 				EndTime:    end,
// 				MaxEndTime: end,
// 			}},
// 			args{end - 1, buyer2, c("usdx", 100), c("kava", 10)},
// 			[]BankOutput{{buyer2, c("kava", 10)}},
// 			[]BankInput{{buyer1, c("kava", 6)}, {seller, c("kava", 4)}},
// 			end, // end time should be capped at MaxEndTime
// 			buyer2,
// 			c("kava", 10),
// 			true,
// 		},
// 	}
// for _, tc := range tests {
// 	t.Run(tc.name, func(t *testing.T) {
// 		// update auction and return in/outputs
// 		outputs, inputs, err := tc.auction.PlaceBid(tc.args.currentBlockHeight, tc.args.bidder, tc.args.lot, tc.args.bid)

// 		// check for err
// 		if tc.expectpass {
// 			require.Nil(t, err)
// 		} else {
// 			require.NotNil(t, err)
// 		}
// 		// check for correct in/outputs
// 		require.Equal(t, tc.expectedOutputs, outputs)
// 		require.Equal(t, tc.expectedInputs, inputs)
// 		// check for correct EndTime, bidder, bid
// require.Equal(t, tc.expectedEndTime, tc.auction.EndTime)
// 		require.Equal(t, tc.expectedBidder, tc.auction.Bidder)
// 		require.Equal(t, tc.expectedBid, tc.auction.Bid)
// 	})
// }
// }

// func TestReverseAuction_PlaceBid(t *testing.T) {
// 	buyer := sdk.AccAddress([]byte("a_buyer"))
// 	seller1 := sdk.AccAddress([]byte("seller1"))
// 	seller2 := sdk.AccAddress([]byte("seller2"))
// 	end := EndTime(10000)
// 	now := EndTime(10)

// 	type args struct {
// 		currentBlockHeight EndTime
// 		bidder             sdk.AccAddress
// 		lot                sdk.Coin
// 		bid                sdk.Coin
// 	}
// 	tests := []struct {
// 		name            string
// 		auction         ReverseAuction
// 		args            args
// 		expectedOutputs []BankOutput
// 		expectedInputs  []BankInput
// 		expectedEndTime EndTime
// 		expectedBidder  sdk.AccAddress
// 		expectedLot     sdk.Coin
// 		expectpass      bool
// 	}{
// 		{
// 			"normal",
// 			ReverseAuction{BaseAuction{
// 				Initiator:  buyer,
// 				Lot:        c("kava", 10),
// 				Bidder:     seller1,
// 				Bid:        c("usdx", 100),
// 				EndTime:    end,
// 				MaxEndTime: end,
// 			}},
// 			args{now, seller2, c("kava", 9), c("usdx", 100)},
// 			[]BankOutput{{seller2, c("usdx", 100)}},
// 			[]BankInput{{seller1, c("usdx", 100)}, {buyer, c("kava", 1)}},
// 			now + DefaultMaxBidDuration,
// 			seller2,
// 			c("kava", 9),
// 			true,
// 		},
// 		{
// 			"highBid",
// 			ReverseAuction{BaseAuction{
// 				Initiator:  buyer,
// 				Lot:        c("kava", 10),
// 				Bidder:     seller1,
// 				Bid:        c("usdx", 100),
// 				EndTime:    end,
// 				MaxEndTime: end,
// 			}},
// 			args{now, seller2, c("kava", 11), c("usdx", 100)},
// 			[]BankOutput{},
// 			[]BankInput{},
// 			end,
// 			seller1,
// 			c("kava", 10),
// 			false,
// 		},
// 		{
// 			"equalBid",
// 			ReverseAuction{BaseAuction{
// 				Initiator:  buyer,
// 				Lot:        c("kava", 10),
// 				Bidder:     seller1,
// 				Bid:        c("usdx", 100),
// 				EndTime:    end,
// 				MaxEndTime: end,
// 			}},
// 			args{now, seller2, c("kava", 10), c("usdx", 100)},
// 			[]BankOutput{},
// 			[]BankInput{},
// 			end,
// 			seller1,
// 			c("kava", 10),
// 			false,
// 		},
// 		{
// 			"timeout",
// 			ReverseAuction{BaseAuction{
// 				Initiator:  buyer,
// 				Lot:        c("kava", 10),
// 				Bidder:     seller1,
// 				Bid:        c("usdx", 100),
// 				EndTime:    end,
// 				MaxEndTime: end,
// 			}},
// 			args{end + 1, seller2, c("kava", 9), c("usdx", 100)},
// 			[]BankOutput{},
// 			[]BankInput{},
// 			end,
// 			seller1,
// 			c("kava", 10),
// 			false,
// 		},
// 		{
// 			"hitMaxEndTime",
// 			ReverseAuction{BaseAuction{
// 				Initiator:  buyer,
// 				Lot:        c("kava", 10),
// 				Bidder:     seller1,
// 				Bid:        c("usdx", 100),
// 				EndTime:    end,
// 				MaxEndTime: end,
// 			}},
// 			args{end - 1, seller2, c("kava", 9), c("usdx", 100)},
// 			[]BankOutput{{seller2, c("usdx", 100)}},
// 			[]BankInput{{seller1, c("usdx", 100)}, {buyer, c("kava", 1)}},
// 			end, // end time should be capped at MaxEndTime
// 			seller2,
// 			c("kava", 9),
// 			true,
// 		},
// 	}
// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// update auction and return in/outputs
// 			outputs, inputs, err := tc.auction.PlaceBid(tc.args.currentBlockHeight, tc.args.bidder, tc.args.lot, tc.args.bid)

// 			// check for err
// 			if tc.expectpass {
// 				require.Nil(t, err)
// 			} else {
// 				require.NotNil(t, err)
// 			}
// 			// check for correct in/outputs
// 			require.Equal(t, tc.expectedOutputs, outputs)
// 			require.Equal(t, tc.expectedInputs, inputs)
// 			// check for correct EndTime, bidder, bid
// 			require.Equal(t, tc.expectedEndTime, tc.auction.EndTime)
// 			require.Equal(t, tc.expectedBidder, tc.auction.Bidder)
// 			require.Equal(t, tc.expectedLot, tc.auction.Lot)
// 		})
// 	}
// }

// func TestForwardReverseAuction_PlaceBid(t *testing.T) {
// 	cdpOwner := sdk.AccAddress([]byte("a_cdp_owner"))
// 	seller := sdk.AccAddress([]byte("a_seller"))
// 	buyer1 := sdk.AccAddress([]byte("buyer1"))
// 	buyer2 := sdk.AccAddress([]byte("buyer2"))
// 	end := EndTime(10000)
// 	now := EndTime(10)

// 	type args struct {
// 		currentBlockHeight EndTime
// 		bidder             sdk.AccAddress
// 		lot                sdk.Coin
// 		bid                sdk.Coin
// 	}
// 	tests := []struct {
// 		name            string
// 		auction         ForwardReverseAuction
// 		args            args
// 		expectedOutputs []BankOutput
// 		expectedInputs  []BankInput
// 		expectedEndTime EndTime
// 		expectedBidder  sdk.AccAddress
// 		expectedLot     sdk.Coin
// 		expectedBid     sdk.Coin
// 		expectpass      bool
// 	}{
// 		{
// 			"normalForwardBid",
// 			ForwardReverseAuction{BaseAuction: BaseAuction{
// 				Initiator:  seller,
// 				Lot:        c("xrp", 100),
// 				Bidder:     buyer1,
// 				Bid:        c("usdx", 5),
// 				EndTime:    end,
// 				MaxEndTime: end},
// 				MaxBid:      c("usdx", 10),
// 				OtherPerson: cdpOwner,
// 			},
// 			args{now, buyer2, c("xrp", 100), c("usdx", 6)},
// 			[]BankOutput{{buyer2, c("usdx", 6)}},
// 			[]BankInput{{buyer1, c("usdx", 5)}, {seller, c("usdx", 1)}},
// 			now + DefaultMaxBidDuration,
// 			buyer2,
// 			c("xrp", 100),
// 			c("usdx", 6),
// 			true,
// 		},
// 		{
// 			"normalSwitchOverBid",
// 			ForwardReverseAuction{BaseAuction: BaseAuction{
// 				Initiator:  seller,
// 				Lot:        c("xrp", 100),
// 				Bidder:     buyer1,
// 				Bid:        c("usdx", 5),
// 				EndTime:    end,
// 				MaxEndTime: end},
// 				MaxBid:      c("usdx", 10),
// 				OtherPerson: cdpOwner,
// 			},
// 			args{now, buyer2, c("xrp", 99), c("usdx", 10)},
// 			[]BankOutput{{buyer2, c("usdx", 10)}},
// 			[]BankInput{{buyer1, c("usdx", 5)}, {seller, c("usdx", 5)}, {cdpOwner, c("xrp", 1)}},
// 			now + DefaultMaxBidDuration,
// 			buyer2,
// 			c("xrp", 99),
// 			c("usdx", 10),
// 			true,
// 		},
// 		{
// 			"normalReverseBid",
// 			ForwardReverseAuction{BaseAuction: BaseAuction{
// 				Initiator:  seller,
// 				Lot:        c("xrp", 99),
// 				Bidder:     buyer1,
// 				Bid:        c("usdx", 10),
// 				EndTime:    end,
// 				MaxEndTime: end},
// 				MaxBid:      c("usdx", 10),
// 				OtherPerson: cdpOwner,
// 			},
// 			args{now, buyer2, c("xrp", 90), c("usdx", 10)},
// 			[]BankOutput{{buyer2, c("usdx", 10)}},
// 			[]BankInput{{buyer1, c("usdx", 10)}, {cdpOwner, c("xrp", 9)}},
// 			now + DefaultMaxBidDuration,
// 			buyer2,
// 			c("xrp", 90),
// 			c("usdx", 10),
// 			true,
// 		},
// 		// TODO more test cases
// 	}
// 	for _, tc := range tests {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// update auction and return in/outputs
// 			outputs, inputs, err := tc.auction.PlaceBid(tc.args.currentBlockHeight, tc.args.bidder, tc.args.lot, tc.args.bid)

// 			// check for err
// 			if tc.expectpass {
// 				require.Nil(t, err)
// 			} else {
// 				require.NotNil(t, err)
// 			}
// 			// check for correct in/outputs
// 			require.Equal(t, tc.expectedOutputs, outputs)
// 			require.Equal(t, tc.expectedInputs, inputs)
// 			// check for correct EndTime, bidder, bid
// 			require.Equal(t, tc.expectedEndTime, tc.auction.EndTime)
// 			require.Equal(t, tc.expectedBidder, tc.auction.Bidder)
// 			require.Equal(t, tc.expectedLot, tc.auction.Lot)
// 			require.Equal(t, tc.expectedBid, tc.auction.Bid)
// 		})
// 	}
// }

// defined to avoid cluttering test cases with long function name
func c(denom string, amount int64) sdk.Coin {
	return sdk.NewInt64Coin(denom, amount)
}
