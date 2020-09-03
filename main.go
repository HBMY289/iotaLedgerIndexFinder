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
)

var wg = sync.WaitGroup{}
var scanner = bufio.NewScanner(os.Stdin)

func main() {

	// mnemonic := getMnemonic()
	mnemonic := "wheel mosquito enroll illness stamp vote tomorrow mandate powder armed fortune buffalo rack mirror elder fun paper between cheap present vast unlock detect birth" //targetAddr := "WNMNYMAOBGWPNFYZHSQHNXMOOIURFWAZUVVCNVZCKNKBH9XOGWUPFRWPSFAHBMMMZKXDJJGIOTERPSEUB" // addindex #3 accountindex #8
	for {
		targetAddr := getTargetAddress()
		// targetAddr := "SZE9WDWHUUYGOXQRMZWKHFHSQCVU9NROSNFERAJMT9YFIHHRCKRFSDESFWDPCLPMJFFXLXZISLWKBSKTC"                                                                               // addindex #8 acc index #98
		// targetAddr := "CRICOFALQY9XBDSPOJAID9TMKMUNYWVN99WEUFOTCNBYZCNALGUCDDMQTHYWZVFMNWBYGBBBDUWKJPAFZ" //addindex #1 accindex 9
		// targetAddr := "JWTWV9KLWZRORTCQGBHEYZFQLZUIGLGJASFDGQOKAVSYIBKOGONQDZZTLM9IYE9GVBTPBSXEWLIDBQYF9" //addindex #2 accountindex 99
		accStart := 0
		accEnd := 1000
		addrsPerSeed := 20
		pageStart := 0
		pageEnd := 0
		startTime := time.Now()
		matchedIndex := getMatchingIndex(mnemonic, targetAddr, addrsPerSeed, accStart, accEnd, pageStart, pageEnd)
		deltaT := time.Now().Sub(startTime)

		if matchedIndex >= 0 {
			fmt.Printf("\nFound matching address for ACCOUNT INDEX '%d' after %v.\n", matchedIndex, deltaT.Round(time.Second))
			break
		}
		fmt.Printf("\nCould not find a match after %v testing the first %d addresses of indexes %d to %d.\nCheck target address or retry with larger values for maximum index and addresses per seed.", deltaT.Round(time.Second), addrsPerSeed, accStart, accEnd)
		if !again() {
			break
		}

	}
	fmt.Println("\nPress Enter to close the program.")
	scanner.Scan()
}

func getMatchingIndex(mnemonic, targetAddr string, addrsPerSeed, accStart, accEnd, pageStart, pageEnd int) int {

	seedChan := make(chan mySeed)
	stopChan := make(chan struct{}, 1)
	workers := runtime.GOMAXPROCS(-1)
	matchedIndex := -1
	wg.Add(1)
	go generateSeeds(mnemonic, seedChan, accStart, accEnd, stopChan)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go checkAddresses(targetAddr, seedChan, addrsPerSeed, &matchedIndex, stopChan)
	}
	fmt.Println("start searching")
	wg.Wait()
	fmt.Println("\nstopped searching")
	return matchedIndex

}

func generateSeeds(mnemonic string, seedChan chan<- mySeed, accStart, accEnd int, stopChan <-chan struct{}) {
	defer close(seedChan)
	defer wg.Done()
	for i := accStart; i <= accEnd; i++ {
		select {
		case <-stopChan:
			return
		default:
		}
		seed := mySeed{mnemonicToSeed(mnemonic, i, 0), i, 0}
		seedChan <- seed
	}
}

func checkAddresses(targetAddr string, seedChan <-chan mySeed, addrsPerSeed int, matchedIndex *int, stopChan chan<- struct{}) {
	for seed := range seedChan {
		addrs := getAddrsOfSeed(seed.string, addrsPerSeed)
		if seed.accountIndex%10 == 0 {
			fmt.Printf("\rchecking index #%d", seed.accountIndex)
		}
		for _, addr := range addrs {
			if strings.Contains(addr, targetAddr) {
				*matchedIndex = int(seed.accountIndex)
				stopChan <- struct{}{}
			}
		}
	}
	wg.Done()
}

func mnemonicToSeed(mnemonic string, accountIndex, pageIndex int) string {
	seed := bip39.NewSeed(mnemonic, "")
	bip32RootKey, _ := bip32.NewMasterKey(seed)
	bip44PuposeKey, _ := bip32RootKey.NewChildKey(0x8000002C)
	bip44CoinTypeKey, _ := bip44PuposeKey.NewChildKey(0x8000107A)
	bip44AccountKey, _ := bip44CoinTypeKey.NewChildKey(0x80000000 + uint32(accountIndex))
	bip44PageIndexKey, _ := bip44AccountKey.NewChildKey(0x80000000 + uint32(pageIndex))
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

func getTargetAddress() string {

	var addr string
	for {
		fmt.Print("Enter target address: ")
		scanner.Scan()
		addr = scanner.Text()
		fmt.Println()
		if err := address.ValidAddress(addr); err == nil {
			break
		}
		fmt.Println("\n\nInvalid address entered.")
	}

	return addr[0:81]
}

func getMnemonic() string {

	var words [24]string
	i := 1
	for i <= 24 {
		fmt.Printf("Enter mnemonic word #%d: ", i)
		scanner.Scan()
		word := scanner.Text()
		fmt.Println()
		if isValidWord(word) {
			words[i-1] = word
			i++

		} else {
			fmt.Printf("'%v' is not a valid BIP 44 word. Try again.\n", word)
		}

	}
	return strings.Join(words[:], " ")
}

func again() bool {

	fmt.Print("\nDo you want to try again using the same 24 words (y/n)?: ")
	scanner.Scan()
	answer := scanner.Text()
	if answer == "y" {
		fmt.Println()
		return true
	}
	return false
}
func isValidWord(word string) bool {
	_, valid := bip39.GetWordIndex(word)
	return valid
}

func getAddrsOfSeed(seed string, addCount int) []string {
	addrs, _ := address.GenerateAddresses(seed, 0, uint64(addCount)-1, 2, false)
	return addrs
}

type mySeed struct {
	string                  string
	accountIndex, pageIndex int
}
