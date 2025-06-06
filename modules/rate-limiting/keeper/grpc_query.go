package keeper

import (
	"context"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibctmtypes "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

var _ types.QueryServer = Keeper{}

// Query all rate limits
func (k Keeper) AllRateLimits(c context.Context, req *types.QueryAllRateLimitsRequest) (*types.QueryAllRateLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	rateLimits := k.GetAllRateLimits(ctx)
	return &types.QueryAllRateLimitsResponse{RateLimits: rateLimits}, nil
}

// Query a rate limit by denom and channelId
func (k Keeper) RateLimit(c context.Context, req *types.QueryRateLimitRequest) (*types.QueryRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	rateLimit, found := k.GetRateLimit(ctx, req.Denom, req.ChannelOrClientId)
	if !found {
		return &types.QueryRateLimitResponse{}, nil
	}
	return &types.QueryRateLimitResponse{RateLimit: &rateLimit}, nil
}

// Query all rate limits for a given chain
func (k Keeper) RateLimitsByChainId(c context.Context, req *types.QueryRateLimitsByChainIdRequest) (*types.QueryRateLimitsByChainIdResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	rateLimits := []types.RateLimit{}
	for _, rateLimit := range k.GetAllRateLimits(ctx) {

		// Determine the client state from the channel Id
		_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, transfertypes.PortID, rateLimit.Path.ChannelOrClientId)
		if err != nil {
			var ok bool
			clientState, ok = k.clientKeeper.GetClientState(ctx, rateLimit.Path.ChannelOrClientId)
			if !ok {
				return &types.QueryRateLimitsByChainIdResponse{}, errorsmod.Wrapf(types.ErrInvalidClientState, "Unable to fetch client state from channel or client Id")
			}
		}
		client, ok := clientState.(*ibctmtypes.ClientState)
		if !ok {
			// If the client state is not a tendermint client state, we don't return the rate limit from this query
			continue
		}

		// If the chain ID matches, add the rate limit to the returned list
		if client.ChainId == req.ChainId {
			rateLimits = append(rateLimits, rateLimit)
		}
	}

	return &types.QueryRateLimitsByChainIdResponse{RateLimits: rateLimits}, nil
}

// Query all rate limits for a given channel
func (k Keeper) RateLimitsByChannelOrClientId(c context.Context, req *types.QueryRateLimitsByChannelOrClientIdRequest) (*types.QueryRateLimitsByChannelOrClientIdResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	rateLimits := []types.RateLimit{}
	for _, rateLimit := range k.GetAllRateLimits(ctx) {
		// If the channel ID matches, add the rate limit to the returned list
		if rateLimit.Path.ChannelOrClientId == req.ChannelOrClientId {
			rateLimits = append(rateLimits, rateLimit)
		}
	}

	return &types.QueryRateLimitsByChannelOrClientIdResponse{RateLimits: rateLimits}, nil
}

// Query all blacklisted denoms
func (k Keeper) AllBlacklistedDenoms(c context.Context, req *types.QueryAllBlacklistedDenomsRequest) (*types.QueryAllBlacklistedDenomsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	blacklistedDenoms := k.GetAllBlacklistedDenoms(ctx)
	return &types.QueryAllBlacklistedDenomsResponse{Denoms: blacklistedDenoms}, nil
}

// Query all whitelisted addresses
func (k Keeper) AllWhitelistedAddresses(c context.Context, req *types.QueryAllWhitelistedAddressesRequest) (*types.QueryAllWhitelistedAddressesResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	whitelistedAddresses := k.GetAllWhitelistedAddressPairs(ctx)
	return &types.QueryAllWhitelistedAddressesResponse{AddressPairs: whitelistedAddresses}, nil
}
