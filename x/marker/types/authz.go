package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var (
	_ authz.Authorization = &MarkerTransferAuthorization{}
)

// NewMarkerTransferAuthorization creates a new MarkerTransferAuthorization object.
func NewMarkerTransferAuthorization(transferLimit sdk.Coins) *MarkerTransferAuthorization {
	return &MarkerTransferAuthorization{
		TransferLimit: transferLimit,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a MarkerTransferAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&MsgTransferRequest{})
}

// Accept implements Authorization.Accept.
func (a MarkerTransferAuthorization) Accept(ctx sdk.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	switch msg := msg.(type) {
	case *MsgTransferRequest:
		limitLeft, isNegative := a.DecreaseTransferLimit(msg.Amount)
		if isNegative {
			return authz.AcceptResponse{}, sdkerrors.ErrInsufficientFunds.Wrapf("requested amount is more than spend limit")
		}
		shouldDelete := false
		if limitLeft.IsZero() {
			shouldDelete = true
		}
		return authz.AcceptResponse{Accept: true, Delete: shouldDelete, Updated: &MarkerTransferAuthorization{TransferLimit: limitLeft}}, nil
	default:
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
	}
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a MarkerTransferAuthorization) ValidateBasic() error {
	if a.TransferLimit == nil {
		return sdkerrors.ErrInvalidCoins.Wrap("spend limit cannot be nil")
	}
	if !a.TransferLimit.IsAllPositive() {
		return sdkerrors.ErrInvalidCoins.Wrap("spend limit cannot be negitive")
	}
	return nil
}

// DecreaseTransferLimit will return the decreased transfer limit and if it is negative
func (a MarkerTransferAuthorization) DecreaseTransferLimit(amount sdk.Coin) (sdk.Coins, bool) {
	return a.TransferLimit.SafeSub(amount)
}
