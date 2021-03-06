# iotaLedgerIndexFinder

I mainly wrote this tool to learn about programming in general and golang in specific. I am also using it as an exercise to become acquainted with git and github. However, I am sure there are a few of people that might actually profit from using this tool.  

## What this tool does
When using a Ledger hardware wallet to secure your Iota tokens the actual Iota seed is calculated based on the 24-word recovery phrase and an index value you chose when you set up the account in the Trinity wallet. Every index value will generate a completely independent seed and only the correct value will give access to your funds. This tool enables Iota token holders to gain access to their funds again if they do no remember the account index anymore.

## How it works
The program requires you to enter the 24 word secret passphrase and any address that belongs to the seed of interest. It then sequentially calculates the seed for each account index and generates addresses for that seed. If any of these addresses matches the one given by you the correct index is found and reported. It is not necessary to have any funds on this address and it is not necessary to connect to any tangle node or even to be online at all. Once the correct index is found you can create a new account in Trinity and will have access to your Iota tokens again.

# Warning
You should never share your seed or your 24-word recovery phrase with anyone. If anybody asks for this kind of information it is definitely a scam no matter who they claim to be. You should also never enter this sensitive information in any software that you do not trust. However, since you probably do not have much reason to trust me as well I recommend reducing your risk by running this program on an air-gapped/offline computer. This way no malicious party including myself could get hold of your secret words. Once the correct index is found you should use another device to access and transfer your funds to a new seed. After that you should completely reset your Ledger device to generate a new 24-word recovery phrase (be careful if you also use the device to store other tokens).

## How to start the tool
The simplest way is to download the appropriate binary executable for your operating system from [releases](https://github.com/HBMY289/iotaLedgerIndexFinder/releases) and then start it. You can also build the tool from source, which is rather easy as well. Assuming you have [go](https://golang.org/doc/install) and [git](https://www.atlassian.com/git/tutorials/install-git) installed already you can just execute this command for example in your user folder to get a copy of the source code.
```
git clone https://github.com/HBMY289/iotaLedgerIndexFinder.git
```

Then you change into the new folder and build the excutable.
```
cd iotaLedgerIndexFinder
go build
```
After that you can just start the newly created binary file by typing
```
./iotaLedgerIndexFinder
```
or on Windows
```
iotaLedgerIndexFinder.exe
```


## How to use the tool
Once the program is running you will have to enter the required information to find your account index.

##### Mnemonic
Enter your 24 recovery words that are required to calculate the seeds. The words have to be entered one-by-one and are automatically checked against the BIP39 word list, so no typos will happen.

##### Target address
Enter any address that was generated from the seed you are looking for. It can be an old address that you might find in your exchange's withdrawal history or a current one that you wrote down somewhere. It is OK if there is no balance on this address.

##### Addresses per seed
Enter the number of addresses that should be generated for each calculated seed. You can press Enter to use the default of 20. If you used your seed a lot for sending or generated a lot of different receive address you might want to increase this value to make sure your entered target address will be found. Generating addresses is a time consuming step, so entering a high number here will result in a lower number of account indexes that can be tested per second.

##### Account index search range
*Account index start* und *Account index end* define the search range that will be covered. You can hit *Enter* for both to use the respective default values of 0 and 1000. If you for example remember that you definetly used a 3 digit number you can change the start index to 100. If you might have used a very large index number you can increase the maximum value accordingly.

##### The game is on
The program now starts the seed and address generation and reports the current status. Once a match is found it will automatically stop and report the found account index. Depending on the hardware you use the tool can check 1000 account indexes per minute or even more. 

## What to do if no address of the seed is known?
It would be possible to check all generated addresses against the tangle and check for balance, but you would need to be online for that. To avoid this and enable running the tool offline I added a special option for this case that matches addresses against a tangle snapshot file with all addresses and their current balances.
You will need to ask someone who runs a Iota node to run this command
```
curl -H 'X-IOTA-API-VERSION: 1' -d '{"command":"getLedgerState"}' localhost:14265 >  snapshot.txt
```
and send you the resulting file "snapshot.txt".
Place this file in the same folder as your IotaLedgerIndexFinder executable and start the program with the snapshot option "-s" like this:
```
./iotaLedgerIndexFinder -s
```
or on Windows
```
iotaLedgerIndexFinder.exe -s
```



## Disclaimer
Thanks go ot to [Tyler Smith](https://github.com/tyler-smith) for providing the used cryptography libraries and to [muXxer](https://github.com/muXxer) as I learned from one of his Python projects how to actually [calculate the seed from the recovery phrase](https://github.com/muXxer/recover-iota-seed-from-ledger-mnemonics).

If you need any help with this tool or require specific customizations you can contact me (HBMY289) via the official [Iota Discord server](https://discord.iota.org/).



