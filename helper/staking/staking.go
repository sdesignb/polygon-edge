package staking

import (
	"fmt"
	"math/big"

	"github.com/sdesignb/polygon-edge/chain"
	"github.com/sdesignb/polygon-edge/helper/hex"
	"github.com/sdesignb/polygon-edge/helper/keccak"
	"github.com/sdesignb/polygon-edge/types"
)

// PadLeftOrTrim left-pads the passed in byte array to the specified size,
// or trims the array if it exceeds the passed in size
func PadLeftOrTrim(bb []byte, size int) []byte {
	l := len(bb)
	if l == size {
		return bb
	}

	if l > size {
		return bb[l-size:]
	}

	tmp := make([]byte, size)
	copy(tmp[size-l:], bb)

	return tmp
}

// getAddressMapping returns the key for the SC storage mapping (address => something)
//
// More information:
// https://docs.soliditylang.org/en/latest/internals/layout_in_storage.html
func getAddressMapping(address types.Address, slot int64) []byte {
	bigSlot := big.NewInt(slot)

	finalSlice := append(
		PadLeftOrTrim(address.Bytes(), 32),
		PadLeftOrTrim(bigSlot.Bytes(), 32)...,
	)
	keccakValue := keccak.Keccak256(nil, finalSlice)

	return keccakValue
}

// getIndexWithOffset is a helper method for adding an offset to the already found keccak hash
func getIndexWithOffset(keccakHash []byte, offset int64) []byte {
	bigOffset := big.NewInt(offset)
	bigKeccak := big.NewInt(0).SetBytes(keccakHash)

	bigKeccak.Add(bigKeccak, bigOffset)

	return bigKeccak.Bytes()
}

// getStorageIndexes is a helper function for getting the correct indexes
// of the storage slots which need to be modified during bootstrap.
//
// It is SC dependant, and based on the SC located at:
// https://github.com/0xPolygon/staking-contracts/
func getStorageIndexes(address types.Address, index int64) *StorageIndexes {
	storageIndexes := StorageIndexes{}

	// Get the indexes for the mappings
	// The index for the mapping is retrieved with:
	// keccak(address . slot)
	// . stands for concatenation (basically appending the bytes)
	storageIndexes.AddressToIsValidatorIndex = getAddressMapping(address, addressToIsValidatorSlot)
	storageIndexes.AddressToStakedAmountIndex = getAddressMapping(address, addressToStakedAmountSlot)
	storageIndexes.AddressToValidatorIndexIndex = getAddressMapping(address, addressToValidatorIndexSlot)

	// Get the indexes for _validators, _stakedAmount
	// Index for regular types is calculated as just the regular slot
	storageIndexes.StakedAmountIndex = big.NewInt(stakedAmountSlot).Bytes()

	// Index for array types is calculated as keccak(slot) + index
	// The slot for the dynamic arrays that's put in the keccak needs to be in hex form (padded 64 chars)
	storageIndexes.ValidatorsIndex = getIndexWithOffset(
		keccak.Keccak256(nil, PadLeftOrTrim(big.NewInt(validatorsSlot).Bytes(), 32)),
		index,
	)

	// For any dynamic array in Solidity, the size of the actual array should be
	// located on slot x
	storageIndexes.ValidatorsArraySizeIndex = []byte{byte(validatorsSlot)}

	return &storageIndexes
}

// StorageIndexes is a wrapper for different storage indexes that
// need to be modified
type StorageIndexes struct {
	ValidatorsIndex              []byte // []address
	ValidatorsArraySizeIndex     []byte // []address size
	AddressToIsValidatorIndex    []byte // mapping(address => bool)
	AddressToStakedAmountIndex   []byte // mapping(address => uint256)
	AddressToValidatorIndexIndex []byte // mapping(address => uint256)
	StakedAmountIndex            []byte // uint256
}

// Slot definitions for SC storage
var (
	validatorsSlot              = int64(0) // Slot 0
	addressToIsValidatorSlot    = int64(1) // Slot 1
	addressToStakedAmountSlot   = int64(2) // Slot 2
	addressToValidatorIndexSlot = int64(3) // Slot 3
	stakedAmountSlot            = int64(4) // Slot 4
)

