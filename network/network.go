// Copyright © 2017 PolySwarm <info@polyswarm.io>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package network

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/spf13/viper"
	"gitlab.inn4science.com/gophers/perigord/project"
)

type NetworkConfig struct {
	url          string
	keystorePath string
	passphrase   string
	mnemonic     string
	numAccounts  int64
}

var networks map[string]NetworkConfig

func InitNetworks() error {
	prj, err := project.FindProject()
	if err != nil {
		return errors.New("Could not find project")
	}

	viper.SetConfigFile(filepath.Join(prj.AbsPath(), project.ProjectConfigFilename))
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	configs := viper.GetStringMap("networks")
	networks = make(map[string]NetworkConfig)

	for key, _ := range configs {
		config := viper.GetStringMapString("networks." + key)
		var count int64
		if config["mnemonic"] != "" {
			count, err = strconv.ParseInt(config["num_accounts"], 10, 64)
			if err != nil {
				count = 10
			}
		}

		networks[key] = NetworkConfig{
			url:          config["url"],
			keystorePath: config["keystore"],
			passphrase:   config["passphrase"],
			mnemonic:     config["mnemonic"],
			numAccounts:  count,
		}

	}

	return nil
}

type Network struct {
	name       string
	rpc_client *rpc.Client
	client     *ethclient.Client
	keystore   *keystore.KeyStore
	hdWallet   *hdwallet.Wallet
}

func Dial(name string) (*Network, error) {
	config, ok := networks[name]
	if !ok {
		return nil, errors.New("No such network " + name)
	}

	rpc_client, err := rpc.Dial(config.url)
	if err != nil {
		return nil, err
	}

	client := ethclient.NewClient(rpc_client)
	ret := &Network{
		name:       name,
		rpc_client: rpc_client,
		client:     client,
	}

	if config.keystorePath != "" {
		ret.keystore = ret.initKeystore(config.keystorePath)

	}

	if config.mnemonic != "" {
		ret.hdWallet = ret.initHDWallet()
	}

	if len(ret.Accounts()) == 0 {
		return nil, errors.New("no accounts configured for this network, did you set the keystore path or mnemonic in perigord.yaml")
	}

	return ret, nil

}

func (n *Network) Name() string {
	return n.name
}

func (n *Network) Url() string {
	return networks[n.name].url
}

func (n *Network) Passphrase() string {
	return networks[n.name].passphrase
}

func (n *Network) Mnemonic() string {
	return networks[n.name].mnemonic
}

func (n *Network) NumAccounts() int64 {
	return networks[n.name].numAccounts
}

func (n *Network) KeystorePath() string {
	return networks[n.name].keystorePath
}

func (n *Network) RpcClient() *rpc.Client {
	return n.rpc_client
}

func (n *Network) Client() *ethclient.Client {
	return n.client
}

func (n *Network) Keystore() *keystore.KeyStore {
	return n.keystore
}

func (n *Network) Accounts() []accounts.Account {
	if n.keystore != nil {
		return n.keystore.Accounts()
	}
	if n.hdWallet != nil {
		return n.hdWallet.Accounts()
	}
	return nil
}

func (n *Network) Unlock(a accounts.Account) error {
	if n.keystore != nil {
		return n.keystore.Unlock(a, n.Passphrase())
	}
	return nil
}

func (n *Network) initKeystore(keystorePath string) *keystore.KeyStore {
	return keystore.NewKeyStore(keystorePath, keystore.StandardScryptN, keystore.StandardScryptP)
}

func (n *Network) initHDWallet() *hdwallet.Wallet {
	mnemonic := n.Mnemonic()
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	var i int64
	for i = 0; i < n.NumAccounts(); i++ {
		path := hdwallet.MustParseDerivationPath(fmt.Sprintf("m/44'/60'/0'/0/%d", i))
		account, err := wallet.Derive(path, false)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(account.Address.Hex())

	}

	return wallet
}

func (n *Network) NewTransactor(a accounts.Account) *bind.TransactOpts {
	return &bind.TransactOpts{
		From:  a.Address,
		Nonce: nil,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != a.Address {
				return nil, errors.New("not authorized to sign this account")
			}
			if n.keystore != nil {
				signature, err := n.keystore.SignHash(a, signer.Hash(tx).Bytes())
				if err != nil {
					return nil, err
				}
				return tx.WithSignature(signer, signature)
			}

			signature, err := n.hdWallet.SignHash(a, signer.Hash(tx).Bytes())
			if err != nil {
				return nil, err
			}
			return tx.WithSignature(signer, signature)
		},
		Value:    nil,
		GasPrice: nil,
		GasLimit: 0,
		Context:  nil,
	}
}
