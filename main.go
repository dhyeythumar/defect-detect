package main

import (
	"context"
	"net/http"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/dmdhrumilmistry/defect-detect/pkg/analyzer/osv"
	"github.com/dmdhrumilmistry/defect-detect/pkg/config"
	"github.com/dmdhrumilmistry/defect-detect/pkg/db"
	"github.com/dmdhrumilmistry/defect-detect/pkg/service/component"
	"github.com/dmdhrumilmistry/defect-detect/pkg/service/sbom"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type SBOMData struct {
	Components []cyclonedx.Component `json:"components"`
}

var sboms = make(map[string]SBOMData) // In-memory storage for simplicity

// curl "http://localhost:8080/api/components?sbom_id=bom.json"
func listComponents(c *gin.Context) {
	sbomID := c.Query("sbom_id")
	sbomData, exists := sboms[sbomID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "SBOM not found"})
		return
	}

	c.JSON(http.StatusOK, sbomData.Components)
}

func main() {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r = r.With()
	r.SetTrustedProxies(nil)

	mgo, err := db.NewMongo(config.DefaultConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get db connection")
	}
	defer mgo.Client.Disconnect(context.TODO())

	if !config.DefaultConfig.IsDevEnv {
		gin.SetMode(gin.ReleaseMode)
	}

	// Analyzers
	osvAnalyzer := osv.NewOsvAnalyzer()

	// create stores
	log.Info().Msg("Registering Routes")
	sbomStore := sbom.NewComponentSbomStore(mgo.Db)
	sbomHandler := sbom.NewComponentSbomHandler(sbomStore)
	sbomHandler.RegisterRoutes(r)

	componentStore := component.NewComponentStore(mgo.Db, osvAnalyzer)
	componentHandler := component.NewComponentHandler(componentStore, sbomStore)
	componentHandler.RegisterRoutes(r)

	// Start the server
	if err := r.Run(":" + config.DefaultConfig.HostPort); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start server on port %s", config.DefaultConfig.HostPort)
	}
}
