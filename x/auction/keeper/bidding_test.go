package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava/app"
	"github.com/kava-labs/kava/x/auction/types"
	"github.com/kava-labs/kava/x/liquidator"
)

type AuctionType int

const (
	Invalid              AuctionType = 0
	Forward              AuctionType = 1
	Reverse              AuctionType = 2
	ForwardReversePhase1 AuctionType = 3
	ForwardReversePhase2 AuctionType = 4
)

func TestAuctionBidding(t *testing.T) {
	// TODO: Block time
	someTime := time.Date(0001, time.January, 1, 0, 0, 0, 0, time.UTC)
	// now := time.Now()
	// end := now.Add(1000000)

	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	buyer := addrs[0]
	modName := liquidator.ModuleName
	forwardReverseAddrs := addrs[1:]
	forwardReverseWeights := is(30, 20, 10)

	tApp := app.NewTestApp()

	// Set up seller account
	sellerAcc := supply.NewEmptyModuleAccount(modName, supply.Minter, supply.Burner) // Forward auctions burn proceeds
	require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 1000), c("token2", 1000))))

	// Initialize genesis accounts
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(buyer, cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
			auth.NewBaseAccount(forwardReverseAddrs[0], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
			auth.NewBaseAccount(forwardReverseAddrs[1], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
			auth.NewBaseAccount(forwardReverseAddrs[2], cs(c("token1", 1000), c("token2", 1000)), nil, 0, 0),
			sellerAcc,
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	type auctionArgs struct {
		auctionType AuctionType
		seller      string
		lot         sdk.Coin
		bid         sdk.Coin
		addresses   []sdk.AccAddress
		weights     []sdk.Int
	}

	type bidArgs struct {
		// currentBlockHeight time.Time
		bidder sdk.AccAddress
		lot    sdk.Coin
		bid    sdk.Coin
	}

	tests := []struct {
		name            string
		auctionArgs     auctionArgs
		bidArgs         bidArgs
		expectedError   string
		expectedEndTime time.Time
		expectedBidder  sdk.AccAddress
		expectedBid     sdk.Coin
		expectpass      bool
	}{
		{
			"forward: normal",
			auctionArgs{Forward, modName, c("token1", 100), c("token2", 10), []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token1", 100), c("token2", 10)},
			"",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			true,
		},
		{
			"forward: invalid bid denom",
			auctionArgs{Forward, modName, c("token1", 100), c("token2", 10), []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token1", 100), c("badtoken", 10)},
			// TODO: Unreachable code: "bid denom doesn't match auction",
			"bid has incorrect denom",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"forward: invalid bid (equal)",
			auctionArgs{Forward, modName, c("token1", 100), c("token2", 10), []sdk.AccAddress{}, []sdk.Int{}},
			bidArgs{buyer, c("token1", 100), c("token2", 0)},
			"bid not greater than last bid",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"reverse: normal",
			auctionArgs{Reverse, modName, c("token2", 100), c("token1", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			bidArgs{buyer, c("token2", 20), c("token1", 10)},                                                  // lot, bid
			"",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token1", 20),
			true,
		},
		{
			"reverse: invalid lot denom",
			auctionArgs{Reverse, modName, c("token2", 100), c("token1", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			bidArgs{buyer, c("badtoken", 20), c("token1", 10)},                                                // lot, bid
			// TODO: Unreachable code: "lot denom doesn't match auction",
			"lot has incorrect denom",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token1", 20),
			false,
		},
		// TODO: PANIC if test is run (coin.go validates positive coin amount)
		// {
		// 	"reverse: negative lot amount",
		// 	auctionArgs{Reverse, modName, c("token2", 199), c("token1", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
		// 	bidArgs{buyer, c("token2", -20), c("token1", 10)},                // lot, bid
		// 	// TODO: Unreachable code: "lot less than 0"
		// 	"negative coin amount:",
		// 	someTime.Add(types.DefaultBidDuration),
		// 	buyer,
		// 	c("token1", 20),
		// 	false,
		// },
		{
			"reverse: invalid lot size (larger)",
			auctionArgs{Reverse, modName, c("token2", 100), c("token1", 20), []sdk.AccAddress{}, []sdk.Int{}}, // initial bid, lot
			bidArgs{buyer, c("token2", 101), c("token1", 10)},                                                 // lot, bid
			"lot not smaller than last lot",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token1", 20),
			false,
		},
		{
			"[forward] reverse: normal",
			auctionArgs{ForwardReversePhase1, modName, c("token1", 20), c("token2", 100), forwardReverseAddrs, forwardReverseWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 20), c("token2", 10)}, // lot, bid
			"",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			true,
		},
		{
			"[forward] reverse: invalid bid size (smaller)",
			auctionArgs{ForwardReversePhase1, modName, c("token1", 20), c("token2", 100), forwardReverseAddrs, forwardReverseWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 20), c("token2", 0)}, // lot, bid
			"auction in forward phase, new bid not higher than last bid",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"[forward] reverse: invalid bid size (greater than max bid)",
			auctionArgs{ForwardReversePhase1, modName, c("token1", 20), c("token2", 100), forwardReverseAddrs, forwardReverseWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 20), c("token2", 101)},                                                                         // lot, bid
			"",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"[forward] reverse: invalid lot size (greater)",
			auctionArgs{ForwardReversePhase1, modName, c("token1", 20), c("token2", 100), forwardReverseAddrs, forwardReverseWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 30), c("token2", 10)}, // lot, bid
			"lot out of bounds",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"[forward] reverse: invalid lot size (smaller)",
			auctionArgs{ForwardReversePhase1, modName, c("token1", 20), c("token2", 100), forwardReverseAddrs, forwardReverseWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 15), c("token2", 10)}, // lot, bid
			"auction cannot enter reverse phase without bidding max bid",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 10),
			false,
		},
		{
			"forward [reverse]: normal",
			auctionArgs{ForwardReversePhase2, modName, c("token1", 20), c("token2", 50), forwardReverseAddrs, forwardReverseWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 15), c("token2", 50)}, // lot, bid
			"",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			true,
		},
		// TODO: PANIC if test is run (coin.go validates positive coin amount)
		// {
		// 	"forward [reverse]: negative lot",
		// 	auctionArgs{ForwardReversePhase2, modName, c("token1", 20), c("token2", 50), forwardReverseAddrs, forwardReverseWeights}, // lot, max bid
		// 	bidArgs{buyer, c("token1", -1), c("token2", 50)}, // lot, bid
		// 	"can't bid negative amount",
		// 	someTime.Add(types.DefaultBidDuration),
		// 	buyer,
		// 	c("token2", 50),
		// 	false,
		// },
		{
			"forward [reverse]: invalid lot size (equal)",
			auctionArgs{ForwardReversePhase2, modName, c("token1", 20), c("token2", 50), forwardReverseAddrs, forwardReverseWeights}, // lot, max bid
			bidArgs{buyer, c("token1", 20), c("token2", 50)}, // lot, bid
			"auction in reverse phase, new bid not less than previous amount",
			someTime.Add(types.DefaultBidDuration),
			buyer,
			c("token2", 50),
			false,
		},
		// 	TODO: "timeout", "hitMaxEndTime"
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Start Auction
			var id uint64
			var err error
			switch tc.auctionArgs.auctionType {
			case Forward:
				id, err = keeper.StartForwardAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid.Denom) // lot, bid denom
				require.NoError(t, err)
			case Reverse:
				id, err = keeper.StartReverseAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.bid, tc.auctionArgs.lot)
				require.NoError(t, err)
			case ForwardReversePhase1, ForwardReversePhase2:
				id, err = keeper.StartForwardReverseAuction(ctx, tc.auctionArgs.seller, tc.auctionArgs.lot, tc.auctionArgs.bid, tc.auctionArgs.addresses, tc.auctionArgs.weights) // seller, lot, maxBid, otherPerson
				require.NoError(t, err)

				// Move ForwardReverseAuction to reverse phase by placing max bid
				if tc.auctionArgs.auctionType == ForwardReversePhase2 {
					err = keeper.PlaceBid(ctx, id, tc.bidArgs.bidder, tc.auctionArgs.bid, tc.auctionArgs.lot)
					require.NoError(t, err)
				}
			default:
				t.Fail()
			}

			// Place bid on auction
			err = keeper.PlaceBid(ctx, id, tc.bidArgs.bidder, tc.bidArgs.bid, tc.bidArgs.lot)

			// Check for Place Bid err
			if tc.expectpass {
				require.Nil(t, err)

				// Get auction from store
				auction, found := keeper.GetAuction(ctx, id)
				require.True(t, found)

				// Check auction values
				require.Equal(t, tc.expectedBidder, auction.GetBidder())
				require.Equal(t, tc.expectedBid, auction.GetBid())
				require.Equal(t, tc.expectedEndTime, auction.GetEndTime())
			} else {
				// Check expected error message
				require.Contains(t, err.Error(), tc.expectedError)

			}
		})
	}
}
