import NonFungibleToken from 0x{{.NonFungibleToken}}
import ExampleNFT from 0x{{.ExampleNFT}}

transaction (NFTProviderPath: PrivatePath) {
    prepare(signer: AuthAccount) {
        // Return early if the account already has a collection
        if signer.borrow<&ExampleNFT.Collection>(from: ExampleNFT.CollectionStoragePath) != nil {
            return
        }

        // create a new empty collection
        let collection <- ExampleNFT.createEmptyCollection()

        // save it to the account
        signer.save(<-collection, to: ExampleNFT.CollectionStoragePath)

        // create a public capability for the collection
        signer.link<&NonFungibleToken.Collection{NonFungibleToken.CollectionPublic}>(ExampleNFT.CollectionPublicPath, target: ExampleNFT.CollectionStoragePath)
        assert(signer.getCapability<&{NonFungibleToken.CollectionPublic}>(ExampleNFT.CollectionPublicPath).check(), message: "did not link pub cap");

         // This needs to be used to allow for PDS to withdraw
        signer.link<&{NonFungibleToken.Provider}>( NFTProviderPath, target: ExampleNFT.CollectionStoragePath)
        assert(signer.getCapability<&{NonFungibleToken.Provider}>(NFTProviderPath).check(), message: "did not link provider cap");
    }
}
