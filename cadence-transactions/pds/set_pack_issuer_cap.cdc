import PDS from 0x{{.PDS}}

transaction (issuer: Address) {
    prepare(pds: AuthAccount) {
        let cap = pds.getCapability<&PDS.DistributionCreator{PDS.IDistCreator}>(PDS.distCreatorPrivPath)
        if !cap.check() {
            panic ("cannot borrow such capability") 
        } else {
            let setCapRef = getAccount(issuer).getCapability<&PDS.PackIssuer{PDS.PackIssuerCapReciever}>(PDS.packIssuerCapRecv).borrow()
                ?? panic("no cap for setting distCap")
            setCapRef.setDistCap(cap: cap);
        }
    }

}

