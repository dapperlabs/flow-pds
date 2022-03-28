import PackNFT from 0x{{.PackNFT}}
import NonFungibleToken from 0x{{.NonFungibleToken}}

transaction() {
    prepare (issuer: AuthAccount) {
        
         // Only initialize account if doesn't have a PackNFT.Collection resource
        if issuer.borrow<&PackNFT.Collection>(from: PackNFT.CollectionStoragePath) == nil {
            issuer.save(<-PackNFT.createEmptyCollection(), to: PackNFT.CollectionStoragePath);
            issuer.link<&{NonFungibleToken.CollectionPublic}>(PackNFT.CollectionPublicPath, target: PackNFT.CollectionStoragePath)
                ?? panic("Could not link PackNFT.Collection Pub Path");
        }
    }
}
 
