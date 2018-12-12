project: {{.project}}
license: {{.license.Name}}

networks:
    dev:
        url: /tmp/geth_private_testnet/geth.ipc
        keystore: /tmp/geth_private_testnet/keystore
        passphrase: blah
        mnemonic: candy maple cake sugar pudding cream honey rich smooth crumble sweet treat
