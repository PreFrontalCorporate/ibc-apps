package packetforward_test

import (
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"testing"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/keeper"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/packetforward/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v10/test"
	"github.com/iancoleman/orderedmap"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
)

var (
	testDenom     = "uatom"
	testAmount    = "100"
	testAmount256 = "100000000000000000000"

	testSourcePort         = "transfer"
	testSourceChannel      = "channel-10"
	testDestinationPort    = "transfer"
	testDestinationChannel = "channel-11"

	senderAddr        = "cosmos1wnlew8ss0sqclfalvj6jkcyvnwq79fd74qxxue"
	hostAddr          = "cosmos1vzxkv3lxccnttr9rs0002s93sgw72h7ghukuhs"
	intermediateAddr  = "cosmos1v954djef63x2lqj8yy7r3r487heg0exdmkj0sr"
	hostAddr2         = "cosmos1q4p4gx889lfek5augdurrjclwtqvjhuntm6j4m"
	intermediateAddr2 = "cosmos1eadmq78mkhg6lrk87lxgateketvz44crq45jpe"
	destAddr          = "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k"
	port              = "transfer"
	channel           = "channel-0"
	channel2          = "channel-1"
)

func makeIBCDenom(port, channel, denom string) string {
	prefixedDenom := transfertypes.GetDenomPrefix(port, channel) + denom
	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}

func emptyPacket() channeltypes.Packet {
	return channeltypes.Packet{}
}

func transferPacket(t *testing.T, sender string, receiver string, metadata any) channeltypes.Packet {
	t.Helper()
	transferPacket := transfertypes.FungibleTokenPacketData{
		Denom:    testDenom,
		Amount:   testAmount,
		Sender:   sender,
		Receiver: receiver,
	}

	if metadata != nil {
		if mStr, ok := metadata.(string); ok {
			transferPacket.Memo = mStr
		} else {
			memo, err := json.Marshal(metadata)
			require.NoError(t, err)
			transferPacket.Memo = string(memo)
		}
	}

	transferData, err := transfertypes.ModuleCdc.MarshalJSON(&transferPacket)
	require.NoError(t, err)

	return channeltypes.Packet{
		SourcePort:         testSourcePort,
		SourceChannel:      testSourceChannel,
		DestinationPort:    testDestinationPort,
		DestinationChannel: testDestinationChannel,
		Data:               transferData,
	}
}

func transferPacket256(t *testing.T, sender string, receiver string, metadata any) channeltypes.Packet {
	t.Helper()
	transferPacket := transfertypes.FungibleTokenPacketData{
		Denom:    testDenom,
		Amount:   testAmount256,
		Sender:   sender,
		Receiver: receiver,
	}

	if metadata != nil {
		if mStr, ok := metadata.(string); ok {
			transferPacket.Memo = mStr
		} else {
			memo, err := json.Marshal(metadata)
			require.NoError(t, err)
			transferPacket.Memo = string(memo)
		}
	}

	transferData, err := transfertypes.ModuleCdc.MarshalJSON(&transferPacket)
	require.NoError(t, err)

	return channeltypes.Packet{
		SourcePort:         testSourcePort,
		SourceChannel:      testSourceChannel,
		DestinationPort:    testDestinationPort,
		DestinationChannel: testDestinationChannel,
		Data:               transferData,
	}
}

func TestOnRecvPacket_EmptyPacket(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	senderAccAddr := test.AccAddress()
	packet := emptyPacket()

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr).
			Return(channeltypes.NewResultAcknowledgement([]byte(""))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr)
	require.True(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
	require.Equal(t, "", expectedAck.GetError())
}

func TestOnRecvPacket_InvalidReceiver(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	senderAccAddr := test.AccAddress()
	packet := transferPacket(t, test.AccAddress().String(), "", nil)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr).
			Return(channeltypes.NewResultAcknowledgement([]byte("test"))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr)
	require.True(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
}

func TestOnRecvPacket_NoForward(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	senderAccAddr := test.AccAddress()
	packet := transferPacket(t, test.AccAddress().String(), "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", nil)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr).
			Return(channeltypes.NewResultAcknowledgement([]byte("test"))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr)
	require.True(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
	require.Equal(t, "test", string(expectedAck.GetResult()))
}

func TestOnRecvPacket_NoMemo(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	// Test data
	senderAccAddr := test.AccAddress()
	packet := transferPacket(t, test.AccAddress().String(), "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", "{}")

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr).
			Return(channeltypes.NewResultAcknowledgement([]byte("test"))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr)
	require.True(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
	require.Equal(t, "test", string(expectedAck.GetResult()))
}

