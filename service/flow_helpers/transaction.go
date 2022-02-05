package flow_helpers

import (
	"context"
	"fmt"
	"github.com/flow-hydraulics/flow-pds/service/transactions"
	"gorm.io/gorm"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

func SignProposeAndPayAs(ctx context.Context, flowClient *client.Client, account *Account, tx *flow.Transaction) (UnlockKeyFunc, error) {

	signer, err := account.GetSigner()
	if err != nil {
		return EmptyUnlockKey, err
	}

	key, unlock, err := account.GetProposalKey(ctx, flowClient)
	if err != nil {
		return unlock, err
	}

	tx.
		SetProposalKey(account.Address, key.Index, key.SequenceNumber).
		SetPayer(account.Address).
		AddAuthorizer(account.Address)

	if err := tx.SignEnvelope(account.Address, key.Index, signer); err != nil {
		return unlock, err
	}

	return unlock, nil
}

func SignWithDBProposeAndPayAs(ctx context.Context, db *gorm.DB, flowClient *client.Client, account *Account, tx *flow.Transaction) (*transactions.AccountKey, error) {

	signer, err := account.GetSigner()
	if err != nil {
		return nil, err
	}

	acctKey, err := account.GetProposalKeyFromDB(ctx, db, flowClient)

	if err != nil {
		return nil, err
	}

	tx.
		SetProposalKey(account.Address, acctKey.FlowKey.Index, acctKey.FlowKey.SequenceNumber).
		SetPayer(account.Address).
		AddAuthorizer(account.Address)

	if err := tx.SignEnvelope(account.Address, acctKey.FlowKey.Index, signer); err != nil {
		return nil, err
	}

	return acctKey, nil
}

// WaitForSeal blocks until
// - an error occurs while fetching the transaction result
// - the transaction gets an error status
// - the transaction gets a "TransactionStatusSealed" or "TransactionStatusExpired" status
// - timeout is reached
func WaitForSeal(ctx context.Context, c *client.Client, id flow.Identifier, timeout time.Duration) (*flow.TransactionResult, error) {
	var (
		result *flow.TransactionResult
		err    error
	)

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	for {
		result, err = c.GetTransactionResult(ctx, id)
		if err != nil {
			return nil, err
		}

		if result.Error != nil {
			return result, result.Error
		}

		switch result.Status {
		default:
			// Not an interesting state, exit switch and continue loop
		case flow.TransactionStatusExpired:
			// Expired, handle as an error
			return result, fmt.Errorf("transaction expired")
		case flow.TransactionStatusSealed:
			// Sealed, all good
			return result, nil
		}

		time.Sleep(time.Second)
	}
}
