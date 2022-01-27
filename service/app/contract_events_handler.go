package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/flow-hydraulics/flow-pds/service/common"
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

func (ev *EventHandler) PollByEventName(ctx context.Context, wg *sync.WaitGroup, cpc *CirculatingPackContract, cpcCursor *CirculatingPackContractBlockCursor, eventName string, begin uint64, end uint64) error {
	defer wg.Done()

	arr, err := ev.flowClient.GetEventsForHeightRange(ctx, client.EventRangeQuery{
		Type:        cpc.EventName(eventName),
		StartHeight: begin,
		EndHeight:   end,
	})

	if err != nil {
		return fmt.Errorf("error querying events for height range. err=%w", err)
	}

	for _, be := range arr {
		for _, e := range be.Events {
			eventLogger := ev.eventLogger.WithFields(log.Fields{"eventType": e.Type, "eventID": e.ID()})

			eventLogger.Debug("Handling event")

			evtValueMap := flow_helpers.EventValuesToMap(e)

			packFlowIDCadence, ok := evtValueMap["id"]
			if !ok {
				return fmt.Errorf("could not read 'id' from event %s", e)
			}

			packFlowID, err := common.FlowIDFromCadence(packFlowIDCadence)
			if err != nil {
				return err // rollback
			}

			addressLocation := AddressLocation{
				Name:    cpc.Name,
				Address: cpc.Address,
			}
			pack, err := GetPackByContractAndFlowID(ev.db, addressLocation, packFlowID)
			if err != nil {
				return fmt.Errorf("error retrieving pack from db err=%w", err) // rollback
			}

			distribution, err := GetDistributionSmall(ev.db, pack.DistributionID)
			if err != nil {
				return fmt.Errorf("error retrieving distribution from db err=%w", err)
			}

			eventLogger = eventLogger.WithFields(log.Fields{
				"distID":     distribution.ID,
				"distFlowID": distribution.FlowID,
				"packID":     pack.ID,
				"packFlowID": pack.FlowID,
			})

			eventLogger.Info("handling event...")

			if err := ev.HandleEvent(ctx, eventName, &e, pack, distribution); err != nil {
				eventLogger.Warnf("distID:%s distFlowID:%s packID:%s packFlowID:%s err:%s", distribution.ID, distribution.FlowID, pack.ID, pack.FlowID, err.Error())
				continue
			}

			eventLogger.Trace("Handling event complete")
		}
	}

	cpcCursor.StartAtBlock = end

	// Update the CirculatingPackContract in database
	if err := UpdateCirculatingPackContractBlockCursor(ev.db, cpcCursor); err != nil {
		return err // rollback
	}

	return nil
}

func (ev *EventHandler) HandleEvent(ctx context.Context, eventName string, event *flow.Event, pack *Pack, distribution *Distribution) error {
	switch eventName {
	// -- REVEAL_REQUEST, Owner has requested to reveal a pack ------------
	case REVEAL_REQUEST:
		if err := ev.HandleRevealRequestEvent(ctx, event, pack, distribution); err != nil {
			return err
		}

	// -- REVEALED, Pack has been revealed onchain ------------------------
	case REVEALED:
		// Make sure the pack is in correct state
		if err := ev.HandleRevealedEvent(ctx, event, pack, distribution); err != nil {
			return err
		}

	// -- OPEN_REQUEST, Owner has requested to open a pack ----------------
	case OPEN_REQUEST:
		if err := ev.HandleOpenRequestEvent(ctx, event, pack, distribution); err != nil {
			return err
		}

	// -- OPENED, Pack has been opened onchain ----------------------------
	case OPENED:
		if err := ev.HandleOpenedEvent(ctx, event, pack, distribution); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported event: %s", eventName)
	}

	return nil
}
