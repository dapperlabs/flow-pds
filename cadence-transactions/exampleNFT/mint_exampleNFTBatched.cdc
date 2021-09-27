import NonFungibleToken from 0x{{.NonFungibleToken}}
import ExampleNFT from 0x{{.ExampleNFT}}

transaction(recipient: Address, batchSize: Int) {

    let minter: &ExampleNFT.NFTMinter

    prepare(signer: AuthAccount) {
        self.minter = signer
            .borrow<&ExampleNFT.NFTMinter>(from: ExampleNFT.MinterStoragePath)!
    }

    execute {
        let receiver = getAccount(recipient)
            .getCapability(ExampleNFT.CollectionPublicPath)!
            .borrow<&{NonFungibleToken.CollectionPublic}>()!

        self.minter.mintNFTBatched(recipient: receiver, batchSize: batchSize)
    }
}
