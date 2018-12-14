# NOT ACTIVELY MAINTAINED

**The PolySwarm team is currently unable to actively maintain Perigord at the moment.**

The PolySwarm team's priority is to faithfully deliver on our roadmap (https://polyswarm.io/road_map/index.html), which regrettably demands we turn our attention away from Perigord in the short term.
The Ethereum smart contract development space is simply moving too quickly right now for us to support a Truffle alternative in addition to hitting our development milestones.

**We plan to return to active Perigord maintenance in the future and will continue to welcome external contributions until we're able to resume work.**

Thank you for your understanding!

~Paul


# Perigord: Golang Tools for Ethereum Development

*Note:* Perigord is in development and it's API is subject to change.

![Perigord Image (not covered by LICENSE)](https://polyswarm.io/img/perigord-logo-small.jpg)

This image is not covered by LICENSE. 

## Install

There is a Dockerfile in `docker/Dockerfile` to build a perigord image, to build
run

```
$ pushd docker
$ docker build -t perigord .
$ popd
```

These instructions assume an Ubuntu 16.04 x86\_64 environment.

### Prerequisite: Golang 1.8

Some dependencies require Go 1.7+, but Go 1.6 is in Ubuntu 16.04's default repos.
The below will install Go 1.8.


```
$ sudo add-apt-repository -y ppa:longsleep/golang-backports
$ sudo apt-get update
$ sudo apt-get install -y golang-go
$ mkdir $HOME/golang
$ echo "export GOPATH=$HOME/golang" >> ~/.bashrc
$ echo "export PATH=$PATH:$HOME/golang/bin" >> ~/.bashrc
```

Close / re-open your terminal or re-`source` your `.bashrc`.

### Prerequisite: `solc`

```
$ sudo add-apt-repository -y ppa:ethereum/ethereum
$ sudo apt-get update
$ sudo apt-get install -y solc
```

### Prerequisite: `abigen`

```
$ go get github.com/ethereum/go-ethereum
$ pushd $GOPATH/src/github.com/ethereum/go-ethereum
$ go install ./cmd/abigen
$ popd
```

## Setup

```
$ go get -u gitlab.inn4science.com/gophers/perigord/...
```

### Dev Dependency: `go-bindata`

```
$ go get -u github.com/jteeuwen/go-bindata/...
```

## Usage

Run for usage information:

```
$ perigord
```

## Tutorial

[Refer to our introductory blog post for now.](https://medium.com/@swarmmarket/introducing-perigord-golang-tools-for-ethereum-dapp-development-60556c2d9fd)


- Navigate to the directory where you want to start a new project. Open a terminal in this directory and type

````
$ perigord init projectName
````

Where projectName is the name of the project you want to initialize

- Check that all perigord imports are using the following path

````
gitlab.inn4science.com/gophers/perigord/{...}

````

- By default, config.yaml file has the following structure

````
networks:
    dev:
        url: /tmp/geth_private_testnet/geth.ipc
        keystore: /tmp/geth_private_testnet/keystore
        passphrase: blah
        mnemonic: candy maple cake sugar pudding cream honey rich smooth crumble sweet treat
        num_accounts: 10
````

You will ether use testnet with 5 pre-generated accounts or use truffle and an arbitrary number of accounts.
To use testnet, leave url, keystore and passphrase and comment out other lines. Run launch_geth_testnet.sh from scripts directory.

To use truffle, run ganache, comment out keystore and passphrase. url should point to the port where ganache is running. For example:

````
 url: HTTP://127.0.0.1:7545
````

Mnemonic is the standard mnemonic phrase used in ganache by default. Number of accounts for these tests is indicated in num_accounts. Pay attention that if you want to make transactions with these accounts, you should indicate in ganache settings that they have certain amount of ETH.

- Add your contract(s) to the project with

````
$ perigord add contract NewContract
````

Omit .sol, it will be added automatically.

- Generate EVM bytecode, ABI and Go bindings with

````
$ perigord build
````
- Add migrations for all added contracts

````
$ perigord add migration NewContract
````

If token contract has a constructor that takes parameters, add them to Deploy{NewContract}Deployer function in the migrations file

````
address, transaction, contract, err := bindings.Deploy{NewContract}(auth, backend, big.NewInt(1337), "FOO", "BAR")
````

- Add tests for all the contracts you want to test

````
$ perigord add test NewContract

````

The simplest test case looks like this

````
func (s *FooSuite) TestName(c *C) {
	session := contract.Session("Foo")
	c.Assert(session, NotNil)
	token_session, ok := session.(*bindings.FooSession)
	c.Assert(ok, Equals, true)
	c.Assert(token_session, NotNil)
	ret, _ := token_session.Bar()
	c.Assert(ret.Cmp(big.NewInt(1337)), Equals, 0)
}
````