func TestOnRecvPacket_RecvPacketFailed(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	senderAccAddr := test.AccAddress()
	packet := transferPacket(t, test.AccAddress().String(), "cosmos16plylpsgxechajltx9yeseqexzdzut9g8vla4k", nil)

	// Expected mocks
	gomock.InOrder(
		// We return a failed OnRecvPacket
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr).
			Return(channeltypes.NewErrorAcknowledgement(fmt.Errorf("test"))),
	)

	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packet, senderAccAddr)
	require.False(t, ack.Success())

	expectedAck := &channeltypes.Acknowledgement{}
	err := cdc.UnmarshalJSON(ack.Acknowledgement(), expectedAck)
	require.NoError(t, err)
	require.Equal(t, "ABCI code: 1: error handling packet: see events for details", expectedAck.GetError())
}

func TestOnRecvPacket_ForwardNoFee(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()
	testCoin := sdk.NewCoin(denom, sdkmath.NewInt(100))
	metadata := &types.PacketMetadata{Forward: &types.ForwardMetadata{
		Receiver: destAddr,
		Port:     port,
		Channel:  channel,
	}}
	packetOrig := transferPacket(t, senderAddr, hostAddr, metadata)
	packetModifiedSender := transferPacket(t, senderAddr, intermediateAddr, nil)
	packetFwd := transferPacket(t, intermediateAddr, destAddr, nil)

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packetModifiedSender, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			ctx,
			transfertypes.NewMsgTransfer(
				port,
				channel,
				testCoin,
				intermediateAddr,
				destAddr,
				keeper.DefaultTransferPacketTimeoutHeight,
				uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
				"",
			),
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, transfertypes.V1, packetFwd, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with packetforward module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// ack returned from chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, transfertypes.V1, packetFwd, successAck, senderAccAddr)
	require.NoError(t, err)
}

func TestOnRecvPacket_ForwardAmountInt256(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()

	amount256, ok := sdkmath.NewIntFromString(testAmount256)
	require.True(t, ok)

	testCoin := sdk.NewCoin(denom, amount256)
	metadata := &types.PacketMetadata{Forward: &types.ForwardMetadata{
		Receiver: destAddr,
		Port:     port,
		Channel:  channel,
	}}

	packetOrig := transferPacket256(t, senderAddr, hostAddr, metadata)
	packetModifiedSender := transferPacket256(t, senderAddr, intermediateAddr, nil)
	packetFwd := transferPacket256(t, intermediateAddr, destAddr, nil)

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packetModifiedSender, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			ctx,
			transfertypes.NewMsgTransfer(
				port,
				channel,
				testCoin,
				intermediateAddr,
				destAddr,
				keeper.DefaultTransferPacketTimeoutHeight,
				uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
				"",
			),
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, transfertypes.V1, packetFwd, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with packetforward module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// ack returned from chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, transfertypes.V1, packetFwd, successAck, senderAccAddr)
	require.NoError(t, err)
}

func TestOnRecvPacket_ForwardMultihopStringNext(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()
	senderAccAddr2 := test.AccAddress()
	testCoin := sdk.NewCoin(denom, sdkmath.NewInt(100))
	nextMetadata := &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel2,
		},
	}
	nextBz, err := json.Marshal(nextMetadata)
	require.NoError(t, err)

	metadata := &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: hostAddr2,
			Port:     port,
			Channel:  channel,
			Next:     types.NewJSONObject(false, nextBz, orderedmap.OrderedMap{}),
		},
	}

	packetOrig := transferPacket(t, senderAddr, hostAddr, metadata)
	packetModifiedSender := transferPacket(t, senderAddr, intermediateAddr, nil)
	packet2 := transferPacket(t, intermediateAddr, hostAddr2, nextMetadata)
	packet2ModifiedSender := transferPacket(t, intermediateAddr, intermediateAddr2, nil)
	packetFwd := transferPacket(t, intermediateAddr2, destAddr, nil)

	memo1, err := json.Marshal(nextMetadata)
	require.NoError(t, err)

	msgTransfer1 := transfertypes.NewMsgTransfer(
		port,
		channel,
		testCoin,
		intermediateAddr,
		hostAddr2,
		keeper.DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
		string(memo1),
	)

	// no memo on final forward
	msgTransfer2 := transfertypes.NewMsgTransfer(
		port,
		channel2,
		testCoin,
		intermediateAddr2,
		destAddr,
		keeper.DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
		"",
	)

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packetModifiedSender, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			ctx,
			msgTransfer1,
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packet2ModifiedSender, senderAccAddr2).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			ctx,
			msgTransfer2,
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, transfertypes.V1, packetFwd, successAck, senderAccAddr2).
			Return(nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, transfertypes.V1, packet2, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with packetforward module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// chain C with packetforward module receives packet and forwards. ack should be nil so that it is not written yet.
	ack = forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packet2, senderAccAddr2)
	require.Nil(t, ack)

	// ack returned from chain D to chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, transfertypes.V1, packetFwd, successAck, senderAccAddr2)
	require.NoError(t, err)

	// ack returned from chain C to chain B
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, transfertypes.V1, packet2, successAck, senderAccAddr)
	require.NoError(t, err)
}