const (
	DefaultStakedBalance = "0x1A784379D99DB42000000" // 10 ETH
	//nolint: lll
	StakingSCBytecode = "0x6080604052600436106100745760003560e01c8063ca1e78191161004e578063ca1e781914610121578063f2888dbb1461014c578063f90ecacc14610175578063facd743b146101b257610084565b80632367f6b51461008957806326476204146100c6578063373d6132146100f657610084565b3661008457610082336101ef565b005b600080fd5b34801561009557600080fd5b506100b060048036038101906100ab9190610d3c565b6104ec565b6040516100bd9190610fac565b60405180910390f35b6100e060048036038101906100db9190610d3c565b610535565b6040516100ed9190610f11565b60405180910390f35b34801561010257600080fd5b5061010b6105a9565b6040516101189190610fac565b60405180910390f35b34801561012d57600080fd5b506101366105b3565b6040516101439190610eef565b60405180910390f35b34801561015857600080fd5b50610173600480360381019061016e9190610d3c565b610641565b005b34801561018157600080fd5b5061019c60048036038101906101979190610d69565b61071c565b6040516101a99190610ed4565b60405180910390f35b3480156101be57600080fd5b506101d960048036038101906101d49190610d3c565b61075b565b6040516101e69190610f11565b60405180910390f35b34600460008282546102019190611011565b9250508190555034600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546102579190611011565b92505081905550600160008273ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff1615801561031457506a01a784379d99db420000006fffffffffffffffffffffffffffffffff16600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410155b1561049b5760018060008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff021916908315150217905550600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000208190555033600560008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506000819080600181540180825580915050600190039060005260206000200160009091909190916101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505b8073ffffffffffffffffffffffffffffffffffffffff167f9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d346040516104e19190610fac565b60405180910390a250565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b60006a01a784379d99db420000006fffffffffffffffffffffffffffffffff16341015610597576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161058e90610f8c565b60405180910390fd5b6105a0826101ef565b60019050919050565b6000600454905090565b6060600080548060200260200160405190810160405280929190818152602001828054801561063757602002820191906000526020600020905b8160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190600101908083116105ed575b5050505050905090565b803373ffffffffffffffffffffffffffffffffffffffff16600560008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff161461070f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161070690610f4c565b60405180910390fd5b610718826107b1565b5050565b6000818154811061072c57600080fd5b906000526020600020016000915054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b6000600160008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff169050919050565b600463ffffffff16600080549050116107ff576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107f690610f2c565b60405180910390fd5b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205490506000600560008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050600160008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060009054906101000a900460ff16156109035761090283610a5d565b5b81600460008282546109159190611067565b925050819055506000600260008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550600560008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690558273ffffffffffffffffffffffffffffffffffffffff167f0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f7583604051610a099190610fac565b60405180910390a28073ffffffffffffffffffffffffffffffffffffffff166108fc839081150290604051600060405180830381858888f19350505050158015610a57573d6000803e3d6000fd5b50505050565b600080549050600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205410610ae3576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610ada90610f6c565b60405180910390fd5b6000600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054905060006001600080549050610b3b9190611067565b9050808214610c29576000808281548110610b5957610b58611141565b5b9060005260206000200160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1690508060008481548110610b9b57610b9a611141565b5b9060005260206000200160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555082600360008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002081905550505b6000600160008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055506000600360008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506000805480610cd857610cd7611112565b5b6001900381819060005260206000200160006101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690559055505050565b600081359050610d218161128b565b92915050565b600081359050610d36816112a2565b92915050565b600060208284031215610d5257610d51611170565b5b6000610d6084828501610d12565b91505092915050565b600060208284031215610d7f57610d7e611170565b5b6000610d8d84828501610d27565b91505092915050565b6000610da28383610dae565b60208301905092915050565b610db78161109b565b82525050565b610dc68161109b565b82525050565b6000610dd782610fd7565b610de18185610fef565b9350610dec83610fc7565b8060005b83811015610e1d578151610e048882610d96565b9750610e0f83610fe2565b925050600181019050610df0565b5085935050505092915050565b610e33816110ad565b82525050565b6000610e46604783611000565b9150610e5182611175565b606082019050919050565b6000610e69601d83611000565b9150610e74826111ea565b602082019050919050565b6000610e8c601283611000565b9150610e9782611213565b602082019050919050565b6000610eaf602683611000565b9150610eba8261123c565b604082019050919050565b610ece816110d9565b82525050565b6000602082019050610ee96000830184610dbd565b92915050565b60006020820190508181036000830152610f098184610dcc565b905092915050565b6000602082019050610f266000830184610e2a565b92915050565b60006020820190508181036000830152610f4581610e39565b9050919050565b60006020820190508181036000830152610f6581610e5c565b9050919050565b60006020820190508181036000830152610f8581610e7f565b9050919050565b60006020820190508181036000830152610fa581610ea2565b9050919050565b6000602082019050610fc16000830184610ec5565b92915050565b6000819050602082019050919050565b600081519050919050565b6000602082019050919050565b600082825260208201905092915050565b600082825260208201905092915050565b600061101c826110d9565b9150611027836110d9565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561105c5761105b6110e3565b5b828201905092915050565b6000611072826110d9565b915061107d836110d9565b9250828210156110905761108f6110e3565b5b828203905092915050565b60006110a6826110b9565b9050919050565b60008115159050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600080fd5b7f4e756d626572206f662076616c696461746f72732063616e2774206265206c6560008201527f7373207468616e204d494e494d554d5f52455155495245445f4e554d5f56414c60208201527f494441544f525300000000000000000000000000000000000000000000000000604082015250565b7f4f6e6c79207374616b65722063616e2063616c6c2066756e6374696f6e000000600082015250565b7f696e646578206f7574206f662072616e67650000000000000000000000000000600082015250565b7f596f75206e656564206d6f72652066756e647320696e206f7264657220746f2060008201527f7374616b65210000000000000000000000000000000000000000000000000000602082015250565b6112948161109b565b811461129f57600080fd5b50565b6112ab816110d9565b81146112b657600080fd5b5056fea2646970667358221220b65faaebb9b1c92b6d8cadc98da4c86a50d3a97d7807a896f5b83e0f2926b08d64736f6c63430008070033"
)

