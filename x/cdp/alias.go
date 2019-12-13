// nolint
// autogenerated code using github.com/rigelrozanski/multitool
// aliases generated for the following subdirectories:
// ALIASGEN: github.com/kava-labs/kava/x/cdp/types/
// ALIASGEN: github.com/kava-labs/kava/x/cdp/keeper/
package cdp

import (
	"github.com/kava-labs/kava/x/cdp/keeper"
	"github.com/kava-labs/kava/x/cdp/types"
)

const (
	CodeCdpAlreadyExists            = types.CodeCdpAlreadyExists
	CodeCollateralLengthInvalid     = types.CodeCollateralLengthInvalid
	CodeCollateralNotSupported      = types.CodeCollateralNotSupported
	CodeDebtNotSupported            = types.CodeDebtNotSupported
	CodeExceedsDebtLimit            = types.CodeExceedsDebtLimit
	CodeInvalidCollateralRatio      = types.CodeInvalidCollateralRatio
	CodeCdpNotFound                 = types.CodeCdpNotFound
	CodeDepositNotFound             = types.CodeDepositNotFound
	CodeInvalidDepositDenom         = types.CodeInvalidDepositDenom
	CodeInvalidPaymentDenom         = types.CodeInvalidPaymentDenom
	CodeDepositNotAvailable         = types.CodeDepositNotAvailable
	CodeInvalidCollateralDenom      = types.CodeInvalidCollateralDenom
	EventTypeCreateCdp              = types.EventTypeCreateCdp
	EventTypeCdpDeposit             = types.EventTypeCdpDeposit
	EventTypeCdpDraw                = types.EventTypeCdpDraw
	EventTypeCdpRepay               = types.EventTypeCdpRepay
	EventTypeCdpClose               = types.EventTypeCdpClose
	EventTypeCdpWithdrawal          = types.EventTypeCdpWithdrawal
	AttributeKeyCdpID               = types.AttributeKeyCdpID
	AttributeValueCategory          = types.AttributeValueCategory
	ModuleName                      = types.ModuleName
	StoreKey                        = types.StoreKey
	RouterKey                       = types.RouterKey
	DefaultParamspace               = types.DefaultParamspace
	QueryGetCdp                     = types.QueryGetCdp
	QueryGetCdps                    = types.QueryGetCdps
	QueryGetCdpsByCollateralization = types.QueryGetCdpsByCollateralization
	QueryGetParams                  = types.QueryGetParams
	RestOwner                       = types.RestOwner
	RestCollateralDenom             = types.RestCollateralDenom
	RestRatio                       = types.RestRatio
)

