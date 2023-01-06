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

func HandleCreatePacks(logger *log.Logger, a *app.App) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		/*
			func (dist *Distribution) Resolve() error {
		*/
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

		pack := app.Pack{
			DistributionID: reqCreatePack.DistributionID,
			ContractReference: app.AddressLocation{
				Name:    reqCreatePack.PackReference.Name,
				Address: reqCreatePack.PackReference.Address,
			},
			State:          common.PackStateInit,
			Salt:           nil,
			CommitmentHash: nil,
			Collectibles:   nil,
		}

		fmt.Printf("p: %+v", pack)

		rw.WriteHeader(http.StatusOK)
	}
}
