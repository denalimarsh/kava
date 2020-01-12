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

// ----------------------------------------------------
// 					Forward Auctions
// 		1. Starting. 2. Execute basic. 3. Bidding.
// ----------------------------------------------------
func TestStartForwardAuction(t *testing.T) {
	someTime := time.Date(1998, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Set up ForwardAuction params
	type args struct {
		seller   string
		lot      sdk.Coin
		bidDenom string
	}

	// Set up ForwardAuction test cases
	testCases := []struct {
		name       string
		blockTime  time.Time
		args       args
		expectPass bool
	}{
		{
			"normal",
			someTime,
			args{liquidator.ModuleName, c("stable", 10), "gov"},
			true,
		},
		{
			"no module account",
			someTime,
			args{"nonExistentModule", c("stable", 10), "gov"},
			false,
		},
		{
			"not enough coins",
			someTime,
			args{liquidator.ModuleName, c("stable", 101), "gov"},
			false,
		},
		{
			"incorrect denom",
			someTime,
			args{liquidator.ModuleName, c("notacoin", 10), "gov"},
			false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up
			initialLiquidatorCoins := cs(c("stable", 100))
			liqAcc := supply.NewEmptyModuleAccount(liquidator.ModuleName, supply.Burner) // TODO: could add test to check for burner permissions
			require.NoError(t, liqAcc.SetCoins(initialLiquidatorCoins))

			tApp := app.NewTestApp()

			tApp.InitializeFromGenesisStates(
				NewAuthGenStateFromAccs(authexported.GenesisAccounts{liqAcc}),
			)
			ctx := tApp.NewContext(false, abci.Header{}).WithBlockTime(tc.blockTime)
			keeper := tApp.GetAuctionKeeper()

			// Execute StartForwardAuction under test
			id, err := keeper.StartForwardAuction(ctx, tc.args.seller, tc.args.lot, tc.args.bidDenom)

			// Get the stored auction and Liquidator module's current coins
			sk := tApp.GetSupplyKeeper()
			liquidatorCoins := sk.GetModuleAccount(ctx, liquidator.ModuleName).GetCoins()
			actualAuc, found := keeper.GetAuction(ctx, id)

			if tc.expectPass {
				require.NoError(t, err)
				// Check coins moved
				require.Equal(t, initialLiquidatorCoins.Sub(cs(tc.args.lot)), liquidatorCoins)
				// Check auction in store and is correct
				require.True(t, found)
				expectedAuction := types.Auction(types.ForwardAuction{BaseAuction: types.BaseAuction{
					ID:         0,
					Initiator:  tc.args.seller,
					Lot:        tc.args.lot,
					Bidder:     nil,
					Bid:        c(tc.args.bidDenom, 0),
					EndTime:    tc.blockTime.Add(types.DefaultMaxAuctionDuration),
					MaxEndTime: tc.blockTime.Add(types.DefaultMaxAuctionDuration),
				}})
				require.Equal(t, expectedAuction, actualAuc)
			} else {
				require.Error(t, err)
				// Check coins not moved
				require.Equal(t, initialLiquidatorCoins, liquidatorCoins)
				// Check auction not in store
				require.False(t, found)
			}
		})
	}
}

func TestExecuteBasicForwardAuction(t *testing.T) {
	// Set up
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	buyer := addrs[0]
	sellerModName := liquidator.ModuleName
	sellerAddr := supply.NewModuleAddress(sellerModName)

	tApp := app.NewTestApp()

	sellerAcc := supply.NewEmptyModuleAccount(sellerModName, supply.Burner) // Forward auctions burn proceeds
	require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 100), c("token2", 100))))
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(buyer, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			sellerAcc,
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Create an auction (lot: 20 token1, initialBid: 0 token2)
	auctionID, err := keeper.StartForwardAuction(ctx, sellerModName, c("token1", 20), "token2") // lot, bid denom
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// PlaceBid (bid: 10 token, lot: same as starting)
	require.NoError(t, keeper.PlaceBid(ctx, auctionID, buyer, c("token2", 10), c("token1", 20))) // bid, lot
	// Check buyer's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have not increased (because proceeds are burned)
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// Close auction at just at auction expiry time
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 120), c("token2", 90)))
}

