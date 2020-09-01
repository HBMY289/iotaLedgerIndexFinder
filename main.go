// main
package main

import (
	"fmt"
	"time"

	// "github.com/dongri/go-mnemonic"
	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/kerl"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func main() {

	// var bip44AccountIndex uint32 = 0x00000000
	var bip44PageIndex uint32 = 0x00000000

	mnemonic := "witch collapse practice feed shame open despair creek road again ice least witch collapse practice feed shame open despair creek road again ice least"
	seed := ""
	// var addrs = []string{}
	startTime := time.Now()
	for i := 0; i < 100; i++ {
		seed = mnemonicToSeed(mnemonic, uint32(i), bip44PageIndex)
		//addrs = getAddrsOfSeed(seed, 10)
	}
	endTime := time.Now()
	fmt.Println("Seed: ", seed)
	// fmt.Println(addrs)
	fmt.Println(endTime.Sub(startTime))

}

func mnemonicToSeed(mnemonic string, accountIndex, pageIndex uint32) string {
	seed := bip39.NewSeed(mnemonic, "")
	bip32RootKey, _ := bip32.NewMasterKey(seed)
	bip44PuposeKey, _ := bip32RootKey.NewChildKey(0x8000002C)
	bip44CoinTypeKey, _ := bip44PuposeKey.NewChildKey(0x8000107A)
	bip44AccountKey, _ := bip44CoinTypeKey.NewChildKey(0x80000000 + accountIndex)
	bip44PageIndexKey, _ := bip44AccountKey.NewChildKey(0x80000000 + pageIndex)
	privateKey := bip44PageIndexKey.Key
	chainCode := bip44PageIndexKey.ChainCode
	kerlInput := append(privateKey[0:32], chainCode[0:16]...)
	kerlInput = append(kerlInput, privateKey[16:32]...)
	kerlInput = append(kerlInput, chainCode[0:32]...)
	myKerl := kerl.NewKerl()
	myKerl.Hash.Write(kerlInput)
	trytes, _ := myKerl.SqueezeTrytes(243)
	return trytes
}

func getAddrsOfSeed(seed string, addCount int) []string {
	addrs, _ := address.GenerateAddresses(seed, 0, uint64(addCount), 2, false)
	return addrs
}
