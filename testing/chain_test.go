package ibctesting_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/x/staking/types"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
)

func TestChangeValSet(t *testing.T) {
	coord := ibctesting.NewCoordinator(t, 2)
	chainA := coord.GetChain(ibctesting.GetChainID(1))
	chainB := coord.GetChain(ibctesting.GetChainID(2))

	path := ibctesting.NewPath(chainA, chainB)
	coord.Setup(path)

	amount, ok := sdkmath.NewIntFromString("10000000000000000000")
	require.True(t, ok)
	amount2, ok := sdkmath.NewIntFromString("30000000000000000000")
	require.True(t, ok)

	val := chainA.GetSimApp().StakingKeeper.GetValidators(chainA.GetContext(), 4)

	fmt.Println("DELEGATE")
	// TODO add a send/receive packet to trigger call to update client with updated valset

	fmt.Println("CHAINA TS", chainA.GetContext().BlockTime())
	timeoutHeight := clienttypes.ZeroHeight()
	// timeoutHeight := clienttypes.GetSelfHeight(chainB.GetContext())
	timeoutTimestamp := uint64(chainA.GetContext().BlockTime().Add(time.Hour * 24).UnixNano())
	// timeoutTimestamp := uint64(0)
	sequence, err := path.EndpointB.SendPacket(timeoutHeight, timeoutTimestamp, ibctesting.MockPacketData)
	packet := channeltypes.NewPacket(ibctesting.MockPacketData, sequence, path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, timeoutHeight, timeoutTimestamp)
	err = path.EndpointA.RecvPacket(packet)
	require.NoError(t, err)

	// coord.CommitBlock(chainA)
	// coord.CommitBlock(chainB)
	err = path.EndpointA.UpdateClient()
	require.NoError(t, err)
	err = path.EndpointB.UpdateClient()
	require.NoError(t, err)

	coord.CommitBlock(chainA)
	coord.CommitBlock(chainB)
	fmt.Println("COMMIT")

	chainA.GetSimApp().StakingKeeper.Delegate(chainA.GetContext(), chainA.SenderAccounts[1].SenderAccount.GetAddress(), //nolint:errcheck // ignore error for test
		amount, types.Unbonded, val[1], true)
	chainA.GetSimApp().StakingKeeper.Delegate(chainA.GetContext(), chainA.SenderAccounts[3].SenderAccount.GetAddress(), //nolint:errcheck // ignore error for test
		amount2, types.Unbonded, val[3], true)

	coord.CommitBlock(chainA)
	fmt.Println("COMMIT")

	// verify that update clients works even after validator update goes into effect
	err = path.EndpointB.UpdateClient()
	require.NoError(t, err)
	err = path.EndpointB.UpdateClient()
	require.NoError(t, err)
}
