import PackNFT from 0x{{.PackNFT}}
import NonFungibleToken from 0x{{.NonFungibleToken}}

transaction() {
    prepare (issuer: AuthAccount) {
        
        // Check if account already have a PackIssuer resource, if so destroy it
        if issuer.borrow<&PackNFT.Collection>(from: PackNFT.CollectionStoragePath) == nil {
            issuer.save(<-PackNFT.createEmptyCollection(), to: PackNFT.CollectionStoragePath);
            issuer.link<&{NonFungibleToken.CollectionPublic}>(PackNFT.CollectionPublicPath, target: PackNFT.CollectionStoragePath)
                ?? panic("Could not link PackNFT.Collection Pub Path");
        }
        
        issuer.save(<- PackNFT.createEmptyCollection(), to: PackNFT.CollectionStoragePath);
        
        issuer.link<&{NonFungibleToken.CollectionPublic}>(PackNFT.CollectionPublicPath, target: PackNFT.CollectionStoragePath)
        ??  panic("Could not link Collection Pub Path");
    } 
}
 
