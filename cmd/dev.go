package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Palats/mapshot/factorio"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

func dev(ctx context.Context, factorioSettings *factorio.Settings) error {
	fact, err := factorio.New(factorioSettings)
	if err != nil {
		return err
	}

	tmpdir, err := ioutil.TempDir("", "mapshot")
	if err != nil {
		return fmt.Errorf("unable to create temp dir: %w", err)
	}
	glog.Info("temp dir: ", tmpdir)

	// Copy mods
	dstMods := path.Join(tmpdir, "mods")
	if err := fact.CopyMods(dstMods, []string{"mapshot"}); err != nil {
		return err
	}

	// Add the mod itself.
	dstMapshot := path.Join(dstMods, "mapshot")
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("unable to find current directory: %w", err)
	}
	modDir := path.Join(cwd, "mod")
	if err := os.Symlink(modDir, dstMapshot); err != nil {
		return fmt.Errorf("unable to symlink %q: %w", modDir, err)
	}
	glog.Infof("mod linked at %q", dstMapshot)

	factorioArgs := []string{
		"--disable-audio",
		"--disable-prototype-history",
		"--mod-directory", dstMods,
	}
	glog.Infof("Factorio args: %v", factorioArgs)

	fmt.Println("Starting Factorio...")
	if err := fact.Run(ctx, factorioArgs); err != nil {
		return fmt.Errorf("error while running Factorio: %w", err)
	}

	// Remove temporary directory.
	if err := os.RemoveAll(tmpdir); err != nil {
		return fmt.Errorf("unable to remove temp dir %q: %w", tmpdir, err)
	}
	glog.Infof("temp dir %q removed", tmpdir)

	return nil
}

var cmdDev = &cobra.Command{
	Use:   "dev",
	Short: "Run Factorio to work on the mod code.",
	Long: `Run Factorio using the mod files in this directory.
The mod files are linked. This means that changes to the mod will be taken
into account when Factorio reads them - i.e., when loading a save.

No override file is created, beside the default one - all rendering
parameters come from the game.
	`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return dev(cmd.Context(), factorioSettings)
	},
}

func init() {
	renderFlags.Register(cmdDev.PersistentFlags(), "")
	cmdRoot.AddCommand(cmdDev)
}
