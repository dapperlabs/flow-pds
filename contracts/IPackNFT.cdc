import Crypto
import FungibleToken from "./FungibleToken.cdc"
import NonFungibleToken from "./NonFungibleToken.cdc"


pub contract interface IPackNFT{
    /// Status of the PackNFTs
    /// 
    /// map of NFT id : Status {"Created", "Revealed", "Opened"}
    access(contract) let status: {UInt64: String}
    /// StoragePath for Collection Resource
    /// 
    pub let collectionStoragePath: StoragePath 
    /// PublicPath expected for deposit
    /// 
    pub let collectionPublicPath: PublicPath 
    /// Request for Reveal
    ///
    pub event RevealRequest(id: UInt64)
    /// Request for Open
    ///
    /// This is emitted when owner of a PackNFT request for the entitled NFT to be
    /// deposited to its account
    pub event OpenPackRequest(id: UInt64) 
    /// New Pack NFT
    ///
    /// Emitted when a new PackNFT has been minted
    pub event Mint(id: UInt64, commitHash: String) 

    /// Public function to get status
    pub fun getStatus(id: UInt64): String

    access(contract) fun reveal(id: UInt64) {
        pre {
            self.status[id] != nil : "No such PackNFT"
        }
        post {
            self.status[id] == "Revealed": "PackNFT status must be Revealed"
        }
    }
    
    access(contract) fun open(id: UInt64) {
        pre {
            self.status[id] == "Revealed": "PackNFT not yet revealed"
        }
        post {
            self.status[id] == "Opened": "PackNFT status must be Opened"
        }
        
    }
    
    /// PackNFTMinter specific interfaces
    pub resource interface IMinter {
         pub fun mint(commitHash: String, issuer: Address)
    }
    
    pub resource interface IPackNFTToken {
        pub fun reveal()
        pub fun open() 
    }
}
