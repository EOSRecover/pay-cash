# EOSIO EVM AdminCall Proposal Validator

## Case Background
Pcash (paycash.app) is a DeFi project on EOS that provides stablecoin exchange services. The main smart contract (swap.pcash) of the project was attacked on May 6, 2023. The attacker drained the liquidity pool, swapped the stolen funds into roughly 2 million EOS, and hid them in 6000+ newly created EOSEVM accounts in order to escape from the mainnet governance actions. The attacker soon managed to transfer a small portion of the stolen funds out of the EOS network, but the majority of the funds are still sitting in the 6000+ EOSEVM addresses. The onchain transaction history indicates that this is a clear hacking incident that causes significant losses. Over the past 9 months, we (Recover+ team) closely communicated with the Pcash team founders and its community users. In fact, the Pcash team took their responsibility by covering the loss for its users and maintaining the same DeFi infrastructure. Based on the information collected, we decided to approve this case and initiate governance proposals to freeze and recover the stolen funds; not only save funds for the Pcash team, but also save the Pcash project for its EOS users. Because of the complexity of this case (6000+ EVM hacker addresses

## Background
The `eosio.evm` contract on the EOS blockchain has an `admincall` function that allows EOS block producers (BPs) to initiate proposals to execute this function. This can directly manipulate addresses on the EVM (Ethereum Virtual Machine). The main purpose of this project is to use proposals to execute the `admincall` function on `eosio.evm` to transfer EOS from hacker addresses to the `eos.recover` account. The corresponding EVM address for `eos.recover` is `0xbbbbbbbbbbbbbbbbbbbbbbbb55300ba914daae00`.

Seizing funds from an account is done through the `admincall` action. More details can be found in the [eosio.evm contract documentation](https://github.com/eosnetworkfoundation/eos-evm-contract/blob/d6ea964cf256a4bfbd849bddf4c757ceaaafd4d4/include/evm_runtime/evm_contract.hpp#L89).

## Main Features
The project needs to validate two key points:
- **Address Validation**: All EVM addresses in `account.csv` should belong to the hacker. The known EOS account of the hacker is `nrwgthbeupex`. The format of `account.csv` is as follows:
```
"tx_id","address","quantity","token" 
"d81c20796d8f5ba167de355bbffef22fdb20f72e18e7ab09b4c1afbfb816e151","0xb81340266E1781750411240E555fc78D033a42ba","256","EOS" 
"d1e170f54c4a87926231a1ff90d5f007c565fe923e56888caada7de1fb511a8a","0x2dD6c369d6A07Cde69712EDCb01f669d52872127","393","EOS" 
"71284ad2ad3e53a41831410d2caeb3a055e6bde1faa8c17d436b7c2bf5d099eb","0x5386E3402aa6Ff6f8998E731292448cCe06eD866","304","EOS"
```
The first column `tx_id` is the transaction ID on the EOS blockchain. The validation process involves querying the `tx_id` using dfuse's API to confirm that the transaction action is a `transfer` operation, with `from` being the hacker's address and `to` being `eosio.evm`.

- **Proposal Validation**: This is to prevent the proposal from containing other malicious proposals. Due to the need to operate on more than 6000+ addresses, a single proposal cannot complete the operation. Therefore, secondary proposals are sent, and the proposals are for BPs to execute the `approve` operation on secondary proposals. The secondary proposals need to read all actions of the proposal to ensure that they are operating the `admincall` action of `eosio.evm`, with the `from` address being the EVM address in `account.csv` and the `to` address being the EVM address of `eos.recover`.

- **Quantity Validation**: In addition to address validation, the project also needs to ensure that the quantity of EOS being transferred in the proposals is correct. Since a portion of EOS is reserved for paying EVM gas fees, the quantity transferred should be slightly less than the total balance in `account.csv`. The following code snippet demonstrates this validation:

    ```go
    // Convert the transfer quantity from byte slice to big.Int
    value := new(big.Int).SetBytes(action.Value)

    // Base precision: 1x10^17
    precision := new(big.Int).SetString("100000000000000000", 10)

    // Calculate the real value by dividing by precision
    value.Div(value, precision)

    // Convert the balance from `account.csv` to an integer
    balance, err := strconv.ParseInt(addr.Quantity, 10, 64)
    if err != nil {
        fmt.Println(addr.Address, "balance error:", err)
        return false
    }

    // Adjust the balance to match the precision
    balance *= 10

    // Calculate the transfer quantity as an int64
    quantity := value.Int64()

    // Reserve 0.1 EOS for EVM gas fees and validate the quantity
    if balance > quantity && balance-quantity > 1 {
  
        fmt.Println(addr.Address, "incorrect transfer quantity")
        return false
    }
    ```

-----

#### ENF Blog
https://eosnetwork.com/blog/advancing-blockchain-defense

#### Paycash Open Letter
##### Website
https://paycashswap.com/open-letters

##### Facebook
https://www.facebook.com/share/p/szeQCHMj1Ssq5cuP/?mibextid=xfxF2i

##### Insta
https://www.instagram.com/p/C3zb-zgMz0h/?igsh=Y2htbGRiYW8zbmRu

##### Twitter
https://twitter.com/PayCashTweet/status/1762032720501203036?t=Gktov0UOVj_dp0bfY4JDCQ&s=19