// TODO: Implement TestForwardAuctionBidding with error edge cases
func TestForwardAuctionBidding(t *testing.T) {
	buyer1 := sdk.AccAddress([]byte("buyer1"))
	buyer2 := sdk.AccAddress([]byte("buyer2"))
	now := time.Now()
	end := now.Add(1000000)

	type args struct {
		currentBlockHeight time.Time
		bidder             sdk.AccAddress
		lot                sdk.Coin
		bid                sdk.Coin
	}
	tests := []struct {
		name    string
		auction types.ForwardAuction
		args    args
		// TODO: update
		// expectedOutputs []BankOutput
		// expectedInputs  []BankInput
		expectedEndTime time.Time
		expectedBidder  sdk.AccAddress
		expectedBid     sdk.Coin
		expectpass      bool
	}{
		{
			"normal",
			types.ForwardAuction{types.BaseAuction{
				Initiator:  liquidator.ModuleName,
				Lot:        c("usdx", 100),
				Bidder:     buyer1,
				Bid:        c("kava", 6),
				EndTime:    end,
				MaxEndTime: end,
			}},
			args{now, buyer2, c("usdx", 100), c("kava", 10)},
			// []BankOutput{{buyer2, c("kava", 10)}},
			// []BankInput{{buyer1, c("kava", 6)}, {seller, c("kava", 4)}},
			now + defaultMaxBidDuration,
			buyer2,
			c("kava", 10),
			true,
		},
		// {
		// 	"lowBid",
		// 	ForwardAuction{BaseAuction{
		// 		Initiator:  seller,
		// 		Lot:        c("usdx", 100),
		// 		Bidder:     buyer1,
		// 		Bid:        c("kava", 6),
		// 		EndTime:    end,
		// 		MaxEndTime: end,
		// 	}},
		// 	args{now, buyer2, c("usdx", 100), c("kava", 5)},
		// 	[]BankOutput{},
		// 	[]BankInput{},
		// 	end,
		// 	buyer1,
		// 	c("kava", 6),
		// 	false,
		// },
		// {
		// 	"equalBid",
		// 	ForwardAuction{BaseAuction{
		// 		Initiator:  seller,
		// 		Lot:        c("usdx", 100),
		// 		Bidder:     buyer1,
		// 		Bid:        c("kava", 6),
		// 		EndTime:    end,
		// 		MaxEndTime: end,
		// 	}},
		// 	args{now, buyer2, c("usdx", 100), c("kava", 6)},
		// 	[]BankOutput{},
		// 	[]BankInput{},
		// 	end,
		// 	buyer1,
		// 	c("kava", 6),
		// 	false,
		// },
		// {
		// 	"timeout",
		// 	ForwardAuction{BaseAuction{
		// 		Initiator:  seller,
		// 		Lot:        c("usdx", 100),
		// 		Bidder:     buyer1,
		// 		Bid:        c("kava", 6),
		// 		EndTime:    end,
		// 		MaxEndTime: end,
		// 	}},
		// 	args{end + 1, buyer2, c("usdx", 100), c("kava", 10)},
		// 	[]BankOutput{},
		// 	[]BankInput{},
		// 	end,
		// 	buyer1,
		// 	c("kava", 6),
		// 	false,
		// },
		// {
		// 	"hitMaxEndTime",
		// 	ForwardAuction{BaseAuction{
		// 		Initiator:  seller,
		// 		Lot:        c("usdx", 100),
		// 		Bidder:     buyer1,
		// 		Bid:        c("kava", 6),
		// 		EndTime:    end,
		// 		MaxEndTime: end,
		// 	}},
		// 	args{end - 1, buyer2, c("usdx", 100), c("kava", 10)},
		// 	[]BankOutput{{buyer2, c("kava", 10)}},
		// 	[]BankInput{{buyer1, c("kava", 6)}, {seller, c("kava", 4)}},
		// 	end, // end time should be capped at MaxEndTime
		// 	buyer2,
		// 	c("kava", 10),
		// 	true,
		// },
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// update auction and return in/outputs
			outputs, inputs, err := tc.auction.PlaceBid(tc.args.currentBlockHeight, tc.args.bidder, tc.args.lot, tc.args.bid)

			// check for err
			if tc.expectpass {
				require.Nil(t, err)
			} else {
				require.NotNil(t, err)
			}
			// check for correct in/outputs
			// require.Equal(t, tc.expectedOutputs, outputs)
			// require.Equal(t, tc.expectedInputs, inputs)
			// check for correct EndTime, bidder, bid
			require.Equal(t, tc.expectedEndTime, tc.auction.EndTime)
			require.Equal(t, tc.expectedBidder, tc.auction.Bidder)
			require.Equal(t, tc.expectedBid, tc.auction.Bid)
		})
	}
}

