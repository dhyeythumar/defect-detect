package main

import (
	"context"

	anz "github.com/dmdhrumilmistry/defect-detect/pkg/analyzer"
	"github.com/dmdhrumilmistry/defect-detect/pkg/config"
	"github.com/dmdhrumilmistry/defect-detect/pkg/db"
	"github.com/dmdhrumilmistry/defect-detect/pkg/service/auth"
	"github.com/dmdhrumilmistry/defect-detect/pkg/service/component"
	"github.com/dmdhrumilmistry/defect-detect/pkg/service/project"
	"github.com/dmdhrumilmistry/defect-detect/pkg/service/sbom"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

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
	analyzer := anz.NewAnalyzer()

	// create stores
	log.Info().Msg("Registering Routes")

	authStore := auth.NewAuthStore(mgo.Db)
	authHandler := auth.NewAuthHandler(authStore)
	authHandler.RegisterRoutes(r)

	// add auth for remaining service endpoints
	r.Use(authStore.WithJwtAuth())

	sbomStore := sbom.NewComponentSbomStore(mgo.Db)
	sbomHandler := sbom.NewComponentSbomHandler(sbomStore, authStore)
	sbomHandler.RegisterRoutes(r, authStore)

	componentStore := component.NewComponentStore(mgo.Db, analyzer)
	componentHandler := component.NewComponentHandler(componentStore, sbomStore, authStore)
	componentHandler.RegisterRoutes(r)

	projectStore := project.NewProjectStore(mgo.Db)
	projectHandler := project.NewProjectHandler(projectStore, sbomStore, componentStore, authStore)
	projectHandler.RegisterRoutes(r)
	log.Info().Msg("Routes Registered Successfully")

	// Start the server
	if err := r.Run(":" + config.DefaultConfig.HostPort); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start server on port %s", config.DefaultConfig.HostPort)
	}
}
