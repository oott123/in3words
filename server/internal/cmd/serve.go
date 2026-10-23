package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/oott123/in3words/server/internal/api"
	"github.com/oott123/in3words/server/internal/config"
	"github.com/oott123/in3words/server/internal/logging"
	"github.com/oott123/in3words/server/internal/wpdb"
	"github.com/oott123/in3words/server/internal/wpfmt"
)

const (
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 10 * time.Second
)

func newServeCommand(v *viper.Viper, configFile *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "启动 HTTP 服务",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(v, *configFile)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			return serve(cmd.Context(), cfg)
		},
	}
	cmd.Flags().String("listen", "", "监听地址，例如 :8080")
	_ = v.BindPFlag("listen", cmd.Flags().Lookup("listen"))
	return cmd
}

func serve(ctx context.Context, cfg *config.Config) error {
	if err := logging.Setup(cfg.Log.Level, cfg.Log.Format); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := wpdb.Open(ctx, cfg.DB)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer func() { _ = db.Close() }()

	fmtOptions := wpfmt.DefaultOptions()
	fmtOptions.HTML5Caption = cfg.Render.HTML5Caption
	fmtOptions.UploadsBaseURL = cfg.Render.UploadsBaseURL
	handler, err := api.New(wpdb.NewQueries(db, cfg.DB.TablePrefix), api.Options{
		Format:    fmtOptions,
		CacheSize: cfg.Render.CacheSize,
	})
	if err != nil {
		return fmt.Errorf("create api: %w", err)
	}

	srv := &http.Server{
		Addr:              cfg.Listen,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}
	errCh := make(chan error, 1)
	go func() {
		logrus.WithField("listen", cfg.Listen).Info("server started")
		errCh <- srv.ListenAndServe()
	}()

	select {
	case serveErr := <-errCh:
		return fmt.Errorf("serve http: %w", serveErr)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err = srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	if err = <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve http: %w", err)
	}
	return nil
}
