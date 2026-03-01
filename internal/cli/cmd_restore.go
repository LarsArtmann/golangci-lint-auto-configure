package cli

import (
	"errors"
	"os"

	"github.com/charmbracelet/log"
	"github.com/larsartmann/golangcli-linter-auto-configure/pkg/config"
	"github.com/spf13/cobra"
)

// newRestoreCommand creates the restore command.
func newRestoreCommand(
	logger *log.Logger,
	configLoader *config.Loader,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore configuration from backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			if verbose {
				logger.SetLevel(log.DebugLevel)
			}

			// Get backup path from flag or argument
			backupPath, _ := cmd.Flags().GetString("backup-path")
			if backupPath == "" && len(args) > 0 {
				return errors.New("backup path required (use --backup-path flag or provide as argument)")
			}

			if backupPath == "" {
				backupPath = args[0]
			}

			logger.Infof("Restoring configuration from: %s", backupPath)

			// Check if backup file exists
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				return err
			}

			// Determine target config path
			targetPath := configPath
			if targetPath == "" {
				var err error

				targetPath, err = configLoader.FindConfigFile(".")
				if err != nil {
					return err
				}
			}

			logger.Infof("Restoring to: %s", targetPath)

			// Perform restore
			err := configLoader.RestoreConfig(backupPath, targetPath)
			if err != nil {
				return err
			}

			logger.Infof("Configuration restored successfully")

			return nil
		},
	}

	cmd.Flags().String("backup-path", "", "Path to backup file to restore from")

	return cmd
}
