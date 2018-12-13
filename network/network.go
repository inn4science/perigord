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
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/jbenet/go-base58"
	"github.com/tyler-smith/go-bip32"
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
	"github.com/spf13/viper"

	"github.com/miguelmota/go-ethereum-hdwallet"
	"gitlab.inn4science.com/gophers/perigord/project"
)

type NetworkConfig struct {
	url           string
	keystore_path string
	passphrase    string
	mnemonic      string
	num_accounts  string
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
		networks[key] = NetworkConfig{
			url:           config["url"],
			keystore_path: config["keystore"],
			passphrase:    config["passphrase"],
			mnemonic:      config["mnemonic"],
			num_accounts:  config["num_accounts"],
		}
	}

	return nil
}

type Network struct {
	name       string
	rpc_client *rpc.Client
	client     *ethclient.Client
	keystore   *keystore.KeyStore
}

func Dial(name string) (*Network, error) {
	if config, ok := networks[name]; ok {
		rpc_client, err := rpc.Dial(config.url)
		if err != nil {
			return nil, err
		}
		var ks *keystore.KeyStore
		client := ethclient.NewClient(rpc_client)
		if config.keystore_path != "" {
			ks = keystore.NewKeyStore(config.keystore_path, keystore.StandardScryptN, keystore.StandardScryptP)
			if len(ks.Accounts()) == 0 {
				return nil, errors.New("No accounts configured for this network, did you set the keystore path in perigord.yaml?")
			}
		}
		if config.keystore_path == "" {
			ks = nil
		}

		ret := &Network{
			name:       name,
			rpc_client: rpc_client,
			client:     client,
			keystore:   ks,
		}

		return ret, nil
	}

	return nil, errors.New("No such network " + name)
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

func (n *Network) NumAccounts() int {
	accountsNumber, err := strconv.Atoi(networks[n.name].num_accounts)
	if err != nil {
		log.Fatal(err)
	}
	return accountsNumber
}

func (n *Network) KeystorePath() string {
	return networks[n.name].keystore_path
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
	return n.GenerateWalletsFromMnemonic(n.NumAccounts())
}

func (n *Network) Unlock(a accounts.Account, passphrase string) error {
	if n.keystore != nil {
		return n.keystore.Unlock(a, passphrase)
	}
	//log.Fatal("CANNOT UNLOCK NIL")
	return nil
}

func (n *Network) GenerateWalletsFromMnemonic(amount int) []accounts.Account {
	accounts := []accounts.Account{}
	mnemonic := n.Mnemonic()
	//fmt.Print(mnemonic)
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < amount; i++ {
		path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/" + strconv.Itoa(i))
		account, err := wallet.Derive(path, false)
		if err != nil {
			log.Fatal(err)
		}
		accounts = append(accounts, account)
		//use to get hex representation of an address
		//fmt.Println(account.Address.Hex())
	}
	return accounts
}

func (n *Network) UnlockWithPrompt(a accounts.Account) error {
	if n.keystore != nil {
		passphrase := n.Passphrase()
		return n.Unlock(a, string(passphrase))
	}
	mnemonic := n.Mnemonic()
	return n.Unlock(a, string(mnemonic))
}

func (n *Network) NewTransactor(a accounts.Account) *bind.TransactOpts {
	return &bind.TransactOpts{
		From:  a.Address,
		Nonce: nil,
		Signer: func(signer types.Signer, address common.Address, tx *types.Transaction) (*types.Transaction, error) {
			if address != a.Address {
				return nil, errors.New("not authorized to sign this account")
			}
			fmt.Println(a)
			if n.keystore != nil {
				signature, err := n.keystore.SignHash(a, signer.Hash(tx).Bytes())
				if err != nil {
					return nil, err
				}
				return tx.WithSignature(signer, signature)
			}


			//seed, err := bip32.NewSeed()
			//if err != nil {
			//	log.Fatal(err)
			//}

			master, err := bip32.NewMasterKey([]byte(n.Mnemonic()))
			if err != nil {
				log.Fatal(err)
			}

			// m/44'
			key, err := master.NewChildKey(2147483648 + 44)
			if err != nil {
				log.Fatal(err)
			}

			decoded := base58.Decode(key.B58Serialize())
			privateKey := decoded[46:78]
			fmt.Println(hexutil.Encode(privateKey)) // 0x801f14cc6b5f2b0785916685c838c8e64f7f4529a9ca7507c90e5f9078cefc07

			// Hex private key to ECDSA private key
			privateKeyECDSA, err := crypto.ToECDSA(privateKey)
			if err != nil {
				log.Fatal(err)
			}

			// ECDSA private key to hex private key
			privateKey = crypto.FromECDSA(privateKeyECDSA)
			//fmt.Println(hexutil.Encode(privateKey)) // 0x801f14cc6b5f2b0785916685c838c8e64f7f4529a9ca7507c90e5f9078cefc07

			//transaction := types.NewTransaction(nonce, recipient, value, gasLimit, gasPrice, input)
			//signature, _ := crypto.Sign(transaction.SigHash().Bytes(), key)


			sig, _ := crypto.Sign(signer.Hash(tx).Bytes(), privateKeyECDSA)
			//signed, _ := tx.WithSignature(signer, signature)

			return tx.WithSignature(signer, sig)
		},
		Value:    nil,
		GasPrice: nil,
		GasLimit: 0,
		Context:  nil,
	}
}
