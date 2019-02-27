package migrations

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"gitlab.com/go-truffle/enhanced-perigord/contract"
	"gitlab.com/go-truffle/enhanced-perigord/migration"
	"gitlab.com/go-truffle/enhanced-perigord/network"

	"{{.project}}/bindings"
)

type {{.contract}}Deployer struct{}

func (d *{{.contract}}Deployer) Deploy(ctx context.Context, network *network.Network) (common.Address, *types.Transaction, interface{}, error) {
	account := network.Accounts()[0]
	err := network.Unlock(account)
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	auth := network.NewTransactor(account)
	address, transaction, contract, err := bindings.Deploy{{.contract}}(auth, network.Client())
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	session := &bindings.{{.contract}}Session{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: *auth,
	}

	return address, transaction, session, nil
}

func (d *{{.contract}}Deployer) Bind(ctx context.Context, network *network.Network, address common.Address) (interface{}, error) {
	account := network.Accounts()[0]
	err := network.Unlock(account)
	if err != nil {
		return common.Address{}, err
	}

	auth := network.NewTransactor(account)
	contract, err := bindings.New{{.contract}}(address, network.Client())
	if err != nil {
		return nil, err
	}

	session := &bindings.{{.contract}}Session{
		Contract: contract,
		CallOpts: bind.CallOpts{
			Pending: true,
		},
		TransactOpts: *auth,
	}

	return session, nil
}

func init() {
	contract.AddContract("{{.contract}}", &{{.contract}}Deployer{})

	migration.AddMigration(&migration.Migration{
		Number: {{.number}},
		F: func(ctx context.Context, network *network.Network) error {
			if err := contract.Deploy(ctx, "{{.contract}}", network); err != nil {
				return err
			}

			return nil
		},
	})
}
