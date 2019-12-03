package nodeconfig

import (
	"math/big"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/harmony-one/bls/ffi/go/bls"

	mock_shardingconfig "github.com/harmony-one/harmony/internal/configs/sharding/mock"
	"github.com/harmony-one/harmony/internal/params"
)

func TestNodeConfigSingleton(t *testing.T) {
	// init 3 configs
	_ = GetShardConfig(2)

	// get the singleton variable
	c := GetShardConfig(Global)

	c.SetBeaconGroupID(GroupIDBeacon)

	d := GetShardConfig(Global)

	g := d.GetBeaconGroupID()

	if g != GroupIDBeacon {
		t.Errorf("GetBeaconGroupID = %v, expected = %v", g, GroupIDBeacon)
	}
}

func TestNodeConfigMultiple(t *testing.T) {
	// init 3 configs
	d := GetShardConfig(1)
	e := GetShardConfig(0)
	f := GetShardConfig(42)

	if f != nil {
		t.Errorf("expecting nil, got: %v", f)
	}

	d.SetShardGroupID("abcd")
	if d.GetShardGroupID() != "abcd" {
		t.Errorf("expecting abcd, got: %v", d.GetShardGroupID())
	}

	e.SetClientGroupID("client")
	if e.GetClientGroupID() != "client" {
		t.Errorf("expecting client, got: %v", d.GetClientGroupID())
	}

	e.SetIsClient(false)
	if e.IsClient() != false {
		t.Errorf("expecting false, got: %v", e.IsClient())
	}
}

func blsPubKeyFromHex(hex string) *bls.PublicKey {
	var k bls.PublicKey
	if err := k.DeserializeHexStr(hex); err != nil {
		panic(err)
	}
	return &k
}

func TestConfigType_ShardIDFromConsensusKey(t *testing.T) {
	type fields struct {
		ConsensusPubKey *bls.PublicKey
		networkType     NetworkType
	}
	tests := []struct {
		name   string
		fields fields
		epoch  *big.Int
		shards uint32
		want   uint32
	}{
		{
			"Mainnet",
			fields{
				blsPubKeyFromHex("ca23704be46ce9c4704681ac9c08ddc644f1858a5c28ce236e1b5d9dee67c1f5a28075b5ef089adeffa8a372c1762007"),
				"mainnet",
			},
			params.MainnetChainConfig.StakingEpoch,
			4,
			3,
		},
		{
			"Testnet",
			fields{
				blsPubKeyFromHex("e7f54994bc5c02edeeb178ce2d34db276a893bab5c59ac3d7eb9f077c893f9e31171de6236ba0e21be415d8631e45b91"),
				"testnet",
			},
			params.TestnetChainConfig.StakingEpoch,
			3,
			1,
		},
		{
			"Devnet",
			fields{
				blsPubKeyFromHex("e7f54994bc5c02edeeb178ce2d34db276a893bab5c59ac3d7eb9f077c893f9e31171de6236ba0e21be415d8631e45b91"),
				"devnet",
			},
			params.TestnetChainConfig.StakingEpoch,
			2,
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := gomock.NewController(t)
			defer mc.Finish()
			instance := mock_shardingconfig.NewMockInstance(mc)
			instance.EXPECT().NumShards().Return(tt.shards)
			schedule := mock_shardingconfig.NewMockSchedule(mc)
			schedule.EXPECT().InstanceForEpoch(tt.epoch).Return(instance)
			conf := &ConfigType{
				ConsensusPubKey:  tt.fields.ConsensusPubKey,
				networkType:      tt.fields.networkType,
				shardingSchedule: schedule,
			}
			got, err := conf.ShardIDFromConsensusKey()
			if err != nil {
				t.Errorf("ShardIDFromConsensusKey() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ShardIDFromConsensusKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}