var (
	// functions aliases
	NewCDP                      = types.NewCDP
	RegisterCodec               = types.RegisterCodec
	NewDeposit                  = types.NewDeposit
	ErrCdpAlreadyExists         = types.ErrCdpAlreadyExists
	ErrInvalidCollateralLength  = types.ErrInvalidCollateralLength
	ErrCollateralNotSupported   = types.ErrCollateralNotSupported
	ErrDebtNotSupported         = types.ErrDebtNotSupported
	ErrExceedsDebtLimit         = types.ErrExceedsDebtLimit
	ErrInvalidCollateralRatio   = types.ErrInvalidCollateralRatio
	ErrCdpNotFound              = types.ErrCdpNotFound
	ErrDepositNotFound          = types.ErrDepositNotFound
	ErrInvalidDepositDenom      = types.ErrInvalidDepositDenom
	ErrInvalidPaymentDenom      = types.ErrInvalidPaymentDenom
	ErrDepositNotAvailable      = types.ErrDepositNotAvailable
	ErrInvalidCollateralDenom   = types.ErrInvalidCollateralDenom
	DefaultGenesisState         = types.DefaultGenesisState
	ValidateGenesis             = types.ValidateGenesis
	GetCdpIDBytes               = types.GetCdpIDBytes
	GetCdpIDFromBytes           = types.GetCdpIDFromBytes
	CdpKey                      = types.CdpKey
	SplitCdpKey                 = types.SplitCdpKey
	DepositKey                  = types.DepositKey
	SplitDepositKey             = types.SplitDepositKey
	CollateralRatioBytes        = types.CollateralRatioBytes
	CollateralRatioKey          = types.CollateralRatioKey
	SplitCollateralRatioKey     = types.SplitCollateralRatioKey
	CollateralRatioIterKey      = types.CollateralRatioIterKey
	SplitCollateralRatioIterKey = types.SplitCollateralRatioIterKey
	NewMsgCreateCDP             = types.NewMsgCreateCDP
	NewMsgDeposit               = types.NewMsgDeposit
	NewMsgWithdraw              = types.NewMsgWithdraw
	NewMsgDrawDebt              = types.NewMsgDrawDebt
	NewMsgRepayDebt             = types.NewMsgRepayDebt
	NewParams                   = types.NewParams
	DefaultParams               = types.DefaultParams
	ParamKeyTable               = types.ParamKeyTable
	NewQueryCdpsParams          = types.NewQueryCdpsParams
	NewQueryCdpParams           = types.NewQueryCdpParams
	NewQueryCdpsByRatioParams   = types.NewQueryCdpsByRatioParams
	ValidSortableDec            = types.ValidSortableDec
	SortableDecBytes            = types.SortableDecBytes
	ParseDecBytes               = types.ParseDecBytes
	NewKeeper                   = keeper.NewKeeper
	NewQuerier                  = keeper.NewQuerier

	// variable aliases
	ModuleCdc                  = types.ModuleCdc
	CdpIdKeyPrefix             = types.CdpIdKeyPrefix
	CdpKeyPrefix               = types.CdpKeyPrefix
	CollateralRatioIndexPrefix = types.CollateralRatioIndexPrefix
	CdpIdKey                   = types.CdpIdKey
	DebtDenomKey               = types.DebtDenomKey
	DepositKeyPrefix           = types.DepositKeyPrefix
	PrincipalKeyPrefix         = types.PrincipalKeyPrefix
	AccumulatorKeyPrefix       = types.AccumulatorKeyPrefix
	KeyGlobalDebtLimit         = types.KeyGlobalDebtLimit
	KeyCollateralParams        = types.KeyCollateralParams
	KeyDebtParams              = types.KeyDebtParams
	KeyCircuitBreaker          = types.KeyCircuitBreaker
	DefaultGlobalDebt          = types.DefaultGlobalDebt
	DefaultCircuitBreaker      = types.DefaultCircuitBreaker
	DefaultCollateralParams    = types.DefaultCollateralParams
	DefaultDebtParams          = types.DefaultDebtParams
	DefaultCdpStartingID       = types.DefaultCdpStartingID
	DefaultDebtDenom           = types.DefaultDebtDenom
	MaxSortableDec             = types.MaxSortableDec
)

type (
	CDP                    = types.CDP
	CDPs                   = types.CDPs
	CollateralState        = types.CollateralState
	Deposit                = types.Deposit
	Deposits               = types.Deposits
	PricefeedKeeper        = types.PricefeedKeeper
	GenesisState           = types.GenesisState
	MsgCreateCDP           = types.MsgCreateCDP
	MsgDeposit             = types.MsgDeposit
	MsgWithdraw            = types.MsgWithdraw
	MsgDrawDebt            = types.MsgDrawDebt
	MsgRepayDebt           = types.MsgRepayDebt
	MsgTransferCDP         = types.MsgTransferCDP
	Params                 = types.Params
	CollateralParam        = types.CollateralParam
	CollateralParams       = types.CollateralParams
	DebtParam              = types.DebtParam
	DebtParams             = types.DebtParams
	QueryCdpsParams        = types.QueryCdpsParams
	QueryCdpParams         = types.QueryCdpParams
	QueryCdpsByRatioParams = types.QueryCdpsByRatioParams
	Keeper                 = keeper.Keeper
)
