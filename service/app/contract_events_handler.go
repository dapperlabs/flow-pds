package app

import (
	"context"
	"fmt"

	"github.com/flow-hydraulics/flow-pds/service/flow_helpers"
	"github.com/flow-hydraulics/flow-pds/service/transactions"
	"github.com/google/uuid"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type EventHandler struct {
	db          *gorm.DB
	eventLogger *log.Entry
	flowClient  *client.Client
}

func (ev *EventHandler) HandleRevealRequestEvent(ctx context.Context, event *flow.Event, pack *Pack, distribution *Distribution) error {

	evtValueMap := flow_helpers.EventValuesToMap(*event)

	// Make sure the pack is in correct state
	if err := pack.RevealRequestHandled(); err != nil {
		return fmt.Errorf("error while handling %s: %w", REVEAL_REQUEST, err)
	}

	// Update the pack in database
	if err := UpdatePack(ev.db, pack); err != nil {
		return err // rollback
	}

	// Get the owner of the pack from the transaction that emitted the open request event
	tx, err := ev.flowClient.GetTransaction(ctx, event.TransactionID)
	if err != nil {
		return err // rollback
	}
	owner := tx.Authorizers[0]

	// NOTE: this only handles one collectible contract per pack
	contract := pack.Collectibles[0].ContractReference

	collectibleCount := len(pack.Collectibles)
	collectibleContractAddresses := make([]cadence.Value, collectibleCount)
	collectibleContractNames := make([]cadence.Value, collectibleCount)
	collectibleIDs := make([]cadence.Value, collectibleCount)

	for i, c := range pack.Collectibles {
		collectibleContractAddresses[i] = cadence.Address(c.ContractReference.Address)
		collectibleContractNames[i] = cadence.String(c.ContractReference.Name)
		collectibleIDs[i] = cadence.UInt64(c.FlowID.Int64)
	}

	openRequestValue, ok := evtValueMap["openRequest"]
	if !ok { // TODO(nanuuki): rollback or use a default value for openRequest?
		err := fmt.Errorf("could not read 'openRequest' from event %s", event)
		return err // rollback
	}

	openRequest := openRequestValue.ToGoValue().(bool)
	eventLogger := ev.eventLogger.WithFields(log.Fields{"openRequest": openRequest})

	arguments := []cadence.Value{
		cadence.UInt64(distribution.FlowID.Int64),
		cadence.UInt64(pack.FlowID.Int64),
		cadence.NewArray(collectibleContractAddresses),
		cadence.NewArray(collectibleContractNames),
		cadence.NewArray(collectibleIDs),
		cadence.String(pack.Salt.String()),
		cadence.Address(owner),
		cadence.NewBool(openRequest),
		cadence.Path{Domain: "private", Identifier: contract.ProviderPath()},
	}

	// NOTE: this only handles one collectible contract per pack
	txScript, err := flow_helpers.ParseCadenceTemplate(
		REVEAL_SCRIPT,
		&flow_helpers.CadenceTemplateVars{
			PackNFTName:           pack.ContractReference.Name,
			PackNFTAddress:        pack.ContractReference.Address.String(),
			CollectibleNFTName:    contract.Name,
			CollectibleNFTAddress: contract.Address.String(),
		},
	)
	if err != nil {
		return err // rollback
	}

	t, err := transactions.NewTransactionWithDistributionID(REVEAL_SCRIPT, txScript, arguments, distribution.ID)
	if err != nil {
		return err // rollback
	}

	if err := t.Save(ev.db); err != nil {
		return err // rollback
	}

	if openRequest { // NOTE: This block should run only if we want to reveal AND open the pack
		// Reset the ID to save a second indentical transaction
		t.ID = uuid.Nil
		if err := t.Save(ev.db); err != nil {
			return err // rollback
		}
	}

	eventLogger.Info("Pack reveal transaction created")
	return nil
}

func (ev *EventHandler) HandleRevealedEvent(ctx context.Context, event *flow.Event, pack *Pack, _ *Distribution) error {
	// Make sure the pack is in correct state
	if err := pack.Reveal(); err != nil {
		return fmt.Errorf("error while handling %s: %w", REVEALED, err)
	}

	// Update the pack in database
	if err := UpdatePack(ev.db, pack); err != nil {
		return err // rollback
	}
	return nil
}

func (ev *EventHandler) HandleOpenRequestEvent(ctx context.Context, event *flow.Event, pack *Pack, distribution *Distribution) error {
	// Make sure the pack is in correct state
	if err := pack.OpenRequestHandled(); err != nil {
		return fmt.Errorf("error while handling %s: %w", OPEN_REQUEST, err)
	}

	// Update the pack in database
	if err := UpdatePack(ev.db, pack); err != nil {
		return err // rollback
	}

	// Get the owner of the pack from the transaction that emitted the open request event
	tx, err := ev.flowClient.GetTransaction(ctx, event.TransactionID)
	if err != nil {
		return err // rollback
	}
	owner := tx.Authorizers[0]

	// NOTE: this only handles one collectible contract per pack
	contract := pack.Collectibles[0].ContractReference

	collectibleCount := len(pack.Collectibles)

	collectibleContractAddresses := make([]cadence.Value, collectibleCount)
	collectibleContractNames := make([]cadence.Value, collectibleCount)
	collectibleIDs := make([]cadence.Value, collectibleCount)

	for i, c := range pack.Collectibles {
		collectibleContractAddresses[i] = cadence.Address(c.ContractReference.Address)
		collectibleContractNames[i] = cadence.String(c.ContractReference.Name)
		collectibleIDs[i] = cadence.UInt64(c.FlowID.Int64)
	}

	arguments := []cadence.Value{
		cadence.UInt64(distribution.FlowID.Int64),
		cadence.UInt64(pack.FlowID.Int64),
		cadence.NewArray(collectibleContractAddresses),
		cadence.NewArray(collectibleContractNames),
		cadence.NewArray(collectibleIDs),
		cadence.Address(owner),
		cadence.Path{Domain: "private", Identifier: contract.ProviderPath()},
	}

	txScript, err := flow_helpers.ParseCadenceTemplate(
		OPEN_SCRIPT,
		&flow_helpers.CadenceTemplateVars{
			CollectibleNFTName:    contract.Name,
			CollectibleNFTAddress: contract.Address.String(),
		},
	)
	if err != nil {
		return err // rollback
	}

	t, err := transactions.NewTransactionWithDistributionID(OPEN_SCRIPT, txScript, arguments, distribution.ID)
	if err != nil {
		return err // rollback
	}

	if err := t.Save(ev.db); err != nil {
		return err // rollback
	}

	ev.eventLogger.Info("Pack open transaction created")
	return nil
}

func (ev *EventHandler) HandleOpenedEvent(ctx context.Context, event *flow.Event, pack *Pack, distribution *Distribution) error {

	// Make sure the pack is in correct state
	if err := pack.Open(); err != nil {
		return fmt.Errorf("error while handling %s: %w", OPENED, err)
	}

	// Update the pack in database
	if err := UpdatePack(ev.db, pack); err != nil {
		return err // rollback
	}
	return nil
}