func TestOnRecvPacket_ForwardMultihopJSONNext(t *testing.T) {
	var err error
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	setup := test.NewTestSetup(t, ctl)
	ctx := setup.Initializer.Ctx
	cdc := setup.Initializer.Marshaler
	forwardMiddleware := setup.ForwardMiddleware

	denom := makeIBCDenom(testDestinationPort, testDestinationChannel, testDenom)
	senderAccAddr := test.AccAddress()
	senderAccAddr2 := test.AccAddress()
	testCoin := sdk.NewCoin(denom, sdkmath.NewInt(100))
	nextMetadata := &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: destAddr,
			Port:     port,
			Channel:  channel2,
		},
	}
	nextBz, err := json.Marshal(nextMetadata)
	require.NoError(t, err)

	nextJSONObject := new(types.JSONObject)
	err = json.Unmarshal(nextBz, nextJSONObject)
	require.NoError(t, err)

	metadata := &types.PacketMetadata{
		Forward: &types.ForwardMetadata{
			Receiver: hostAddr2,
			Port:     port,
			Channel:  channel,
			Next:     nextJSONObject,
		},
	}
	packetOrig := transferPacket(t, senderAddr, hostAddr, metadata)
	packetModifiedSender := transferPacket(t, senderAddr, intermediateAddr, nil)
	packet2 := transferPacket(t, intermediateAddr, hostAddr2, nextMetadata)
	packet2ModifiedSender := transferPacket(t, intermediateAddr, intermediateAddr2, nil)
	packetFwd := transferPacket(t, intermediateAddr2, destAddr, nil)

	memo1, err := json.Marshal(nextMetadata)
	require.NoError(t, err)

	msgTransfer1 := transfertypes.NewMsgTransfer(
		port,
		channel,
		testCoin,
		intermediateAddr,
		hostAddr2,
		keeper.DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
		string(memo1),
	)

	// no memo on final forward
	msgTransfer2 := transfertypes.NewMsgTransfer(
		port,
		channel2,
		testCoin,
		intermediateAddr2,
		destAddr,
		keeper.DefaultTransferPacketTimeoutHeight,
		uint64(ctx.BlockTime().UnixNano())+uint64(keeper.DefaultForwardTransferPacketTimeoutTimestamp.Nanoseconds()),
		"",
	)

	acknowledgement := channeltypes.NewResultAcknowledgement([]byte("test"))
	successAck := cdc.MustMarshalJSON(&acknowledgement)

	// Expected mocks
	gomock.InOrder(
		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packetModifiedSender, senderAccAddr).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			ctx,
			msgTransfer1,
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnRecvPacket(ctx, transfertypes.V1, packet2ModifiedSender, senderAccAddr2).
			Return(acknowledgement),

		setup.Mocks.TransferKeeperMock.EXPECT().Transfer(
			ctx,
			msgTransfer2,
		).Return(&transfertypes.MsgTransferResponse{Sequence: 0}, nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, transfertypes.V1, packetFwd, successAck, senderAccAddr2).
			Return(nil),

		setup.Mocks.IBCModuleMock.EXPECT().OnAcknowledgementPacket(ctx, transfertypes.V1, packet2, successAck, senderAccAddr).
			Return(nil),
	)

	// chain B with packetforward module receives packet and forwards. ack should be nil so that it is not written yet.
	ack := forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packetOrig, senderAccAddr)
	require.Nil(t, ack)

	// chain C with packetforward module receives packet and forwards. ack should be nil so that it is not written yet.
	ack = forwardMiddleware.OnRecvPacket(ctx, transfertypes.V1, packet2, senderAccAddr2)
	require.Nil(t, ack)

	// ack returned from chain D to chain C
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, transfertypes.V1, packetFwd, successAck, senderAccAddr2)
	require.NoError(t, err)

	// ack returned from chain C to chain B
	err = forwardMiddleware.OnAcknowledgementPacket(ctx, transfertypes.V1, packet2, successAck, senderAccAddr)
	require.NoError(t, err)
}