// ----------------------------------------------------
// 					Reverse Auctions
// 		1. Starting. 2. Execute basic. 3. Bidding.
// ----------------------------------------------------
func TestStartReverseAuction(t *testing.T) {
	// Set up ReverseAuction params
	type args struct {
		buyer      string
		bid        sdk.Coin
		initialLot sdk.Coin
	}

	// Set up ReverseAuction test cases
	testCases := []struct {
		name       string
		args       args
		expectPass bool
	}{
		{
			"normal",
			args{liquidator.ModuleName, c("token1", 10), c("token2", 99999)},
			true,
		},
		// TODO: Add starting ReverseAuction error edge cases
		// {
		// 	"no module account",
		// 	args{"nonExistentModule", c("token1", 10), c("token2", 99999)},
		// 	false,
		// },
		// {
		// 	"not enough coins",
		// 	args{liquidator.ModuleName, c("token1", 101), c("token2", 99999)},
		// 	false,
		// },
		// {
		// 	"incorrect bid denom",
		// 	args{liquidator.ModuleName, c("wrong_denom", 10), c("token2", 99999)},
		// 	false,
		// },
		// 	"incorrect lot denom",
		// 	args{liquidator.ModuleName, c("token1", 10), c("wrong_denom", 99999)},
		// 	false,
		// },
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up seller
			_, addrs := app.GeneratePrivKeyAddressPairs(1)
			seller := addrs[0]

			// Set up buyer (Liquidator module account)
			buyerModName := liquidator.ModuleName
			buyerAddr := supply.NewModuleAddress(buyerModName)
			buyerAcc := supply.NewEmptyModuleAccount(buyerModName, supply.Minter) // Reverse auctions mint payout

			// Set up initial module coins
			initialLiquidatorCoins := cs(
				c("token1", 100),
				c("token2", 100),
			)

			tApp := app.NewTestApp()

			tApp.InitializeFromGenesisStates(
				NewAuthGenStateFromAccs(authexported.GenesisAccounts{
					auth.NewBaseAccount(seller, initialLiquidatorCoins, nil, 0, 0),
					buyerAcc,
				}),
			)
			ctx := tApp.NewContext(false, abci.Header{})
			keeper := tApp.GetAuctionKeeper()

			// Start auction under test
			auctionID, err := keeper.StartReverseAuction(ctx, tc.args.buyer, tc.args.bid, tc.args.initialLot)

			// Get stored auction and Liquidator module's current coins
			// TODO: Coin count
			// sk := tApp.GetSupplyKeeper()
			// liquidatorCoins := sk.GetModuleAccount(ctx, liquidator.ModuleName).GetCoins()
			actualAuc, found := keeper.GetAuction(ctx, auctionID)

			if tc.expectPass {
				require.NoError(t, err)
				// TODO: Check coins moved
				// require.Equal(t, initialLiquidatorCoins.Sub(cs(tc.args.initialLot)), liquidatorCoins)

				// Check auction in store and is correct
				require.True(t, found)
				expectedAuction := types.Auction(types.ReverseAuction{BaseAuction: types.BaseAuction{
					ID:         0,
					Initiator:  tc.args.buyer,
					Lot:        tc.args.initialLot,
					Bidder:     buyerAddr,
					Bid:        tc.args.bid,
					EndTime:    ctx.BlockTime().Add(types.DefaultMaxAuctionDuration),
					MaxEndTime: ctx.BlockTime().Add(types.DefaultMaxAuctionDuration),
				}})
				require.Equal(t, expectedAuction, actualAuc)
			} else {
				require.Error(t, err)
				// TODO: Check coins not moved
				// require.Equal(t, initialLiquidatorCoins, liquidatorCoins)

				// check auction not in store
				require.False(t, found)
			}
		})
	}
}

