package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/flow-hydraulics/flow-pds/service/app"
	"github.com/flow-hydraulics/flow-pds/service/common"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Set distribution capability
func HandleSetDistCap(logger *log.Logger, app *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Check body is not empty
		if err := checkNonEmptyBody(r); err != nil {
			handleError(rw, logger, err)
			return
		}

		var reqData ReqSetDistCap

		// Decode JSON
		if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
			handleError(rw, logger, err)
			return
		}

		if err := app.SetDistCap(r.Context(), reqData.Issuer); err != nil {
			handleError(rw, logger, err)
			return
		}

		handleJsonResponse(rw, http.StatusOK, "Ok")

	}
}

// Create a distribution
func HandleCreateDistribution(logger *log.Logger, app *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Check body is not empty
		if err := checkNonEmptyBody(r); err != nil {
			handleError(rw, logger, err)
			return
		}

		var reqDist ReqCreateDistribution

		// Decode JSON
		if err := json.NewDecoder(r.Body).Decode(&reqDist); err != nil {
			handleError(rw, logger, err)
			return
		}

		// Create new distribution
		appDist := reqDist.ToApp()
		if err := app.CreateDistribution(r.Context(), &appDist); err != nil {
			handleError(rw, logger, err)
			return
		}

		res := ResCreateDistribution{
			ID:     appDist.ID,
			FlowID: appDist.FlowID,
		}

		handleJsonResponse(rw, http.StatusCreated, res)
	}
}

// List distributions
func HandleListDistributions(logger *log.Logger, app *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		limit, err := strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			limit = 0
		}

		offset, err := strconv.Atoi(r.FormValue("offset"))
		if err != nil {
			offset = 0
		}

		list, err := app.ListDistributions(r.Context(), limit, offset)
		if err != nil {
			handleError(rw, logger, err)
			return
		}

		res := ResDistributionListFromApp(list)

		handleJsonResponse(rw, http.StatusOK, res)
	}
}

// Get distribution details
func HandleGetDistribution(logger *log.Logger, app *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id, err := uuid.Parse(vars["id"])
		if err != nil {
			handleError(rw, logger, err)
			return
		}

		dist, err := app.GetDistribution(r.Context(), id)
		if err != nil {
			handleError(rw, logger, err)
			return
		}

		res := ResGetDistributionFromApp(dist)

		handleJsonResponse(rw, http.StatusOK, res)
	}
}

// Abort a distribution
func HandleAbortDistribution(logger *log.Logger, app *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id, err := uuid.Parse(vars["id"])
		if err != nil {
			handleError(rw, logger, err)
			return
		}

		if err := app.AbortDistribution(r.Context(), id); err != nil {
			handleError(rw, logger, err)
			return
		}

		handleJsonResponse(rw, http.StatusOK, "Ok")
	}
}

func HandleHealthReady() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}
}

/*
	HandleCreatePacks allows the creation of packs with pre-existing content.
	It assumes the existence of a long-running distribution (ie: MINTING status).
	It requires either the nft_flow_ids _or_ commitmentHash to be sent as part of the request. If the nft_flow_ids are
	sent, this endpoint will calculate the commitment hash
*/
func HandleCreatePacks(logger *log.Logger, a *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		// Check body is not empty
		if err := checkNonEmptyBody(r); err != nil {
			handleError(rw, logger, err)
			return
		}

		var reqCreatePack ReqCreatePack

		// Decode JSON
		if err := json.NewDecoder(r.Body).Decode(&reqCreatePack); err != nil {
			handleError(rw, logger, err)
			return
		}

		if err := reqCreatePack.Validate(); err != nil {
			handleError(rw, logger, err)
			return
		}

		distributionState, err := a.GetDistributionState(r.Context(), reqCreatePack.DistributionID)
		if err != nil {
			handleError(rw, logger, fmt.Errorf("error retrieving distribution with id %s :%w", reqCreatePack.DistributionID, err))
			return
		}

		if distributionState != common.DistributionStateMinting {
			handleError(rw, logger, fmt.Errorf("distribution is in invalid state with id %s :%w", distributionState, err))
			return
		}

		addressLocation := app.AddressLocation{
			Name:    reqCreatePack.PackReference.Name,
			Address: reqCreatePack.PackReference.Address,
		}

		var collectibles []app.Collectible

		for _, flowID := range reqCreatePack.NFTFlowIDs {
			collectibles = append(collectibles, app.Collectible{
				ContractReference: addressLocation,
				FlowID:            flowID,
			})
		}

		pack := app.Pack{
			DistributionID:    reqCreatePack.DistributionID,
			ContractReference: addressLocation,
			State:             common.PackStateInit,
			Collectibles:      collectibles,
		}

		if reqCreatePack.CommitmentHash == nil {
			err = pack.SetCommitmentHash()
			if err != nil {
				handleError(rw, logger, fmt.Errorf("error setting commitment hash: %w", err))
				return
			}
		} else {
			pack.CommitmentHash = []byte(*reqCreatePack.CommitmentHash)
		}

		err = a.InsertPack(r.Context(), &pack)
		if err != nil {
			handleError(rw, logger, fmt.Errorf("error inserting pack: %w", err))
			return
		}

		rw.WriteHeader(http.StatusOK)
		return
	}
}