// PredeployStakingSC is a helper method for setting up the staking smart contract account,
// using the passed in validators as prestaked validators
func PredeployStakingSC(
	validators []types.Address,
) (*chain.GenesisAccount, error) {
	// Set the code for the staking smart contract
	// Code retrieved from https://github.com/0xPolygon/staking-contracts
	scHex, _ := hex.DecodeHex(StakingSCBytecode)
	stakingAccount := &chain.GenesisAccount{
		Code: scHex,
	}

	// Parse the default staked balance value into *big.Int
	val := DefaultStakedBalance
	bigDefaultStakedBalance, err := types.ParseUint256orHex(&val)

	if err != nil {
		return nil, fmt.Errorf("unable to generate DefaultStatkedBalance, %w", err)
	}

	// Generate the empty account storage map
	storageMap := make(map[types.Hash]types.Hash)
	bigTrueValue := big.NewInt(1)
	stakedAmount := big.NewInt(0)

	for indx, validator := range validators {
		// Update the total staked amount
		stakedAmount.Add(stakedAmount, bigDefaultStakedBalance)

		// Get the storage indexes
		storageIndexes := getStorageIndexes(validator, int64(indx))

		// Set the value for the validators array
		storageMap[types.BytesToHash(storageIndexes.ValidatorsIndex)] =
			types.BytesToHash(
				validator.Bytes(),
			)

		// Set the value for the address -> validator array index mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToIsValidatorIndex)] =
			types.BytesToHash(bigTrueValue.Bytes())

		// Set the value for the address -> staked amount mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToStakedAmountIndex)] =
			types.StringToHash(hex.EncodeBig(bigDefaultStakedBalance))

		// Set the value for the address -> validator index mapping
		storageMap[types.BytesToHash(storageIndexes.AddressToValidatorIndexIndex)] =
			types.StringToHash(hex.EncodeUint64(uint64(indx)))

		// Set the value for the total staked amount
		storageMap[types.BytesToHash(storageIndexes.StakedAmountIndex)] =
			types.BytesToHash(stakedAmount.Bytes())

		// Set the value for the size of the validators array
		storageMap[types.BytesToHash(storageIndexes.ValidatorsArraySizeIndex)] =
			types.StringToHash(hex.EncodeUint64(uint64(indx + 1)))
	}

	// Save the storage map
	stakingAccount.Storage = storageMap

	// Set the Staking SC balance to numValidators * defaultStakedBalance
	stakingAccount.Balance = stakedAmount

	return stakingAccount, nil
}
