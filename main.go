// main
package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/iotaledger/iota.go/address"
	"github.com/iotaledger/iota.go/kerl"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"github.com/tyler-smith/go-bip44"
)

var wg = sync.WaitGroup{}

func main() {
	getInputs()
	mnemonic := "witch collapse practice feed shame open despair creek road again ice least witch collapse practice feed shame open despair creek road again ice least"
	//targetAddr := "WNMNYMAOBGWPNFYZHSQHNXMOOIURFWAZUVVCNVZCKNKBH9XOGWUPFRWPSFAHBMMMZKXDJJGIOTERPSEUB" // addindex #3 accountindex #8
	targetAddr := "SZE9WDWHUUYGOXQRMZWKHFHSQCVU9NROSNFERAJMT9YFIHHRCKRFSDESFWDPCLPMJFFXLXZISLWKBSKTC" // addindex #8 acc index #98
	startTime := time.Now()
	startSearch(mnemonic, targetAddr, 50, 0, 100, 0, 0)
	endTime := time.Now()
	fmt.Println(endTime.Sub(startTime))
	// seed := ""
	// var addrs = []string{}
	// startTime := time.Now()
	// for i := 0; i < 200; i++ {
	// 	seed = mnemonicToSeed(mnemonic, uint32(i), pageIndex)
	// 	addrs = getAddrsOfSeed(seed, 10)
	// }
	// endTime := time.Now()
	// fmt.Println("Seed: ", seed)
	// fmt.Println(addrs)
	// fmt.Println(endTime.Sub(startTime))

}

func startSearch(mnemonic, targetAddr string, addrsPerSeed int, accIndexStart, accIndexEnd, pageIndexStart, pageIndexEnd uint32) {

	seedChan := make(chan mySeed, 100)

	workers := runtime.GOMAXPROCS(-1)
	wg.Add(1)
	go generateSeeds(mnemonic, seedChan, accIndexStart, accIndexEnd)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go checkAddresses(targetAddr, seedChan, addrsPerSeed)
	}
	go checkAddresses(targetAddr, seedChan, addrsPerSeed)
	fmt.Println("start searching")
	wg.Wait()
	fmt.Println("finished searching")
}

func generateSeeds(mnemonic string, seedChan chan mySeed, accIndexStart, accIndexEnd uint32) {
	for i := accIndexStart; i <= accIndexEnd; i++ {
		seed := mySeed{mnemonicToSeed(mnemonic, i, 0), i, 0}
		seedChan <- seed
		fmt.Println(len(seedChan))
	}
	close(seedChan)
	wg.Done()
}

func checkAddresses(targetAddr string, seedChan chan mySeed, addrsPerSeed int) {
	for seed := range seedChan {
		addrs := getAddrsOfSeed(seed.string, addrsPerSeed)
		// fmt.Println(addrs[8])
		for _, addr := range addrs {
			if strings.Contains(addr, targetAddr) {
				fmt.Printf("found match for account: %d\n", seed.accountIndex)
			}
		}
	}
	wg.Done()
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

func getInputs() string {
	scanner := bufio.NewScanner(os.Stdin)
	mnemonic := getMnemonic(scanner)
	return mnemonic
}

func getMnemonic(scanner *bufio.Scanner) string {
	var words [23]string
	i := 1
	for i <= 24 {
		fmt.Printf("Enter mnemonic word #%d:", i)
		scanner.Scan()
		word := scanner.Text()
		fmt.Println()
		if isValidWord(word) {
			words[i] = word
			i++

		} else {
			fmt.Printf("'%v' is not a valid BIP 44 word. Try again.\n", word)
		}

	}
	return strings.Join(words[:], " ")
}

func isValidWord(word string) bool {
	_, valid := bip39.GetWordIndex(word)
	return valid
}

func getAddrsOfSeed(seed string, addCount int) []string {
	addrs, _ := address.GenerateAddresses(seed, 0, uint64(addCount), 2, false)
	return addrs
}

type mySeed struct {
	string                  string
	accountIndex, pageIndex uint32
}