func TestExecuteBasicReverseAuction(t *testing.T) {
	// Set up
	_, addrs := app.GeneratePrivKeyAddressPairs(1)
	seller := addrs[0]
	buyerModName := liquidator.ModuleName
	buyerAddr := supply.NewModuleAddress(buyerModName)

	tApp := app.NewTestApp()

	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(seller, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			supply.NewEmptyModuleAccount(buyerModName, supply.Minter), // Reverse auctions mint payout
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartReverseAuction(ctx, buyerModName, c("token1", 20), c("token2", 99999)) // buyer, bid, initialLot
	require.NoError(t, err)
	// Check buyer's coins have not decreased, as lot is minted at the end
	tApp.CheckBalance(t, ctx, buyerAddr, nil) // zero coins

	// Place a bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, seller, c("token1", 20), c("token2", 10))) // bid, lot
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 100)))
	// Check buyer's coins have increased
	tApp.CheckBalance(t, ctx, buyerAddr, cs(c("token1", 20)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check seller's coins increased
	tApp.CheckBalance(t, ctx, seller, cs(c("token1", 80), c("token2", 110)))
}

// TODO: Implement TestReverseAuctionBidding with error edge cases
// func TestReverseAuctionBidding(t *testing.T) {
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

// ----------------------------------------------------
// 					Reverse Auctions
// 		1. Starting. 2. Execute basic. 3. Bidding.
// ----------------------------------------------------
func TestStartForwardReverseAuction(t *testing.T) {
	// TODO: Implement TestStartForwardReverseAuction
}

func TestExecuteBasicForwardReverseAuction(t *testing.T) {
	// Setup
	_, addrs := app.GeneratePrivKeyAddressPairs(4)
	buyer := addrs[0]
	returnAddrs := addrs[1:]
	returnWeights := is(30, 20, 10)
	sellerModName := liquidator.ModuleName
	sellerAddr := supply.NewModuleAddress(sellerModName)

	tApp := app.NewTestApp()
	sellerAcc := supply.NewEmptyModuleAccount(sellerModName)
	require.NoError(t, sellerAcc.SetCoins(cs(c("token1", 100), c("token2", 100))))
	tApp.InitializeFromGenesisStates(
		NewAuthGenStateFromAccs(authexported.GenesisAccounts{
			auth.NewBaseAccount(buyer, cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			auth.NewBaseAccount(returnAddrs[0], cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			auth.NewBaseAccount(returnAddrs[1], cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			auth.NewBaseAccount(returnAddrs[2], cs(c("token1", 100), c("token2", 100)), nil, 0, 0),
			sellerAcc,
		}),
	)
	ctx := tApp.NewContext(false, abci.Header{})
	keeper := tApp.GetAuctionKeeper()

	// Start auction
	auctionID, err := keeper.StartForwardReverseAuction(ctx, sellerModName, c("token1", 20), c("token2", 50), returnAddrs, returnWeights) // seller, lot, maxBid, otherPerson
	require.NoError(t, err)
	// Check seller's coins have decreased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 100)))

	// Place a forward bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 10), c("token1", 20))) // bid, lot
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 90)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 110)))
	// Check return addresses have not received coins
	for _, ra := range returnAddrs {
		tApp.CheckBalance(t, ctx, ra, cs(c("token1", 100), c("token2", 100)))
	}

	// Place a reverse bid
	require.NoError(t, keeper.PlaceBid(ctx, 0, buyer, c("token2", 50), c("token1", 15))) // bid, lot
	// Check bidder's coins have decreased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 100), c("token2", 50)))
	// Check seller's coins have increased
	tApp.CheckBalance(t, ctx, sellerAddr, cs(c("token1", 80), c("token2", 150)))
	// Check return addresses have received coins
	tApp.CheckBalance(t, ctx, returnAddrs[0], cs(c("token1", 102), c("token2", 100)))
	tApp.CheckBalance(t, ctx, returnAddrs[1], cs(c("token1", 102), c("token2", 100)))
	tApp.CheckBalance(t, ctx, returnAddrs[2], cs(c("token1", 101), c("token2", 100)))

	// Close auction at just after auction expiry
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(types.DefaultBidDuration))
	require.NoError(t, keeper.CloseAuction(ctx, auctionID))
	// Check buyer's coins increased
	tApp.CheckBalance(t, ctx, buyer, cs(c("token1", 115), c("token2", 50)))
}

// TODO: Implement TestForwardReverseAuctionBidding with error edge cases
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
